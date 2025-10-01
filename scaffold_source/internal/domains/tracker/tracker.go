package tracker

import (
	"encoding/json"
	"time"

	"github.com/PROJECT_NAME/internal/validator"
	"github.com/google/uuid"
)

type Tracker struct {
	ID             uuid.UUID       `db:"id" json:"id"`                             // UUID
	IsSuccess      *bool           `db:"is_success" json:"is_success"`             // NULL means unknown
	Path           *string         `db:"path" json:"path"`                         // Can be NULL
	Method         *string         `db:"method" json:"method"`                     // Can be NULL
	RequestBody    json.RawMessage `db:"request_body" json:"request_body"`         // JSONB, must not be NULL
	ResponseBody   string          `db:"response_body" json:"response_body"`       // TEXT, must not be NULL
	ResponseTimeMs int64           `db:"response_time_ms" json:"response_time_ms"` // INTEGER, must not be NULL
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`             // Timestamp, not NULL
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`             // Timestamp, not NULL
}

func (t *Tracker) TableName() string {
	return "vendor_tracker"
}

type SaveCallDto struct {
	RequestID    *uuid.UUID    // Optional request ID, can be nil
	IsSuccess    *bool         // Optional success status, can be nil
	ReqBody      []byte        `validate:"required"` // Raw request body bytes
	Path         *string       // Optional path, can be nil
	Method       *string       // Optional method, can be nil
	RespBody     []byte        `validate:"required"` // Raw response body bytes
	ResponseTime time.Duration `validate:"required"` // Response time in milliseconds
}

func (dto *SaveCallDto) Validate() error {
	return validator.Validate(dto)
}
