package validation

import (
	"net/http"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestResolveRequestPath(t *testing.T) {

	manager := oas.NewOASManager(nil)

	// Create test API spec
	content := []byte(`{
        "openapi": "3.0.0",
        "info": {
            "title": "Test API",
            "version": "1.0.0"
        },
        "paths": {
            "/pets": {
                "get": {}
            },
            "/pets/{petId}": {
                "get": {},
                "post": {},
                "delete": {}
            }
        }
    }`)

	err := manager.LoadAPI("test", content)
	assert.NoError(t, err)

	validator := NewValidator(manager)
	err = validator.SetCurrentAPI("test")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "exact match",
			path:     "/pets",
			expected: "/pets",
			wantErr:  false,
		},
		{
			name:     "template match",
			path:     "/pets/123",
			expected: "/pets/{petId}",
			wantErr:  false,
		},
		{
			name:     "unknown path",
			path:     "/unknown",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			route, err := validator.ResolveRequestPath(req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, route)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, route)
			}
		})
	}
}

func TestValidateRequestPath(t *testing.T) {

	manager := oas.NewOASManager(nil)

	// Load test API spec
	content := []byte(`{
        "openapi": "3.0.0",
        "info": {
            "title": "Test API",
            "version": "1.0.0"
        },
        "paths": {
            "/pets": {
                "get": {}
            },
            "/pets/{petId}": {
                "get": {},
                "post": {},
                "delete": {}
            }
        }
    }`)

	err := manager.LoadAPI("test", content)
	assert.NoError(t, err)

	validator := NewValidator(manager)
	err = validator.SetCurrentAPI("test")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		route    string
		expected bool
		errMsg   string
	}{
		{
			name:     "exact match",
			path:     "/pets",
			route:    "/pets",
			expected: true,
			errMsg:   "",
		},
		{
			name:     "non-existent route",
			path:     "/pets",
			route:    "/prouts",
			expected: false,
			errMsg:   "path '/prouts' doesn't exist in oas schema",
		},
		{
			name:     "path collision",
			path:     "/pets",
			route:    "/pets/{petId}",
			expected: false,
			errMsg:   "request path '/pets' is not matching with oas schema path '/pets/{petId}' as there is another path with same name",
		},
		{
			name:     "valid template match",
			path:     "/pets/123",
			route:    "/pets/{petId}",
			expected: true,
			errMsg:   "",
		},
		{
			name:     "invalid path for route",
			path:     "/pets/123",
			route:    "/pets",
			expected: false,
			errMsg:   "request path '/pets/123' is not matching with oas schema path '/pets'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			ok, err := validator.ValidateRequestPath(req, tt.route)

			assert.Equal(t, tt.expected, ok)
			if !tt.expected {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateRequestMethod(t *testing.T) {
	// Setup manager
	config := &oas.CacheConfig{
		MaxAPIs:     10,
		MinPathHits: 5,
	}

	manager := oas.NewOASManager(config)

	// Load test API spec
	content := []byte(`{
        "openapi": "3.0.0",
        "info": {
            "title": "Test API",
            "version": "1.0.0"
        },
        "paths": {
            "/pets": {
                "get": {}
            },
            "/pets/{petId}": {
                "get": {},
                "post": {},
                "delete": {}
            }
        }
    }`)

	err := manager.LoadAPI("test", content)
	assert.NoError(t, err)

	validator := NewValidator(manager)
	err = validator.SetCurrentAPI("test")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		route    string
		method   string
		expected bool
		errMsg   string
	}{
		{
			name:     "valid GET method",
			path:     "/pets",
			route:    "/pets",
			method:   http.MethodGet,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "invalid POST method",
			path:     "/pets",
			route:    "/pets",
			method:   http.MethodPost,
			expected: false,
			errMsg:   "method 'POST' not allowed for path '/pets'",
		},
		{
			name:     "valid GET with path param",
			path:     "/pets/123",
			route:    "/pets/{petId}",
			method:   http.MethodGet,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "valid POST with path param",
			path:     "/pets/123",
			route:    "/pets/{petId}",
			method:   http.MethodPost,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "valid DELETE with path param",
			path:     "/pets/123",
			route:    "/pets/{petId}",
			method:   http.MethodDelete,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "invalid PATCH with path param",
			path:     "/pets/123",
			route:    "/pets/{petId}",
			method:   http.MethodPatch,
			expected: false,
			errMsg:   "method 'PATCH' not allowed for path '/pets/{petId}'",
		},
		{
			name:     "non-existent path",
			path:     "/unknown",
			route:    "/unknown",
			method:   http.MethodGet,
			expected: false,
			errMsg:   "path '/unknown' doesn't exist in oas schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			ok, err := validator.ValidateRequestMethod(req, tt.route)

			assert.Equal(t, tt.expected, ok)
			if !tt.expected {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}
