package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/project-name/internal/config"
)

type (
	DBProvider interface {
		DB() *sqlx.DB
	}

	dbDependencies interface {
		config.ConfigProvider
	}
)

func Connect(d dbDependencies) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", d.Config().DatabaseHost, d.Config().DatabaseUsername, d.Config().DatabasePassword, d.Config().DatabaseName, d.Config().DatabasePort)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
