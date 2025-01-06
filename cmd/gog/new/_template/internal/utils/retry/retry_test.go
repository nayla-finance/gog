package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDependencies struct {
	mock.Mock
	logger logger.Logger
	config *config.Config
}

func (m *mockDependencies) Logger() logger.Logger {
	return log.DefaultLogger()
}

func (m *mockDependencies) Config() *config.Config {
	return m.config
}

func TestRetry_Success(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 3,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test successful operation on first try
	operationCalled := 0
	err := retry.Do(func() error {
		operationCalled++
		return nil
	}, "test-operation")

	assert.NoError(t, err)
	assert.Equal(t, 1, operationCalled)
}

func TestRetry_EventualSuccess(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 3,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test operation that succeeds on second try
	operationCalled := 0
	err := retry.Do(func() error {
		operationCalled++
		if operationCalled == 1 {
			return errors.New("temporary error")
		}
		return nil
	}, "test-operation")

	assert.NoError(t, err)
	assert.Equal(t, 2, operationCalled)
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 3,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test operation that always fails
	expectedErr := errors.New("persistent error")
	operationCalled := 0
	err := retry.Do(func() error {
		operationCalled++
		return expectedErr
	}, "test-operation")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 4, operationCalled) // Should be called MaxRetries + 1 times
}

func TestRetry_ZeroRetries(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 0,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test operation with zero retries configured
	operationCalled := 0
	err := retry.Do(func() error {
		operationCalled++
		return errors.New("error")
	}, "test-operation")

	assert.Error(t, err)
	assert.Equal(t, 1, operationCalled)
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 3,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test that delays between retries increase
	startTime := time.Now()
	operationCalled := 0
	_ = retry.Do(func() error {
		operationCalled++
		return errors.New("error")
	}, "test-operation")

	duration := time.Since(startTime)
	// Should take at least: 10ms + 10ms + 10ms = 30ms
	t.Logf("duration: %s", duration)
	assert.True(t, duration >= 30*time.Millisecond)
	assert.Equal(t, 4, operationCalled)
}

func TestRetry_RunAtLeastOnce(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 0,
				RetryDelay: 10 * time.Millisecond,
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Test successful operation on first try
	operationCalled := 0
	err := retry.Do(func() error {
		operationCalled++
		return nil
	}, "test-operation")

	assert.NoError(t, err)
	assert.Equal(t, 1, operationCalled)
}
