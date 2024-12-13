package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMiddleware(t *testing.T) {
	// Load configuration from YAML file
	filePath := filepath.Join("..", "config.yaml")
	config, err := LoadConfigFromFile(filePath)
	if err != nil {
		panic(err)
	}

	// Create your next handler (the final handler in the chain)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Create the middleware
	middleware, err := New(nextHandler, config)
	if err != nil {
		panic(err)
	}

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
