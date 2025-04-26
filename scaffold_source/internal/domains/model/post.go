package model

import (
	"time"

	"github.com/PROJECT_NAME/internal/validator"
	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	AuthorID  uuid.UUID `db:"author_id" json:"author_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Author User `db:"author" json:"author"`
}

func (p *Post) TableName() string {
	return "posts"
}

type CreatePostDTO struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	AuthorID string `json:"author_id" validate:"required"`
}

func (dto *CreatePostDTO) Validate() error {
	return validator.Validate(dto)
}
