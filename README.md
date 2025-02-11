# validate-api-request
validate-api-request is a middleware for validating HTTP requests against OpenAPI specifications. This ensures that incoming requests conform to the defined API contract, improving the reliability and robustness of your API.

validate-api-request stands out by working directly with the authentic OpenAPI Specification (OAS) structure, avoiding any conversion to a different format. This direct approach simplifies configuration, debugging, and integration of new OAS functionalities while ensuring accurate validations and maintaining compatibility with evolving industry standards.


## Features

- Validate requests against OpenAPI 3.0 specifications
- Support for multiple APIs
- Configurable via YAML
- Supports both YAML and JSON OpenAPI specification formats
- Flexible API selection mechanisms


## High-Level Functionality

- The middleware starts by loading OpenAPI specifications that are either stored locally (via a file) or provided as inline text.
- It supports loading multiple specifications concurrently, making it suitable for environments with several APIs.
- Upon initialization, the middleware loads and validates the OAS definitions to ensure they are correctly formatted and accessible.
- For each incoming HTTP request, the middleware identifies the appropriate OAS based on criteria like host, header, or path prefix.
- The middleware then checks the details of the request against the corresponding OpenAPI specification.
- If there is any discrepancy or mismatch between the HTTP request and the OAS, the middleware returns a failure response to ensure API contract adherence.
- This setup ensures reliability by validating requests in real-time against defined API contracts.


## Installation

To install the middleware, use the following command:

```bash
go get github.com/lionelgarnier/validate-api-request
```

## Configuration

The middleware is configured via a YAML file. Below is an example configuration file (`config.yaml`):

```yaml
apis:
        - name: petstore
                specFile: "../oas_files/petstore3.swagger.io_api_json.json"
        - name: advancedoas
                specFile: "../oas_files/advancedoas.swagger.io.json"
        - name: inlineapi
                specText: |
                        {
                                "openapi": "3.0.0",
                                "info": {
                                                "title": "Test API",
                                                "version": "1.0.0"
                                },
                                "paths": {
                                                "/pets": {
                                                                "get": {}
                                                },
                                                "/pets/{petId}": {
                                                                "get": {},
                                                                "post": {},
                                                                "delete": {}
                                                }
                                }
                        }
selectorType: "host"
selector:
        api.pets.com: petstore
        api.users.com: userapi
        api.inline.com: inlineapi
cacheConfig:
        maxAPIs: 10
        maxPathsPerAPI: 1000
        pathExpiryTime: 24h
        apiExpiryTime: 72h
        minPathHits: 5
```

### Parameters

- `apis`: List of APIs to be loaded. `specFile` or `specText` must be specified for each API
        - `name`: Name of the API.
        - `specFile`: Path to the OpenAPI specification file.
        - `specText`: Inline OpenAPI specification text.
- `selectorType`: Type of selector to use for API selection. Possible values are `host`, `header`, `pathprefix`, and `fixed`.
- `selector`: Mapping for the selector type.
- `cacheConfig`: Configuration for caching API specifications.
        - `maxAPIs`: Maximum number of APIs to cache.
        - `maxPathsPerAPI`: Maximum number of paths per API to cache.
        - `pathExpiryTime`: Expiry time for cached paths.
        - `apiExpiryTime`: Expiry time for cached APIs.
        - `minPathHits`: Minimum number of hits for a path to be cached.

### Selectors

Selectors determine which API specification to use based on the incoming request. The following selector types are supported:

- `host`: Selects the API based on the request host.
- `header`: Selects the API based on a specific request header.
- `pathprefix`: Selects the API based on the request path prefix.
- `fixed`: Always selects a fixed API.

### Loading OpenAPI Specifications

OpenAPI specifications can be loaded from files or inline text. The middleware supports both JSON and YAML formats.

## Usage

To use the middleware, create a new instance and attach it to your HTTP server:

```go
package main

import (
        "net/http"
        "path/filepath"

        "github.com/lionelgarnier/validate-api-request/middleware"
)

func main() {
        // Load configuration from YAML file
        configPath := filepath.Join("config.yaml")
        config, err := middleware.LoadConfigFromFile(configPath)
        if err != nil {
                panic(err)
        }

        // Create your next handler (the final handler in the chain)
        nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte("Hello, World!"))
        })

        // Create the middleware
        middleware, err := middleware.New(nextHandler, config)
        if err != nil {
                panic(err)
        }

        // Create HTTP server
        http.Handle("/", middleware)
        http.ListenAndServe(":8080", nil)
}
```

## Testing

To test the middleware, you can use the provided test file (`middleware_test.go`):

```go
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
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

