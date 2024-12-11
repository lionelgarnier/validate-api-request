package validation

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequest(t *testing.T) {
	manager := oas.NewOASManager(nil, oas.FixedSelector("test"))
	filePath := filepath.Join("..", "test_data", "petstore3.swagger.io_api_json.json")
	manager.LoadAPIFromFile("test", filePath)

	tests := []struct {
		name       string
		path       string
		method     string
		headers    map[string]string
		query      map[string]string
		body       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "valid pet create request",
			path:    "/pet",
			method:  http.MethodPost,
			headers: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer valid-oauth2-token"},
			body:    `{"name": "doggie", "photoUrls": ["http://example.com"]}`,
			wantErr: false,
		},
		{
			name:    "invalid pet create request - missing required fields",
			path:    "/pet",
			method:  http.MethodPost,
			headers: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer valid-oauth2-token"},
			body:    `{"id": 1}`,
			wantErr: true,
			//wantErrMsg: "missing required fields: name,photoUrls",
			wantErrMsg: "request body does not match schema",
		},
		{
			name:    "valid pet get by status",
			path:    "/pet/findByStatus",
			method:  http.MethodGet,
			headers: map[string]string{"Authorization": "Bearer valid-oauth2-token"},
			query:   map[string]string{"status": "available"},
			wantErr: false,
		},
		{
			name:       "invalid pet get by status",
			path:       "/pet/findByStatus",
			method:     http.MethodGet,
			headers:    map[string]string{"Authorization": "Bearer valid-oauth2-token"},
			query:      map[string]string{"status": "invalid_status"},
			wantErr:    true,
			wantErrMsg: "invalid type for parameter 'status'",
			//wantErrMsg: "parameter status value 'invalid_status' is not one of allowed values",
		},
		{
			name:       "missing security token format pet create request",
			path:       "/pet",
			method:     http.MethodPost,
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       `{"name": "doggie", "photoUrls": ["http://example.com"]}`,
			wantErr:    true,
			wantErrMsg: "request does not satisfy any security requirements",
		},
	}

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			assert.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if tt.query != nil {
				q := req.URL.Query()
				for k, v := range tt.query {
					q.Add(k, v)
				}
				req.URL.RawQuery = q.Encode()
			}

			oasRequest := oas.NewOASRequest(req)

			_, err = validator.ValidateRequest(oasRequest)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComplexRequest(t *testing.T) {
	manager := oas.NewOASManager(nil, oas.FixedSelector("test"))
	filePath := filepath.Join("..", "test_data", "advancedoas.swagger.io.json")
	manager.LoadAPIFromFile("test", filePath)

	tests := []struct {
		name       string
		path       string
		method     string
		headers    map[string]string
		query      map[string]string
		body       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid - Cat with age",
			path:    "/validateAllOf",
			method:  http.MethodPatch,
			headers: map[string]string{"Content-Type": "application/json"},
			body: `{
                "pet_type": "Cat",
                "age": 3
            }`,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:    "Valid - Dog with bark",
			path:    "/validateAllOf",
			method:  http.MethodPatch,
			headers: map[string]string{"Content-Type": "application/json"},
			body: `{
                "pet_type": "Dog",
                "bark": true
            }`,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:    "Valid - Dog with bark and breed",
			path:    "/validateAllOf",
			method:  http.MethodPatch,
			headers: map[string]string{"Content-Type": "application/json"},
			body: `{
                "pet_type": "Dog",
                "bark": false,
                "breed": "Dingo"
            }`,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:    "Invalid - Missing pet_type",
			path:    "/validateAllOf",
			method:  http.MethodPatch,
			headers: map[string]string{"Content-Type": "application/json"},
			body: `{
                "age": 3
            }`,
			wantErr:    true,
			wantErrMsg: "request body does not match schema",
		},
		{
			name:    "Invalid - Cat with missing age",
			path:    "/validateAllOf",
			method:  http.MethodPatch,
			headers: map[string]string{"Content-Type": "application/json"},
			body: `{
                "pet_type": "Cat",
                "bark": true
            }`,
			wantErr:    true,
			wantErrMsg: "request body does not match schema",
		},
	}

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			assert.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if tt.query != nil {
				q := req.URL.Query()
				for k, v := range tt.query {
					q.Add(k, v)
				}
				req.URL.RawQuery = q.Encode()
			}

			oasRequest := oas.NewOASRequest(req)

			_, err = validator.ValidateRequest(oasRequest)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithDiscriminator(t *testing.T) {
	manager := oas.NewOASManager(nil, oas.FixedSelector("test"))
	filePath := filepath.Join("..", "test_data", "advancedoas.swagger.io.json")
	manager.LoadAPIFromFile("test", filePath)

	spec, _ := manager.GetApiSpec("test")
	validator := NewValidator(spec)

	dogJson := `{
        "pet_type": "Dog",
        "bark": true,
        "breed": "Husky"
    }`

	var dog interface{}
	json.Unmarshal([]byte(dogJson), &dog)

	schema := oas.Schema{
		OneOf: []oas.Schema{
			{Ref: "#/components/schemas/Dog"},
			{Ref: "#/components/schemas/Cat"},
		},
		Discriminator: &oas.Discriminator{
			PropertyName: "pet_type",
			Mapping: map[string]string{
				"Dog": "#/components/schemas/Dog",
				"Cat": "#/components/schemas/Cat",
			},
		},
	}
	result := validator.ValidateSchema(dog, &schema)
	assert.True(t, result)
}
