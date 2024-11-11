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
	"github.com/jmoiron/sqlx"
)

// Ensure that Registry implements RegistryProvider
var _ RegistryProvider = new(Registry)

type Registry struct {
	db     *sqlx.DB
	config *config.Config
	logger logger.Logger

	// errors
	errorHandler *errors.Handler

	// domains
	userRepository *user.Repository
	userService    *user.Service
	userHandler    *user.Handler

	postRepository *post.Repository
	postService    *post.Service
	postHandler    *post.Handler
}

func NewRegistry(c *config.Config) (*Registry, error) {
	r := &Registry{
		config: c,
	}

	if err := r.initialize(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Registry) initialize() error {
	db, err := db.Connect(r)
	if err != nil {
		return err
	}

	r.db = db

	return nil
}

func (r *Registry) RegisterMiddlewares(app *fiber.App) {
	// Global middlewares apply to all routes
	app.Use(middleware.NewRequestIDMiddleware().Handle)
	app.Use(middleware.NewLoggingMiddleware(r).Handle)
	app.Use(middleware.NewAuthMiddleware(r).Handle)
	// register other middlewares
}

func (r *Registry) RegisterApiRoutes(api fiber.Router) {
	// health check
	api.Get("/health", health.NewHealthHandler(r).HealthCheck)

	// user routes
	r.UserHandler().RegisterRoutes(api)

	// post routes
	r.PostHandler().RegisterRoutes(api)

	// register other routes
}
