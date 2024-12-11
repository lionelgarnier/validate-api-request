package validation

import (
	"net/http"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestResolveRequestPath(t *testing.T) {

	manager := oas.NewOASManager(nil, oas.FixedSelector("test"))

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

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

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

			oasRequest := oas.NewOASRequest(req)

			pathCache, err := validator.ResolveRequestPath(oasRequest)

			var route string
			if pathCache != nil {
				route = pathCache.Route
			}

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

	manager := oas.NewOASManager(nil, oas.FixedSelector("test"))

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

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
		errMsg   string
	}{
		{
			name:     "exact match",
			path:     "/pets",
			expected: true,
			errMsg:   "",
		},
		{
			name:     "non-existent route",
			path:     "/prouts",
			expected: false,
			errMsg:   "no schema found for path '/prouts'",
		},
		{
			name:     "valid template match",
			path:     "/pets/123",
			expected: true,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)

			oasRequest := oas.NewOASRequest(req)

			ok, err := validator.ValidateRequestPath(oasRequest)

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

	manager := oas.NewOASManager(config, oas.FixedSelector("test"))

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

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

	assert.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		method   string
		expected bool
		errMsg   string
	}{
		{
			name:     "valid GET method",
			path:     "/pets",
			method:   http.MethodGet,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "invalid POST method",
			path:     "/pets",
			method:   http.MethodPost,
			expected: false,
			errMsg:   "method 'POST' not allowed for path '/pets'",
		},
		{
			name:     "valid GET with path param",
			path:     "/pets/123",
			method:   http.MethodGet,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "valid POST with path param",
			path:     "/pets/123",
			method:   http.MethodPost,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "valid DELETE with path param",
			path:     "/pets/123",
			method:   http.MethodDelete,
			expected: true,
			errMsg:   "",
		},
		{
			name:     "invalid PATCH with path param",
			path:     "/pets/123",
			method:   http.MethodPatch,
			expected: false,
			errMsg:   "method 'PATCH' not allowed for path '/pets/{petId}'",
		},
		{
			name:     "non-existent path",
			path:     "/unknown",
			method:   http.MethodGet,
			expected: false,
			errMsg:   "no schema found for path '/unknown'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)

			oasRequest := oas.NewOASRequest(req)

			ok, err := validator.ValidateRequestMethod(oasRequest)

			assert.Equal(t, tt.expected, ok)
			if !tt.expected {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}
