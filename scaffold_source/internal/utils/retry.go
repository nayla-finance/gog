package utils

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/getsentry/sentry-go"
)

type (
	Retry interface {
		Do(operation func() error, name string) error
		DoHttp(name string, c *http.Client, req *http.Request) (*http.Response, error)
	}

	RetryProvider interface {
		Retry() Retry
	}

	retryDependencies interface {
		logger.LoggerProvider
		config.ConfigProvider
	}

	retry struct {
		d retryDependencies
	}
)

func NewRetry(d retryDependencies) *retry {
	return &retry{d: d}
}

// Do executes the operation with retries and returns both the result and error
func (r *retry) Do(operation func() error, name string) error {
	var lastErr error
	attempts := 1

	for {
		if err := operation(); err != nil {
			lastErr = err

			if attempts > r.d.Config().App.MaxRetries {
				r.d.Logger().Error("Operation ", name, " failed after max retries", " error ", lastErr)
				return lastErr
			}

			attempts++
			r.d.Logger().Warn("Operation ", name, " failed, retrying... ",
				" attempt ", attempts,
				" maxRetries ", r.d.Config().App.MaxRetries,
				" error ", err)
			time.Sleep(r.d.Config().App.RetryDelay)
			continue
		}
		return nil
	}
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for url.Error (wraps most HTTP client errors)
	if urlErr, ok := err.(*url.Error); ok {
		err = urlErr.Err
	}

	// Type-based error detection (most reliable)
	switch err := err.(type) {
	case *net.OpError:
		// Check for connection-related network operation errors
		if syscallErr, ok := err.Err.(*os.SyscallError); ok {
			if errno, ok := syscallErr.Err.(syscall.Errno); ok {
				switch errno {
				case syscall.ECONNREFUSED, // Connection refused
					syscall.ECONNRESET,   // Connection reset by peer
					syscall.ECONNABORTED, // Connection aborted
					syscall.ETIMEDOUT,    // Connection timed out
					syscall.EHOSTUNREACH, // Host unreachable
					syscall.ENETUNREACH,  // Network unreachable
					syscall.ENOTCONN,     // Socket not connected
					syscall.EPIPE:        // Broken pipe
					return true
				}
			}
		}
		// Also retry general network operation errors that don't have specific syscall errors
		return true
	case *net.DNSError:
		// Only retry temporary DNS errors
		return err.Temporary()
	case *os.SyscallError:
		// Direct syscall errors
		if errno, ok := err.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ECONNREFUSED, // Connection refused
				syscall.ECONNRESET,   // Connection reset by peer
				syscall.ECONNABORTED, // Connection aborted
				syscall.ETIMEDOUT,    // Connection timed out
				syscall.EHOSTUNREACH, // Host unreachable
				syscall.ENETUNREACH,  // Network unreachable
				syscall.ENOTCONN,     // Socket not connected
				syscall.EPIPE:        // Broken pipe
				return true
			}
		}
	}

	// Fallback: string-based detection for edge cases
	errMsg := strings.ToLower(err.Error())
	retryableMessages := []string{
		"connection refused",
		"connection reset by peer",
		"connection aborted",
		"broken pipe",
		"no such host", // DNS resolution failures
		"dial tcp",     // General dial errors
		"eof",          // Unexpected connection close
		"host unreachable",
		"network unreachable",
		"socket not connected",
	}

	for _, msg := range retryableMessages {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}

	return false
}

// This method will retry only on 500 errors and retryable network errors
func (r *retry) DoHttp(name string, c *http.Client, req *http.Request) (*http.Response, error) {
	for i := 0; i < r.d.Config().Clients.Retry.MaxRetries; i++ {
		response, err := c.Do(req)

		// If there's an error, check if it's retryable
		if err != nil {
			if !isRetryableError(err) {
				// Non-retryable error, return immediately
				return nil, err
			}
			// Retryable network error, continue to retry logic
			// Problem: Connection pool might have stale/broken connections
			// Solution: Clear the pool to force fresh connections
			// Network errors indicate connection-level issues (not server-level issues)
			// Clear idle connections to establish fresh network paths on retry
			c.CloseIdleConnections()
		} else if response.StatusCode != http.StatusInternalServerError {
			// No error and not 500, return the response
			return response, nil
		}

		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}

		// Report the error to sentry
		sentry.CaptureException(fmt.Errorf("❌ HTTP %s %s failed: %w", req.Method, req.URL.String(), err))

		// We either got a 500 response or a retryable network error
		// Apply backoff delay before retrying
		delay := r.d.Config().Clients.Retry.Delay * time.Duration(i+1)
		r.d.Logger().Warnw("⚠️ HTTP request failed, retrying... ", "method", req.Method, "url", req.URL.String(), "delay_ms", delay.Milliseconds(), "status", statusCode, "error", err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("❌ HTTP %s %s failed after %d retries", req.Method, req.URL.String(), r.d.Config().Clients.Retry.MaxRetries)
}
