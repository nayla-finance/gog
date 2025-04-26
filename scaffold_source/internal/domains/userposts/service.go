package userpost

import (
	"context"

	"github.com/PROJECT_NAME/internal/domains/interfaces"
	"github.com/PROJECT_NAME/internal/domains/model"
)

var _ Service = new(svc)

type (
	UserPost struct {
		User  model.User
		Posts []model.Post
	}

	Service interface {
		GetUserPosts(ctx context.Context, userID string) (*UserPost, error)
		DeleteUserPosts(ctx context.Context, userID string) error
	}

	serviceDependencies interface {
		interfaces.UserServiceProvider
		interfaces.PostServiceProvider
	}

	svc struct {
		d serviceDependencies
	}
)

func NewService(d serviceDependencies) *svc {
	return &svc{
		d: d,
	}
}

func (s *svc) GetUserPosts(ctx context.Context, userID string) (*UserPost, error) {
	user := &model.User{}

	err := s.d.UserService().GetUserByID(ctx, userID, user)
	if err != nil {
		return nil, err
	}

	posts := []model.Post{}
	if err := s.d.PostService().GetPostsByUserID(ctx, user.ID, &posts); err != nil {
		return nil, err
	}

	return &UserPost{
		User:  *user,
		Posts: posts,
	}, nil
}

func (s *svc) DeleteUserPosts(ctx context.Context, userID string) error {
	user := &model.User{}
	if err := s.d.UserService().GetUserByID(ctx, userID, user); err != nil {
		return err
	}

	return s.d.PostService().DeletePostsByUserID(ctx, user.ID)
}
