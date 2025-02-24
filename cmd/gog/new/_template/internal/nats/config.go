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
