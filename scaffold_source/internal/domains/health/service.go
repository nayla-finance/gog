package health

import (
	"context"
	"fmt"
	"strings"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/getsentry/sentry-go"
	"github.com/nayla-finance/go-nayla/clients/rest/kyc"
	"github.com/nayla-finance/go-nayla/clients/rest/los"
	"github.com/nayla-finance/go-nayla/logger"
	"github.com/nayla-finance/go-nayla/nats"
)

var _ Service = new(svc)

type (
	Service interface {
		ReadinessCheck(ctx context.Context) error
		LivenessCheck(ctx context.Context) error
	}

	ServiceProvider interface {
		HealthService() Service
	}

	svcDependencies interface {
		errors.ErrorProvider
		logger.Provider
		db.DBProvider
		nats.ServiceProvider
		config.ConfigProvider
		kyc.ClientProvider
		los.ClientProvider
	}

	svc struct {
		d                     svcDependencies
		readinessCheckSkipped int
		livenessCheckSkipped  int
	}
)

func NewService(d svcDependencies) *svc {
	return &svc{
		d:                     d,
		readinessCheckSkipped: 0,
		livenessCheckSkipped:  0,
	}
}

func (s *svc) ReadinessCheck(ctx context.Context) error {
	isVerbose := s.d.Config().Health.Readiness.VerboseLog
	if isVerbose {
		s.PrintServiceDependenciesHealth(ctx)
	}

	dbConfig, ok := s.d.Config().Health.Dependencies["database"]
	if ok && dbConfig.ReadinessCheck {
		if err := s.d.DB().Ping(); err != nil {
			s.d.Logger().Errorw(ctx, "❌ Database is not healthy", "error", err)

			sentry.CaptureException(fmt.Errorf("❌ Database is not healthy: %w", err))
			// 🚨 Readiness check for internal dependencies should return an error if they fail
			return err
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ Database is ready and caffeinated! ☕ It's got its schemas in order and its transactions committed.")
		}
	}

	natsConfig, ok := s.d.Config().Health.Dependencies["nats"]
	if ok && natsConfig.ReadinessCheck {
		if !s.d.NatsService().Ping(ctx) {
			s.d.Logger().Errorw(ctx, "❌ Nats connection is not ready")

			sentry.CaptureException(fmt.Errorf("❌ Nats connection is not ready"))
			// 🚨 Readiness check for internal dependencies should return an error if they fail
			return fmt.Errorf("❌ Nats connection is not ready")
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ NATS is ready to deliver! 📮 Like a postal service that actually works on time.")
		}
	}

	kycConfig, ok := s.d.Config().Health.Dependencies["kyc"]
	if ok && kycConfig.ReadinessCheck {
		if err := s.d.KYCClient().Ping(ctx); err != nil {
			s.d.Logger().Errorw(ctx, "❌ KYC client is not ready", "error", err)

			sentry.CaptureException(fmt.Errorf("❌ KYC client is not ready: %w", err))
			// 🚨 Readiness check for external dependencies should return an error if they fail
			return err
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ KYC client is ready for battle! ⚔️ All identities are accounted for and customer data is verified.")
		}
	}

	losConfig, ok := s.d.Config().Health.Dependencies["los"]
	if ok && losConfig.ReadinessCheck {
		if err := s.d.LOSClient().Ping(ctx); err != nil {
			s.d.Logger().Errorw(ctx, "❌ LOS client is not ready", "error", err)

			sentry.CaptureException(fmt.Errorf("❌ LOS client is not ready: %w", err))
			// 🚨 Readiness check for external dependencies should return an error if they fail
			return err
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ LOS client is ready for battle! ⚔️ All identities are accounted for and customer data is verified.")
		}
	}

	s.d.Logger().Infow(ctx, "✅ All service dependencies are healthy and having a great day! 🎉 Time to get back to some serious SMS business!")

	return nil
}

func (s *svc) LivenessCheck(ctx context.Context) error {
	isVerbose := s.d.Config().Health.Liveness.VerboseLog
	if isVerbose {
		s.PrintServiceDependenciesHealth(ctx)
	}

	var failedServices []string

	dbConfig, ok := s.d.Config().Health.Dependencies["database"]
	if ok && dbConfig.LivenessCheck {
		if err := s.d.DB().Ping(); err != nil {
			s.d.Logger().Errorw(ctx, "❌ Database is not healthy", "error", err)
			sentry.CaptureException(fmt.Errorf("❌ Database is not healthy: %w", err))
			return fmt.Errorf("❌ Critical service Database is not healthy: %w", err)
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ Database is alive! 🧟‍♂️ It just told me a joke about SQL injections. Don't worry, I didn't laugh.")
		}
	}

	natsConfig, ok := s.d.Config().Health.Dependencies["nats"]
	if ok && natsConfig.LivenessCheck {
		if !s.d.NatsService().Ping(ctx) {
			s.d.Logger().Errorw(ctx, "❌ Nats connection is not healthy")
			sentry.CaptureException(fmt.Errorf("❌ Nats connection is not healthy"))
			return fmt.Errorf("❌ Critical service NATS is not healthy")
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ NATS is buzzing with life! 🐝 Messages are flowing faster than gossip in a small town.")
		}
	}

	kycConfig, ok := s.d.Config().Health.Dependencies["kyc"]
	if ok && kycConfig.LivenessCheck {
		if err := s.d.KYCClient().Ping(ctx); err != nil {
			s.d.Logger().Errorw(ctx, "❌ KYC client is not healthy", "error", err)
			sentry.CaptureException(fmt.Errorf("❌ KYC client is not healthy: %w", err))
			failedServices = append(failedServices, "KYC client")
		} else if isVerbose {
			s.d.Logger().Infow(ctx, "✅ KYC client is alive and verifying! 🆔 All identities are properly checked.")
		}
	}

	// Only log success if no services failed
	if len(failedServices) == 0 {
		s.d.Logger().Infow(ctx, "✅ All service dependencies are healthy and having a great day! 🎉 Time to get back to some serious business!")
	} else {
		s.d.Logger().Warnw(ctx, "⚠️ Some services are not healthy, but service is still operational",
			"failed_services", failedServices,
			"total_failed", len(failedServices))
	}

	return nil
}

func (s *svc) PrintServiceDependenciesHealth(ctx context.Context) error {
	s.d.Logger().Infow(ctx, "This Service Depends on the following dependencies:")

	for dep, config := range s.d.Config().Health.Dependencies {
		depName := strings.ToLower(dep)
		s.d.Logger().Infow(ctx, "🔗 Dependency", "name", depName, "readiness_check", config.ReadinessCheck, "liveness_check", config.LivenessCheck)
	}

	return nil
}
