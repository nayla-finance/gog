package logger

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/project-name/internal/config"
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
	switch d.Config().LogLevel {
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
