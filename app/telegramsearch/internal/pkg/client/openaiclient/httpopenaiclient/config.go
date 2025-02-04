package httpopenaiclient

import (
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/resttooling"
	"github.com/yanakipre/bot/internal/resttooling/ratelimiter"
	"github.com/yanakipre/bot/internal/secret"
)

type Config struct {
	EmbeddingConfig EmbeddingConfig
	// httpClient
	//
	// Override HTTP client.
	httpClient *http.Client
	ApiKey     secret.String
	Transport  resttooling.Config `yaml:"transport"`
	Retries    resttooling.RetriesConfig
	// Rate limits by handlers(slug)
	RateLimiters   []ratelimiter.RateLimitByHandlersConfig `yaml:"rate_limiters"`
	AskingAbout    string                                  `yaml:"asking_about"`
	DoNotHighlight string                                  `json:"do_not_highlight"`
}

type EmbeddingConfig struct {
	Model openai.EmbeddingModel
}

func DefaultConfig() Config {
	tr := resttooling.DefaultTransportConfig()
	tr.ResponseHeaderTimeout = encodingtooling.Duration{Duration: time.Minute}
	tr.ClientName = "openapi"
	return Config{
		DoNotHighlight: "Cyprus",
		AskingAbout:    "Cyprus",
		EmbeddingConfig: EmbeddingConfig{
			Model: openai.SmallEmbedding3,
		},
		Transport: tr,
		Retries: resttooling.RetriesConfig{
			Backoff:  encodingtooling.Duration{Duration: 100 * time.Millisecond},
			Attempts: 10,
		},
	}
}
