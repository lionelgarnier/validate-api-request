package validation

import (
	"encoding/json"
	"fmt"

	"github.com/lionelgarnier/validate-api-request/oas"
)

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateRequestBody(req *oas.OASRequest) (bool, error) {
	if req.PathItem == nil || req.Route == "" || req.Operation == nil {
		_, err := v.ValidateRequestMethod(req)
		if err != nil {
			return false, err
		}
	}

	operation := req.Operation

	requestBody := operation.RequestBody
	// Skip validation if no request body defined
	if requestBody == nil {
		return true, nil
	}

	// Check if request body is required
	if requestBody.Required && req.Request.ContentLength == 0 {
		return false, fmt.Errorf("request body is required")
	}

	// Get content type from request
	contentType := req.Request.Header.Get("Content-Type")
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
	if err := json.NewDecoder(req.Request.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("invalid request body: %v", err)
	}

	// Validate request body against schema
	if !v.ValidateSchema(body, mediaType.Schema) {
		return false, fmt.Errorf("request body does not match schema")
	}

	return true, nil
}
