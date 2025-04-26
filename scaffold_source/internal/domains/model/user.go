package model

import (
	"time"

	"github.com/PROJECT_NAME/internal/validator"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	FirstName string    `db:"first_name" json:"firstName"`
	LastName  string    `db:"last_name" json:"lastName"`
	Email     string    `db:"email" json:"email"`
	Phone     *string   `db:"phone" json:"phone"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	Posts     []Post    `db:"-" json:"posts"`
}

func (u *User) TableName() string {
	return "users"
}

type CreateUserDTO struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required"`
}

func (dto *CreateUserDTO) Validate() error {
	return validator.Validate(dto)
}

type UpdateUserDTO struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Phone     *string `json:"phone"`
}

func (dto *UpdateUserDTO) Validate() error {
	return validator.Validate(dto)
}
