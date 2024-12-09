package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/lionelgarnier/validate-api-request/oas"
)

func TestMiddleware(t *testing.T) {
	selector := oas.HostBasedSelector(map[string]string{
		"api.pets.com":  "petstore",
		"api.users.com": "userapi",
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Create middleware
	middleware := NewMiddleware(
		handler,
		oas.DefaultCacheConfig(),
		selector,
	)

	//load API OAS from file
	var filePath string

	filePath = filepath.Join("..", "test_data", "petstore3.swagger.io_api_json.json")
	middleware.manager.LoadAPIFromFile("petstore", filePath)

	filePath = filepath.Join("..", "test_data", "advancedoas.swagger.io.json")
	middleware.manager.LoadAPIFromFile("advancedoas", filePath)

	// Create test request
	req, err := http.NewRequest("GET", "/pet/10", nil)
	if err != nil {
		panic(err)
	}
	req.Host = "api.pets.com"

	// Test middleware
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Output:
	// OK
}
