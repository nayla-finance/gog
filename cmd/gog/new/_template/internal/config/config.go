package config

import (
	"fmt"
	"strings"

	"github.com/PROJECT_NAME/internal/validator"
	"github.com/spf13/viper"
)

type ConfigProvider interface {
	Config() *Config
}

type Config struct {
	App struct {
		Name         string `mapstructure:"name"`
		Env          string `mapstructure:"env"`
		Port         int    `mapstructure:"port"`
		LogLevel     string `mapstructure:"log_level"`
		ReadTimeout  int    `mapstructure:"read_timeout"`
		WriteTimeout int    `mapstructure:"write_timeout"`
		MaxRetries   int    `mapstructure:"max_retries"`
	} `mapstructure:"app"`

	Api struct {
		Key string `mapstructure:"key" validate:"required"`
	} `mapstructure:"api"`

	Database struct {
		Host          string `mapstructure:"host" validate:"required"`
		Port          int    `mapstructure:"port" validate:"required"`
		Name          string `mapstructure:"name" validate:"required"`
		Username      string `mapstructure:"username" validate:"required"`
		Password      string `mapstructure:"password" validate:"required"`
		Synchronize   bool   `mapstructure:"synchronize"`
		Ssl           bool   `mapstructure:"ssl"`
		SSLMode       string // only "require" (default), "verify-full", "verify-ca", and "disable" supported
		MigrationsDir string `mapstructure:"migrations_dir"`
		Driver        string `mapstructure:"driver"`
		MigrateTable  string `mapstructure:"migrate_table"`
	} `mapstructure:"database"`
}

func Load(configFile string) (*Config, error) {
	fmt.Println("🔄 Loading configuration from file: ", configFile)

	v := viper.New()
	// Allow config file to be specified via -c flag
	v.SetConfigFile(configFile) // Default config file

	v.AddConfigPath(".")
	v.AutomaticEnv()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))

	// err is ignored to allow reading from os env
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Set default values
	v.SetDefault("app.name", "PROJECT_NAME")
	v.SetDefault("app.env", "production")
	v.SetDefault("app.port", 3000)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.read_timeout", 60)
	v.SetDefault("app.write_timeout", 60)
	v.SetDefault("app.max_retries", 3)
	v.SetDefault("database.migrations_dir", "migrations")
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.migrate_table", "schema_migrations")

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.Database.Ssl {
		config.Database.SSLMode = "require"
	} else {
		config.Database.SSLMode = "disable"
	}

	if err := validator.Validate(config); err != nil {
		return nil, err
	}

	return &config, nil
}
