package model

type Signal chan SignalPayload
type SignalType int

const (
	SignalTypeNatsConsumerRestart SignalType = iota + 1
)

type SignalPayload struct {
	Type SignalType
}
