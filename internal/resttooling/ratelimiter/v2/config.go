package ratelimiter

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"

	"github.com/yanakipre/bot/internal/rate"
)

// Mode is the mode the rate limit manager is in.
type Mode string

const (
	ModeReporting Mode = "reporting"
	ModeEnforcing Mode = "enforcing"
)

// Config is the configuration for the rate limit manager.
type Config struct {
	// Disabling the rate limiter turns off tracking of requests all together, regardless of the `mode` set.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Either "reporting" or "enforcing"
	// - "reporting" keeps track of requests, doesn't enforce any limits, but increases metric counters
	// - "enforcing" keeps track of requests and enforces limits, denying requests that exceed them
	Mode Mode `yaml:"mode" json:"mode"`

	// Path patterns to apply rate limits to
	Paths map[string][]rate.WindowConfig `yaml:"paths" json:"paths"`
}

func (cfg Config) hash() (uint64, error) {
	hash, err := hashstructure.Hash(cfg, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to hash the rate limiter config: %w", err)
	}

	return hash, nil
}
