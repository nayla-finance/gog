package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

type ConsumerHandler func(ctx context.Context, msg jetstream.Msg) error
