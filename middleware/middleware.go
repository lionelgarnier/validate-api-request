package middleware

import (
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/lionelgarnier/validate-api-request/oas"
	"github.com/lionelgarnier/validate-api-request/validation"
)

// APIConfig represents the configuration for an API
type APIConfig struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	SpecFile string `json:"specFile,omitempty" yaml:"specFile,omitempty"`
	SpecText string `json:"specText,omitempty" yaml:"specText,omitempty"`
}

// Config represents the configuration for the OAS middleware
type Config struct {
	APIs         []APIConfig       `json:"apis,omitempty" yaml:"apis,omitempty"`
	SelectorType string            `json:"selectorType,omitempty" yaml:"selectorType,omitempty"`
	Selector     map[string]string `json:"selector,omitempty" yaml:"selector,omitempty"`
	CacheConfig  *oas.CacheConfig  `json:"cacheConfig,omitempty" yaml:"cacheConfig,omitempty"`
}

// CreateConfig creates a new Config with default values
func CreateConfig() *Config {
	return &Config{
		APIs:        []APIConfig{},
		Selector:    map[string]string{},
		CacheConfig: oas.DefaultCacheConfig(),
	}
}

// OASMiddleware validates requests against OpenAPI specs
type OASMiddleware struct {
	next      http.Handler
	manager   *oas.OASManager
	validator validation.Validator
}

// NewMiddleware creates a new OASMiddleware
func New(next http.Handler, config *Config) (*OASMiddleware, error) {
	// Create API selector based on the configuration
	var selector oas.APISelector
	switch config.SelectorType {
	case "host":
		selector = oas.HostSelector(config.Selector)
	case "header":
		selector = oas.HeaderSelector(config.Selector)
	case "pathprefix":
		selector = oas.PathPrefixSelector(config.Selector)
	case "fixed":
		selector = oas.FixedSelector(config.Selector)
	default:
		return nil, fmt.Errorf("unknown selector type '%s'", config.SelectorType)
	}

	// Create OAS manager with cache config and selector
	manager := oas.NewOASManager(config.CacheConfig, selector)

	// Load APIs from the configuration
	for _, apiConfig := range config.APIs {
		if apiConfig.SpecFile != "" {
			// Load from file
			if err := manager.LoadAPIFromFile(apiConfig.Name, apiConfig.SpecFile); err != nil {
				return nil, fmt.Errorf("failed to load OAS file '%s': %w", apiConfig.SpecFile, err)
			}
		} else if apiConfig.SpecText != "" {
			// Load from text
			if err := manager.LoadAPI(apiConfig.Name, []byte(apiConfig.SpecText)); err != nil {
				return nil, fmt.Errorf("failed to load OAS text for API '%s': %w", apiConfig.Name, err)
			}
		} else {
			return nil, fmt.Errorf("API '%s' must have either specFile or specText", apiConfig.Name)
		}
	}

	// Create validator
	validator := validation.NewValidator(nil)

	return &OASMiddleware{
		next:      next,
		manager:   manager,
		validator: validator,
	}, nil
}

// ServeHTTP validates the request against the OpenAPI spec
func (m *OASMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get API spec for request
	spec, err := m.manager.GetApiSpecForRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set spec in validator
	m.validator.SetApiSpec(spec)

	oasRequest := oas.NewOASRequest(r)

	// Validate request
	if ok, err := m.validator.ValidateRequest(oasRequest); !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call next handler
	m.next.ServeHTTP(w, r)
}

func LoadConfigFromFile(configPath string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create an instance of Config
	config := &Config{}

	// Unmarshal YAML into the Config struct
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
