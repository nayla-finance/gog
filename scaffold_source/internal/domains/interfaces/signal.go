package interfaces

import "github.com/PROJECT_NAME/internal/domains/model"

type SignalProvider interface {
	SendSignal(signal model.SignalPayload)
}
