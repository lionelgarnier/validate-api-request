package validation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lionelgarnier/validate-api-request/oas"
)

// ValidateRequestPath validates the request path
func (v *DefaultValidator) ValidateSecurity(req *oas.OASRequest) (bool, error) {
	if req.PathItem == nil || req.Route == "" || req.Operation == nil {
		_, err := v.ValidateRequestMethod(req)
		if err != nil {
			return false, err
		}
	}

	operation := req.Operation

	// Get security requirements (operation-level or global)
	securityRequirements := operation.Security
	if securityRequirements == nil {
		securityRequirements = v.apiSpec.Security
	}

	if len(securityRequirements) == 0 {
		// No security requirements; request is valid
		return true, nil
	}

	// Check if the request satisfies at least one security requirement
	for _, secReq := range securityRequirements {
		if v.validateSecurityRequirement(req.Request, secReq) {
			// At least one requirement satisfied
			return true, nil
		}
	}

	return false, fmt.Errorf("request does not satisfy any security requirements")
}

func (v *DefaultValidator) validateSecurityRequirement(r *http.Request, secReq map[string][]string) bool {
	for secSchemeName := range secReq {
		secScheme, exists := v.apiSpec.Components.SecuritySchemes[secSchemeName]
		if !exists {
			// Security scheme not defined
			return false
		}

		switch secScheme.Type {
		case "apiKey":
			if !v.validateAPIKeySecurity(r, secScheme) {
				return false
			}
		case "http":
			if !v.validateHTTPSecurity(r, secScheme) {
				return false
			}
		case "oauth2":
			if !v.validateOAuth2Security(r) {
				return false
			}
		case "openIdConnect":
			if !v.validateOpenIdConnectSecurity(r) {
				return false
			}
		default:
			return false
		}
	}
	// All security schemes in this requirement are satisfied
	return true
}

func (v *DefaultValidator) validateAPIKeySecurity(r *http.Request, secScheme *oas.SecurityScheme) bool {
	var value string
	switch secScheme.In {
	case "header":
		value = r.Header.Get(secScheme.Name)
	case "query":
		value = r.URL.Query().Get(secScheme.Name)
	case "cookie":
		cookie, err := r.Cookie(secScheme.Name)
		if err == nil {
			value = cookie.Value
		}
	}
	return value != ""
}

func (v *DefaultValidator) validateHTTPSecurity(r *http.Request, secScheme *oas.SecurityScheme) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	scheme := strings.ToLower(secScheme.Scheme)
	switch scheme {
	case "basic":
		return strings.HasPrefix(authHeader, "Basic ")
	case "bearer":
		return strings.HasPrefix(authHeader, "Bearer ")
	case "digest":
		return strings.HasPrefix(authHeader, "Digest ")
	case "apikey":
		return strings.HasPrefix(authHeader, "ApiKey ")
	default:
		return false
	}
}

func (v *DefaultValidator) validateOAuth2Security(r *http.Request) bool {
	// Check for access token in Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}
	accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	return accessToken != ""
}

func (v *DefaultValidator) validateOpenIdConnectSecurity(r *http.Request) bool {
	// Check for ID token in Authorization header or specific parameter
	authHeader := r.Header.Get("Authorization")
	var idToken string
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		idToken = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	} else {
		// Alternatively, check for token in a query parameter or cookie
		idToken = r.URL.Query().Get("id_token")
	}
	return idToken != ""
}
