package httpopenaiclient

import (
	"net/http"

	"github.com/sashabaranov/go-openai"
	"github.com/yanakipe/bot/internal/resttooling"
	"github.com/yanakipe/bot/internal/resttooling/restretries"
)

type Client struct {
	c   *openai.Client
	cfg Config
}

func NewClient(cfg Config) *Client {
	oaiCfg := openai.DefaultConfig(cfg.ApiKey.Unmask())
	if cfg.httpClient != nil {
		oaiCfg.HTTPClient = cfg.httpClient
	} else {
		oaiCfg.HTTPClient = defaultHTTPClient(cfg)
	}
	client := openai.NewClientWithConfig(oaiCfg)
	return &Client{c: client, cfg: cfg}
}

func defaultHTTPClient(cfg Config) *http.Client {
	return resttooling.NewHTTPClientFromConfig(
		cfg.Transport,
		resttooling.WithMetrics(resttooling.MetricFromContext(resttooling.MetricReportCfg{
			ClientName: cfg.Transport.ClientName,
		})),
		resttooling.WithTracing(cfg.Transport.ClientName),
		resttooling.WithRetries(
			// retry all network errors
			restretries.UnconditionalNetworkRetry(),
			restretries.RetryByStatusCode(
				http.StatusServiceUnavailable,
				http.StatusTooManyRequests,
				http.StatusGatewayTimeout,
			),
			restretries.FibonacciBackoffWithJitterRetryStrategy(
				cfg.Retries.Backoff.Duration,
				cfg.Retries.Attempts,
			),
		),
	)
}
