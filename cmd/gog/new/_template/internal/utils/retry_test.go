package utils

import (
	"errors"
	"net/http"
	"net/url"
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

// mockRoundTripper allows us to mock HTTP responses
type mockRoundTripper struct {
	responses []mockResponse
	callCount int
}

type mockResponse struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() { m.callCount++ }()

	if m.callCount >= len(m.responses) {
		// Return the last response if we exceed the planned responses
		lastIdx := len(m.responses) - 1
		return m.responses[lastIdx].response, m.responses[lastIdx].err
	}

	resp := m.responses[m.callCount]
	return resp.response, resp.err
}

func (m *mockRoundTripper) GetCallCount() int {
	return m.callCount
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

func TestRetry_DoHttp_Success(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with successful response
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, transport.GetCallCount())
}

func TestRetry_DoHttp_EventualSuccess(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with 500 then success
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.GetCallCount())
}

func TestRetry_DoHttp_MaxRetriesExceeded(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client that always returns 500
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP GET https://example.com failed after 3 retries")
	assert.Equal(t, 3, transport.GetCallCount()) // Should be called MaxRetries times
}

func TestRetry_DoHttp_ErrorFromDoHttp(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client that returns an error
	expectedErr := errors.New("network error")
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: expectedErr},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error()) // HTTP client wraps errors in url.Error
	assert.Nil(t, resp)
	assert.Equal(t, 1, transport.GetCallCount()) // Should return immediately on error
}

func TestRetry_DoHttp_Non500ErrorCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"BadGateway", http.StatusBadGateway},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deps := &mockDependencies{
				config: &config.Config{
					Clients: config.Client{
						Retry: config.ClientRetry{
							MaxRetries: 3,
							Delay:      10 * time.Millisecond,
						},
					},
				},
			}
			mockLogger := logger.NewLogger(deps)
			deps.logger = mockLogger
			retry := NewRetry(deps)

			// Mock HTTP client with non-500 status code
			transport := &mockRoundTripper{
				responses: []mockResponse{
					{response: &http.Response{StatusCode: tc.statusCode}, err: nil},
				},
			}
			client := &http.Client{Transport: transport}
			req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

			resp, err := retry.DoHttp("test-operation", client, req)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.statusCode, resp.StatusCode)
			assert.Equal(t, 1, transport.GetCallCount()) // Should return immediately
		})
	}
}

func TestRetry_DoHttp_ExponentialBackoff(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      50 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client that always returns 500
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	startTime := time.Now()
	_, err := retry.DoHttp("test-operation", client, req)

	duration := time.Since(startTime)
	// Should take at least: 50ms*(1) + 50ms*(2) + 50ms*(3) = 300ms for 3 retries
	expectedMinDuration := 300 * time.Millisecond
	t.Logf("duration: %s, expected min: %s", duration, expectedMinDuration)
	assert.True(t, duration >= expectedMinDuration)
	assert.Error(t, err)
	assert.Equal(t, 3, transport.GetCallCount())
}

func TestRetry_DoHttp_ZeroRetries(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 0,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP GET https://example.com failed after 0 retries")
	assert.Equal(t, 0, transport.GetCallCount()) // Should not be called at all with 0 retries
}

func TestRetry_DoHttp_ZeroRetriesSuccess(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 0,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with successful response
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.Error(t, err) // Should still error because loop doesn't run at all
	assert.Nil(t, resp)
	assert.Equal(t, 0, transport.GetCallCount()) // Should not be called at all with 0 retries
}

func TestRetry_DoHttp_MixedResponsesEventualSuccess(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 5,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with various 500 responses followed by success
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusCreated}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, 4, transport.GetCallCount()) // Should succeed on 4th try
}

func TestRetry_DoHttp_ErrorAfterSuccessfulResponse(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with 500 response followed by error
	expectedErr := errors.New("connection reset")
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: nil, err: expectedErr},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error()) // HTTP client wraps errors in url.Error
	assert.Nil(t, resp)
	assert.Equal(t, 2, transport.GetCallCount()) // Should be called twice
}

func TestRetry_IsRetryableError(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"connection reset by peer", errors.New("connection reset by peer"), true},
		{"broken pipe", errors.New("broken pipe"), true},
		{"no such host", errors.New("no such host"), true},
		{"dial tcp error", errors.New("dial tcp: connection failed"), true},
		{"EOF error", errors.New("EOF"), true},
		{"i/o timeout", errors.New("i/o timeout"), false}, // Changed: timeouts should not retry
		{"network unreachable", errors.New("network unreachable"), true},
		{"host unreachable", errors.New("host unreachable"), true},
		{"invalid URL", errors.New("invalid URL"), false},
		{"context canceled", errors.New("context canceled"), false},
		{"TLS handshake failure", errors.New("tls: handshake failure"), false},
		{"invalid request", errors.New("invalid request format"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isRetryableError(tc.err)
			assert.Equal(t, tc.shouldRetry, result, "Error: %v", tc.err)
		})
	}
}

