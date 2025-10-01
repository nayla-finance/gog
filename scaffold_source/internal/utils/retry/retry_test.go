package retry

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type mockDependencies struct {
	mock.Mock
	config *config.Config
	logger logger.Logger
}

func (m *mockDependencies) Config() *config.Config {
	return m.config
}

func (m *mockDependencies) Logger() logger.Logger {
	if m.logger == nil {
		m.logger = logger.NewMockLogger()
	}
	return m.logger
}

// Mock HTTP Transport for testing HTTP retries
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
		// Return the last response if we exceed planned responses
		lastIdx := len(m.responses) - 1
		return m.responses[lastIdx].response, m.responses[lastIdx].err
	}

	resp := m.responses[m.callCount]
	return resp.response, resp.err
}

func (m *mockRoundTripper) GetCallCount() int {
	return m.callCount
}

func createTestRetry(maxRetries int, retryDelay time.Duration, httpMaxRetries int, httpDelay time.Duration) *retry {
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: maxRetries,
				RetryDelay: retryDelay,
			},
			Clients: config.Client{
				Retry: config.ClientRetry{
					MaxRetries: httpMaxRetries,
					Delay:      httpDelay,
				},
			},
		},
	}

	return NewRetry(deps)
}

func TestRetry_Do_Success(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	operationCalled := 0
	err := r.Do(func() error {
		operationCalled++
		return nil
	}, "test-operation")

	assert.NoError(t, err)
	assert.Equal(t, 1, operationCalled)
}

func TestRetry_Do_EventualSuccess(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	operationCalled := 0
	err := r.Do(func() error {
		operationCalled++
		if operationCalled == 1 {
			return errors.New("temporary error")
		}
		return nil
	}, "test-operation")

	assert.NoError(t, err)
	assert.Equal(t, 2, operationCalled)
}

func TestRetry_Do_MaxRetriesExceeded(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	expectedErr := errors.New("persistent error")
	operationCalled := 0
	err := r.Do(func() error {
		operationCalled++
		return expectedErr
	}, "test-operation")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 4, operationCalled) // Should be called MaxRetries + 1 times
}

func TestRetry_Do_ZeroRetries(t *testing.T) {
	r := createTestRetry(0, 10*time.Millisecond, 3, 10*time.Millisecond)

	operationCalled := 0
	err := r.Do(func() error {
		operationCalled++
		return errors.New("error")
	}, "test-operation")

	assert.Error(t, err)
	assert.Equal(t, 1, operationCalled)
}

func TestRetry_Do_DelayTiming(t *testing.T) {
	r := createTestRetry(2, 50*time.Millisecond, 3, 10*time.Millisecond)

	startTime := time.Now()
	operationCalled := 0
	_ = r.Do(func() error {
		operationCalled++
		return errors.New("error")
	}, "test-operation")

	duration := time.Since(startTime)

	// Should take at least: 50ms + 50ms = 100ms for 2 retries
	assert.True(t, duration >= 100*time.Millisecond)
	assert.Equal(t, 3, operationCalled) // Initial + 2 retries
}

func TestIsRetryableError(t *testing.T) {
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

func TestIsRetryableError_WithUrlError(t *testing.T) {
	innerErr := errors.New("connection refused")
	urlErr := &url.Error{
		Op:  "Get",
		URL: "https://example.com",
		Err: innerErr,
	}

	result := isRetryableError(urlErr)
	assert.True(t, result, "Should recognize retryable error even when wrapped in url.Error")
}

func TestIsRetryableError_NetworkErrors(t *testing.T) {
	// Test *net.OpError
	opErr := &net.OpError{
		Err: &os.SyscallError{
			Syscall: "connect",
			Err:     syscall.ECONNREFUSED,
		},
	}
	assert.True(t, isRetryableError(opErr))

	// Test *net.DNSError (temporary)
	dnsErr := &net.DNSError{
		Err:         "no such host",
		Name:        "example.com",
		IsTemporary: true,
	}
	assert.True(t, isRetryableError(dnsErr))

	// Test *net.DNSError (not temporary)
	dnsErrNotTemp := &net.DNSError{
		Err:         "no such host",
		Name:        "example.com",
		IsTemporary: false,
	}
	assert.False(t, isRetryableError(dnsErrNotTemp))
}

func TestRetry_DoHttp_Success(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, transport.GetCallCount())
}

