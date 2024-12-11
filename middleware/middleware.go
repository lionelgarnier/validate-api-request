package middleware

import (
	"net/http"
	"path/filepath"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/validation"
)

// OASMiddleware validates requests against OpenAPI specs
type OASMiddleware struct {
	next      http.Handler
	manager   *oas.OASManager
	validator validation.Validator
}

type Config struct {
	selector    oas.APISelector
	cacheconfig *oas.CacheConfig
}

func CreateConfig() *Config {
	return &Config{}
}

// NewMiddleware creates a new OASMiddleware
func New(next http.Handler, config *Config) *OASMiddleware {
	manager := oas.NewOASManager(config.cacheconfig, config.selector)

	//load API OAS from file
	var filePath string

	filePath = filepath.Join("..", "test_data", "petstore3.swagger.io_api_json.json")
	manager.LoadAPIFromFile("petstore", filePath)

	filePath = filepath.Join("..", "test_data", "advancedoas.swagger.io.json")
	manager.LoadAPIFromFile("advancedoas", filePath)

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

	oasRequest := oas.NewOASRequest(r)

	// Validate request
	if ok, err := m.validator.ValidateRequest(oasRequest); !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call next handler
	m.next.ServeHTTP(w, r)
}
