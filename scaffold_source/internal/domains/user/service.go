package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

var _ Service = new(svc)

type (
	Service interface {
		CreateUser(ctx context.Context, dto *CreateUserDTO) error
		GetUsers(ctx context.Context) ([]User, error)
		GetUserByID(ctx context.Context, id string, user *User) error
		UpdateUser(ctx context.Context, id string, dto *UpdateUserDTO) error
		DeleteUser(ctx context.Context, id string) error
	}

	ServiceProvider interface {
		UserService() Service
	}

	serviceDependencies interface {
		RepositoryProvider
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

func (s *svc) CreateUser(ctx context.Context, dto *CreateUserDTO) error {
	u := &User{
		ID:        uuid.Must(uuid.NewV7()),
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Phone:     &dto.Phone,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return s.d.UserRepository().createUser(ctx, u)
}

func (s *svc) GetUsers(ctx context.Context) ([]User, error) {
	users := []User{}

	if err := s.d.UserRepository().getUsers(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *svc) GetUserByID(ctx context.Context, id string, user *User) error {
	return s.d.UserRepository().getUserByID(ctx, id, user)
}

func (s *svc) UpdateUser(ctx context.Context, id string, dto *UpdateUserDTO) error {
	u := &User{}

	if err := s.d.UserRepository().getUserByID(ctx, id, u); err != nil {
		return err
	}

	if dto.FirstName != nil {
		u.FirstName = *dto.FirstName
	}

	if dto.LastName != nil {
		u.LastName = *dto.LastName
	}

	if dto.Phone != nil {
		u.Phone = dto.Phone
	}

	u.UpdatedAt = time.Now().UTC()

	return s.d.UserRepository().updateUser(ctx, u)
}

func (s *svc) DeleteUser(ctx context.Context, id string) error {
	return s.d.UserRepository().deleteUser(ctx, &User{ID: uuid.Must(uuid.Parse(id))})
}
