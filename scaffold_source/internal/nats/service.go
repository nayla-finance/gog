package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/utils"
	"github.com/getsentry/sentry-go"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ Service = new(svc)

type (
	Service interface {
		Publish(ctx context.Context, subject string, payload interface{}) error
		Consume(name string, subjects []string, handler ConsumerHandler, opts ...jetstream.ConsumerConfig) (jetstream.ConsumeContext, error)
		Cleanup() error
		Reconnect() error
		HealthCheck() bool
	}

	ServiceProvider interface {
		NatsService() Service
	}

	serviceDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		errors.ErrorProvider
		utils.RetryProvider
		interfaces.SignalProvider
	}

	svc struct {
		d   serviceDependencies
		nc  *nats.Conn
		js  jetstream.JetStream
		cfg *NatsConfig
	}
)

var tracer trace.Tracer

func init() {
	// this seems to work even if the init happens before setting up the trace provider
	tracer = otel.Tracer("github.com/PROJECT_NAME/internal/nats")
}

func NewService(d serviceDependencies) (*svc, error) {
	svc := &svc{d: d, cfg: LoadConfig(d)}
	if err := svc.Setup(); err != nil {
		return nil, err
	}

	NewMonitoring(d, svc).Start()

	return svc, nil
}

func (s *svc) Publish(ctx context.Context, subject string, payload interface{}) error {
	if s.js == nil {
		return s.d.NewError(errors.ErrInternal, "nats jetstream is not initialized")
	}

	p, err := json.Marshal(payload)
	if err != nil {
		return s.d.NewError(errors.ErrInternal, "failed to marshal data: "+err.Error())
	}

	err = s.d.Retry().Do(func() error {
		_, err := s.js.Publish(ctx, subject, p)
		return err
	}, "publish-message")

	if err != nil {
		sentry.CaptureMessage(fmt.Sprintf("❌ Failed to publish message to %s with payload %s", subject, string(p)))
		sentry.CaptureException(err)

		s.d.Logger().Errorw("❌ Failed to publish message", "subject", subject, "error", err)
		return err
	}

	s.d.Logger().Debugw("✅ Published message", "subject", subject)

	return nil
}

func (s *svc) Consume(name string, subjects []string, handler ConsumerHandler, opts ...jetstream.ConsumerConfig) (jetstream.ConsumeContext, error) {
	if s.js == nil {
		return nil, s.d.NewError(errors.ErrInternal, "nats jetstream is not initialized")
	}

	s.d.Logger().Info("Consuming messages", " name ", name, " subjects ", subjects)

	cfg := s.cfg.DefaultConsumerConfig

	if len(opts) > 0 {
		c := opts[0]
		if c.AckWait != 0 {
			cfg.AckWait = c.AckWait
		}
		if len(c.BackOff) > 0 {
			cfg.BackOff = c.BackOff
		}
	}

	cfg.Name = name
	cfg.Durable = name
	cfg.FilterSubjects = subjects

	var consumer jetstream.Consumer
	err := s.d.Retry().Do(func() error {
		var err error
		consumer, err = s.js.CreateOrUpdateConsumer(context.Background(), s.d.Config().Nats.DefaultStreamName, cfg)
		return err
	}, "create-or-update-consumer")
	if err != nil {
		return nil, s.d.NewError(errors.ErrInternal, "failed to get consumer: "+err.Error())
	}

	s.d.Logger().Info("Consuming messages", " by ", name, " on subjects ", subjects)

	return consumer.Consume(func(msg jetstream.Msg) {
		_, span := tracer.Start(context.Background(), name)
		span.SetAttributes(attribute.String("subject", msg.Subject()))
		defer span.End()

		if err := handler(msg); err != nil {
			s.d.Logger().Error("Failed to handle message in consumer ", name, " error ", err)
			sentry.CaptureException(err)
			span.RecordError(err)
			span.SetAttributes(attribute.String("payload", string(msg.Data())))
			span.SetAttributes(attribute.String("error", err.Error()))
			span.SetStatus(codes.Error, "An error occurred while processing the message")
		} else {
			s.d.Logger().Debug("Acked message", " subject ", msg.Subject())
			span.SetStatus(codes.Ok, "message processed successfully")
			msg.Ack()
		}
	})
}

func (s *svc) HealthCheck() bool {
	return s.nc.IsConnected()
}

func (s *svc) GetJetStream() jetstream.JetStream {
	return s.js
}
