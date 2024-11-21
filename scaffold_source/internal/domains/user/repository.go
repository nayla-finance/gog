package user

import (
	"context"

	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/logger"
)

var _ Repository = new(repo)

type (
	Repository interface {
		createUser(ctx context.Context, user *User) error
		getUsers(ctx context.Context, users *[]User) error
		getUserByID(ctx context.Context, id string, user *User) error
		updateUser(ctx context.Context, user *User) error
		deleteUser(ctx context.Context, user *User) error
	}

	RepositoryProvider interface {
		UserRepository() Repository
	}

	repositoryDependencies interface {
		logger.LoggerProvider
		db.DBProvider
	}

	repo struct {
		d repositoryDependencies
	}
)

func NewRepository(d repositoryDependencies) *repo {
	return &repo{
		d: d,
	}
}

func (r *repo) createUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().GetConn().NamedExecContext(ctx, "INSERT INTO users (id, first_name, last_name, email, phone, created_at, updated_at) VALUES (:id, :first_name, :last_name, :email, :phone, :created_at, :updated_at)", user); err != nil {
		return err
	}

	return nil
}

func (r *repo) getUsers(ctx context.Context, users *[]User) error {
	if err := r.d.DB().GetConn().SelectContext(ctx, users, "SELECT * FROM users"); err != nil {
		return err
	}

	return nil
}

func (r *repo) getUserByID(ctx context.Context, id string, user *User) error {
	if err := r.d.DB().GetConn().GetContext(ctx, user, "SELECT * FROM users WHERE id = :id", id); err != nil {
		return err
	}

	return nil
}

func (r *repo) updateUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().GetConn().NamedExecContext(ctx, "UPDATE users SET first_name = :first_name, last_name = :last_name, phone = :phone, updated_at = :updated_at WHERE id = :id", user); err != nil {
		return err
	}

	return nil
}

func (r *repo) deleteUser(ctx context.Context, user *User) error {
	if _, err := r.d.DB().GetConn().NamedExecContext(ctx, "DELETE FROM users WHERE id = :id", user); err != nil {
		return err
	}

	return nil
}
