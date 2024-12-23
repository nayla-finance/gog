package nats

import "github.com/nats-io/nats.go/jetstream"

type ConsumerHandler func(msg jetstream.Msg) error
