package tracker

import (
	"context"

	"github.com/PROJECT_NAME/internal/db"
	"github.com/google/uuid"
)

var _ Repository = &repo{}

type (
	Repository interface {
		Create(ctx context.Context, tracker *Tracker) error
		GetByID(ctx context.Context, id uuid.UUID) (*Tracker, error)
	}

	RepositoryProvider interface {
		TrackerRepository() Repository
	}

	repoDependencies interface {
		db.DBProvider
	}

	repo struct {
		d repoDependencies
	}
)

func NewRepository(d repoDependencies) Repository {
	return &repo{d: d}
}

func (r *repo) Create(ctx context.Context, tracker *Tracker) error {
	query := `
		INSERT INTO vendor_tracker (id, is_success, path, method, request_body, response_body, response_time_ms)
		VALUES (:id, :is_success, :path, :method, :request_body, :response_body, :response_time_ms)
	`

	_, err := r.d.DB().GetConn().NamedExecContext(ctx, query, tracker)

	return err
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*Tracker, error) {
	query := `
		SELECT * FROM vendor_tracker WHERE id = $1
	`

	var tracker Tracker
	if err := r.d.DB().GetConn().GetContext(ctx, &tracker, query, id); err != nil {
		return nil, err
	}

	return &tracker, nil
}
