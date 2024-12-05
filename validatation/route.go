package validation

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/lionelgarnier/validate-api-request/oas"
)

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ResolveRequestPath(req *http.Request) (string, error) {
	path := req.URL.Path

	var route string

	// Look for exact match
	if _, exists := v.spec.Paths[path]; exists {
		route = path
	} else {
		// Look for path template match
		for pathTemplate := range v.spec.Paths {
			regexPattern := pathTemplateToRegex(pathTemplate)
			matched := matchPath(path, regexPattern)
			if matched {
				route = pathTemplate
				break
			}
		}
	}
	if route == "" {
		return "", fmt.Errorf("no schema found for path '%s'", path)
	}
	return route, nil
}

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateRequestPath(req *http.Request, route string) (bool, error) {
	path := req.URL.Path

	// Look for route in spec
	if _, exists := v.spec.Paths[route]; !exists {
		return false, fmt.Errorf("path '%s' doesn't exist in oas schema", route)
	}

	// Look for exact match
	if path == route {
		return true, nil
	} else
	// Check that path doesn't exists without templating in spec
	if _, exists := v.spec.Paths[path]; exists {
		return false, fmt.Errorf("request path '%s' is not matching with oas schema path '%s' as there is another path with same name", path, route)
	} else
	// Look for path template match
	{
		regexPattern := pathTemplateToRegex(route)
		matched := matchPath(path, regexPattern)
		if matched {
			return true, nil
		}
	}

	return false, fmt.Errorf("request path '%s' is not matching with oas schema path '%s'", path, route)
}

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateRequestMethod(req *http.Request, route string) (bool, error) {
	method := req.Method
	var pathItem oas.PathItem

	// Look for exact match
	if item, exists := v.spec.Paths[route]; !exists {
		return false, fmt.Errorf("path '%s' doesn't exist in oas schema", route)
	} else {
		pathItem = item
	}

	// Check if method is allowed for path
	methodMap := map[string]*oas.Operation{
		http.MethodGet:     pathItem.Get,
		http.MethodPut:     pathItem.Put,
		http.MethodPost:    pathItem.Post,
		http.MethodDelete:  pathItem.Delete,
		http.MethodOptions: pathItem.Options,
		http.MethodHead:    pathItem.Head,
		http.MethodPatch:   pathItem.Patch,
		http.MethodTrace:   pathItem.Trace,
	}
	if operation := methodMap[method]; operation != nil {
		return true, nil
	}
	return false, fmt.Errorf("method '%s' not allowed for path '%s'", method, route)

}

// pathTemplateToRegex converts a path template to a regex pattern
func pathTemplateToRegex(pathTemplate string) string {
	// Replace path parameters with regex patterns
	regexPattern := regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllString(pathTemplate, `([^/]+)`)
	return "^" + regexPattern + "$"
}

// matchPath matches the request path against the regex pattern and extracts parameters
func matchPath(requestPath, regexPattern string) bool {
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(requestPath)
	return matches != nil
}

func (v *DefaultValidator) GetOperation(route, method string) (*oas.Operation, error) {
	if pathItem, exists := v.spec.Paths[route]; !exists {
		return nil, fmt.Errorf("path '%s' doesn't exist in oas schema", route)
	} else {

		// Check if method is allowed for path
		methodMap := map[string]*oas.Operation{
			http.MethodGet:     pathItem.Get,
			http.MethodPut:     pathItem.Put,
			http.MethodPost:    pathItem.Post,
			http.MethodDelete:  pathItem.Delete,
			http.MethodOptions: pathItem.Options,
			http.MethodHead:    pathItem.Head,
			http.MethodPatch:   pathItem.Patch,
			http.MethodTrace:   pathItem.Trace,
		}
		if operation := methodMap[method]; operation != nil {
			return operation, nil
		}
		return nil, fmt.Errorf("method '%s' not allowed for path '%s'", method, route)
	}
}
