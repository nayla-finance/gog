package logger

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/gofiber/fiber/v2/log"
)

type (
	Logger log.AllLogger

	LoggerProvider interface {
		Logger() Logger
	}

	loggerDependencies interface {
		config.ConfigProvider
	}
)

func NewLogger(d loggerDependencies) Logger {
	switch d.Config().App.LogLevel {
	case "debug":
		log.SetLevel(log.LevelDebug)
	case "trace":
		log.SetLevel(log.LevelTrace)
	case "error":
		log.SetLevel(log.LevelError)
	case "warn":
		log.SetLevel(log.LevelWarn)
	case "fatal":
		log.SetLevel(log.LevelFatal)
	default:
		log.SetLevel(log.LevelInfo)
	}

	return log.DefaultLogger()
}

func NewMockLogger() Logger {
	return log.DefaultLogger()
}
