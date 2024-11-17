package db

import (
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/jmoiron/sqlx"
)

type (
	DBProvider interface {
		DB() *Connection
	}

	dbDependencies interface {
		config.ConfigProvider
	}

	Connection struct {
		*sqlx.DB
	}
)

func Connect(d dbDependencies) (*Connection, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", d.Config().DatabaseHost, d.Config().DatabaseUsername, d.Config().DatabasePassword, d.Config().DatabaseName, d.Config().DatabasePort)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{db}, nil
}

// Transaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If the function executes successfully, the transaction is committed.
// If a rollback fails after a function error, both errors are returned.
// Returns any error that occurred during transaction handling.
func (c *Connection) Transaction(fn func(tx *sqlx.Tx) error) error {
	tx, err := c.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback is called on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback() // Ignore rollback error on panic
			panic(p)          // Re-panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
		}
		return fmt.Errorf("tx failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
