package config

import (
	"os"
	"reflect"
	"strings"

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
	viper.SetDefault("DATABASE_MIGRATIONS_DIR", "migrations")
	viper.SetDefault("DATABASE_DRIVER", "postgres")
	viper.SetDefault("DATABASE_MIGRATE_TABLE", "schema_migrations")

	// override values from env if present
	t := reflect.TypeOf(Config{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envKey := strings.Split(field.Tag.Get("mapstructure"), ",")[0]
		if envKey == "" {
			continue
		}

		if val := getEnvCaseInsensitive(envKey); val != "" {
			viper.Set(envKey, val)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
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

func getEnvCaseInsensitive(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	if val := os.Getenv(strings.ToUpper(key)); val != "" {
		return val
	}

	if val := os.Getenv(strings.ToLower(key)); val != "" {
		return val
	}

	return ""
}
