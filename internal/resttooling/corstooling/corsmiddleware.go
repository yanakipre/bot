package corstooling

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"github.com/yanakipe/bot/internal/logger"
)

type Config struct {
	// AllowedHosts
	//
	// List of URLs from which CORS requests are allowed
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func DefaultConfig() Config {
	return Config{
		AllowedOrigins: []string{
			"https://neon.tech",
			"http://localhost:8080",
		},
	}
}

type log struct {
	logger logger.Logger
}

func (l log) Printf(s string, i ...any) {
	l.logger.Info(fmt.Sprintf(strings.TrimSpace(s), i...))
}

var _ cors.Logger = &log{}

func CorsFromConfig(logger logger.Logger, cfg Config) *cors.Cors {
	middleware := cors.New(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: []string{"POST", "OPTIONS", "GET", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
		},
		AllowCredentials:   true,
		OptionsPassthrough: false,
		Debug:              false,
	})
	middleware.Log = log{logger: logger.Named("cors")}
	return middleware
}

func Middleware(logger logger.Logger, cfg Config) func(next http.Handler) http.Handler {
	corsHandler := CorsFromConfig(logger, cfg)
	return func(next http.Handler) http.Handler {
		return corsHandler.Handler(next)
	}
}
