package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	ServiceProvider interface {
		UserService() *Service
	}

	serviceDependencies interface {
		RepositoryProvider
	}

	Service struct {
		d serviceDependencies
	}
)

func NewService(d serviceDependencies) *Service {
	return &Service{
		d: d,
	}
}

func (s *Service) CreateUser(ctx context.Context, dto *CreateUserDTO) error {
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

func (s *Service) GetUsers(ctx context.Context) ([]User, error) {
	users := []User{}

	if err := s.d.UserRepository().getUsers(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}

	if err := s.d.UserRepository().getUserByID(ctx, id, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, dto *UpdateUserDTO) error {
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

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.d.UserRepository().deleteUser(ctx, &User{ID: uuid.Must(uuid.Parse(id))})
}
