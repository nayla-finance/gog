package post

import (
	"context"

	"github.com/project-name/internal/db"
)

type (
	RepositoryProvider interface {
		PostRepository() *Repository
	}

	repositoryDependencies interface {
		db.DBProvider
	}

	Repository struct {
		d repositoryDependencies
	}
)

func NewRepository(d repositoryDependencies) *Repository {
	return &Repository{d: d}
}

func (r *Repository) createPost(ctx context.Context, post *Post) error {
	if _, err := r.d.DB().NamedExecContext(ctx, "INSERT INTO posts (id, title, content, author_id) VALUES (:title, :content, :author_id)", post); err != nil {
		return err
	}

	return nil
}

func (r *Repository) getPostByID(ctx context.Context, p *Post) error {
	return r.d.DB().GetContext(ctx, p, "SELECT * FROM posts WHERE id = :id", p)
}
