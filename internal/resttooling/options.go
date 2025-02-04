package resttooling

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/time/rate"

	"github.com/yanakipre/bot/internal/resttooling/http2tooling"
	"github.com/yanakipre/bot/internal/resttooling/ratelimiter"
	"github.com/yanakipre/bot/internal/resttooling/restretries"
	"github.com/yanakipre/bot/internal/resttooling/roundtripper"
)

// WithHTTP2Support adds support for HTTP2 protocol retries.
//
// At the moment there is no opt-out from it. It's enabled automatically.
// There seems to be no added negative impact by always enabling it.
func WithHTTP2Support() Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.HTTP2 = func(c *http.Client) *http.Client {
			c.Transport = http2tooling.RoundTripper(c.Transport)
			return c
		}
		return cfg
	}
}

// WithTimeout adds a timeout to http client.
//
// From net/http lib source code:
// """
// Timeout specifies a time limit for individual requests made by this
// Client. The timeout includes connection time, any
// redirects, and reading the response body. The timer remains
// running after Get, Head, Post, or Do return and will
// interrupt reading of the Response.Body.
// """
func WithTimeout(timeout time.Duration) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.OrderlessOptions = append(cfg.OrderlessOptions, func(c *http.Client) *http.Client {
			c.Timeout = timeout
			return c
		})
		return cfg
	}
}

// WithMetrics adds observability for outgoing HTTP requests.
func WithMetrics(
	eventDescription func(ctx context.Context) (e roundtripper.MetricEvent),
) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Metrics = func(c *http.Client) *http.Client {
			c.Transport = roundtripper.MetricsRoundTripper(c.Transport, eventDescription)
			return c
		}
		return cfg
	}
}

// WithRetries adds retry attempts if request fails.
func WithRetries(
	retryNetwork restretries.RetryByNetwork,
	retryByResponse restretries.RetryByResponse,
	s restretries.RetryStrategy,
) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Retries = func(c *http.Client) *http.Client {
			c.Transport = restretries.NewRetryableRoundTripper(
				c.Transport,
				retryNetwork,
				s,
				retryByResponse,
			)
			return c
		}
		return cfg
	}
}

// WithResponseBasedRetries adds retry attempts if request fails.
func WithResponseBasedRetries(
	retryByResponse restretries.RetryByResponse,
	s restretries.RetryStrategy,
) Option {
	return WithRetries(restretries.StraightforwardNetworkRetry(), retryByResponse, s)
}

// WithRequestID adds request ID to http client.
func WithRequestID() Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.RequestID = func(c *http.Client) *http.Client {
			c.Transport = roundtripper.RequestIDRoundTripper(c.Transport)
			return c
		}
		return cfg
	}
}

// WithLogging adds logging to http client.
// WithRequestID should come AFTER that option in list of applied options.
func WithLogging(clientName string) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Logging = func(c *http.Client) *http.Client {
			c.Transport = roundtripper.LoggingRoundTripper(c.Transport, clientName, false)
			return c
		}
		return cfg
	}
}

// WithBodyLogging adds logging to http client including logging of request and response bodies.
// WithRequestID should come AFTER that option in list of applied options.
func WithBodyLogging(clientName string) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Logging = func(c *http.Client) *http.Client {
			c.Transport = roundtripper.LoggingRoundTripper(c.Transport, clientName, true)
			return c
		}
		return cfg
	}
}

// WithTracing adds OpenTelemetry tracing headers to all requests made with the client.
func WithTracing(clientName string) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Tracing = func(c *http.Client) *http.Client {
			c.Transport = otelhttp.NewTransport(
				c.Transport,
				otelhttp.WithSpanNameFormatter(
					func(operation string, r *http.Request) string {
						slug := URISlugFromContext(r.Context())
						if slug == "" {
							slug = r.Method
						}
						return fmt.Sprintf("%s %s %s", "HTTP", clientName, slug)
					},
				),
			)
			return c
		}
		return cfg
	}
}

// WithHeadersFromContext adds custom headers derived from the context
func WithHeadersFromContext(
	setHeaders func(ctx context.Context, r http.Request) (map[string]string, error),
) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.OrderlessOptions = append(cfg.OrderlessOptions, func(c *http.Client) *http.Client {
			c.Transport = roundtripper.HeadersRoundTripper(c.Transport, setHeaders)
			return c
		})
		return cfg
	}
}

// WithHeaderForJWT adds the JWT authorization header to all
// requests.
func WithHeaderForJWT(
	token string,
) Option {
	return WithHeadersFromContext(func(context.Context, http.Request) (map[string]string, error) {
		if token == "" {
			return nil, nil
		}
		return map[string]string{AuthorizationHeader: fmt.Sprint("Bearer ", token)}, nil
	})
}

func WithTransportCacheSizeCollector(collector func(int)) Option {
	return func(c optionsConfig) optionsConfig {
		collector(roundTripperCacheSize())
		return c
	}
}

func WithTransport(conf Config) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Transport = func(c *http.Client) *http.Client {
			c.Transport = getHTTPRoundTripper(conf)
			return c
		}
		return cfg
	}
}

func WithRateLimiter(
	rateLimitersConfig []ratelimiter.RateLimitByHandlersConfig,
	workersCount uint,
) Option {
	return func(cfg optionsConfig) optionsConfig {
		limiters := make(map[string]*rate.Limiter, len(rateLimitersConfig))
		for _, cfg := range rateLimitersConfig {
			limiter := rate.NewLimiter(
				rate.Every(
					cfg.Config.Period.Duration/time.Duration(cfg.Config.Requests/workersCount),
				),
				int(cfg.Config.Burst),
			)
			for _, handler := range cfg.Handlers {
				limiters[handler] = limiter
			}
		}
		cfg.OrderlessOptions = append(cfg.OrderlessOptions, func(c *http.Client) *http.Client {
			c.Transport = roundtripper.RateLimiterRoundTripper(
				c.Transport,
				limiters,
				URISlugFromContext,
			)
			return c
		})
		return cfg
	}
}

// WithMock mocks all incoming requests using the given handler
func WithMock(handler http.Handler) Option {
	return func(cfg optionsConfig) optionsConfig {
		cfg.Transport = func(c *http.Client) *http.Client {
			c.Transport = roundtripper.MockRoundTripper(handler)
			return c
		}
		return cfg
	}
}
