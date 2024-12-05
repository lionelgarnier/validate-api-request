package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ValidateRequestBody validates the request body
func (v *DefaultValidator) ValidateRequestBody(req *http.Request, route string) (bool, error) {

	// Look for route & method in spec
	operation, err := v.GetOperation(route, req.Method)
	if err != nil {
		return false, err
	}

	requestBody := operation.RequestBody
	// Skip validation if no request body defined
	if requestBody == nil {
		return true, nil
	}

	// Check if request body is required
	if requestBody.Required && req.ContentLength == 0 {
		return false, fmt.Errorf("request body is required")
	}

	// Get content type from request
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json" // Default to JSON if not specified
	}

	// Check if content type is supported
	mediaType, exists := requestBody.Content[contentType]
	if !exists {
		return false, fmt.Errorf("unsupported content type '%s'", contentType)
	}

	// Skip validation if no schema defined
	if mediaType.Schema == nil {
		return true, nil
	}

	// Parse request body
	var body interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("invalid request body: %v", err)
	}

	// Validate request body against schema
	if !v.validateSchema(body, *mediaType.Schema) {
		return false, fmt.Errorf("request body does not match schema")
	}

	return true, nil
}
