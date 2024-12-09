package validation

import (
	"net/http"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestValidateParameters(t *testing.T) {

	manager := oas.NewOASManager(nil)

	// Load test API spec
	content := []byte(`{
        "openapi": "3.0.0",
        "info": {
            "title": "Test API",
            "version": "1.0.0"
        },
        "paths": {
            "/pet/{petId}": {
                "get": {
                    "parameters": [
                        {
                            "name": "petId",
                            "in": "path",
                            "required": true,
                            "schema": {
                                "type": "integer",
                                "format": "int64"
                            }
                        }
                    ]
                }
            },
            "/pet/findByStatus": {
                "get": {
                    "parameters": [
                        {
                            "name": "status",
                            "in": "query",
                            "required": true,
                            "schema": {
                                "type": "string",
                                "enum": ["available", "pending", "sold"]
                            }
                        }
                    ]
                }
            },
            "/pet/{petId}/owner/{ownerId}": {
                "get": {
                    "parameters": [
                        {
                            "name": "petId",
                            "in": "path",
                            "required": true,
                            "schema": {
                                "type": "integer"
                            }
                        },
                        {
                            "name": "ownerId",
                            "in": "path",
                            "required": true,
                            "schema": {
                                "type": "integer"
                            }
                        }
                    ]
                }
            }
        }
    }`)

	err := manager.LoadAPI("test", content)
	assert.NoError(t, err)

	validator := NewValidator(manager)
	err = validator.SetCurrentAPI("test")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		method        string
		path          string
		route         string
		setupRequest  func(*http.Request)
		expectedError string
	}{
		{
			name:   "Valid path parameter",
			method: http.MethodGet,
			path:   "/pet/123",
			route:  "/pet/{petId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "",
		},
		{
			name:   "Invalid path parameter type",
			method: http.MethodGet,
			path:   "/pet/abc",
			route:  "/pet/{petId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "invalid type for parameter 'petId'",
		},
		{
			name:   "Missing required query parameter",
			method: http.MethodGet,
			path:   "/pet/findByStatus",
			route:  "/pet/findByStatus",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "missing required parameter 'status'",
		},
		{
			name:   "Valid query parameter",
			method: http.MethodGet,
			path:   "/pet/findByStatus",
			route:  "/pet/findByStatus",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("status", "available")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid query parameter value",
			method: http.MethodGet,
			path:   "/pet/findByStatus",
			route:  "/pet/findByStatus",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("status", "invalid")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'status'",
		},
		{
			name:   "Valid multiple path parameters",
			method: http.MethodGet,
			path:   "/pet/123/owner/456",
			route:  "/pet/{petId}/owner/{ownerId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "",
		},
		{
			name:   "Invalid petId path parameter",
			method: http.MethodGet,
			path:   "/pet/abc/owner/456",
			route:  "/pet/{petId}/owner/{ownerId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "invalid type for parameter 'petId'",
		},
		{
			name:   "Invalid ownerId path parameter",
			method: http.MethodGet,
			path:   "/pet/123/owner/abc",
			route:  "/pet/{petId}/owner/{ownerId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "invalid type for parameter 'ownerId'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)

			tt.setupRequest(req)

			ok, err := validator.ValidateParameters(req, tt.route)
			if tt.expectedError == "" {
				assert.True(t, ok)
				assert.NoError(t, err)
			} else {
				assert.False(t, ok)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
