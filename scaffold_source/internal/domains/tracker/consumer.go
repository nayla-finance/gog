package tracker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/nats"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	consumerDependencies interface {
		logger.LoggerProvider
		config.ConfigProvider
		nats.ServiceProvider
		nats.ConsumerNameBuilderProvider
		ServiceProvider
	}

	consumer struct {
		d consumerDependencies
	}
)

func NewConsumer(d consumerDependencies) *consumer {
	return &consumer{d: d}
}

func (c *consumer) RegisterConsumers() error {
	_, err := c.d.NatsService().Consume(
		c.d.ConsumerNameBuilder().Build("sms-calls-completed-tracker"),
		[]string{SubjectCallCompleted},
		c.callCompletedHandler,
	)
	if err != nil {
		c.d.Logger().Errorw("❌ Failed to register calls completed consumer", "error", err)
		return err
	}

	return nil
}

func (c *consumer) callCompletedHandler(ctx context.Context, msg jetstream.Msg) error {
	c.d.Logger().Infow("🚀 Entering callCompletedHandler", "subject", msg.Subject())

	var dto SaveCallDto
	if err := c.unmarshalAndValidate(msg.Data(), &dto); err != nil {
		c.d.Logger().Errorw("❌ Failed to unmarshal and validate payload", "error", err)
		return nil
	}

	if err := c.d.TrackerService().saveCall(ctx, dto); err != nil {
		c.d.Logger().Errorw("❌ Failed to save call", "error", err)
		return err
	}

	return nil
}

func (c *consumer) unmarshalAndValidate(data []byte, p model.Payload) error {
	if err := json.Unmarshal(data, p); err != nil {
		c.d.Logger().Errorw("❌ Failed to unmarshal payload", "error", err, "payload_type", fmt.Sprintf("%T", p))
		return err
	}

	if err := p.Validate(); err != nil {
		c.d.Logger().Errorw("❌ Payload validation failed", "error", err, "payload_type", fmt.Sprintf("%T", p))
		return err
	}

	return nil
}
