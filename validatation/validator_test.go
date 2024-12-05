package validation

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestValidateRequest(t *testing.T) {
	parser := oas.NewParser()
	pwd, _ := os.Getwd()
	t.Logf("Current working directory: %s", pwd)

	filePath := filepath.Join("..", "test_data", "petstore3.swagger.io_api_json.json")

	_, err := parser.LoadFromFile(filePath)
	assert.NoError(t, err)

	spec, err := parser.Parse()
	assert.NoError(t, err)

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
			headers: map[string]string{"Content-Type": "application/json"},
			body:    `{"name": "doggie", "photoUrls": ["http://example.com"]}`,
			wantErr: false,
		},
		{
			name:    "invalid pet create request - missing required fields",
			path:    "/pet",
			method:  http.MethodPost,
			headers: map[string]string{"Content-Type": "application/json"},
			body:    `{"id": 1}`,
			wantErr: true,
			//wantErrMsg: "missing required fields: name,photoUrls",
			wantErrMsg: "request body does not match schema",
		},
		{
			name:    "valid pet get by status",
			path:    "/pet/findByStatus",
			method:  http.MethodGet,
			query:   map[string]string{"status": "available"},
			wantErr: false,
		},
		{
			name:       "invalid pet get by status",
			path:       "/pet/findByStatus",
			method:     http.MethodGet,
			query:      map[string]string{"status": "invalid_status"},
			wantErr:    true,
			wantErrMsg: "invalid type for parameter 'status'",
			//wantErrMsg: "parameter status value 'invalid_status' is not one of allowed values",
		},
	}

	validator := NewValidator(*spec)

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

			_, err = validator.ValidateRequest(req)
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
