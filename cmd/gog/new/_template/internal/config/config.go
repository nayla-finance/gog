package config

import (
	"reflect"

	"github.com/PROJECT_NAME/internal/validator"
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

	DatabaseHost        string `mapstructure:"DATABASE_HOST" validate:"required"`
	DatabasePort        int    `mapstructure:"DATABASE_PORT" validate:"required"`
	DatabaseName        string `mapstructure:"DATABASE_NAME" validate:"required"`
	DatabaseUsername    string `mapstructure:"DATABASE_USERNAME" validate:"required"`
	DatabasePassword    string `mapstructure:"DATABASE_PASSWORD" validate:"required"`
	DatabaseSynchronize bool   `mapstructure:"DATABASE_SYNCHRONIZE"`
	DatabaseSsl         bool   `mapstructure:"DATABASE_SSL"`
	// only "require" (default), "verify-full", "verify-ca", and "disable" supported
	DatabaseSSLMode       string
	DatabaseMigrationsDir string `mapstructure:"DATABASE_MIGRATIONS_DIR"`
	DatabaseDriver        string `mapstructure:"DATABASE_DRIVER"`
	DatabaseMigrateTable  string `mapstructure:"DATABASE_MIGRATE_TABLE"`

	// Add more configs here ...
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.AddConfigPath(".")

	v.AutomaticEnv()

	// err is ignored to allow reading from os env
	_ = v.ReadInConfig()

	// First, bind all possible environment variables based on struct tags
	t := reflect.TypeOf(Config{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if envKey := field.Tag.Get("mapstructure"); envKey != "" {
			err := v.BindEnv(envKey)
			if err != nil {
				return nil, err
			}
		}
	}

	// Set default values
	v.SetDefault("APP_NAME", "Project Name")
	v.SetDefault("ENV", "production")
	v.SetDefault("PORT", 3000)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("READ_TIMEOUT", 60)
	v.SetDefault("WRITE_TIMEOUT", 60)
	v.SetDefault("DATABASE_MIGRATIONS_DIR", "migrations")
	v.SetDefault("DATABASE_DRIVER", "postgres")
	v.SetDefault("DATABASE_MIGRATE_TABLE", "schema_migrations")

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.DatabaseSsl {
		config.DatabaseSSLMode = "require"
	} else {
		config.DatabaseSSLMode = "disable"
	}

	if err := validator.Validate(config); err != nil {
		return nil, err
	}

	return &config, nil
}
