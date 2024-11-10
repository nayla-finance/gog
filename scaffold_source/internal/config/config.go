package config

import (
	"strings"

	"github.com/project-name/internal/validator"
	"github.com/spf13/viper"
)

type ConfigProvider interface {
	Config() *Config
}

type Config struct {
	AppName string `mapstructure:"APP_NAME"`

	Env string `mapstructure:"ENV"`

	LogLevel string `mapstructure:"LOG_LEVEL"`
	Port     int    `mapstructure:"PORT"`

	ReadTimeout  int `mapstructure:"READ_TIMEOUT"`
	WriteTimeout int `mapstructure:"WRITE_TIMEOUT"`

	MaxTries int `mapstructure:"MAX_TRIES"`

	ApiKey string `mapstructure:"API_KEY"`

	DatabaseHost          string `mapstructure:"DATABASE_HOST" validate:"required"`
	DatabasePort          int    `mapstructure:"DATABASE_PORT" validate:"required"`
	DatabaseName          string `mapstructure:"DATABASE_NAME" validate:"required"`
	DatabaseUsername      string `mapstructure:"DATABASE_USERNAME" validate:"required"`
	DatabasePassword      string `mapstructure:"DATABASE_PASSWORD" validate:"required"`
	DatabaseSynchronize   bool   `mapstructure:"DATABASE_SYNCHRONIZE"`
	DatabaseSsl           bool   `mapstructure:"DATABASE_SSL"`
	DatabaseMigrationsDir string `mapstructure:"DATABASE_MIGRATIONS_DIR"`
	DatabaseDriver        string `mapstructure:"DATABASE_DRIVER"`
	DatabaseMigrateTable  string `mapstructure:"DATABASE_MIGRATE_TABLE"`

	// Add more configs here ...
}

func Load() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	// err ignored to allow reading from os env
	_ = viper.ReadInConfig()

	// Load environment variables from the OS
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("APP_NAME", "Project Name")
	viper.SetDefault("ENV", "production")
	viper.SetDefault("PORT", 3000)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("READ_TIMEOUT", 60)
	viper.SetDefault("WRITE_TIMEOUT", 60)
	viper.SetDefault("DATABASE_MIGRATIONS_DIR", "internal/db/migrations")
	viper.SetDefault("DATABASE_DRIVER", "postgres")
	viper.SetDefault("DATABASE_MIGRATE_TABLE", "schema_migrations")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := validator.Validate(config); err != nil {
		return nil, err
	}

	return &config, nil
}
