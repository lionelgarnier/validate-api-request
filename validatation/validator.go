package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/pkg/helpers"
)

// Validator defines the interface for request validation
type Validator interface {
	ValidateRequest(req *http.Request) (bool, error)
	ResolveRequestPath(req *http.Request) (string, error)
	ValidateRequestPath(req *http.Request, route string) (bool, error)
	ValidateRequestMethod(req *http.Request, route string) (bool, error)
	ValidateParameters(req *http.Request, route string) (bool, error)
	ValidateRequestBody(req *http.Request, route string) (bool, error)
	ValidateSecurity(req *http.Request, route string) (bool, error)
}

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	spec *oas.OpenAPI
}

// NewValidator returns a new Validator
func NewValidator(spec oas.OpenAPI) Validator {
	return &DefaultValidator{
		spec: &spec,
	}
}

// ValidateRequest performs full request validation
func (v *DefaultValidator) ValidateRequest(req *http.Request) (bool, error) {

	route, err := v.ResolveRequestPath(req)
	if route == "" {
		return false, err
	}
	/* Not required as resolved paths should always be valid
	if ok, err := v.ValidateRequestPath(req, route); !ok {
		return false, err
	}*/
	if ok, err := v.ValidateRequestMethod(req, route); !ok {
		return false, err
	}
	if ok, err := v.ValidateParameters(req, route); !ok {
		return false, err
	}
	if ok, err := v.ValidateRequestBody(req, route); !ok {
		return false, err
	}
	if ok, err := v.ValidateSecurity(req, route); !ok {
		return false, err
	}
	return true, nil
}

// validateSchema validates the request body against the schema
func (v *DefaultValidator) validateSchema(value interface{}, schema oas.Schema) bool {
	// Resolve the schema reference if necessary
	if schema.Ref != "" {
		resolvedSchema, err := v.resolveSchemaReference(schema.Ref)
		if err != nil {
			return false
		}
		schema = resolvedSchema
	}

	if schema.AllOf != nil {
		for _, subSchema := range schema.AllOf {
			if !v.validateSchema(value, subSchema) {
				return false
			}
		}
		return true
	}

	if schema.OneOf != nil {
		validCount := 0
		for _, subSchema := range schema.OneOf {
			if v.validateSchema(value, subSchema) {
				validCount++
			}
		}
		return validCount == 1
	}

	if schema.AnyOf != nil {
		for _, subSchema := range schema.AnyOf {
			if v.validateSchema(value, subSchema) {
				return true
			}
		}
		return false
	}

	return v.validateSchemaType(value, schema)
}

// validateArray validates an array value against the schema
func (v *DefaultValidator) validateArray(value interface{}, schema oas.Schema) bool {
	// Resolve the schema reference if necessary
	if schema.Ref != "" {
		resolvedSchema, err := v.resolveSchemaReference(schema.Ref)
		if err != nil {
			return false
		}
		schema = resolvedSchema
	}

	var arr []interface{}

	// Check if the value is a JSON string representation of an array
	if str, ok := value.(string); ok {
		if err := json.Unmarshal([]byte(str), &arr); err != nil {
			// If unmarshalling fails, create an array with a single item
			arr = []interface{}{str}
		}
	} else {
		// Otherwise, assert it as a slice of interfaces
		arr, ok = value.([]interface{})
		if !ok {
			return false
		}
	}

	if schema.MinItems != nil && uint64(len(arr)) < *schema.MinItems {
		return false
	}
	if schema.MaxItems != nil && uint64(len(arr)) > *schema.MaxItems {
		return false
	}
	if schema.UniqueItems {
		if !helpers.UniqueItems(arr) {
			return false
		}
	}

	for _, item := range arr {
		if schema.Items.Ref != "" {
			resolvedSchema, err := v.resolveSchemaReference(schema.Items.Ref)
			if err != nil {
				return false
			}
			if !v.validateSchema(item, resolvedSchema) {
				return false
			}
		} else {
			if !v.validateSchema(item, *schema.Items) {
				return false
			}
		}
	}
	return true
}

// validateObject validates an object value against the schema
func (v *DefaultValidator) validateObject(value interface{}, schema oas.Schema) bool {
	// Resolve the schema reference if necessary
	if schema.Ref != "" {
		resolvedSchema, err := v.resolveSchemaReference(schema.Ref)
		if err != nil {
			return false
		}
		schema = resolvedSchema
	}

	var obj map[string]interface{}

	// Check if the value is a JSON string representation of an array
	if str, ok := value.(string); ok {
		if err := json.Unmarshal([]byte(str), &obj); err != nil {
			obj = map[string]interface{}{"value": str}
		}
	} else {
		// Otherwise, assert it as a slice of interfaces
		obj, ok = value.(map[string]interface{})
		if !ok {
			return false
		}
	}

	for propName, propSchema := range schema.Properties {
		propValue, exists := obj[propName]
		if !exists {
			if helpers.Contains(schema.Required, propName) {
				return false
			}
			continue
		}

		if propSchema.Ref != "" {
			resolvedSchema, err := v.resolveSchemaReference(propSchema.Ref)
			if err != nil {
				return false
			}
			if !v.validateSchema(propValue, resolvedSchema) {
				return false
			}
		} else {
			if !v.validateSchema(propValue, propSchema) {
				return false
			}
		}
	}

	if schema.AdditionalProperties != nil {
		for propName := range obj {
			if _, exists := schema.Properties[propName]; !exists {
				additionalPropertiesSchema, ok := schema.AdditionalProperties.(*oas.Schema)
				if !ok {
					return false
				}
				if !v.validateSchema(obj[propName], *additionalPropertiesSchema) {
					return false
				}
			}
		}
	}

	return true
}

