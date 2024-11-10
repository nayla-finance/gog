package user

import "github.com/project-name/internal/validator"

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
