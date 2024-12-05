package oas

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser interface for parsing OpenAPI specifications
type Parser interface {
	Parse(oasText string) (map[string]*Schema, error)
}

// DefaultParser implements the Parser interface
type DefaultParser struct {
	oasText []byte
	isYAML  bool
}

func NewParser() *DefaultParser {
	return &DefaultParser{
		oasText: nil,
		isYAML:  false,
	}
}

// LoaFromFile reads and parses an OpenAPI specification from a file
func (p *DefaultParser) LoadFromFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}

	p.oasText = content
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

	return &openAPI, nil
}
