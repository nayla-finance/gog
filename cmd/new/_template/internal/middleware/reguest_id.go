package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (m *RequestIDMiddleware) Handle(c *fiber.Ctx) error {
	requestID := uuid.New().String()
	c.Locals("RequestID", requestID)

	return c.Next()
}
