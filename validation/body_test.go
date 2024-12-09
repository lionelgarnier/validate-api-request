package validation

import (
	"net/http"
	"strings"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequestBody(t *testing.T) {

	manager := oas.NewOASManager(nil)

	// Load test API spec
	content := []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/pet": {
				"post": {
					"requestBody": {
						"required": true,
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {"type": "string"},
										"age": {"type": "integer"}
									},
									"required": ["name"]
								}
							}
						}
					}
				}
			},
			"/user": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"$ref": "#/components/schemas/User"
								}
							}
						}
					}
				}
			},
			"/noschema": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {}
						}
					}
				}
			},
			"/oneofComponent": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"$ref": "#/components/schemas/OneOfSchema"
								}
							}
						}
					}
				}
			},
			"/anyofComponent": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"$ref": "#/components/schemas/AnyOfSchema"
								}
							}
						}
					}
				}
			},
			"/allofComponent": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"$ref": "#/components/schemas/AllOfSchema"
								}
							}
						}
					}
				}
			},
			"/oneof": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"oneOf": [
										{"type": "string"},
										{"type": "number"}
									]
								}
							}
						}
					}
				}
			},
			"/anyof": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"anyOf": [
										{
											"type": "string",
											"maxLength": 5
										},
										{
											"type": "number",
											"minimum": 0
										}
									]
								}
							}
						}
					}
				}
			},
			"/allof": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"allOf": [
										{
											"type": "object",
											"properties": {
												"id": {"type": "integer"}
											}
										},
										{
											"type": "object",
											"properties": {
												"status": {"type": "string"}
											}
										}
									]
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"username": {"type": "string"},
						"email": {"type": "string", "format": "email"}
					},
					"required": ["username"]
				},
				"OneOfSchema": {
					"oneOf": [
						{"type": "string", "format": "email"},
						{"type": "integer", "minimum": 0, "maximum": 100}
					]
				},
				"AnyOfSchema": {
					"anyOf": [
						{"type": "string", "pattern": "^[A-Z]{3}$"},
						{"type": "number", "multipleOf": 5},
						{"type": "boolean"}
					]
				},
				"AllOfSchema": {
					"allOf": [
						{
							"type": "object",
							"required": ["id"],
							"properties": {
								"id": {"type": "integer"}
							}
						},
						{
							"type": "object",
							"required": ["name", "age"],
							"properties": {
								"name": {"type": "string"},
								"age": {"type": "integer", "minimum": 0}
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
		path          string
		route         string
		method        string
		headers       map[string]string
		body          string
		expectedError string
	}{
		{
			name:          "Valid request body",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", "age": 5}`,
			expectedError: "",
			headers:       map[string]string{"Content-Type": "application/json"},
		},
		{
			name:          "Missing required field",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"age": 5}`,
			expectedError: "request body does not match schema",
		},
		{
			name:          "Invalid type",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", "age": "five"}`,
			expectedError: "request body does not match schema",
		},
		{
			name:          "Missing required body",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          "",
			expectedError: "request body is required",
		},
		{
			name:          "Unsupported content type",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy"}`,
			headers:       map[string]string{"Content-Type": "text/plain"},
			expectedError: "unsupported content type 'text/plain'",
		},
		{
			name:          "Invalid JSON",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", invalid}`,
			expectedError: "invalid request body",
		},
		{
			name:          "Valid schema reference",
			method:        http.MethodPost,
			path:          "/user",
			route:         "/user",
			body:          `{"username": "john", "email": "john@example.com"}`,
			expectedError: "",
		},
		{
			name:          "No schema defined",
			method:        http.MethodPost,
			path:          "/noschema",
			route:         "/noschema",
			body:          `{"anything": "goes"}`,
			expectedError: "",
		},
		{
			name:          "Valid oneOf component",
			method:        http.MethodPost,
			path:          "/oneofComponent",
			route:         "/oneofComponent",
			body:          `"test@example.com"`,
			expectedError: "",
		},
		{
			name:          "Invalid oneOf component",
			method:        http.MethodPost,
			path:          "/oneofComponent",
			route:         "/oneofComponent",
			body:          `"not-an-email"`,
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid anyOf component",
			method:        http.MethodPost,
			path:          "/anyofComponent",
			route:         "/anyofComponent",
			body:          `"ABC"`,
			expectedError: "",
		},
		{
			name:          "Invalid anyOf component",
			method:        http.MethodPost,
			path:          "/anyofComponent",
			route:         "/anyofComponent",
			body:          `"ABCD"`,
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid allOf component",
			method:        http.MethodPost,
			path:          "/allofComponent",
			route:         "/allofComponent",
			body:          `{"id": 1, "name": "John", "age": 25}`,
			expectedError: "",
		},
		{
			name:          "Invalid allOf component",
			method:        http.MethodPost,
			path:          "/allofComponent",
			route:         "/allofComponent",
			body:          `{"id": 1, "name": "John"}`,
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid oneOf",
			method:        http.MethodPost,
			path:          "/oneof",
			route:         "/oneof",
			body:          `"test"`,
			expectedError: "",
		},
		{
			name:          "Valid anyOf",
			method:        http.MethodPost,
			path:          "/anyof",
			route:         "/anyof",
			body:          `"abc"`,
			expectedError: "",
		},
		{
			name:          "Valid allOf",
			method:        http.MethodPost,
			path:          "/allof",
			route:         "/allof",
			body:          `{"id": 1, "status": "active"}`,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			assert.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ok, err := validator.ValidateRequestBody(req, tt.route)
			if tt.expectedError != "" {
				assert.False(t, ok)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.True(t, ok)
				assert.NoError(t, err)
			}
		})
	}
}
