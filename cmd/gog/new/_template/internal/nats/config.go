package nats

import (
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	configDependencies interface {
		config.ConfigProvider
	}

	NatsConfig struct {
		ConnectionOptions     nats.Option
		Streams               []jetstream.StreamConfig
		DefaultConsumerConfig jetstream.ConsumerConfig
	}
)

func LoadConfig(d configDependencies) *NatsConfig {
	cfg := &NatsConfig{}

	cfg.ConnectionOptions = func(o *nats.Options) error {
		o.Name = d.Config().Nats.ClientName
		o.AllowReconnect = true
		return nil
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

	cfg.DefaultConsumerConfig = jetstream.ConsumerConfig{
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		MaxDeliver:    5,
		// MaxAckPending: 1000, // use default 1000
		ReplayPolicy: jetstream.ReplayInstantPolicy,
		BackOff:      []time.Duration{30 * time.Second, 5 * time.Minute, 1 * time.Hour, 12 * time.Hour},
	}

	return cfg
}
