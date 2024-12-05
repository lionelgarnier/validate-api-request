package validation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestValidateParameters(t *testing.T) {
	spec := oas.OpenAPI{
		Paths: map[string]oas.PathItem{
			"/pet/{petId}/owner/{ownerId}": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "petId",
							In:       "path",
							Required: true,
							Schema: &oas.Schema{
								Type: "integer",
							},
						},
						{
							Name:     "ownerId",
							In:       "path",
							Required: true,
							Schema: &oas.Schema{
								Type: "integer",
							},
						},
					},
				},
			},
			"/pet/{petId}": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "petId",
							In:       "path",
							Required: true,
							Schema: &oas.Schema{
								Type: "integer",
							},
						},
					},
				},
			},
			"/pet/findByStatus": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "status",
							In:       "query",
							Required: true,
							Schema: &oas.Schema{
								Type: "string",
								Enum: []interface{}{"available", "pending", "sold"},
							},
						},
					},
				},
			},
			"/user": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "api_key",
							In:       "header",
							Required: true,
							Schema: &oas.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			"/user/login": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "session",
							In:   "cookie",
							Schema: &oas.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			"/store/inventory": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "timestamp",
							In:       "query",
							Required: true,
							Schema: &oas.Schema{
								Type:   "string",
								Format: "date-time",
							},
						},
						{
							Name: "email",
							In:   "query",
							Schema: &oas.Schema{
								Type:   "string",
								Format: "email",
							},
						},
						{
							Name: "ip",
							In:   "query",
							Schema: &oas.Schema{
								Type:   "string",
								Format: "ipv4",
							},
						},
					},
				},
			},
			"/arrayParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "ids",
							In:   "query",
							Schema: &oas.Schema{
								Type: "array",
								Items: &oas.Schema{
									Type: "integer",
								},
							},
						},
					},
				},
			},
			"/optionalParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "optional",
							In:   "query",
							Schema: &oas.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			"/booleanParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "flag",
							In:   "query",
							Schema: &oas.Schema{
								Type: "boolean",
							},
						},
					},
				},
			},
			"/componentParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Schema: &oas.Schema{
								Ref: "#/components/parameters/TestParam",
							},
						},
					},
				},
			},
			"/componentAllOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Schema: &oas.Schema{
								Ref: "#/components/parameters/AllOfParam",
							},
						},
					},
				},
			},
			"/componentOneOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Schema: &oas.Schema{
								Ref: "#/components/parameters/OneOfParam",
							},
						},
					},
				},
			},
			"/componentAnyOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Schema: &oas.Schema{
								Ref: "#/components/parameters/AnyOfParam",
							},
						},
					},
				},
			},
			"/multipleParams": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name:     "timestamp",
							In:       "query",
							Required: true,
							Schema: &oas.Schema{
								Type:   "string",
								Format: "date-time",
							},
						},
						{
							Name:     "uuid",
							In:       "query",
							Required: true,
							Schema: &oas.Schema{
								Type:   "string",
								Format: "uuid",
							},
						},
						{
							Name: "url",
							In:   "query",
							Schema: &oas.Schema{
								Type:   "string",
								Format: "url",
							},
						},
					},
				},
			},
			"/arrayOfComponents": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "items",
							In:   "query",
							Schema: &oas.Schema{
								Type: "array",
								Items: &oas.Schema{
									Ref: "#/components/schemas/Pet",
								},
							},
						},
					},
				},
			},
			"/allOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "obj",
							In:   "query",
							Schema: &oas.Schema{
								AllOf: []oas.Schema{
									{Type: "object", Required: []string{"id"}, Properties: map[string]oas.Schema{
										"id": {Type: "integer"},
									}},
									{Type: "object", Required: []string{"name"}, Properties: map[string]oas.Schema{
										"name": {Type: "string"},
									}},
								},
							},
						},
					},
				},
			},
			"/oneOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "data",
							In:   "query",
							Schema: &oas.Schema{
								OneOf: []oas.Schema{
									{Type: "integer"},
									{Type: "string", Format: "email"},
								},
							},
						},
					},
				},
			},
			"/anyOfParam": {
				Get: &oas.Operation{
					Parameters: []oas.Parameter{
						{
							Name: "value",
							In:   "query",
							Schema: &oas.Schema{
								AnyOf: []oas.Schema{
									{Type: "string", Pattern: "^[A-Z]{3}$"},
									{Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(100)},
								},
							},
						},
					},
				},
			},
		},
		Components: &oas.Components{
			Parameters: map[string]oas.Parameter{
				"TestParam": {
					Name:     "TestParam",
					In:       "query",
					Required: true,
					Schema: &oas.Schema{
						Type: "string",
					},
				},
				"AllOfParam": {
					Name:     "AllOfParam",
					In:       "query",
					Required: true,
					Schema: &oas.Schema{
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
								Required: []string{"name"},
								Properties: map[string]oas.Schema{
									"name": {Type: "string"},
								},
							},
						},
					},
				},
				"OneOfParam": {
					Name:     "OneOfParam",
					In:       "query",
					Required: true,
					Schema: &oas.Schema{
						OneOf: []oas.Schema{
							{Type: "integer"},
							{Type: "string", Format: "email"},
						},
					},
				},
				"AnyOfParam": {
					Name:     "AnyOfParam",
					In:       "query",
					Required: true,
					Schema: &oas.Schema{
						AnyOf: []oas.Schema{
							{Type: "string", Pattern: "^[A-Z]{3}$"},
							{Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(100)},
						},
					},
				},
			},
			Schemas: map[string]oas.Schema{
				"Pet": {
					Type: "object",
					Properties: map[string]oas.Schema{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
					Required: []string{"id", "name"},
				},
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
		setupRequest  func(r *http.Request)
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
			name:   "Valid header parameter",
			method: http.MethodGet,
			path:   "/user",
			route:  "/user",
			setupRequest: func(r *http.Request) {
				r.Header.Set("api_key", "12345")
			},
			expectedError: "",
		},
		{
			name:   "Missing required header parameter",
			method: http.MethodGet,
			path:   "/user",
			route:  "/user",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "missing required parameter 'api_key'",
		},
		{
			name:   "Valid cookie parameter",
			method: http.MethodGet,
			path:   "/user/login",
			route:  "/user/login",
			setupRequest: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
			},
			expectedError: "",
		},
		{
			name:   "Valid array parameter",
			method: http.MethodGet,
			path:   "/arrayParam",
			route:  "/arrayParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("ids", "[1,2,3]")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Optional parameter not provided",
			method: http.MethodGet,
			path:   "/optionalParam",
			route:  "/optionalParam",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "",
		},
		{
			name:   "Valid boolean parameter",
			method: http.MethodGet,
			path:   "/booleanParam",
			route:  "/booleanParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("flag", "true")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid boolean parameter",
			method: http.MethodGet,
			path:   "/booleanParam",
			route:  "/booleanParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("flag", "notabool")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'flag'",
		},
		{
			name:   "Valid component parameter",
			method: http.MethodGet,
			path:   "/componentParam",
			route:  "/componentParam",
			setupRequest: func(r *http.Request) {
				// Assuming TestParam is a string parameter
				q := r.URL.Query()
				q.Add("TestParam", "test123")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Valid multiple parameters",
			method: http.MethodGet,
			path:   "/multipleParams",
			route:  "/multipleParams",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("timestamp", "2023-06-15T10:00:00Z")
				q.Add("uuid", "123e4567-e89b-12d3-a456-426614174000")
				q.Add("url", "https://example.com")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Missing required timestamp in multiple params",
			method: http.MethodGet,
			path:   "/multipleParams",
			route:  "/multipleParams",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("uuid", "123e4567-e89b-12d3-a456-426614174000")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "missing required parameter 'timestamp'",
		},
		{
			name:   "Invalid timestamp format in multiple params",
			method: http.MethodGet,
			path:   "/multipleParams",
			route:  "/multipleParams",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("timestamp", "invalid-date")
				q.Add("uuid", "123e4567-e89b-12d3-a456-426614174000")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'timestamp'",
		},
		{
			name:   "Valid array of components parameter",
			method: http.MethodGet,
			path:   "/arrayOfComponents",
			route:  "/arrayOfComponents",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("items", `[{"id":1,"name":"pet1"},{"id":2,"name":"pet2"}]`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid array of components - missing required field",
			method: http.MethodGet,
			path:   "/arrayOfComponents",
			route:  "/arrayOfComponents",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("items", `[{"id":1},{"id":2,"name":"pet2"}]`) // Missing name field
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'items'",
		},
		{
			name:   "Invalid array of components - wrong type",
			method: http.MethodGet,
			path:   "/arrayOfComponents",
			route:  "/arrayOfComponents",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("items", `[{"id":"not-an-integer","name":"pet1"}]`) // id should be integer
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'items'",
		},
		{
			name:   "Invalid array of components - malformed JSON",
			method: http.MethodGet,
			path:   "/arrayOfComponents",
			route:  "/arrayOfComponents",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("items", `[{"id":1,name:"pet1"}]`) // Malformed JSON
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'items'",
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
		{
			name:   "Invalid both path parameters",
			method: http.MethodGet,
			path:   "/pet/abc/owner/def",
			route:  "/pet/{petId}/owner/{ownerId}",
			setupRequest: func(r *http.Request) {
			},
			expectedError: "invalid type for parameter 'petId'",
		},
		{
			name:   "Valid allOf parameter",
			method: http.MethodGet,
			path:   "/allOfParam",
			route:  "/allOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("obj", `{"id":1,"name":"test"}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid allOf parameter - missing required field",
			method: http.MethodGet,
			path:   "/allOfParam",
			route:  "/allOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("obj", `{"id":1}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'obj'",
		},
		{
			name:   "Valid oneOf parameter",
			method: http.MethodGet,
			path:   "/oneOfParam",
			route:  "/oneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("data", "test@example.com")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid oneOf parameter",
			method: http.MethodGet,
			path:   "/oneOfParam",
			route:  "/oneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("data", "not-an-email")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'data'",
		},
		{
			name:   "Valid anyOf parameter - string pattern",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "ABC")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Valid anyOf parameter - integer range",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "50")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid anyOf parameter",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "invalid")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'value'",
		},
		{
			name:   "Valid allOf component parameter",
			method: http.MethodGet,
			path:   "/componentAllOfParam",
			route:  "/componentAllOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("AllOfParam", `{"id":1,"name":"test"}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid allOf component parameter - missing required field",
			method: http.MethodGet,
			path:   "/componentAllOfParam",
			route:  "/componentAllOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("AllOfParam", `{"id":1}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'AllOfParam'",
		},
		{
			name:   "Valid oneOf component parameter",
			method: http.MethodGet,
			path:   "/componentOneOfParam",
			route:  "/componentOneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("OneOfParam", "test@example.com")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid oneOf component parameter",
			method: http.MethodGet,
			path:   "/componentOneOfParam",
			route:  "/componentOneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("OneOfParam", "not-an-email")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'OneOfParam'",
		},
		{
			name:   "Valid anyOf component parameter - string pattern",
			method: http.MethodGet,
			path:   "/componentAnyOfParam",
			route:  "/componentAnyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("AnyOfParam", "ABC")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Valid anyOf component parameter - integer range",
			method: http.MethodGet,
			path:   "/componentAnyOfParam",
			route:  "/componentAnyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("AnyOfParam", "50")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid anyOf component parameter",
			method: http.MethodGet,
			path:   "/componentAnyOfParam",
			route:  "/componentAnyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("AnyOfParam", "invalid")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'AnyOfParam'",
		},
		{
			name:   "Valid allOf parameter",
			method: http.MethodGet,
			path:   "/allOfParam",
			route:  "/allOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("obj", `{"id":1,"name":"test"}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid allOf parameter - missing required field",
			method: http.MethodGet,
			path:   "/allOfParam",
			route:  "/allOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("obj", `{"id":1}`)
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'obj'",
		},
		{
			name:   "Valid oneOf parameter",
			method: http.MethodGet,
			path:   "/oneOfParam",
			route:  "/oneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("data", "test@example.com")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid oneOf parameter",
			method: http.MethodGet,
			path:   "/oneOfParam",
			route:  "/oneOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("data", "not-an-email")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'data'",
		},
		{
			name:   "Valid anyOf parameter - string pattern",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "ABC")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Valid anyOf parameter - integer range",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "50")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "",
		},
		{
			name:   "Invalid anyOf parameter",
			method: http.MethodGet,
			path:   "/anyOfParam",
			route:  "/anyOfParam",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("value", "invalid")
				r.URL.RawQuery = q.Encode()
			},
			expectedError: "invalid type for parameter 'value'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			tt.setupRequest(req)

			ok, err := validator.ValidateParameters(req, tt.route)

			if tt.expectedError == "" {
				if !ok || err != nil {
					t.Errorf("expected validation to pass, got error: %v", err)
				}
			} else {
				if ok || err == nil || err.Error() != tt.expectedError {
					t.Errorf("expected error '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}
