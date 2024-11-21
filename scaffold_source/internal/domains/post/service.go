package post

import (
	"context"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/user"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/google/uuid"
)

var _ Service = new(svc)

type (
	Service interface {
		CreatePost(ctx context.Context, dto *CreatePostDTO) error
		GetPostByID(ctx context.Context, id uuid.UUID, post *Post) error
		GetPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]Post) error
		DeletePostsByUserID(ctx context.Context, userID uuid.UUID) error
	}

	ServiceProvider interface {
		PostService() Service
	}

	serviceDependencies interface {
		logger.LoggerProvider
		config.ConfigProvider
		errors.ErrorProvider
		RepositoryProvider
		user.ServiceProvider
	}

	svc struct {
		d serviceDependencies
	}
)

func NewService(d serviceDependencies) *svc {
	return &svc{d: d}
}

func (s *svc) CreatePost(ctx context.Context, dto *CreatePostDTO) error {
	user := &user.User{}
	if err := s.d.UserService().GetUserByID(ctx, string(dto.AuthorID), user); err != nil {
		return err
	}

	if err := s.d.PostRepository().createPost(ctx, &Post{
		Title:    dto.Title,
		Content:  dto.Content,
		AuthorID: user.ID,
	}); err != nil {
		return err
	}

	return nil
}

func (s *svc) GetPostByID(ctx context.Context, id uuid.UUID, post *Post) error {
	if err := s.d.PostRepository().getPostByID(ctx, id, post); err != nil {
		return err
	}

	return nil
}

func (s *svc) GetPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]Post) error {
	return s.d.PostRepository().getPostsByUserID(ctx, userID, posts)
}

func (s *svc) DeletePostsByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.d.PostRepository().deletePostsByUserID(ctx, userID)
}
