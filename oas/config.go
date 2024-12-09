package oas

import (
	"time"
)

type CacheConfig struct {
	MaxAPIs         int           `json:"maxAPIs"`         // Maximum number of APIs to cache
	MaxPathsPerAPI  int           `json:"maxPathsPerAPI"`  // Maximum paths per API
	PathExpiryTime  time.Duration `json:"pathExpiryTime"`  // Time before path cache expires
	APIExpiryTime   time.Duration `json:"apiExpiryTime"`   // Time before API cache expires
	MinPathHits     int64         `json:"minPathHits"`     // Minimum hits to keep path cached
	CleanupInterval time.Duration `json:"cleanupInterval"` // How often to run cleanup
}

func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxAPIs:         100,
		MaxPathsPerAPI:  1000,
		PathExpiryTime:  24 * time.Hour,
		APIExpiryTime:   72 * time.Hour,
		MinPathHits:     10,
		CleanupInterval: time.Hour,
	}
}
