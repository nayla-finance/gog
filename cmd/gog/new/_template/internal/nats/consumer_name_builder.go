package nats

import (
	"fmt"
	"strings"
)

type (
	ConsumerNameBuilderProvider interface {
		ConsumerNameBuilder() *ConsumerNameBuilder
	}

	ConsumerNameBuilder struct {
		prefix string
		suffix string
	}
)

func NewConsumerNameBuilder() *ConsumerNameBuilder {
	return &ConsumerNameBuilder{
		prefix: "PROJECT_NAME",
		suffix: "consumer",
	}
}

// Build creates a consumer name by combining the given name with prefix and suffix.
// The name parameter should be the base consumer name without the 'PROJECT_NAME' prefix or 'consumer' suffix.
// For example:
//   - Input: "business-verified"
//   - Output: "PROJECT_NAME-business-verified-consumer"
func (b *ConsumerNameBuilder) Build(name string) string {
	if !strings.HasPrefix(name, b.prefix) {
		name = fmt.Sprintf("%s-%s", b.prefix, name)
	}

	if !strings.HasSuffix(name, b.suffix) {
		name = fmt.Sprintf("%s-%s", name, b.suffix)
	}

	return name
}
