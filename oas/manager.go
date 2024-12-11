package oas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zeebo/xxh3"
)

// APISelector is a function that determines the API specification for a given request.
type Manager interface {
	LoadAPI(name string, content []byte) error
	GetApiSpec(name string) (*APISpec, error)
	CleanApiSpec()
	EvictApiSpec(name string)
	EvictAllApiSpecs()
	GetApiSpecs() map[string]*APISpec
}

// APISelector is a function that determines the API specification for a given request.
type OASManager struct {
	apiSpecs    map[string]*APISpec // Maps API name/version to context
	config      *CacheConfig
	apiSelector APISelector
	mu          sync.RWMutex
}

// APISelector is a function that determines the API specification for a given request.
type APISpec struct {
	openapi      string                // OpenAPI version
	info         json.RawMessage       // Info
	servers      []json.RawMessage     // Servers
	Paths        map[string]*PathCache // Hot paths
	Components   *ComponentCache       // Warm components
	Security     []SecurityRequirement // Security
	tags         []json.RawMessage     // Tags
	externalDocs json.RawMessage       // ExternalDocs
	hash         uint64                // Quick comparison
	LastAccess   time.Time
	HitCount     int64
}

// APISelector is a function that determines the API specification for a given request.
type PathCache struct {
	Item          *PathItem
	CompiledRegex *regexp.Regexp
	Route         string
	LastAccess    time.Time
	HitCount      int64
}

// APISelector is a function that determines the API specification for a given request.
type ComponentCache struct {
	Schemas         map[string]*Schema
	Responses       map[string]*Response
	Parameters      map[string]*Parameter
	Examples        map[string]*Example
	RequestBodies   map[string]*RequestBody
	Headers         map[string]*Header
	SecuritySchemes map[string]*SecurityScheme
	Links           map[string]*Link
	Callbacks       map[string]*Callback
}

type OASRequest struct {
	Request   *http.Request
	Route     string
	PathItem  *PathItem
	Operation *Operation
}

func NewOASRequest(r *http.Request) *OASRequest {
	return &OASRequest{Request: r}
}

// NewOASManager creates a new OAS manager with the given configuration and API selector.
func NewOASManager(config *CacheConfig, selector APISelector) *OASManager {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &OASManager{
		apiSpecs:    make(map[string]*APISpec),
		config:      config,
		apiSelector: selector,
		mu:          sync.RWMutex{},
	}
}

// GetApiSpecForRequest returns the API specification for the given request.
func (m *OASManager) GetApiSpecForRequest(r *http.Request) (*APISpec, error) {
	apiName := m.apiSelector(r)
	if apiName == "" {
		return nil, fmt.Errorf("could not determine API specification")
	}

	spec, err := m.GetApiSpec(apiName)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

// LoadAPI loads an API specification into the manager.
func (m *OASManager) LoadAPI(name string, content []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash := xxh3.HashString(string(content))

	// Check if API exists with same hash
	if existing, exists := m.apiSpecs[name]; exists {
		if existing.hash == hash {
			// Same content, skip loading
			return nil
		}
		// Different content, remove old spec and continue
		delete(m.apiSpecs, name)
	}

	// Parse initial structure
	var raw struct {
		Info         json.RawMessage       `json:"info"`
		OpenAPI      string                `json:"openapi"`
		Servers      []json.RawMessage     `json:"servers"`
		Security     []SecurityRequirement `json:"security"`
		Tags         []json.RawMessage     `json:"tags"`
		ExternalDocs json.RawMessage       `json:"externalDocs"`
	}

	if err := json.Unmarshal(content, &raw); err != nil {
		return fmt.Errorf("failed to parse OAS base structure: %v", err)
	}

	// Parse paths with minimal memory footprint
	paths, err := parsePathsFromRaw(content)
	if err != nil {
		return fmt.Errorf("failed to parse paths: %v", err)
	}

	// Initialize component cache
	components, err := parseComponentHeaders(content)
	if err != nil {
		return fmt.Errorf("failed to parse components: %v", err)
	}

	spec := &APISpec{
		info:         raw.Info,
		openapi:      raw.OpenAPI,
		Paths:        paths,
		Components:   components,
		servers:      raw.Servers,
		Security:     raw.Security,
		tags:         raw.Tags,
		externalDocs: raw.ExternalDocs,
		hash:         hash,
		LastAccess:   time.Now(),
		HitCount:     0,
	}

	m.apiSpecs[name] = spec
	return nil
}

// LoadAPIFromFile loads an API specification from a file into the manager.
func (m *OASManager) LoadAPIFromFile(name, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return m.LoadAPI(name, content)
}

// GetApiSpec returns the API specification for the given name.
func (m *OASManager) GetApiSpec(name string) (*APISpec, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	spec, exists := m.apiSpecs[name]
	if exists {
		spec.HitCount++
		spec.LastAccess = time.Now()
		return spec, nil
	}
	return nil, fmt.Errorf("API spec '%s' not found", name)
}

// EvictApiSpec removes the API specification with the given name.
func (m *OASManager) EvictApiSpec(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.apiSpecs, name)
}

