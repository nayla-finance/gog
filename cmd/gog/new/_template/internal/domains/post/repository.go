package post

import (
	"context"

	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/domains/model"
	"github.com/google/uuid"
)

var _ Repository = new(repo)

type (
	Repository interface {
		createPost(ctx context.Context, post *model.Post) error
		getPostByID(ctx context.Context, id uuid.UUID, post *model.Post) error
		getPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]model.Post) error
		deletePostsByUserID(ctx context.Context, userID uuid.UUID) error
	}

	RepositoryProvider interface {
		PostRepository() Repository
	}

	repositoryDependencies interface {
		db.DBProvider
	}

	repo struct {
		d repositoryDependencies
	}
)

func NewRepository(d repositoryDependencies) *repo {
	return &repo{d: d}
}

func (r *repo) createPost(ctx context.Context, post *model.Post) error {
	if _, err := r.d.DB().GetConn().NamedExecContext(ctx, "INSERT INTO posts (id, title, content, author_id) VALUES (:title, :content, :author_id)", post); err != nil {
		return err
	}

	return nil
}

func (r *repo) getPostByID(ctx context.Context, id uuid.UUID, p *model.Post) error {
	return r.d.DB().GetConn().GetContext(ctx, p, "SELECT * FROM posts WHERE id = $1", id)
}

func (r *repo) getPostsByUserID(ctx context.Context, userID uuid.UUID, posts *[]model.Post) error {
	return r.d.DB().GetConn().SelectContext(ctx, posts, "SELECT * FROM posts WHERE author_id = $1", userID)
}

func (r *repo) deletePostsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.d.DB().GetConn().ExecContext(ctx, "DELETE FROM posts WHERE author_id = $1", userID)
	return err
}
