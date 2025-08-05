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

const (
	OneMB = 1024 * 1024
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

	Dependency struct {
		ReadinessCheck bool `mapstructure:"readiness_check"`
		LivenessCheck  bool `mapstructure:"liveness_check"`
	}

	Dependencies struct {
		Nats       Dependency `mapstructure:"nats"`
		Database   Dependency `mapstructure:"database"`
		TestClient Dependency `mapstructure:"test_client"`
	}

	HealthCheck struct {
		VerboseLog          bool `mapstructure:"verbose_log"`
		InitialChecksToSkip int  `mapstructure:"initial_checks_to_skip"`
	}

	Health struct {
		Liveness     HealthCheck  `mapstructure:"liveness"`
		Readiness    HealthCheck  `mapstructure:"readiness"`
		Dependencies Dependencies `mapstructure:"dependencies"`
	}

	ClientHttpTransport struct {
		DialTimeout         time.Duration `mapstructure:"dial_timeout" validate:"required,gte=1s"`
		DialKeepAlive       time.Duration `mapstructure:"dial_keep_alive" validate:"required,gte=1s"`
		MaxIdleConns        int           `mapstructure:"max_idle_conns" validate:"gt=0"`
		IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout" validate:"required,gte=1s"`
		TLSHandshakeTimeout time.Duration `mapstructure:"tls_handshake_timeout" validate:"required,gte=1s"`
		DisableKeepAlive    bool          `mapstructure:"disable_keep_alive"`
	}

	ClientRetry struct {
		MaxRetries int           `mapstructure:"max_retries" validate:"required,min=1"`
		Delay      time.Duration `mapstructure:"delay" validate:"required,gte=100ms"`
	}

	Client struct {
		MaxRegularBodySizeMB int                 `mapstructure:"max_regular_body_size_mb" validate:"required,min=1"`
		Timeout              time.Duration       `mapstructure:"timeout" validate:"required,gt=1s"`
		HttpTransport        ClientHttpTransport `mapstructure:"http_transport"`
		Retry                ClientRetry         `mapstructure:"retry"`
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
		Timezone      string `mapstructure:"timezone"`
	}

	Api struct {
		Key          string   `mapstructure:"key" validate:"required"`
		PublicRoutes []string `mapstructure:"public_routes"`
	}

	NatsMonitoring struct {
		Enabled                  bool            `mapstructure:"enabled"`
		Interval                 time.Duration   `mapstructure:"interval" validate:"required_if=Enabled true"`
		ExcludedConsumers        map[string]bool `mapstructure:"excluded_consumers"`
		PendingMessagesThreshold int             `mapstructure:"pending_messages_threshold" validate:"required_if=Enabled true,omitempty,gt=0"`
	}

	NatsConsumer struct {
		MaxDeliver             int             `mapstructure:"max_deliver"`
		BackoffDurations       []time.Duration `mapstructure:"backoff_durations"`
		DefaultBackoffDuration time.Duration   `mapstructure:"default_backoff_duration"`
	}

	Nats struct {
		Servers               string         `mapstructure:"servers" validate:"required"`
		ClientName            string         `mapstructure:"client_name" validate:"required"`
		CredsPath             string         `mapstructure:"creds_path" validate:"required"`
		DefaultStreamName     string         `mapstructure:"default_stream_name" validate:"required"`
		DefaultStreamSubjects []string       `mapstructure:"default_stream_subjects" validate:"required"`
		Consumer              NatsConsumer   `mapstructure:"consumer"`
		Monitoring            NatsMonitoring `mapstructure:"monitoring"`
	}

	Sentry struct {
		Dsn              string  `mapstructure:"dsn" validate:"required"`
		TracesSampleRate float64 `mapstructure:"traces_sample_rate" validate:"required"`
	}

	OpenTelemetry struct {
		Enabled        bool     `mapstructure:"enabled"`
		ExcludedRoutes []string `mapstructure:"excluded_routes"`
	}

	Service struct {
		BaseUrl string `mapstructure:"base_url" validate:"required"`
		ApiKey  string `mapstructure:"api_key" validate:"required"`
	}

	Config struct {
		App           App           `mapstructure:"app"`
		Health        Health        `mapstructure:"health"`
		Api           Api           `mapstructure:"api"`
		Database      Database      `mapstructure:"database"`
		Nats          Nats          `mapstructure:"nats"`
		Sentry        Sentry        `mapstructure:"sentry"`
		OpenTelemetry OpenTelemetry `mapstructure:"open_telemetry"`
		Clients       Client        `mapstructure:"clients"`
		TestClient    Service       `mapstructure:"test_client"`
	}
)

func (c *Config) GetClientMaxRegularBodySize() int64 {
	return int64(c.Clients.MaxRegularBodySizeMB) * OneMB
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
	v.SetDefault("app.timezone", "Asia/Riyadh")
	v.SetDefault("app.read_timeout", 60)
	v.SetDefault("app.write_timeout", 60)
	v.SetDefault("app.max_retries", 3)
	v.SetDefault("database.migrations_dir", "migrations")
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.migrate_table", "schema_migrations")
	v.SetDefault("api.public_routes", []string{"/api/healthz/alive", "/api/healthz/ready", "/api/docs"})
	v.SetDefault("nats.consumer.max_deliver", 72)
	v.SetDefault("nats.consumer.backoff_durations", []time.Duration{30 * time.Second, time.Minute, 5 * time.Minute, 15 * time.Minute})
	v.SetDefault("nats.consumer.default_backoff_duration", time.Hour)
	v.SetDefault("open_telemetry.enabled", false)
	v.SetDefault("open_telemetry.excluded_routes", []string{"/api/healthz/alive", "/api/healthz/ready", "/api/docs", "/metrics"})
	v.SetDefault("nats.monitoring.enabled", false)
	v.SetDefault("nats.monitoring.interval", 5*time.Minute)
	v.SetDefault("nats.monitoring.pending_messages_threshold", 2)
	v.SetDefault("clients.max_regular_body_size_mb", 5)
	v.SetDefault("clients.timeout", 30*time.Second)

	// Health check defaults
	v.SetDefault("health.liveness.verbose_log", false)
	v.SetDefault("health.liveness.initial_checks_to_skip", 0)
	v.SetDefault("health.readiness.verbose_log", true)
	v.SetDefault("health.readiness.initial_checks_to_skip", 0)
	v.SetDefault("health.dependencies", Dependencies{
		Nats:       Dependency{ReadinessCheck: true, LivenessCheck: true},
		Database:   Dependency{ReadinessCheck: true, LivenessCheck: true},
		TestClient: Dependency{ReadinessCheck: false, LivenessCheck: true},
	})

	// HTTP transport defaults
	v.SetDefault("clients.http_transport.dial_timeout", 30*time.Second)
	v.SetDefault("clients.http_transport.dial_keep_alive", 30*time.Second)
	v.SetDefault("clients.http_transport.max_idle_conns", 100)
	v.SetDefault("clients.http_transport.idle_conn_timeout", 20*time.Second)
	v.SetDefault("clients.http_transport.tls_handshake_timeout", 10*time.Second)
	v.SetDefault("clients.http_transport.disable_keep_alive", false)

	// Retry defaults
	v.SetDefault("clients.retry.max_retries", 3)
	v.SetDefault("clients.retry.delay", 500*time.Millisecond)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.Database.Timezone == "" {
		config.Database.Timezone = config.App.Timezone
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
