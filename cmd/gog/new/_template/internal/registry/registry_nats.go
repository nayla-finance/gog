package registry

import "github.com/PROJECT_NAME/internal/nats"

func (r *Registry) NatsService() nats.Service {
	return r.natsService
}

func (r *Registry) ConsumerNameBuilder() *nats.ConsumerNameBuilder {
	if r.consumerNameBuilder == nil {
		r.consumerNameBuilder = nats.NewConsumerNameBuilder()
	}

	return r.consumerNameBuilder
}
