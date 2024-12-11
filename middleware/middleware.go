package middleware

import (
	"net/http"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/validation"
)

// OASMiddleware validates requests against OpenAPI specs
type OASMiddleware struct {
	next      http.Handler
	manager   *oas.OASManager
	validator validation.Validator
}

// NewMiddleware creates a new OASMiddleware
func NewMiddleware(next http.Handler, config *oas.CacheConfig, selector oas.APISelector) *OASMiddleware {
	manager := oas.NewOASManager(config, selector)
	validator := validation.NewValidator(nil)

	return &OASMiddleware{
		next:      next,
		manager:   manager,
		validator: validator,
	}
}

// ServeHTTP validates the request against the OpenAPI spec
func (m *OASMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get API spec for request
	spec, err := m.manager.GetApiSpecForRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set spec in validator
	m.validator.SetApiSpec(spec)

	oasRequest := &oas.OASRequest{
		Request: r,
		Route:   "",
	}

	// Validate request
	if ok, err := m.validator.ValidateRequest(oasRequest); !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call next handler
	m.next.ServeHTTP(w, r)
}
