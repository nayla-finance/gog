package registry

import (
	"context"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/domains/health"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/PROJECT_NAME/internal/domains/post"
	"github.com/PROJECT_NAME/internal/domains/user"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/nayla-finance/go-nayla/clients/rest/kyc"
	"github.com/nayla-finance/go-nayla/clients/rest/los"
	"github.com/nayla-finance/go-nayla/logger"
	"github.com/nayla-finance/go-nayla/middleware"
	"github.com/nayla-finance/go-nayla/nats"
	"github.com/nayla-finance/go-nayla/otel"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Ensure that Registry implements RegistryProvider
var _ RegistryProvider = new(Registry)

type Registry struct {
	signal model.Signal

	db     db.Database
	config *config.Config
	logger logger.Logger

	// errors
	errorHandler *errors.Handler

	healthService health.Service

	natsService nats.Service

	// domains
	userRepository user.Repository
	userService    interfaces.UserService

	postRepository post.Repository
	postService    interfaces.PostService

	kycClient kyc.Client
	losClient los.Client

	// otel
	otelClient *otel.Client
}

// Uncomment if you need child spans
// var tracer trace.Tracer

// func init() {
// 	// this seems to work even if the init happens before setting up the trace provider
// 	tracer = otel.Tracer("PROJECT_NAME")
// }

func NewRegistry(c *config.Config) *Registry {
	return &Registry{
		signal: make(model.Signal, 10),
		config: c,
	}
}

func (r *Registry) InitializeWithFiber(app *fiber.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sentry.Init(sentry.ClientOptions{
		Dsn:              r.config.Sentry.Dsn,
		TracesSampleRate: r.config.Sentry.TracesSampleRate,
		Environment:      r.config.App.Env,
	})

	sentryHandler := sentryfiber.New(sentryfiber.Options{
		Repanic:         true,
		WaitForDelivery: true,
	})

	app.Use(sentryHandler)

	var err error
	if r.Config().OpenTelemetry.Enabled {
		r.otelClient, err = otel.NewClient(ctx)
		if err != nil {
			return err
		}

		// skip health check requests
		app.Use(otelfiber.Middleware(otelfiber.WithNext(func(c *fiber.Ctx) bool {
			for _, route := range r.Config().OpenTelemetry.ExcludedRoutes {
				if c.Path() == route {
					return true
				}
			}

			return false
		})))

		serveMetrics(app)
	}

	if err := r.Initialize(ctx); err != nil {
		sentry.CaptureException(err)
		return err
	}

	if err := r.RegisterConsumers(); err != nil {
		return err
	}

	// Register pre middlewares
	if err := r.RegisterPreMiddlewares(app); err != nil {
		return err
	}

	r.RegisterApiRoutes(app.Group("/api"))

	// Register post middlewares
	if err := r.RegisterPostMiddlewares(app); err != nil {
		return err
	}
	// register other "things" (e.g. listeners, consumers, etc.)

	// Register signal after initializing dependencies
	r.RegisterSignalListener()

	return nil
}

func (r *Registry) Initialize(ctx context.Context) error {
	var err error

	r.logger, err = logger.NewLogger(
		logger.WithLogLevel(r.Config().App.LogLevel),
		logger.WithSpanLevel(r.Config().App.LogLevel),
	)
	if err != nil {
		return err
	}

	r.db, err = db.Connect(r)
	if err != nil {
		return err
	}

	r.natsService, err = nats.NewService(
		ctx,
		nats.WithServers([]string{r.config.Nats.Servers}),
		nats.WithAuthProvider(nats.NewCredsAuth(r.config.Nats.CredsPath)),
		nats.WithClientName(r.config.Nats.ClientName),
		nats.WithLogger(r.Logger()),
		nats.WithJetstreamEnabled(true),
		nats.WithStream(nats.Stream{
			Name:     r.config.Nats.DefaultStreamName,
			Subjects: r.config.Nats.DefaultStreamSubjects,
		}),
		nats.WithMonitoringInterval(r.config.Nats.Monitoring.Interval),
		nats.WithMonitoringPendingMessagesThreshold(r.config.Nats.Monitoring.PendingMessagesThreshold),
		nats.WithMonitoringExcludedConsumers(r.config.Nats.Monitoring.ExcludedConsumers),
		nats.WithMonitoringOnConsumerRestart(func(ctx context.Context, consumerName string) {
			r.SendSignal(model.SignalPayload{
				Type: model.SignalTypeNatsConsumerRestart,
			})
		}),
	)
	if err != nil {
		return err
	}

	if err := r.InitializeClients(); err != nil {
		return err
	}

	return nil
}

func (r *Registry) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r.Logger().Debugw(ctx, "🧹 Cleaning up registry")

	r.Logger().Infow(ctx, "🔌 Closing database connection")
	if err := r.db.Close(); err != nil {
		return err
	}

	r.Logger().Infow(ctx, "🔌 Closing NATS connection")
	if err := r.NatsService().Cleanup(ctx); err != nil {
		return err
	}

	if r.otelClient != nil {
		if err := r.otelClient.Shutdown(ctx); err != nil {
			r.Logger().Errorw(ctx, "Error shutting down tracer provider", "error", err)
		}
	}

	if r.signal != nil {
		r.Logger().Debugw(ctx, "🔄 Closing signal channel")
		close(r.signal)
		r.signal = nil
		r.Logger().Debugw(ctx, "✅ Signal channel closed")
	}

	r.Logger().Infow(ctx, "✅ Registry cleaned up successfully")
	// call cleanup funcs (e.g. unsubscribe listeners, etc.)

	return nil
}

func (r *Registry) RegisterConsumers() error {
	// register consumers here

	return nil
}

func (r *Registry) RegisterPreMiddlewares(app *fiber.App) error {
	// Global middlewares apply to all routes

	app.Use(middleware.NewRequestIDMiddleware().Handle)

	requestIDMiddleware, err := middleware.NewLoggingMiddleware(
		middleware.WithLogger(r.Logger()),
	)
	if err != nil {
		return err
	}
	app.Use(requestIDMiddleware.Handle)

	authMiddleware, err := middleware.NewAuthMiddleware(
		middleware.WithAuthAPIKey(r.Config().Api.Key),
		middleware.WithAuthFallbackToXAPIKeyHeader(true),
		middleware.WithAuthLogger(r.Logger()),
		middleware.WithAuthPublicRoutes(r.Config().Api.PublicRoutes),
		middleware.WithAuthOnUnauthorized(func() error {
			return r.NewError(errors.ErrUnauthorized, "unauthorized missing or invalid API key")
		}),
	)
	if err != nil {
		return err
	}
	app.Use(authMiddleware.Handle)

	// register other middlewares

	return nil
}

func (r *Registry) RegisterApiRoutes(api fiber.Router) {
	// health check
	health.NewHandler(r).RegisterRoutes(api)

	// user routes
	user.NewHandler(r).RegisterRoutes(api)

	// post routes
	post.NewHandler(r).RegisterRoutes(api)

	// register other routes
}

func (r *Registry) RegisterPostMiddlewares(app *fiber.App) error {
	notFoundMiddleware, err := middleware.NewNotFoundMiddleware(
		middleware.WithNotFoundLogger(r.Logger()),
		middleware.WithNotFoundOnNotFound(func() error {
			return r.NewError(errors.ErrResourceNotFound, "route not found")
		}),
	)
	if err != nil {
		return err
	}
	app.Use(notFoundMiddleware.Handle)

	return nil
}

func serveMetrics(app *fiber.App) {
	h := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	app.Get("/metrics", func(c *fiber.Ctx) error {
		h(c.Context())
		return nil
	})
}