// EvictAllApiSpecs removes all API specifications from the manager.
func (m *OASManager) EvictAllApiSpecs() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apiSpecs = make(map[string]*APISpec)
}

// GetApiSpecs returns all API specifications in the manager.
func (m *OASManager) GetApiSpecs() map[string]*APISpec {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.apiSpecs
}

// DefaultCacheConfig returns a default cache configuration.
func parsePathsFromRaw(content []byte) (map[string]*PathCache, error) {
	var raw struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}

	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, err
	}

	paths := make(map[string]*PathCache)
	for path, rawPath := range raw.Paths {
		var pathItem PathItem
		if err := json.Unmarshal(rawPath, &pathItem); err != nil {
			return nil, err
		}

		// Initialize PathCache
		pathCache := &PathCache{
			Item:     &pathItem,
			Route:    path,
			HitCount: 0,
		}

		// If the path contains parameters, compile the regex
		if strings.Contains(path, "{") && strings.Contains(path, "}") {
			regexPattern := pathTemplateToRegex(path)
			compiledRegex := regexp.MustCompile(regexPattern)
			pathCache.CompiledRegex = compiledRegex
		}

		paths[path] = pathCache
	}

	return paths, nil
}

// pathTemplateToRegex converts a path template to a regex pattern
func pathTemplateToRegex(pathTemplate string) string {
	// Replace path parameters with regex patterns
	regexPattern := regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllString(pathTemplate, `([^/]+)`)
	return "^" + regexPattern + "$"
}

// parseComponentHeaders parses the components section of an OAS document.
func parseComponentHeaders(content []byte) (*ComponentCache, error) {
	var raw struct {
		Components Components `json:"components"`
	}

	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse components: %v", err)
	}

	// Convert map[string]Schema to map[string]*Schema
	schemaPointers := make(map[string]*Schema)
	for key, schema := range raw.Components.Schemas {
		schemaCopy := schema // Create a copy to avoid issues with loop variable
		schemaPointers[key] = &schemaCopy
	}

	return &ComponentCache{
		Schemas:         mapToPointers(raw.Components.Schemas),
		Responses:       mapToPointers(raw.Components.Responses),
		Parameters:      mapToPointers(raw.Components.Parameters),
		Examples:        mapToPointers(raw.Components.Examples),
		RequestBodies:   mapToPointers(raw.Components.RequestBodies),
		Headers:         mapToPointers(raw.Components.Headers),
		SecuritySchemes: mapToPointers(raw.Components.SecuritySchemes),
		Links:           mapToPointers(raw.Components.Links),
		Callbacks:       mapToPointers(raw.Components.Callbacks),
	}, nil
}

// mapToPointers converts a map of values to a map of pointers.
func mapToPointers[T any](m map[string]T) map[string]*T {
	pointers := make(map[string]*T, len(m))
	for key, value := range m {
		valueCopy := value // Create copy to avoid loop variable issues
		pointers[key] = &valueCopy
	}
	return pointers
}
