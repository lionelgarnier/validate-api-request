package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/pkg/helpers"
)

// Validator defines the interface for request validation
type Validator interface {
	ValidateRequest(req *oas.OASRequest) (bool, error)
	ResolveRequestPath(req *oas.OASRequest) (*oas.PathCache, error)
	ValidateRequestPath(req *oas.OASRequest) (bool, error)
	ValidateRequestMethod(req *oas.OASRequest) (bool, error)
	ValidateParameters(req *oas.OASRequest) (bool, error)
	ValidateRequestBody(req *oas.OASRequest) (bool, error)
	ValidateSecurity(req *oas.OASRequest) (bool, error)
	ValidateSchema(value interface{}, schema *oas.Schema) bool
	SetApiSpec(apiSpec *oas.APISpec)
}

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	apiSpec *oas.APISpec
}

// NewValidator returns a new Validator
func NewValidator(apiSpec *oas.APISpec) Validator {
	return &DefaultValidator{
		apiSpec: apiSpec,
	}
}

// SetApiSpec sets the current API spec to validate against
func (v *DefaultValidator) SetApiSpec(apiSpec *oas.APISpec) {
	v.apiSpec = apiSpec
}

// ValidateRequest performs full request validation
func (v *DefaultValidator) ValidateRequest(req *oas.OASRequest) (bool, error) {

	if v.apiSpec == nil {
		return false, fmt.Errorf("no API spec selected, call SetCurrentAPI first")
	}

	if ok, err := v.ValidateRequestPath(req); !ok {
		return false, err
	}
	if ok, err := v.ValidateRequestMethod(req); !ok {
		return false, err
	}
	if ok, err := v.ValidateParameters(req); !ok {
		return false, err
	}
	if ok, err := v.ValidateRequestBody(req); !ok {
		return false, err
	}
	if ok, err := v.ValidateSecurity(req); !ok {
		return false, err
	}
	return true, nil
}

// ValidateSchema validates the request body against the schema
func (v *DefaultValidator) ValidateSchema(value interface{}, schema *oas.Schema) bool {
	// Handle discriminator first
	if schema.Discriminator != nil {
		resolvedSchema, err := v.resolveDiscriminator(value, schema)
		if err != nil {
			return false
		}
		return v.ValidateSchema(value, resolvedSchema)
	}

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
			schemaCopy := subSchema
			if !v.ValidateSchema(value, &schemaCopy) {
				return false
			}
		}
		return true
	}

	if schema.OneOf != nil {
		validCount := 0
		for _, subSchema := range schema.OneOf {
			schemaCopy := subSchema
			if v.ValidateSchema(value, &schemaCopy) {
				validCount++
			}
		}
		return validCount == 1
	}

	if schema.AnyOf != nil {
		for _, subSchema := range schema.AnyOf {
			schemaCopy := subSchema
			if v.ValidateSchema(value, &schemaCopy) {
				return true
			}
		}
		return false
	}

	return v.ValidateSchemaType(value, schema)
}

// GetRequestOperation returns the operation for a given request
func (v *DefaultValidator) GetRequestOperation(req *oas.OASRequest) (*oas.Operation, error) {
	pathCache, err := v.ResolveRequestPath(req)
	if err != nil {
		return nil, err
	}

	pathItem := pathCache.Item
	method := strings.ToUpper(req.Request.Method)

	// Look for route & method in spec
	operation := v.GetOperation(pathItem, method)
	if operation == nil {
		return nil, fmt.Errorf("method '%s' not allowed for path '%s'", method, pathCache.Route)
	}

	return operation, nil
}

// validateArray validates an array value against the schema
func (v *DefaultValidator) validateArray(value interface{}, schema *oas.Schema) bool {
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
			if !v.ValidateSchema(item, resolvedSchema) {
				return false
			}
		} else {
			if !v.ValidateSchema(item, schema.Items) {
				return false
			}
		}
	}
	return true
}

// validateObject validates an object value against the schema
func (v *DefaultValidator) validateObject(value interface{}, schema *oas.Schema) bool {
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
			if !v.ValidateSchema(propValue, resolvedSchema) {
				return false
			}
		} else {
			if !v.ValidateSchema(propValue, &propSchema) {
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
				if !v.ValidateSchema(obj[propName], additionalPropertiesSchema) {
					return false
				}
			}
		}
	}

	return true
}

// validateParameterType validates the parameter value against the expected type
func (v *DefaultValidator) ValidateSchemaType(value interface{}, paramSchema *oas.Schema) bool {

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
func (v *DefaultValidator) resolveSchemaReference(ref string) (*oas.Schema, error) {

	// Remove the "#/components/schemas/" prefix
	ref = strings.TrimPrefix(ref, "#/components/schemas/")

	// Check if Components or Schemas are nil
	if v.apiSpec.Components == nil || v.apiSpec.Components.Schemas == nil {
		return nil, fmt.Errorf("components or schemas not defined in OAS")
	}

	schema, exists := v.apiSpec.Components.Schemas[ref]
	if !exists {
		return nil, fmt.Errorf("schema reference '%s' not found", ref)
	}

	return schema, nil
}

// resolveParameterReference resolves a parameter reference to its actual definition
func (v *DefaultValidator) resolveParameterReference(ref string) (*oas.Parameter, error) {
	// Remove the "#/components/parameters/" prefix
	ref = strings.TrimPrefix(ref, "#/components/parameters/")

	// Check if Components or Parameters are nil
	if v.apiSpec.Components == nil || v.apiSpec.Components.Parameters == nil {
		return nil, fmt.Errorf("components or parameters not defined in oas")
	}

	param, exists := v.apiSpec.Components.Parameters[ref]
	if !exists {
		return nil, fmt.Errorf("parameter reference '%s' not found", ref)
	}
	return param, nil
}

// resolveDiscriminator resolves discriminator mapping and returns the correct schema
func (v *DefaultValidator) resolveDiscriminator(value interface{}, schema *oas.Schema) (*oas.Schema, error) {
	if schema.Discriminator == nil {
		return schema, nil
	}

	// Get object to check discriminator property
	obj, ok := value.(map[string]interface{})
	if !ok {
		return schema, fmt.Errorf("value must be object when using discriminator")
	}

	// Get discriminator value
	discriminatorValue, ok := obj[schema.Discriminator.PropertyName].(string)
	if !ok {
		return schema, fmt.Errorf("discriminator property '%s' not found or not string",
			schema.Discriminator.PropertyName)
	}

	// Check mapping
	var schemaRef string
	if len(schema.Discriminator.Mapping) > 0 {
		// Use explicit mapping
		if ref, ok := schema.Discriminator.Mapping[discriminatorValue]; ok {
			schemaRef = ref
		}
	} else {
		// Default mapping - append to current schema path
		schemaRef = "#/components/schemas/" + discriminatorValue
	}

	if schemaRef == "" {
		return schema, fmt.Errorf("no schema found for discriminator value '%s'",
			discriminatorValue)
	}

	resolvedSchema, err := v.resolveSchemaReference(schemaRef)
	if err != nil {
		return schema, fmt.Errorf("failed to resolve discriminator schema: %v", err)
	}

	return resolvedSchema, nil
}

// validateString validates a string value against the schema
func validateString(value interface{}, schema *oas.Schema) bool {
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
func validateNumber(value interface{}, schema *oas.Schema) bool {
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
