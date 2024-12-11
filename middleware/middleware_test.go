package middleware

import (
	"net/http"
	"net/http/httptest"
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

	config := CreateConfig()
	config.cacheconfig = oas.DefaultCacheConfig()
	config.selector = selector

	// Create middleware
	middleware := New(
		handler,
		config,
	)

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
