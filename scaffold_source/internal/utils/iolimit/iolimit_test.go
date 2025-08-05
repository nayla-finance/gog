package iolimit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// errorReader is a helper type that always returns an error when reading
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func TestReadAll_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxSize  int64
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			maxSize:  100,
			expected: "",
		},
		{
			name:     "small input within limit",
			input:    "hello world",
			maxSize:  100,
			expected: "hello world",
		},
		{
			name:     "input one byte less than limit",
			input:    "test",
			maxSize:  5,
			expected: "test",
		},
		{
			name:     "single character within limit",
			input:    "a",
			maxSize:  10,
			expected: "a",
		},
		{
			name:     "unicode content within limit",
			input:    "hello 世界 🌍",
			maxSize:  50,
			expected: "hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := ReadAll(reader, tt.maxSize)

			if err != nil {
				t.Errorf("ReadAll() unexpected error = %v", err)
				return
			}

			if string(result) != tt.expected {
				t.Errorf("ReadAll() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestReadAll_SizeExceeded(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxSize int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "input equals limit",
			input:   "hello",
			maxSize: 5,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "input exceeds limit by one byte",
			input:   "hello world",
			maxSize: 10,
			wantErr: true,
			errMsg:  "reader exceeded size limit of 10 bytes",
		},
		{
			name:    "large input with small limit",
			input:   strings.Repeat("a", 1000),
			maxSize: 100,
			wantErr: true,
			errMsg:  "reader exceeded size limit of 100 bytes",
		},
		{
			name:    "zero max size with non-empty input",
			input:   "a",
			maxSize: 0,
			wantErr: true,
			errMsg:  "reader exceeded size limit of 0 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := ReadAll(reader, tt.maxSize)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadAll() expected error but got nil")
					return
				}

				if err.Error() != tt.errMsg {
					t.Errorf("ReadAll() error = %q, want %q", err.Error(), tt.errMsg)
				}

				if result != nil {
					t.Errorf("ReadAll() expected nil result when error occurs, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("ReadAll() unexpected error = %v", err)
					return
				}

				if string(result) != tt.input {
					t.Errorf("ReadAll() = %q, want %q", string(result), tt.input)
				}
			}
		})
	}
}

func TestReadAll_ReaderError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		maxSize int64
	}{
		{
			name:    "generic read error",
			err:     errors.New("read failed"),
			maxSize: 100,
		},
		{
			name:    "custom error with large maxSize",
			err:     fmt.Errorf("network timeout"),
			maxSize: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &errorReader{err: tt.err}
			result, err := ReadAll(reader, tt.maxSize)

			if err == nil {
				t.Errorf("ReadAll() expected error but got nil")
				return
			}

			if !errors.Is(err, tt.err) && err.Error() != tt.err.Error() {
				t.Errorf("ReadAll() error = %v, want %v", err, tt.err)
			}

			if result != nil {
				t.Errorf("ReadAll() expected nil result when error occurs, got %v", result)
			}
		})
	}
}

func TestReadAll_EdgeCases(t *testing.T) {
	t.Run("zero maxSize with empty input", func(t *testing.T) {
		reader := strings.NewReader("")
		result, err := ReadAll(reader, 0)

		if err != nil {
			t.Errorf("ReadAll() unexpected error = %v", err)
			return
		}

		if len(result) != 0 {
			t.Errorf("ReadAll() expected empty result, got %v", result)
		}
	})

	t.Run("very large maxSize", func(t *testing.T) {
		input := "test data"
		reader := strings.NewReader(input)
		result, err := ReadAll(reader, 1<<30) // 1GB limit

		if err != nil {
			t.Errorf("ReadAll() unexpected error = %v", err)
			return
		}

		if string(result) != input {
			t.Errorf("ReadAll() = %q, want %q", string(result), input)
		}
	})
}

func TestReadAll_BytesReader(t *testing.T) {
	t.Run("bytes.Reader within limit", func(t *testing.T) {
		data := []byte("test data with bytes reader")
		reader := bytes.NewReader(data)
		result, err := ReadAll(reader, 100)

		if err != nil {
			t.Errorf("ReadAll() unexpected error = %v", err)
			return
		}

		if !bytes.Equal(result, data) {
			t.Errorf("ReadAll() = %v, want %v", result, data)
		}
	})

	t.Run("bytes.Reader exceeds limit", func(t *testing.T) {
		data := []byte("this data exceeds limit")
		reader := bytes.NewReader(data)
		result, err := ReadAll(reader, 10)

		if err == nil {
			t.Errorf("ReadAll() expected error but got nil")
			return
		}

		expectedErr := "reader exceeded size limit of 10 bytes"
		if err.Error() != expectedErr {
			t.Errorf("ReadAll() error = %q, want %q", err.Error(), expectedErr)
		}

		if result != nil {
			t.Errorf("ReadAll() expected nil result when error occurs, got %v", result)
		}
	})
}
