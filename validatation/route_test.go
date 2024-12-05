package validation

import (
	"net/http"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/stretchr/testify/assert"
)

func TestResolveRequestPath(t *testing.T) {
	spec := oas.OpenAPI{
		Paths: map[string]oas.PathItem{
			"/pets": {
				Get: &oas.Operation{},
			},
			"/pets/{petId}": {
				Get:    &oas.Operation{},
				Post:   &oas.Operation{},
				Delete: &oas.Operation{},
			},
		},
	}

	validator := NewValidator(spec)

	tests := []struct {
		path     string
		expected string
	}{
		{"/pets", "/pets"},
		{"/pets", "/pets"},
		{"/pets/123", "/pets/{petId}"},
		{"/pets/123", "/pets/{petId}"},
		{"/unknown", ""},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodGet, test.path, nil)
		route, _ := validator.ResolveRequestPath(req)
		assert.Equal(t, test.expected, route)
	}
}

func TestValidateRequestPath(t *testing.T) {
	spec := oas.OpenAPI{
		Paths: map[string]oas.PathItem{
			"/pets": {
				Get: &oas.Operation{},
			},
			"/pets/{petId}": {
				Get:    &oas.Operation{},
				Post:   &oas.Operation{},
				Delete: &oas.Operation{},
			},
		},
	}

	validator := NewValidator(spec)

	tests := []struct {
		path     string
		route    string
		expected bool
		err      string
	}{
		{"/pets", "/pets", true, ""},
		{"/pets", "/prouts", false, "path '/prouts' doesn't exist in oas schema"},
		{"/pets", "/pets/{petId}", false, "request path '/pets' is not matching with oas schema path '/pets/{petId}' as there is another path with same name"},
		{"/petards", "/pets", false, "request path '/petards' is not matching with oas schema path '/pets'"},
		{"/pets/123", "/pets/{petId}", true, ""},
		{"/pets/123", "/prouts", false, "path '/prouts' doesn't exist in oas schema"},
		{"/pets/123", "/pets", false, "request path '/pets/123' is not matching with oas schema path '/pets'"},
		{"/unknown", "/unknown", false, "path '/unknown' doesn't exist in oas schema"},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(http.MethodGet, test.path, nil)
		ok, err := validator.ValidateRequestPath(req, test.route)
		assert.Equal(t, test.expected, ok)
		if !ok {
			assert.NotEmpty(t, err)
			assert.Contains(t, err.Error(), test.err)
		}
	}
}

func TestValidateRequestMethod(t *testing.T) {
	spec := oas.OpenAPI{
		Paths: map[string]oas.PathItem{
			"/pets": {
				Get: &oas.Operation{},
			},
			"/pets/{petId}": {
				Get:    &oas.Operation{},
				Post:   &oas.Operation{},
				Delete: &oas.Operation{},
			},
		},
	}

	validator := NewValidator(spec)

	tests := []struct {
		path     string
		route    string
		method   string
		expected bool
		err      string
	}{
		{"/pets", "/pets", http.MethodGet, true, ""},
		{"/pets", "/pets", http.MethodPost, false, "method 'POST' not allowed for path '/pets'"},
		{"/pets/123", "/pets/{petId}", http.MethodGet, true, ""},
		{"/pets/123", "/pets/{petId}", http.MethodPost, true, ""},
		{"/pets/123", "/pets/{petId}", http.MethodDelete, true, ""},
		{"/pets/123", "/pets/{petId}", http.MethodPatch, false, "method 'PATCH' not allowed for path '/pets/{petId}'"},
		{"/unknown", "/unknown", http.MethodGet, false, "path '/unknown' doesn't exist in oas schema"},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		ok, err := validator.ValidateRequestMethod(req, test.route)
		assert.Equal(t, test.expected, ok)
		if !ok {
			assert.NotEmpty(t, err)
			assert.Contains(t, err.Error(), test.err)
		}
	}
}
