package config

import (
	"fmt"
	"os"
	"strings"

	// Embed timezone data
	_ "time/tzdata"

	"github.com/nayla-finance/go-nayla/config"
	"github.com/nayla-finance/go-nayla/validator"
	"github.com/spf13/viper"
)

type ConfigProvider interface {
	Config() *Config
}

type (
	Config struct {
		App           config.App           `mapstructure:"app"`
		Health        config.Health        `mapstructure:"health"`
		Api           config.API           `mapstructure:"api"`
		Database      config.Database      `mapstructure:"database"`
		Nats          config.Nats          `mapstructure:"nats"`
		Sentry        config.Sentry        `mapstructure:"sentry"`
		OpenTelemetry config.OpenTelemetry `mapstructure:"open_telemetry"`
		KYC           config.Service       `mapstructure:"kyc"`
		LOS           config.Service       `mapstructure:"los"`
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

	config.LoadDefaultConfig(v)
	v.SetDefault("health.dependencies", config.Dependencies{
		"nats":     config.Dependency{ReadinessCheck: true, LivenessCheck: true},
		"database": config.Dependency{ReadinessCheck: true, LivenessCheck: true},
		"kyc":      config.Dependency{ReadinessCheck: false, LivenessCheck: true},
		"los":      config.Dependency{ReadinessCheck: false, LivenessCheck: true},
	})

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
