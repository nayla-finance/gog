package retry

import (
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
)

type (
	RetryProvider interface {
		Retry() *Retry
	}

	retryDependencies interface {
		logger.LoggerProvider
		config.ConfigProvider
	}

	Retry struct {
		d retryDependencies
	}
)

func NewRetry(d retryDependencies) *Retry {
	return &Retry{d: d}
}

// Do executes the operation with retries and returns both the result and error
func (r *Retry) Do(operation func() error, name string) error {
	var lastErr error
	attempts := 1

	for {
		if err := operation(); err != nil {
			lastErr = err

			if attempts >= r.d.Config().App.MaxRetries {
				r.d.Logger().Error("Operation ", name, " failed after max retries", " error ", lastErr)
				return lastErr
			}

			attempts++
			r.d.Logger().Warn("Operation ", name, " failed, retrying... ",
				" attempt ", attempts+1,
				" maxRetries ", r.d.Config().App.MaxRetries,
				" error ", err)
			time.Sleep(r.d.Config().App.RetryDelay)
			continue
		}
		return nil
	}
}
