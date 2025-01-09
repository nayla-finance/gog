package utils

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path with leading slash",
			path:     "/api/v1/users",
			expected: "api/v1/users",
		},
		{
			name:     "path with trailing slash",
			path:     "api/v1/users/",
			expected: "api/v1/users",
		},
		{
			name:     "path with leading and trailing slash",
			path:     "/api/v1/users/",
			expected: "api/v1/users",
		},
		{
			name:     "path without leading slash",
			path:     "api/v1/users",
			expected: "api/v1/users",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "single slash",
			path:     "/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.path)
			if got != tt.expected {
				t.Errorf("NormalizePath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
