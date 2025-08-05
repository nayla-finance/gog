package testclient

import (
	"time"

	"github.com/google/uuid"
)

type FancyResponseFromTestClient struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
