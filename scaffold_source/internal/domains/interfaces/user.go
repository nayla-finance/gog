package interfaces

import (
	"context"

	"github.com/PROJECT_NAME/internal/domains/model"
)

type (
	UserService interface {
		CreateUser(ctx context.Context, dto *model.CreateUserDTO) error
		GetUsers(ctx context.Context) ([]model.User, error)
		GetUserByID(ctx context.Context, id string, user *model.User) error
		UpdateUser(ctx context.Context, id string, dto *model.UpdateUserDTO) error
		DeleteUser(ctx context.Context, id string) error
	}

	UserServiceProvider interface {
		UserService() UserService
	}
)
