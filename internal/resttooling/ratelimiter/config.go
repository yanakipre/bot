package ratelimiter

import "github.com/yanakipre/bot/internal/encodingtooling"

// RateLimitConfig is a config for a single rate limiter
// This package uses golang.org/x/time/rate
// It implements a "token bucket" of size b(Burst), initially full and refilled at rate Requests tokens per Period.
//
// So, to configure usual behavior with a limit N requests per second, you need to set
// Period = 1 second, Requests = N, Burst = 1
//
// Burst allows to handle requests spikes (for example, if you have spikes each 10 sec of 100 requests,
// you can set Period = 10 sec, Requests = 100, Burst = 100) so that spike will be served immediately and
// then tokens bucket will be refilled at rate 10 requests per second
type RateLimitConfig struct {
	Period   encodingtooling.Duration `yaml:"period"   json:"period"`
	Requests uint                     `yaml:"requests" json:"requests"`
	Burst    uint                     `yaml:"burst"    json:"burst"`
}

type RateLimitByHandlersConfig struct {
	Handlers []string        `yaml:"handlers" json:"handlers"`
	Config   RateLimitConfig `yaml:"config"   json:"config"`
}

type RateLimitersConfig struct {
	ByHandlers   []RateLimitByHandlersConfig `yaml:"by_handlers"   json:"by_handlers"`
	WorkersCount uint                        `yaml:"workers_count" json:"workers_count"`
}
