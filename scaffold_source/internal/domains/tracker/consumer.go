package tracker

import (
	"context"
	"encoding/json"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nayla-finance/go-nayla/logger"
	"github.com/nayla-finance/go-nayla/nats"
)

type (
	consumerDependencies interface {
		logger.Provider
		config.ConfigProvider
		nats.ServiceProvider
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
		"calls-completed-tracker",
		[]string{SubjectCallCompleted},
		c.callCompletedHandler,
	)
	if err != nil {
		c.d.Logger().Errorw(context.Background(), "❌ Failed to register calls completed consumer", "error", err)
		return err
	}

	return nil
}

func (c *consumer) callCompletedHandler(ctx context.Context, msg jetstream.Msg) error {
	c.d.Logger().Infow(ctx, "🚀 Entering callCompletedHandler", "subject", msg.Subject())

	var dto SaveCallDto
	if err := c.unmarshalAndValidate(msg.Data(), &dto); err != nil {
		c.d.Logger().Errorw(ctx, "❌ Failed to unmarshal and validate payload", "error", err)
		return nil
	}

	if err := c.d.TrackerService().saveCall(ctx, dto); err != nil {
		c.d.Logger().Errorw(ctx, "❌ Failed to save call", "error", err)
		return err
	}

	return nil
}

func (c *consumer) unmarshalAndValidate(data []byte, p interfaces.Payload) error {
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	return nil
}
