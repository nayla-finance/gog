package post

import (
	"github.com/project-name/internal/validator"
)

type CreatePostDTO struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	AuthorID string `json:"author_id" validate:"required"`
}

func (dto *CreatePostDTO) Validate() error {
	return validator.Validate(dto)
}
