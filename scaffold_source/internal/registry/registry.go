package registry

import (
	"context"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/domains/health"
	"github.com/PROJECT_NAME/internal/domains/post"
	"github.com/PROJECT_NAME/internal/domains/user"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/middleware"
	"github.com/PROJECT_NAME/internal/nats"
	"github.com/PROJECT_NAME/internal/utils"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Ensure that Registry implements RegistryProvider
var _ RegistryProvider = new(Registry)

type Registry struct {
	db     db.Database
	config *config.Config
	logger logger.Logger

	// errors
	errorHandler *errors.Handler

	retry *utils.Retry

	natsService         nats.Service
	consumerNameBuilder *nats.ConsumerNameBuilder

	// domains
	userRepository user.Repository
	userService    user.Service

	postRepository post.Repository
	postService    post.Service

	// otel
	tp  *sdktrace.TracerProvider
	exp *otlptrace.Exporter
}

// Uncomment if you need child spans
// var tracer trace.Tracer

// func init() {
// 	// this seems to work even if the init happens before setting up the trace provider
// 	tracer = otel.Tracer("PROJECT_NAME")
// }

func NewRegistry(c *config.Config) *Registry {
	return &Registry{
		config: c,
	}
}

func (r *Registry) InitializeWithFiber(app *fiber.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sentry.Init(sentry.ClientOptions{
		Dsn:              r.config.Sentry.Dsn,
		TracesSampleRate: r.config.Sentry.TracesSampleRate,
	})

	sentryHandler := sentryfiber.New(sentryfiber.Options{
		Repanic:         true,
		WaitForDelivery: true,
	})

	app.Use(sentryHandler)

	var err error
	r.tp, r.exp, err = r.initializeOpenTelemetry(ctx)
	if err != nil {
		return err
	}

	// skip health check requests
	app.Use(otelfiber.Middleware(otelfiber.WithNext(func(c *fiber.Ctx) bool {
		return c.Path() == "/api/health" || c.Path() == "/api/health/ready"
	})))

	if err := r.Initialize(); err != nil {
		sentry.CaptureException(err)
		return err
	}

	r.RegisterPreMiddlewares(app)
	r.RegisterApiRoutes(app.Group("/api"))
	r.RegisterPostMiddlewares(app)
	// register other "things" (e.g. listeners, consumers, etc.)

	return nil
}

func (r *Registry) Initialize() error {
	var err error

	r.db, err = db.Connect(r)
	if err != nil {
		return err
	}

	r.natsService, err = nats.NewService(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Registry) Cleanup() error {
	r.Logger().Debug("🧹 Cleaning up registry")

	r.Logger().Info("🔌 Closing database connection")
	if err := r.db.Close(); err != nil {
		return err
	}

	r.Logger().Info("🔌 Closing NATS connection")
	if err := r.natsService.Cleanup(); err != nil {
		return err
	}

	if err := r.tp.Shutdown(context.Background()); err != nil {
		r.Logger().Error("Error shutting down tracer provider: %v", err)
	}

	if err := r.exp.Shutdown(context.Background()); err != nil {
		r.Logger().Error("Error shutting down exporter: %v", err)
	}

	r.Logger().Info("✅ Registry cleaned up successfully")
	// call cleanup funcs (e.g. unsubscribe listeners, etc.)

	return nil
}

func (r *Registry) RegisterPreMiddlewares(app *fiber.App) {
	// Global middlewares apply to all routes
	app.Use(middleware.NewRequestIDMiddleware().Handle)
	app.Use(middleware.NewLoggingMiddleware(r).Handle)
	app.Use(middleware.NewAuthMiddleware(r).Handle)
	// register other middlewares
}

func (r *Registry) RegisterApiRoutes(api fiber.Router) {
	// health check
	health.NewHealthHandler(r).RegisterRoutes(api)

	// user routes
	user.NewHandler(r).RegisterRoutes(api)

	// post routes
	post.NewHandler(r).RegisterRoutes(api)

	// register other routes
}

func (r *Registry) RegisterPostMiddlewares(app *fiber.App) {
	app.Use(middleware.NewNotFoundMiddleware(r).Handle)
}

func (r *Registry) initializeOpenTelemetry(ctx context.Context) (*sdktrace.TracerProvider, *otlptrace.Exporter, error) {
	// It'll uses these envs to get the endpoint and service name
	// OTEL_SERVICE_NAME=PROJECT_NAME
	// OTEL_EXPORTER_OTLP_ENDPOINT=https://mymonitor.nayla.tech:443
	// OTEL_EXPORTER_OTLP_HEADERS=x-special-key=your_key_here

	// Configure a new OTLP exporter using environment variables for sending data to Honeycomb over gRPC
	clientOTel := otlptracegrpc.NewClient()
	exp, err := otlptrace.New(ctx, clientOTel)
	if err != nil {
		return nil, nil, err
	}

	resource, rErr := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("environment", r.Config().App.Env),
		),
	)

	if rErr != nil {
		return nil, nil, rErr
	}

	// Create a new tracer provider with a batch span processor and the otlp exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
	)

	// Register the global Tracer provider
	otel.SetTracerProvider(tp)

	// Register the W3C trace context and baggage propagators so data is propagated across services/processes
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, exp, nil
}