// validateParameterType validates the parameter value against the expected type
func (v *DefaultValidator) validateSchemaType(value interface{}, paramSchema oas.Schema) bool {

	switch paramSchema.Type {
	case "string":
		return validateString(value, paramSchema)
	case "integer", "number":
		return validateNumber(value, paramSchema)
	case "boolean":
		return helpers.IsBoolean(value)
	case "array":
		return v.validateArray(value, paramSchema)
	case "object", "":
		return v.validateObject(value, paramSchema)
	default:
		return false
	}
}

// resolveSchemaReference resolves a schema reference to its actual definition
func (v *DefaultValidator) resolveSchemaReference(ref string) (oas.Schema, error) {
	// Remove the "#/components/schemas/" prefix
	ref = strings.TrimPrefix(ref, "#/components/schemas/")

	// Check if Components or Schemas are nil
	if v.spec.Components == nil || v.spec.Components.Schemas == nil {
		return oas.Schema{}, fmt.Errorf("components or schemas not defined in OAS")
	}

	schema, exists := v.spec.Components.Schemas[ref]
	if !exists {
		return oas.Schema{}, fmt.Errorf("schema reference '%s' not found", ref)
	}
	return schema, nil
}

// resolveParameterReference resolves a parameter reference to its actual definition
func (v *DefaultValidator) resolveParameterReference(ref string) (oas.Parameter, error) {
	// Remove the "#/components/parameters/" prefix
	ref = strings.TrimPrefix(ref, "#/components/parameters/")

	// Check if Components or Parameters are nil
	if v.spec.Components == nil || v.spec.Components.Parameters == nil {
		return oas.Parameter{}, fmt.Errorf("components or parameters not defined in oas")
	}

	param, exists := v.spec.Components.Parameters[ref]
	if !exists {
		return oas.Parameter{}, fmt.Errorf("parameter reference '%s' not found", ref)
	}
	return param, nil
}

// validateString validates a string value against the schema
func validateString(value interface{}, schema oas.Schema) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}

	if schema.MinLength != nil && uint64(len(str)) < *schema.MinLength {
		return false
	}
	if schema.MaxLength != nil && uint64(len(str)) > *schema.MaxLength {
		return false
	}
	if schema.Pattern != "" {
		if !helpers.MatchPattern(str, schema.Pattern) {
			return false
		}
	}
	if schema.Enum != nil {
		enumStrings := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			enumStrings[i], ok = v.(string)
			if !ok {
				return false
			}
		}
		if !helpers.Contains(enumStrings, str) {
			return false
		}
	}

	switch schema.Format {
	case "uuid":
		return helpers.IsUUID(value)
	case "email":
		return helpers.IsEmail(value)
	case "url", "uri":
		return helpers.IsURL(value)
	case "hostname":
		return helpers.IsHostnameValid(value)
	case "ipv4":
		return helpers.IsIPv4(value)
	case "ipv6":
		return helpers.IsIPv6(value)
	case "byte":
		return helpers.IsByte(value)
	case "date", "date-time":
		return helpers.IsISO8601(value)
	default:
		return true
	}

}

// validateNumber validates a numeric value against the schema
func validateNumber(value interface{}, schema oas.Schema) bool {
	// Try to convert string to number if needed
	if str, ok := value.(string); ok {
		parsed, err := helpers.ParseNumber(str)
		if err != nil {
			return false
		}
		value = parsed
	}

	num, ok := value.(float64)
	if !ok {
		return false
	}

	if schema.Minimum != nil && num < *schema.Minimum {
		return false
	}
	if schema.Maximum != nil && num > *schema.Maximum {
		return false
	}
	if schema.MultipleOf != nil && int(num)%int(*schema.MultipleOf) != 0 {
		return false
	}

	switch schema.Type {
	case "integer":
		switch schema.Format {
		case "int32":
			return helpers.IsInt32(value)
		default:
			return helpers.IsInt64(value)
		}
	case "number":
		switch schema.Format {
		case "double":
			return helpers.IsDouble(value)
		default:
			return helpers.IsFloat(value)
		}
	default:
		return false
	}
}
