package oas

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/zeebo/xxh3"
)

type Manager interface {
	LoadAPI(name string, content []byte) error
	GetApiSpec(name string) (*APISpec, bool)
	CleanApiSpec()
	EvictApiSpec(name string)
	EvictAllApiSpecs()
	GetApiSpecs() map[string]*APISpec
}

type OASManager struct {
	apiSpecs map[string]*APISpec // Maps API name/version to context
	config   *CacheConfig
	mu       sync.RWMutex
}

type APISpec struct {
	openapi      string                // OpenAPI version
	info         json.RawMessage       // Info
	servers      []json.RawMessage     // Servers
	Paths        map[string]*PathCache // Hot paths (L1)
	Components   *ComponentCache       // Warm components (L2)
	security     []json.RawMessage     // Security
	tags         []json.RawMessage     // Tags
	externalDocs json.RawMessage       // ExternalDocs
	rawOAS       []byte                // Original OAS content
	hash         uint64                // Quick comparison
	LastAccess   time.Time
	HitCount     int64
}

type PathCache struct {
	Item     *PathItem
	Schemas  map[string]*Schema // Referenced schemas for this path
	HitCount int64
}

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

func NewOASManager(config *CacheConfig) *OASManager {
	if config == nil {
		config = &CacheConfig{
			MaxAPIs:     10,
			MinPathHits: 5,
		}
	}

	return &OASManager{
		apiSpecs: make(map[string]*APISpec),
		config:   config,
		mu:       sync.RWMutex{},
	}
}

func (m *OASManager) LoadAPI(name string, content []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash := xxh3.HashString(string(content))

	// Parse initial structure
	var raw struct {
		Info         json.RawMessage   `json:"info"`
		OpenAPI      string            `json:"openapi"`
		Servers      []json.RawMessage `json:"servers"`
		Security     []json.RawMessage `json:"security"`
		Tags         []json.RawMessage `json:"tags"`
		ExternalDocs json.RawMessage   `json:"externalDocs"`
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
		security:     raw.Security,
		tags:         raw.Tags,
		externalDocs: raw.ExternalDocs,
		rawOAS:       content,
		hash:         hash,
		LastAccess:   time.Now(),
		HitCount:     0,
	}

	m.apiSpecs[name] = spec
	return nil
}

func (m *OASManager) LoadAPIFromFile(name, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return m.LoadAPI(name, content)
}

func (m *OASManager) GetApiSpec(name string) (*APISpec, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	spec, exists := m.apiSpecs[name]
	return spec, exists
}

func (m *OASManager) CleanApiSpec() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove least used APIs
	for name, spec := range m.apiSpecs {
		if time.Since(spec.LastAccess) > m.config.APIExpiryTime {
			delete(m.apiSpecs, name)
		}
	}

	// Clear cold paths
	for _, spec := range m.apiSpecs {
		for path, cache := range spec.Paths {
			if cache.HitCount < m.config.MinPathHits {
				delete(spec.Paths, path)
			}
		}
	}
}

func (m *OASManager) EvictApiSpec(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.apiSpecs, name)
}

func (m *OASManager) EvictAllApiSpecs() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apiSpecs = make(map[string]*APISpec)
}

func (m *OASManager) GetApiSpecs() map[string]*APISpec {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.apiSpecs
}

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

		paths[path] = &PathCache{
			Item:     &pathItem,
			Schemas:  make(map[string]*Schema),
			HitCount: 0,
		}
	}

	return paths, nil
}

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

func mapToPointers[T any](m map[string]T) map[string]*T {
	pointers := make(map[string]*T, len(m))
	for key, value := range m {
		valueCopy := value // Create copy to avoid loop variable issues
		pointers[key] = &valueCopy
	}
	return pointers
}
