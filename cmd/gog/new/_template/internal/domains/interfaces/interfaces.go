package interfaces

// Just to prevent using these names

type Service interface{}

type ServiceProvider interface{}

type Payload interface {
	Validate() error
}
