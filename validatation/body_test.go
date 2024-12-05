package validation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
)

func TestValidateRequestBody(t *testing.T) {
	spec := oas.OpenAPI{
		Paths: map[string]oas.PathItem{
			"/pet": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Required: true,
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									Type: "object",
									Properties: map[string]oas.Schema{
										"name": {Type: "string"},
										"age":  {Type: "integer"},
									},
									Required: []string{"name"},
								},
							},
						},
					},
				},
			},
			"/user": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Required: false,
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									Ref: "#/components/schemas/User",
								},
							},
						},
					},
				},
			},
			"/noschema": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {},
						},
					},
				},
			},
			"/oneofComponent": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									Ref: "#/components/schemas/OneOfSchema",
								},
							},
						},
					},
				},
			},
			"/anyofComponent": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									Ref: "#/components/schemas/AnyOfSchema",
								},
							},
						},
					},
				},
			},
			"/allofComponent": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									Ref: "#/components/schemas/AllOfSchema",
								},
							},
						},
					},
				},
			},
			"/oneof": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									OneOf: []oas.Schema{
										{Type: "string"},
										{Type: "number"},
									},
								},
							},
						},
					},
				},
			},
			"/anyof": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									AnyOf: []oas.Schema{
										{Type: "string", MaxLength: uint64Ptr(5)},
										{Type: "number", Minimum: float64Ptr(0)},
									},
								},
							},
						},
					},
				},
			},
			"/allof": {
				Post: &oas.Operation{
					RequestBody: &oas.RequestBody{
						Content: map[string]oas.MediaType{
							"application/json": {
								Schema: &oas.Schema{
									AllOf: []oas.Schema{
										{
											Type: "object",
											Properties: map[string]oas.Schema{
												"id": {Type: "integer"},
											},
										},
										{
											Type: "object",
											Properties: map[string]oas.Schema{
												"status": {Type: "string"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Components: &oas.Components{
			Schemas: map[string]oas.Schema{
				"User": {
					Type: "object",
					Properties: map[string]oas.Schema{
						"username": {Type: "string"},
						"email":    {Type: "string", Format: "email"},
					},
					Required: []string{"username"},
				},
				"OneOfSchema": {
					OneOf: []oas.Schema{
						{Type: "string", Format: "email"},
						{Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(100)},
					},
				},
				"AnyOfSchema": {
					AnyOf: []oas.Schema{
						{Type: "string", Pattern: "^[A-Z]{3}$"},
						{Type: "number", MultipleOf: float64Ptr(5)},
						{Type: "boolean"},
					},
				},
				"AllOfSchema": {
					AllOf: []oas.Schema{
						{
							Type:     "object",
							Required: []string{"id"},
							Properties: map[string]oas.Schema{
								"id": {Type: "integer"},
							},
						},
						{
							Type:     "object",
							Required: []string{"name", "age"},
							Properties: map[string]oas.Schema{
								"name": {Type: "string"},
								"age":  {Type: "integer", Minimum: float64Ptr(0)},
							},
						},
					},
				},
			},
		},
	}

	validator := NewValidator(spec)

	tests := []struct {
		name          string
		method        string
		path          string
		route         string
		body          string
		contentType   string
		expectedError string
	}{
		{
			name:          "Valid request body",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", "age": 5}`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Missing required field",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"age": 5}`,
			contentType:   "application/json",
			expectedError: "request body does not match schema",
		},
		{
			name:          "Invalid type",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", "age": "five"}`,
			contentType:   "application/json",
			expectedError: "request body does not match schema",
		},
		{
			name:          "Missing required body",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          "",
			contentType:   "application/json",
			expectedError: "request body is required",
		},
		{
			name:          "Unsupported content type",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy"}`,
			contentType:   "text/plain",
			expectedError: "unsupported content type 'text/plain'",
		},
		{
			name:          "Invalid JSON",
			method:        http.MethodPost,
			path:          "/pet",
			route:         "/pet",
			body:          `{"name": "Fluffy", invalid}`,
			contentType:   "application/json",
			expectedError: "invalid request body",
		},
		{
			name:          "Valid schema reference",
			method:        http.MethodPost,
			path:          "/user",
			route:         "/user",
			body:          `{"username": "john", "email": "john@example.com"}`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "No schema defined",
			method:        http.MethodPost,
			path:          "/noschema",
			route:         "/noschema",
			body:          `{"anything": "goes"}`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Valid oneOf component",
			method:        http.MethodPost,
			path:          "/oneofComponent",
			route:         "/oneofComponent",
			body:          `"test@example.com"`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Invalid oneOf component",
			method:        http.MethodPost,
			path:          "/oneofComponent",
			route:         "/oneofComponent",
			body:          `"not-an-email"`,
			contentType:   "application/json",
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid anyOf component",
			method:        http.MethodPost,
			path:          "/anyofComponent",
			route:         "/anyofComponent",
			body:          `"ABC"`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Invalid anyOf component",
			method:        http.MethodPost,
			path:          "/anyofComponent",
			route:         "/anyofComponent",
			body:          `"ABCD"`,
			contentType:   "application/json",
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid allOf component",
			method:        http.MethodPost,
			path:          "/allofComponent",
			route:         "/allofComponent",
			body:          `{"id": 1, "name": "John", "age": 25}`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Invalid allOf component",
			method:        http.MethodPost,
			path:          "/allofComponent",
			route:         "/allofComponent",
			body:          `{"id": 1, "name": "John"}`,
			contentType:   "application/json",
			expectedError: "request body does not match schema",
		},
		{
			name:          "Valid oneOf",
			method:        http.MethodPost,
			path:          "/oneof",
			route:         "/oneof",
			body:          `"test"`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Valid anyOf",
			method:        http.MethodPost,
			path:          "/anyof",
			route:         "/anyof",
			body:          `"abc"`,
			contentType:   "application/json",
			expectedError: "",
		},
		{
			name:          "Valid allOf",
			method:        http.MethodPost,
			path:          "/allof",
			route:         "/allof",
			body:          `{"id": 1, "status": "active"}`,
			contentType:   "application/json",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			ok, err := validator.ValidateRequestBody(req, tt.route)

			if tt.expectedError == "" {
				if !ok || err != nil {
					t.Errorf("expected validation to pass, got error: %v", err)
				}
			} else {
				if ok || err == nil || !ErrorContains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

func ErrorContains(got, want string) bool {
	return got == want || len(got) >= len(want) && got[0:len(want)] == want
}
