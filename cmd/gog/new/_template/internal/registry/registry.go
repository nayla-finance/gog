package registry

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/domains/health"
	"github.com/PROJECT_NAME/internal/domains/post"
	"github.com/PROJECT_NAME/internal/domains/user"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// Ensure that Registry implements RegistryProvider
var _ RegistryProvider = new(Registry)

type Registry struct {
	db     db.Database
	config *config.Config
	logger logger.Logger

	// errors
	errorHandler *errors.Handler

	// domains
	userRepository user.Repository
	userService    user.Service

	postRepository post.Repository
	postService    post.Service
}

func NewRegistry(c *config.Config) *Registry {
	return &Registry{
		config: c,
	}
}

func (r *Registry) Initialize(app *fiber.App) error {
	var err error

	r.db, err = db.Connect(r)
	if err != nil {
		return err
	}

	r.RegisterMiddlewares(app)
	r.RegisterApiRoutes(app.Group("/api"))
	// register other "things" (e.g. listeners, consumers, etc.)

	return nil
}

func (r *Registry) Cleanup() error {
	r.Logger().Debug("🧹 Cleaning up registry")

	r.Logger().Info("🔌 Closing database connection")
	if err := r.db.Close(); err != nil {
		return err
	}

	r.Logger().Info("✅ Registry cleaned up successfully")
	// call cleanup funcs (e.g. unsubscribe listeners, etc.)

	return nil
}

func (r *Registry) RegisterMiddlewares(app *fiber.App) {
	// Global middlewares apply to all routes
	app.Use(middleware.NewRequestIDMiddleware().Handle)
	app.Use(middleware.NewLoggingMiddleware(r).Handle)
	app.Use(middleware.NewAuthMiddleware(r).Handle)
	app.Use(middleware.NewRequestTimingMiddleware(r).Handle)
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
