package nats

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/getsentry/sentry-go"
	"github.com/nats-io/nats.go"
)

type (
	monitoringDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		interfaces.SignalProvider
	}

	monitoring struct {
		d   monitoringDependencies
		svc *svc
	}
)

func NewMonitoring(d monitoringDependencies, svc *svc) *monitoring {
	return &monitoring{d: d, svc: svc}
}

func (m *monitoring) Start() {
	if !m.d.Config().Nats.Monitoring.Enabled {
		m.d.Logger().Warn("⚠️ NATS monitoring is disabled")
		return
	}

	m.d.Logger().Info("🔄 Starting NATS monitoring")

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		ticker := time.NewTicker(m.d.Config().Nats.Monitoring.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-sigChan:
				m.d.Logger().Info("🔄 NATS monitoring stopped")
				return
			case <-ticker.C:
				m.checkConsumers()
			}
		}
	}()

	m.d.Logger().Info("✅ NATS monitoring started")
}

func (m *monitoring) checkConsumers() {
	m.d.Logger().Debug("🔄 Checking NATS consumers")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := m.svc.GetJetStream().Stream(ctx, m.d.Config().Nats.DefaultStreamName)
	if err != nil {
		m.d.Logger().Errorw("❌ Failed to get stream", "error", err)
		if err == nats.ErrConnectionClosed {
			m.d.Logger().Error("captured connection closed error in monitoring")
			sentry.CaptureMessage("NATS connection closed in monitoring")
			m.d.SendSignal(model.SignalPayload{
				Type: model.SignalTypeNatsConsumerRestart,
			})
		} else {
			m.d.Logger().Errorw("❌ Failed to get stream", "error", err)
			sentry.CaptureException(err)
		}

		return
	}

	for consumer := range stream.ListConsumers(ctx).Info() {
		if _, ok := m.d.Config().Nats.Monitoring.ExcludedConsumers[consumer.Name]; ok {
			m.d.Logger().Debugw("🔄 Skipping consumer as it is excluded", "consumer", consumer.Name, "stream", m.d.Config().Nats.DefaultStreamName)
			continue
		}

		threshold := uint64(m.d.Config().Nats.Monitoring.PendingMessagesThreshold)
		if consumer.NumPending >= threshold {
			m.d.Logger().Errorw("❌ Consumer has too many pending messages", "consumer", consumer.Name, "stream", m.d.Config().Nats.DefaultStreamName, "pending_messages", consumer.NumPending)

			sentry.CaptureMessage(fmt.Sprintf("Consumer %s of stream %s has too many pending messages: %d", consumer.Name, m.d.Config().Nats.DefaultStreamName, consumer.NumPending))
			m.d.SendSignal(model.SignalPayload{
				Type: model.SignalTypeNatsConsumerRestart,
			})
			break
		}
	}

	m.d.Logger().Debug("✅ NATS consumers check finished successfully")
}
