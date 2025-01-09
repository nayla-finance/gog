package middleware

import (
	"testing"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
)

type mockDependencies struct {
	c config.Config
}

func (m *mockDependencies) Config() *config.Config {
	return &m.c
}

func (m *mockDependencies) Logger() logger.Logger {
	return nil
}

func (m *mockDependencies) NewError(err errors.ErrorCode, message string) *errors.AppError {
	return nil
}

func TestIsPublicRoute(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		publicRoutes []string
		expected     bool
	}{
		{
			name:         "exact match should return true",
			path:         "/api/health",
			publicRoutes: []string{"/api/health", "/api/docs"},
			expected:     true,
		},
		{
			name:         "non-matching path should return false",
			path:         "/api/private",
			publicRoutes: []string{"/api/health", "/api/docs"},
			expected:     false,
		},
		{
			name:         "empty public routes should return false",
			path:         "/api/health",
			publicRoutes: []string{},
			expected:     false,
		},
		{
			name:         "trailing slashes should not affect matching",
			path:         "/api/health/",
			publicRoutes: []string{"/api/health"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		authMiddleware := AuthMiddleware{
			d: &mockDependencies{
				c: config.Config{
					Api: config.Api{
						PublicRoutes: tt.publicRoutes,
					},
				},
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			result := authMiddleware.isPublicRoute(tt.path)
			if result != tt.expected {
				t.Errorf("TestIsPublicRoute(%s, %v) = %v; want %v",
					tt.path, tt.publicRoutes, result, tt.expected)
			}
		})
	}
}
