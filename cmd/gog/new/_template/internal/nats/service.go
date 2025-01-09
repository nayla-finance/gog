package nats

import (
	"context"
	"encoding/json"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var _ Service = new(svc)

type (
	Service interface {
		Publish(ctx context.Context, subject string, payload interface{}) error
		Consume(name string, subjects []string, handler ConsumerHandler, opts ...jetstream.ConsumerConfig) (jetstream.ConsumeContext, error)
		Cleanup() error
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
	}

	svc struct {
		d   serviceDependencies
		nc  *nats.Conn
		js  jetstream.JetStream
		cfg *NatsConfig
	}
)

func NewService(d serviceDependencies) (*svc, error) {
	svc := &svc{d: d, cfg: LoadConfig(d)}
	if err := svc.Setup(); err != nil {
		return nil, err
	}
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

	var ack *jetstream.PubAck
	err = s.d.Retry().Do(func() error {
		var err error
		ack, err = s.js.Publish(ctx, subject, p)
		return err
	}, "publish-message")

	if err == nil {
		s.d.Logger().Debug("Published message", " subject ", subject, " ack ", ack.Sequence, " error ", err)
	}

	return err
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
		if err := handler(msg); err != nil {
			s.d.Logger().Error("Failed to handle message in consumer ", name, " error ", err)
		} else {
			s.d.Logger().Debug("Acked message", " subject ", msg.Subject())
			msg.Ack()
		}
	})
}

func (s *svc) HealthCheck() bool {
	return s.nc.IsConnected()
}
