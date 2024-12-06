package db

import (
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/jmoiron/sqlx"
)

var _ Database = new(db)

type (
	Database interface {
		GetConn() *sqlx.DB
		Transaction(fn func(tx *sqlx.Tx) error) error
		Close() error
		Ping() error
	}

	DBProvider interface {
		DB() Database
	}

	dbDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
	}

	db struct {
		conn *sqlx.DB
	}
)

func Connect(d dbDependencies) (*db, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", d.Config().DatabaseHost, d.Config().DatabaseUsername, d.Config().DatabasePassword, d.Config().DatabaseName, d.Config().DatabasePort, d.Config().DatabaseSSLMode)

	d.Logger().Debug(fmt.Sprintf("🔄 Connecting to '%s' database with user '%s'...", d.Config().DatabaseName, d.Config().DatabaseUsername))
	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		d.Logger().Error("❌ Failed to connect to database", "error", err)
		return nil, err
	}

	d.Logger().Info("✅ Successfully connected to database")

	return &db{conn}, nil
}

func (c *db) Ping() error {
	return c.conn.Ping()
}

func (c *db) GetConn() *sqlx.DB {
	return c.conn
}

func (c *db) Close() error {
	return c.conn.Close()
}

// Transaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If the function executes successfully, the transaction is committed.
// If a rollback fails after a function error, both errors are returned.
// Returns any error that occurred during transaction handling.
func (c *db) Transaction(fn func(tx *sqlx.Tx) error) error {
	tx, err := c.conn.Beginx()
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
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
