package oas

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lionelgarnier/validate-api-request/cache"
	"github.com/zeebo/xxh3"
	"gopkg.in/yaml.v3"
)

// Parser interface for parsing OpenAPI specifications
type Parser interface {
	LoadOasFromFile(filePath string) (bool, error)
	Parse(oasText string) (map[string]*Schema, error)
	SetOASText(oasText []byte)
	GetSchema() (*OpenAPI, error)
}

// DefaultParser implements the Parser interface
type DefaultParser struct {
	oasText []byte
	oasHash uint64
	isYAML  bool
	cache   *paserCache
}

type paserCache struct {
	*cache.BaseCache[*OpenAPI]
}

func NewParser() *DefaultParser {
	return &DefaultParser{
		oasText: nil,
		oasHash: 0,
		isYAML:  false,
		cache:   newPaserCache(1000, 1000),
	}
}

func newPaserCache(maxSize int, ttl time.Duration) *paserCache {
	return &paserCache{
		BaseCache: cache.NewBaseCache[*OpenAPI](maxSize, ttl),
	}
}

// LoaFromFile reads and parses an OpenAPI specification from a file
func (p *DefaultParser) LoadOasFromFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}

	p.SetOASText(content)
	p.isYAML = strings.HasSuffix(strings.ToLower(filePath), ".yaml") ||
		strings.HasSuffix(strings.ToLower(filePath), ".yml")

	return true, nil
}

// Parse converts OpenAPI spec text to schemas
func (p *DefaultParser) Parse() (*OpenAPI, error) {

	var openAPI OpenAPI
	var err error

	trimmedText := strings.TrimSpace(string(p.oasText))
	if len(trimmedText) > 0 && trimmedText[0] == '{' {
		// Try to unmarshal as JSON
		if err = json.Unmarshal(p.oasText, &openAPI); err != nil {
			return nil, fmt.Errorf("failed to parse OAS as JSON: %v", err)
		}
	} else {
		// Try to unmarshal as YAML
		if err = yaml.Unmarshal(p.oasText, &openAPI); err != nil {
			return nil, fmt.Errorf("failed to parse OAS as YAML: %v", err)
		}
	}

	if err != nil {
		return nil, err
	}

	// Cache the result
	p.cache.Set(p.oasHash, &openAPI)

	return &openAPI, nil
}

// Set OAS text
func (p *DefaultParser) SetOASText(oasText []byte) {
	p.oasText = oasText
	p.oasHash = xxh3.HashString(string(oasText))
}

// Get OAS schema
func (p *DefaultParser) GetSchema() (*OpenAPI, error) {

	// Check cache first
	if result, found := p.cache.Get(p.oasHash); found {
		return result, nil
	}

	return p.Parse()
}
