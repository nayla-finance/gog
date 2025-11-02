package registry

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/domains/health"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/post"
	"github.com/PROJECT_NAME/internal/domains/user"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/nayla-finance/go-nayla/clients/rest/kyc"
	"github.com/nayla-finance/go-nayla/clients/rest/los"
	"github.com/nayla-finance/go-nayla/logger"
	"github.com/nayla-finance/go-nayla/nats"
)

type RegistryProvider interface {
	interfaces.SignalProvider
	db.DBProvider
	config.ConfigProvider
	logger.Provider

	// errors
	errors.ErrorProvider
	errors.ErrorHandlerProvider

	nats.ServiceProvider

	// domains
	// user
	user.RepositoryProvider
	interfaces.UserServiceProvider

	// post
	post.RepositoryProvider
	interfaces.PostServiceProvider

	kyc.ClientProvider
	los.ClientProvider
}

func (r *Registry) DB() db.Database {
	return r.db
}

func (r *Registry) Config() *config.Config {
	return r.config
}

func (r *Registry) Logger() logger.Logger {
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

func (r *Registry) HealthService() health.Service {
	if r.healthService == nil {
		r.healthService = health.NewService(r)
	}

	return r.healthService
}

func (r *Registry) NatsService() nats.Service {
	return r.natsService
}
