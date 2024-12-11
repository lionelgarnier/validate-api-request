package oas

// Base OAS structure
type OpenAPI struct {
	OpenAPI      string                `json:"openapi" yaml:"openapi"`
	Info         Info                  `json:"info" yaml:"info"`
	Servers      []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        map[string]PathItem   `json:"paths" yaml:"paths"`
	Components   *Components           `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Components is a container for reusable schemas and responses.
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
}

// Components is a container for reusable schemas and responses.
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

// PathItem is a list of operations that can be performed on a path.
type PathItem struct {
	Ref         string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation  `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation  `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation  `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation  `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation  `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation  `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation  `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []Server    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation is a single API operation on a path.
type Operation struct {
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs          `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationId  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []Parameter            `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    map[string]Response    `json:"responses" yaml:"responses"`
	Callbacks    map[string]Callback    `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Extensions   map[string]interface{} `json:"-" yaml:"-"`
}

// Schema is a JSON Schema object.
type Schema struct {
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Properties           map[string]Schema      `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *Schema                `json:"items,omitempty" yaml:"items,omitempty"`
	Required             []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty" yaml:"enum,omitempty"`
	AllOf                []Schema               `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []Schema               `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []Schema               `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty" yaml:"not,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	MultipleOf           *float64               `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	MinLength            *uint64                `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            *uint64                `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinItems             *uint64                `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems             *uint64                `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	Example              interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Deprecated           bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Discriminator        *Discriminator         `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	XML                  *XML                   `json:"xml,omitempty" yaml:"xml,omitempty"`
	Title                string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	ExclusiveMaximum     bool                   `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum     bool                   `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MinProperties        uint64                 `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	MaxProperties        uint64                 `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                   `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Extensions           map[string]interface{} `json:"-" yaml:"-"`
}

// Server is a URL to the target host.
type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable is a server variable for URL template substitution.
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Parameter is a list of parameters that can be used across operations.
type Parameter struct {
	Name            string               `json:"name" yaml:"name"`
	In              string               `json:"in" yaml:"in"`
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// RequestBody is a request body object that can be passed to an operation.
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// MediaType is a media type object that can be referenced by responses and request bodies.
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Encoding is a single encoding definition applied to a single schema property.
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string            `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Response is a list of possible responses as they are returned from executing this operation.
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// Callback is a map of possible out-of band callbacks related to the parent operation.
type Callback map[string]PathItem

// Example is a "example" object that provides an example of the schema.
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Link represents a possible design-time link for a response.
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationId  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
}

// Header is a list of headers that can be used across operations.
type Header struct {
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Tag is a list of tags used by the specification with additional metadata.
type Tag struct {
	Name         string        `json:"name" yaml:"name"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// ExternalDocs is a list of external documentation.
type ExternalDocs struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License information for the exposed API.
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

type SecurityRequirement map[string][]string

// XML is a metadata object that allows for more fine-tuned XML model definitions.
type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
}

// Discriminator adds support for polymorphism.
type Discriminator struct {
	PropertyName string            `json:"propertyName" yaml:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

// OAuth2 configuration
type SecurityScheme struct {
	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIdConnectUrl string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// OAuthFlows configuration
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow configuration
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl" yaml:"authorizationUrl"`
	TokenURL         string            `json:"tokenUrl" yaml:"tokenUrl"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}