func TestRetry_IsRetryableError_WithUrlError(t *testing.T) {
	// Test that url.Error wrapping is handled correctly
	innerErr := errors.New("connection refused")
	urlErr := &url.Error{
		Op:  "Get",
		URL: "https://example.com",
		Err: innerErr,
	}

	result := isRetryableError(urlErr)
	assert.True(t, result, "Should recognize retryable error even when wrapped in url.Error")
}

func TestRetry_DoHttp_RetryableNetworkErrors(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		shouldRetry   bool
		expectedCalls int
	}{
		{
			name:          "connection refused - should retry",
			err:           errors.New("connection refused"),
			shouldRetry:   true,
			expectedCalls: 3, // All retries exhausted
		},
		{
			name:          "connection reset - should retry",
			err:           errors.New("connection reset by peer"),
			shouldRetry:   true,
			expectedCalls: 3, // All retries exhausted
		},
		{
			name:          "timeout error - should not retry", // Changed: timeouts don't retry
			err:           errors.New("i/o timeout"),
			shouldRetry:   false,
			expectedCalls: 1, // Return immediately
		},
		{
			name:          "invalid URL - should not retry",
			err:           errors.New("invalid URL format"),
			shouldRetry:   false,
			expectedCalls: 1, // Return immediately
		},
		{
			name:          "context canceled - should not retry",
			err:           errors.New("context canceled"),
			shouldRetry:   false,
			expectedCalls: 1, // Return immediately
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deps := &mockDependencies{
				config: &config.Config{
					Clients: config.Client{
						Retry: config.ClientRetry{
							MaxRetries: 3,
							Delay:      10 * time.Millisecond,
						},
					},
				},
			}
			mockLogger := logger.NewLogger(deps)
			deps.logger = mockLogger
			retry := NewRetry(deps)

			// Mock HTTP client that always returns the test error
			transport := &mockRoundTripper{
				responses: []mockResponse{
					{response: nil, err: tc.err},
				},
			}
			client := &http.Client{Transport: transport}
			req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

			resp, err := retry.DoHttp("test-operation", client, req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(t, tc.expectedCalls, transport.GetCallCount())

			if tc.shouldRetry {
				// For retryable errors, we should get the final retry failure message
				assert.Contains(t, err.Error(), "failed after 3 retries")
			} else {
				// For non-retryable errors, we should get the original error
				assert.Contains(t, err.Error(), tc.err.Error())
			}
		})
	}
}

func TestRetry_DoHttp_MixedNetworkErrorAndSuccess(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 5,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with network errors followed by success
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: errors.New("connection refused")},
			{response: nil, err: errors.New("connection reset by peer")},
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, transport.GetCallCount()) // Should succeed on 3rd try
}

func TestRetry_DoHttp_Mixed500AndNetworkErrors(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 5,
					Delay:      10 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client with mix of 500 responses and network errors, then success
	// Note: Changed from "i/o timeout" to "connection refused" since timeouts don't retry
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: nil, err: errors.New("connection refused")},
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: nil, err: errors.New("connection reset by peer")},
			{response: &http.Response{StatusCode: http.StatusCreated}, err: nil},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := retry.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, 5, transport.GetCallCount()) // Should succeed on 5th try
}

func TestRetry_DoHttp_NetworkErrorBackoffTiming(t *testing.T) {
	// Setup
	deps := &mockDependencies{
		config: &config.Config{
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: 3,
					Delay:      50 * time.Millisecond,
				},
			},
		},
	}
	mockLogger := logger.NewLogger(deps)
	deps.logger = mockLogger
	retry := NewRetry(deps)

	// Mock HTTP client that always returns network errors
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: errors.New("connection refused")},
		},
	}
	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	startTime := time.Now()
	_, err := retry.DoHttp("test-operation", client, req)

	duration := time.Since(startTime)
	// Should take at least: 50ms*(1) + 50ms*(2) + 50ms*(3) = 300ms for 3 retries
	expectedMinDuration := 300 * time.Millisecond
	t.Logf("duration: %s, expected min: %s", duration, expectedMinDuration)
	assert.True(t, duration >= expectedMinDuration)
	assert.Error(t, err)
	assert.Equal(t, 3, transport.GetCallCount())
}
