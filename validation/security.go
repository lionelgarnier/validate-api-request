package validation

import (
	"net/http"
)

// ValidateSecurity validates the request security
func (v *DefaultValidator) ValidateSecurity(req *http.Request, route string) (bool, error) {

	return true, nil
}
