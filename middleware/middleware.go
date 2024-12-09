package middleware

import (
	"net/http"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/validation"
)

// OASMiddleware validates requests against OpenAPI specs
type OASMiddleware struct {
	next        http.Handler
	manager     *oas.OASManager
	validator   validation.Validator
	apiSelector oas.APISelector
}

func NewMiddleware(next http.Handler, config *oas.CacheConfig, selector oas.APISelector) *OASMiddleware {
	manager := oas.NewOASManager(config)
	validator := validation.NewValidator(manager)

	return &OASMiddleware{
		next:        next,
		manager:     manager,
		validator:   validator,
		apiSelector: selector,
	}
}

func (m *OASMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiName := m.apiSelector(r)
	if apiName == "" {
		http.Error(w, "could not determine API specification", http.StatusBadRequest)
		return
	}

	err := m.validator.SetCurrentAPI(apiName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = m.validator.ValidateRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m.next.ServeHTTP(w, r)
}
