package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `db:"id"`

	FirstName string `db:"first_name"`

	LastName string `db:"last_name"`

	Email string `db:"email"`

	Phone *string `db:"phone"`

	CreatedAt time.Time `db:"created_at"`

	UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
