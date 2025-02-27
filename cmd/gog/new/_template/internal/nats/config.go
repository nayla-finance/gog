package nats

import (
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	configDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
	}

	NatsConfig struct {
		ConnectionOptions     []nats.Option
		Streams               []jetstream.StreamConfig
		DefaultConsumerConfig jetstream.ConsumerConfig
	}
)

func LoadConfig(d configDependencies) *NatsConfig {
	cfg := &NatsConfig{}

	cfg.ConnectionOptions = []nats.Option{
		nats.UserCredentials(d.Config().Nats.CredsPath),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			d.Logger().Errorf("❌ NATS connection error: %s", err)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			d.Logger().Info("✅ NATS connection closed")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			d.Logger().Infof("🔄 Reconnected [%s]", nc.ConnectedUrl())
		}),
		func(o *nats.Options) error {
			// provide a unique name for each connection
			o.Name = d.Config().Nats.ClientName + "_" + uuid.New().String()
			o.AllowReconnect = true
			return nil
		},
	}

	cfg.Streams = []jetstream.StreamConfig{
		{
			Name:        d.Config().Nats.DefaultStreamName,
			Replicas:    1,
			MaxMsgs:     -1,
			MaxBytes:    -1,
			Compression: jetstream.S2Compression,
			Discard:     jetstream.DiscardOld,
			MaxAge:      0,
			Storage:     jetstream.FileStorage,
			Subjects:    d.Config().Nats.DefaultStreamSubjects,
			Retention:   jetstream.LimitsPolicy,
		},
	}

	backoffDurations := d.Config().Nats.Consumer.BackoffDurations
	// NOTE(🚨): max deliver is required to be > length of backoff values
	for len(backoffDurations) < d.Config().Nats.Consumer.MaxDeliver-1 {
		backoffDurations = append(backoffDurations, d.Config().Nats.Consumer.DefaultBackoffDuration)
	}

	cfg.DefaultConsumerConfig = jetstream.ConsumerConfig{
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		MaxDeliver:    d.Config().Nats.Consumer.MaxDeliver,
		// MaxAckPending: 1000, // use default 1000
		ReplayPolicy: jetstream.ReplayInstantPolicy,
		BackOff:      backoffDurations,
	}

	return cfg
}
