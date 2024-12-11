package validation

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lionelgarnier/validate-api-request/oas"
)

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ResolveRequestPath(req *oas.OASRequest) (*oas.PathCache, error) {

	var pathCache *oas.PathCache
	var exists bool
	var path string

	if req.Route != "" {
		path = req.Route
	} else {
		path = req.Request.URL.Path
	}

	// Look for exact match
	pathCache, exists = v.apiSpec.Paths[path]
	if !exists {
		// Iterate over paths with precompiled regex
		for _, pathItem := range v.apiSpec.Paths {
			if pathItem.CompiledRegex != nil && pathItem.CompiledRegex.MatchString(path) {
				pathCache = pathItem
				break
			}
		}
	}

	if pathCache == nil {
		return nil, fmt.Errorf("no schema found for path '%s'", path)
	}

	// Update cache stats
	pathCache.HitCount++
	pathCache.LastAccess = time.Now()

	// Set route in request
	req.Route = pathCache.Route
	req.PathItem = pathCache.Item
	return pathCache, nil

}

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateRequestPath(req *oas.OASRequest) (bool, error) {
	_, err := v.ResolveRequestPath(req)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateRequestMethod(req *oas.OASRequest) (bool, error) {

	if req.PathItem == nil || req.Route == "" {
		_, err := v.ResolveRequestPath(req)
		if err != nil {
			return false, err
		}
	}

	method := strings.ToUpper(req.Request.Method)
	pathItem := req.PathItem
	route := req.Route

	operation := v.GetOperation(pathItem, method)
	if operation == nil {
		return false, fmt.Errorf("method '%s' not allowed for path '%s'", method, route)
	}

	req.Operation = operation
	return true, nil

}

// GetOperation returns the operation for a given route and method
func (v *DefaultValidator) GetOperation(pathItem *oas.PathItem, method string) *oas.Operation {

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
		return operation
	}
	return nil
}
