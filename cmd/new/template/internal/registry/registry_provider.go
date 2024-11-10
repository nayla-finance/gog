package registry

import (
	"github.com/jmoiron/sqlx"
	"github.com/project-name/internal/config"
	"github.com/project-name/internal/db"
	"github.com/project-name/internal/domains/post"
	"github.com/project-name/internal/domains/user"
	"github.com/project-name/internal/errors"
	"github.com/project-name/internal/logger"
)

type RegistryProvider interface {
	db.DBProvider
	config.ConfigProvider
	logger.LoggerProvider

	// errors
	errors.ErrorProvider
	errors.ErrorHandlerProvider

	// domains
	// user
	user.RepositoryProvider
	user.ServiceProvider
	user.HandlerProvider

	// post
	post.RepositoryProvider
	post.ServiceProvider
	post.HandlerProvider
}

func (r *Registry) DB() *sqlx.DB {
	return r.db
}

func (r *Registry) Config() *config.Config {
	return r.config
}

func (r *Registry) Logger() logger.Logger {
	if r.logger == nil {
		r.logger = logger.NewLogger(r)
	}

	return r.logger
}

func (r *Registry) NewError(c errors.ErrorCode, m string) *errors.AppError {
	return &errors.AppError{
		Code:    c,
		Message: m,
	}
}

func (r *Registry) ErrorHandler() *errors.Handler {
	if r.errorHandler == nil {
		r.errorHandler = errors.NewErrorHandler(r)
	}

	return r.errorHandler
}
