package oas

import (
	"time"
)

type Duration struct {
	time.Duration
}

type CacheConfig struct {
	MaxAPIs         int      `yaml:"maxAPIs" json:"maxAPIs"`
	MaxPathsPerAPI  int      `yaml:"maxPathsPerAPI" json:"maxPathsPerAPI"`
	PathExpiryTime  Duration `yaml:"pathExpiryTime" json:"pathExpiryTime"`
	APIExpiryTime   Duration `yaml:"apiExpiryTime" json:"apiExpiryTime"`
	MinPathHits     int64    `yaml:"minPathHits" json:"minPathHits"`
	CleanupInterval Duration `yaml:"cleanupInterval" json:"cleanupInterval"`
}

// DefaultCacheConfig returns a default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxAPIs:         100,
		MaxPathsPerAPI:  1000,
		PathExpiryTime:  Duration{24 * time.Hour},
		APIExpiryTime:   Duration{72 * time.Hour},
		MinPathHits:     10,
		CleanupInterval: Duration{time.Hour},
	}
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}