func TestRetry_DoHttp_EventualSuccess(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.GetCallCount())
}

func TestRetry_DoHttp_MaxRetriesExceeded(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP GET https://example.com failed after 3 retries")
	assert.Equal(t, 3, transport.GetCallCount())
}

func TestRetry_DoHttp_NonRetryableError(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	expectedErr := errors.New("invalid URL format")
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: expectedErr},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
	assert.Nil(t, resp)
	assert.Equal(t, 1, transport.GetCallCount()) // Should return immediately
}

func TestRetry_DoHttp_RetryableNetworkError(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

	retryableErr := errors.New("connection refused")
	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: retryableErr},
			{response: &http.Response{StatusCode: http.StatusOK}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.GetCallCount())
}

func TestRetry_DoHttp_Non500StatusCodes(t *testing.T) {
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
			r := createTestRetry(3, 10*time.Millisecond, 3, 10*time.Millisecond)

			transport := &mockRoundTripper{
				responses: []mockResponse{
					{response: &http.Response{StatusCode: tc.statusCode}, err: nil},
				},
			}

			client := &http.Client{Transport: transport}
			req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

			resp, err := r.DoHttp("test-operation", client, req)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tc.statusCode, resp.StatusCode)
			assert.Equal(t, 1, transport.GetCallCount()) // Should return immediately
		})
	}
}

func TestRetry_DoHttp_ExponentialBackoff(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 50*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	startTime := time.Now()
	_, err := r.DoHttp("test-operation", client, req)
	duration := time.Since(startTime)

	// Should take at least: 50ms*(1) + 50ms*(2) + 50ms*(3) = 300ms for 3 retries
	expectedMinDuration := 300 * time.Millisecond
	assert.True(t, duration >= expectedMinDuration, "Duration: %v, expected min: %v", duration, expectedMinDuration)
	assert.Error(t, err)
	assert.Equal(t, 3, transport.GetCallCount())
}

func TestRetry_DoHttp_ZeroRetries(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 0, 10*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP GET https://example.com failed after 0 retries")
	assert.Equal(t, 0, transport.GetCallCount()) // Should not be called at all with 0 retries
}

func TestRetry_DoHttp_MixedErrorsEventualSuccess(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 5, 10*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: nil, err: errors.New("connection refused")},
			{response: &http.Response{StatusCode: http.StatusInternalServerError}, err: nil},
			{response: &http.Response{StatusCode: http.StatusCreated}, err: nil},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	resp, err := r.DoHttp("test-operation", client, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, 4, transport.GetCallCount()) // Should succeed on 4th try
}

func TestNewRetry(t *testing.T) {
	deps := &mockDependencies{
		config: &config.Config{
			App: config.App{
				MaxRetries: 5,
				RetryDelay: 100 * time.Millisecond,
			},
		},
	}

	r := NewRetry(deps)

	assert.NotNil(t, r)
	assert.Equal(t, deps, r.d)
}

// Benchmark tests
func BenchmarkRetry_Do_Success(b *testing.B) {
	r := createTestRetry(3, 1*time.Millisecond, 3, 1*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(func() error {
			return nil
		}, "benchmark-operation")
	}
}

func BenchmarkRetry_Do_WithRetries(b *testing.B) {
	r := createTestRetry(3, 1*time.Millisecond, 3, 1*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callCount := 0
		_ = r.Do(func() error {
			callCount++
			if callCount < 2 {
				return errors.New("temporary error")
			}
			return nil
		}, "benchmark-operation")
	}
}

func TestRetry_DoHttp_NetworkErrorBackoffTiming(t *testing.T) {
	r := createTestRetry(3, 10*time.Millisecond, 3, 50*time.Millisecond)

	transport := &mockRoundTripper{
		responses: []mockResponse{
			{response: nil, err: errors.New("connection refused")},
		},
	}

	client := &http.Client{Transport: transport}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)

	startTime := time.Now()
	_, err := r.DoHttp("test-operation", client, req)
	duration := time.Since(startTime)

	// Should take at least: 50ms*(1) + 50ms*(2) + 50ms*(3) = 300ms for 3 retries
	expectedMinDuration := 300 * time.Millisecond
	assert.True(t, duration >= expectedMinDuration, "Duration: %v, expected min: %v", duration, expectedMinDuration)
	assert.Error(t, err)
	assert.Equal(t, 3, transport.GetCallCount())
}
