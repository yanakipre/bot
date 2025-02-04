package resttooling

import (
	"net/http"
	"strings"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

type Option func(cfg optionsConfig) optionsConfig

type OptionFunc func(c *http.Client) *http.Client

func getOptionsCfg(options ...Option) optionsConfig {
	// no-opt out options go here.
	// we add no-opt out option into beginning (prepending)
	// because in the for loop below last applied option wins.
	// So, if client defines an option,
	// then the option supplied by client will override the default one.
	options = append([]Option{}, options...)
	cfg := optionsConfig{}
	for _, o := range options {
		cfg = o(cfg)
	}
	return cfg
}

// applyOptionsCfg controls the order of applied options
// last applied is executed first.
func applyOptionsCfg(c *http.Client, cfg optionsConfig) *http.Client {
	if cfg.Transport != nil {
		c = cfg.Transport(c)
	}

	if cfg.Metrics != nil {
		c = cfg.Metrics(c)
	}

	// create a separate tracing span for each retry
	if cfg.Tracing != nil {
		c = cfg.Tracing(c)
	}

	// NOTE: anything above Retries is called separately for each
	// retry, while anything below is called only once for whole
	// overall operation!
	if cfg.Retries != nil {
		c = cfg.Retries(c)
	}

	if cfg.Logging != nil {
		c = cfg.Logging(c)
	}
	if cfg.RequestID != nil {
		c = cfg.RequestID(c)
	}

	for _, o := range cfg.OrderlessOptions {
		c = o(c)
	}

	if cfg.HTTP2 != nil {
		c = cfg.HTTP2(c)
	}
	return c
}

// NewHTTPClient creates HttpClient configured with options.
// Order of options is important.
func NewHTTPClient(options ...Option) *http.Client {
	return WrapHTTPClientWithOptions(&http.Client{}, options...)
}

type optionsConfig struct {
	HTTP2            OptionFunc
	Logging          OptionFunc
	Tracing          OptionFunc
	RequestID        OptionFunc
	Metrics          OptionFunc
	Retries          OptionFunc
	Transport        OptionFunc
	OrderlessOptions []OptionFunc
}

type RetriesConfig struct {
	Backoff  encodingtooling.Duration `yaml:"backoff"  json:"backoff"`
	Attempts uint                     `yaml:"attempts" json:"attempts"`
}

// NewHTTPClientFromConfig creates HttpClient configured from Config
func NewHTTPClientFromConfig(config Config, userOpts ...Option) *http.Client {
	options := make([]Option, 0, 4+len(userOpts))

	if config.ClientTimeout != nil && config.ClientTimeout.Duration > 0 {
		options = append(options, WithTimeout(config.ClientTimeout.Duration))
	}
	if config.ClientName == "" {
		config.ClientName = "UNKNOWN CLIENT"
	}

	// backward compatibility for clients
	// who still use RequestTimeout and not using ResponseHeaderTimeout
	if config.ResponseHeaderTimeout.Duration == 0 && config.RequestTimeout.Duration != 0 {
		config.ResponseHeaderTimeout = config.RequestTimeout
	}

	options = append(options,
		WithLogging(config.ClientName),
		WithRequestID(),
		WithTransport(config),
	)

	return NewHTTPClient(append(options, userOpts...)...)
}

func WrapHTTPClientWithOptions(c *http.Client, options ...Option) *http.Client {
	return applyOptionsCfg(c, getOptionsCfg(options...))
}

// GetClientAddr returns client IP address or empty string. This returns
// an untrustworthy IP address which should not be used for security purposes.
// Ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For#selecting_an_ip_address
func GetClientAddr(req *http.Request) string {
	value := req.Header.Get("X-Forwarded-For")
	ips := strings.Split(value, ",")
	if len(ips) > 0 {
		return ips[0]
	}

	return ""
}
