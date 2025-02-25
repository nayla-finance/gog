package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	// Embed timezone data
	_ "time/tzdata"

	"github.com/PROJECT_NAME/internal/validator"
	"github.com/spf13/viper"
)

type ConfigProvider interface {
	Config() *Config
}

type (
	App struct {
		Name         string        `mapstructure:"name"`
		Env          string        `mapstructure:"env"`
		Port         int           `mapstructure:"port"`
		LogLevel     string        `mapstructure:"log_level"`
		ReadTimeout  int           `mapstructure:"read_timeout"`
		WriteTimeout int           `mapstructure:"write_timeout"`
		MaxRetries   int           `mapstructure:"max_retries"`
		RetryDelay   time.Duration `mapstructure:"retry_delay"`
		Timezone     string        `mapstructure:"timezone"`
	}

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
	}

	Api struct {
		Key          string   `mapstructure:"key" validate:"required"`
		PublicRoutes []string `mapstructure:"public_routes"`
	}

	NatsConsumer struct {
		MaxDeliver             int             `mapstructure:"max_deliver"`
		BackoffDurations       []time.Duration `mapstructure:"backoff_durations"`
		DefaultBackoffDuration time.Duration   `mapstructure:"default_backoff_duration"`
	}

	Nats struct {
		Servers               string       `mapstructure:"servers" validate:"required"`
		ClientName            string       `mapstructure:"client_name" validate:"required"`
		CredsPath             string       `mapstructure:"creds_path" validate:"required"`
		DefaultStreamName     string       `mapstructure:"default_stream_name" validate:"required"`
		DefaultStreamSubjects []string     `mapstructure:"default_stream_subjects" validate:"required"`
		Consumer              NatsConsumer `mapstructure:"consumer"`
	}

	Sentry struct {
		Dsn              string  `mapstructure:"dsn" validate:"required"`
		TracesSampleRate float64 `mapstructure:"traces_sample_rate" validate:"required"`
	}

	Config struct {
		App      App      `mapstructure:"app"`
		Api      Api      `mapstructure:"api"`
		Database Database `mapstructure:"database"`
		Nats     Nats     `mapstructure:"nats"`
		Sentry   Sentry   `mapstructure:"sentry"`
	}
)

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
	v.SetDefault("app.timezone", "Asia/Riyadh")
	v.SetDefault("app.read_timeout", 60)
	v.SetDefault("app.write_timeout", 60)
	v.SetDefault("app.max_retries", 3)
	v.SetDefault("database.migrations_dir", "migrations")
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.migrate_table", "schema_migrations")
	v.SetDefault("api.public_routes", []string{"/api/health", "/api/health/ready", "/api/docs"})
	v.SetDefault("nats.consumer.max_deliver", 72)
	v.SetDefault("nats.consumer.backoff_durations", []time.Duration{30 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute})
	v.SetDefault("nats.consumer.default_backoff_duration", time.Hour)

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

	// 🚨 This only works if os.Setenv is called before any time.Now() is called
	// issue: https://stackoverflow.com/questions/54363451/setting-timezone-globally-in-golang
	// Make sure to embed timezone data in the binary to be able to load the desired timezone
	// Add _ "time/tzdata" at the top of the file to embed timezone data
	os.Setenv("TZ", config.App.Timezone)

	return &config, nil
}
