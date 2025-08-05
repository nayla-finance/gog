package health

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/PROJECT_NAME/internal/clients/testclient"
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/nats"
	"github.com/getsentry/sentry-go"
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
		logger.LoggerProvider
		db.DBProvider
		nats.ServiceProvider
		config.ConfigProvider
		testclient.ClientProvider
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
	if s.readinessCheckSkipped < s.d.Config().Health.Readiness.InitialChecksToSkip {
		s.readinessCheckSkipped++
		s.d.Logger().Infow("✅ Readiness check skipped", "skipped_checks", s.readinessCheckSkipped)
		return nil
	}

	isVerbose := s.d.Config().Health.Readiness.VerboseLog
	if isVerbose {
		s.PrintServiceDependenciesHealth(ctx)
	}

	if s.d.Config().Health.Dependencies.Database.ReadinessCheck {
		if err := s.d.DB().Ping(); err != nil {
			s.d.Logger().Errorw("❌ Database is not healthy", "error", err)

			sentry.CaptureException(fmt.Errorf("❌ Database is not healthy: %w", err))
			// 🚨 Readiness check for internal dependencies should return an error if they fail
			return err
		} else if isVerbose {
			s.d.Logger().Info("✅ Database is ready and caffeinated! ☕ It's got its schemas in order and its transactions committed.")
		}
	}

	if s.d.Config().Health.Dependencies.Nats.ReadinessCheck {
		if !s.d.NatsService().HealthCheck() {
			s.d.Logger().Error("❌ Nats connection is not ready")

			sentry.CaptureException(fmt.Errorf("❌ Nats connection is not ready"))
			// 🚨 Readiness check for internal dependencies should return an error if they fail
			return fmt.Errorf("❌ Nats connection is not ready")
		} else if isVerbose {
			s.d.Logger().Info("✅ NATS is ready to deliver! 📮 Like a postal service that actually works on time.")
		}
	}

	if s.d.Config().Health.Dependencies.TestClient.ReadinessCheck {
		if err := s.d.TestClient().IsReady(); err != nil {
			s.d.Logger().Errorw("❌ TestClient is not ready", "error", err)

			sentry.CaptureException(fmt.Errorf("❌ TestClient is not ready: %w", err))
			// 🚨 Readiness check for external dependencies should return an error if they fail
			return err
		} else if isVerbose {
			s.d.Logger().Info("✅ TestClient is ready for battle! ⚔️ All identities are accounted for and customer data is verified.")
		}
	}

	s.d.Logger().Info("✅ All service dependencies are healthy and having a great day! 🎉 Time to get back to some serious SMS business!")

	return nil
}

func (s *svc) LivenessCheck(ctx context.Context) error {
	if s.livenessCheckSkipped < s.d.Config().Health.Liveness.InitialChecksToSkip {
		s.livenessCheckSkipped++
		s.d.Logger().Infow("✅ Liveness check skipped", "skipped_checks", s.livenessCheckSkipped)
		return nil
	}

	isVerbose := s.d.Config().Health.Liveness.VerboseLog
	if isVerbose {
		s.PrintServiceDependenciesHealth(ctx)
	}

	var failedServices []string

	if s.d.Config().Health.Dependencies.Database.LivenessCheck {
		if err := s.d.DB().Ping(); err != nil {
			s.d.Logger().Errorw("❌ Database is not healthy", "error", err)
			sentry.CaptureException(fmt.Errorf("❌ Database is not healthy: %w", err))
			return fmt.Errorf("❌ Critical service Database is not healthy: %w", err)
		} else if isVerbose {
			s.d.Logger().Info("✅ Database is alive! 🧟‍♂️ It just told me a joke about SQL injections. Don't worry, I didn't laugh.")
		}
	}

	if s.d.Config().Health.Dependencies.Nats.LivenessCheck {
		if !s.d.NatsService().HealthCheck() {
			s.d.Logger().Error("❌ Nats connection is not healthy")
			sentry.CaptureException(fmt.Errorf("❌ Nats connection is not healthy"))
			return fmt.Errorf("❌ Critical service NATS is not healthy")
		} else if isVerbose {
			s.d.Logger().Info("✅ NATS is buzzing with life! 🐝 Messages are flowing faster than gossip in a small town.")
		}
	}

	if s.d.Config().Health.Dependencies.TestClient.LivenessCheck {
		if err := s.d.TestClient().IsReady(); err != nil {
			s.d.Logger().Errorw("❌ TestClient is not healthy", "error", err)
			sentry.CaptureException(fmt.Errorf("❌ TestClient is not healthy: %w", err))
			failedServices = append(failedServices, "TestClient")
		} else if isVerbose {
			s.d.Logger().Info("✅ TestClient is alive and verifying! 🆔 All identities are properly checked.")
		}
	}

	// Only log success if no services failed
	if len(failedServices) == 0 {
		s.d.Logger().Info("✅ All service dependencies are healthy and having a great day! 🎉 Time to get back to some serious business!")
	} else {
		s.d.Logger().Warnw("⚠️ Some services are not healthy, but service is still operational",
			"failed_services", failedServices,
			"total_failed", len(failedServices))
	}

	return nil
}

func (s *svc) PrintServiceDependenciesHealth(ctx context.Context) error {
	s.d.Logger().Info("This Service Depends on the following dependencies:")

	// Use reflection to dynamically get dependency names and values
	deps := s.d.Config().Health.Dependencies
	depsValue := reflect.ValueOf(deps)
	depsType := reflect.TypeOf(deps)

	for i := 0; i < depsValue.NumField(); i++ {
		field := depsType.Field(i)
		value := depsValue.Field(i)

		// Get the mapstructure tag name, fallback to field name
		depName := field.Tag.Get("mapstructure")
		if depName == "" {
			depName = strings.ToLower(field.Name)
		}

		// Get the Dependency struct value
		depConfig := value.Interface().(config.Dependency)

		s.d.Logger().Infof("🔗 %s, Readiness check enabled: %t, Liveness check enabled: %t", depName, depConfig.ReadinessCheck, depConfig.LivenessCheck)
	}

	return nil
}
