// Package application provides the application configuration for standard cases.
package openapiapp

import "time"

type Config struct {
	// BaseURL
	//
	// Base URL to serve API from. Example: /billing/api/v1
	BaseURL string `json:"base_url"`
	// Addr
	//
	// Address to bind and listen on. Example: 0.0.0.0:9085
	Addr string
	// Name of the application will be present in the logs and metrics exposed.
	Name string `yaml:"name"`
	// ReadHeaderTimeout
	//
	// Timeout of http server reading the headers, as in the stdlib http.Server.
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
}

// DefaultConfig returns default configuration for application.
func DefaultConfig(baseURL string, addr string, name string) Config {
	return Config{
		BaseURL:           baseURL,
		Addr:              addr,
		Name:              name,
		ReadHeaderTimeout: time.Second,
	}
}
