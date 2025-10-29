package post

import (
	"context"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/google/uuid"
	"github.com/nayla-finance/go-nayla/logger"
)

var _ interfaces.PostService = new(svc)

type (
	serviceDependencies interface {
		logger.Provider
		config.ConfigProvider
		errors.ErrorProvider
		RepositoryProvider
		interfaces.UserServiceProvider
	}

	svc struct {
		d serviceDependencies
	}
)

func NewService(d serviceDependencies) *svc {
	return &svc{d: d}
}

func (s *svc) CreatePost(ctx context.Context, dto *model.CreatePostDTO) error {
	user := &model.User{}
	if err := s.d.UserService().GetUserByID(ctx, string(dto.AuthorID), user); err != nil {
		return err
	}

	if err := s.d.PostRepository().createPost(ctx, &model.Post{
		Title:    dto.Title,
		Content:  dto.Content,
		AuthorID: user.ID,
	}); err != nil {
		return err
	}

	return nil
}

func (s *svc) GetPostByID(ctx context.Context, id uuid.UUID, post *model.Post) error {
	if err := s.d.PostRepository().getPostByID(ctx, id, post); err != nil {
		return err
	}

	return nil
}

func (s *svc) GetPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]model.Post) error {
	return s.d.PostRepository().getPostsByUserID(ctx, userID, posts)
}

func (s *svc) DeletePostsByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.d.PostRepository().deletePostsByUserID(ctx, userID)
}
