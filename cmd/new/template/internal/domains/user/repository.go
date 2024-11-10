package user

import (
	"context"

	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/logger"
)

type (
	RepositoryProvider interface {
		UserRepository() *Repository
	}

	repositoryDependencies interface {
		logger.LoggerProvider
		db.DBProvider
	}

	Repository struct {
		d repositoryDependencies
	}
)

func NewRepository(d repositoryDependencies) *Repository {
	return &Repository{
		d: d,
	}
}

func (r *Repository) createUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().NamedExecContext(ctx, "INSERT INTO users (id, first_name, last_name, email, phone, created_at, updated_at) VALUES (:id, :first_name, :last_name, :email, :phone, :created_at, :updated_at)", user); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getUsers(ctx context.Context, users *[]User) error {
	if err := r.d.DB().SelectContext(ctx, users, "SELECT * FROM users"); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getUserByID(ctx context.Context, id string, user *User) error {
	if err := r.d.DB().GetContext(ctx, user, "SELECT * FROM users WHERE id = :id", id); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().NamedExecContext(ctx, "UPDATE users SET first_name = :first_name, last_name = :last_name, phone = :phone, updated_at = :updated_at WHERE id = :id", user); err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleteUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().NamedExecContext(ctx, "DELETE FROM users WHERE id = :id", user); err != nil {
		return err
	}

	return nil
}
