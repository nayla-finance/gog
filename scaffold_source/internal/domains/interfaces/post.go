package interfaces

import (
	"context"

	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/google/uuid"
)

type (
	PostService interface {
		CreatePost(ctx context.Context, dto *model.CreatePostDTO) error
		GetPostByID(ctx context.Context, id uuid.UUID, post *model.Post) error
		GetPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]model.Post) error
		DeletePostsByUserID(ctx context.Context, userID uuid.UUID) error
	}

	PostServiceProvider interface {
		PostService() PostService
	}
)
