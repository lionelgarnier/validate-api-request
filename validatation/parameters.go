package validation

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidateParameters validates the request parameters
func (v *DefaultValidator) ValidateParameters(req *http.Request, route string) (bool, error) {

	// Look for route & method in spec
	operation, err := v.GetOperation(route, req.Method)
	if err != nil {
		return false, err
	}

	parameters := operation.Parameters

	for _, param := range parameters {
		// Resolve parameter reference if necessary
		if param.Schema.Ref != "" {
			param, err = v.resolveParameterReference(param.Schema.Ref)
			if err != nil {
				return false, err
			}
		}

		var value string
		switch param.In {
		case "query":
			value = req.URL.Query().Get(param.Name)
		case "header":
			value = req.Header.Get(param.Name)
		case "path":
			value = extractPathParam(req.URL.Path, route, param.Name)
		case "cookie":
			cookie, err := req.Cookie(param.Name)
			if err != nil {
				return false, fmt.Errorf("missing cookie parameter '%s'", param.Name)
			}
			value = cookie.Value
		}

		if value == "" && param.Required {
			return false, fmt.Errorf("missing required parameter '%s'", param.Name)
		}

		if value != "" {
			if !v.validateSchema(value, *param.Schema) {
				return false, fmt.Errorf("invalid type for parameter '%s'", param.Name)
			}
		}
	}

	return true, nil
}

// extractPathParam extracts the value of a path parameter from the request path
func extractPathParam(requestPath, route, paramName string) string {
	routeParts := strings.Split(route, "/")
	pathParts := strings.Split(requestPath, "/")
	for i, part := range routeParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := part[1 : len(part)-1]
			if name == paramName {
				return pathParts[i]
			}
		}
	}
	return ""
}
