package validation

import (
	"fmt"
	"strings"

	"github.com/lionelgarnier/validate-api-request/oas"
)

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateParameters(req *oas.OASRequest) (bool, error) {
	pathCache, err := v.ResolveRequestPath(req)
	if err != nil {
		return false, err
	}
	return v.ValidateParametersForPath(req, pathCache)
}

// ValidateParameters validates the request parameters for a given pathCache
func (v *DefaultValidator) ValidateParametersForPath(req *oas.OASRequest, pathCache *oas.PathCache) (bool, error) {
	pathItem := pathCache.Item
	method := strings.ToUpper(req.Request.Method)

	// Look for route & method in spec
	operation := v.GetOperation(pathItem, method)
	if operation == nil {
		return false, fmt.Errorf("method '%s' not allowed for path '%s'", method, pathCache.Route)
	}

	parameters := mergeParameters(pathItem.Parameters, operation.Parameters)
	var err error

	for i := range parameters {
		param := &parameters[i]
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
			value = req.Request.URL.Query().Get(param.Name)
		case "header":
			value = req.Request.Header.Get(param.Name)
		case "path":
			value = extractPathParam(req.Request.URL.Path, pathCache, param.Name)
		case "cookie":
			cookie, err := req.Request.Cookie(param.Name)
			if err != nil {
				return false, fmt.Errorf("missing cookie parameter '%s'", param.Name)
			}
			value = cookie.Value
		}

		if value == "" && param.Required {
			return false, fmt.Errorf("missing required parameter '%s'", param.Name)
		}

		if value != "" {
			if !v.ValidateSchema(value, param.Schema) {
				return false, fmt.Errorf("invalid type for parameter '%s'", param.Name)
			}
		}
	}

	return true, nil
}

func mergeParameters(pathParams, opParams []oas.Parameter) []oas.Parameter {
	paramMap := make(map[string]oas.Parameter)

	// Add PathItem parameters to the map
	for _, param := range pathParams {
		key := param.In + ":" + param.Name
		paramMap[key] = param
	}

	// Add Operation parameters, overriding if necessary
	for _, param := range opParams {
		key := param.In + ":" + param.Name
		paramMap[key] = param // Operation-level param overrides path-level
	}

	// Convert map back to slice
	mergedParams := make([]oas.Parameter, 0, len(paramMap))
	for _, param := range paramMap {
		mergedParams = append(mergedParams, param)
	}
	return mergedParams
}

// extractPathParam extracts the value of a path parameter from the request path
func extractPathParam(requestPath string, pathCache *oas.PathCache, paramName string) string {
	routeParts := strings.Split(pathCache.Route, "/")
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
