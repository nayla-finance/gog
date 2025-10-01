package model

type Payload interface {
	Validate() error
}
