package roundtripper

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

const DefaultSlug = "default"

func RateLimiterRoundTripper(
	rt http.RoundTripper,
	limiters map[string]*rate.Limiter,
	uriSlug func(ctx context.Context) (uri string),
) http.RoundTripper {
	return &rateLimiterRoundTripper{rt: rt, limiters: limiters, uriSlug: uriSlug}
}

type rateLimiterRoundTripper struct {
	rt http.RoundTripper
	// slug -> limiter. One limiter can be referenced by several slugs
	limiters map[string]*rate.Limiter
	uriSlug  func(ctx context.Context) (uri string)
}

// RoundTrip implements http.RoundTripper
// Does not error out when fails to log request or response
func (rt *rateLimiterRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	limiter := rt.limiters[rt.uriSlug(ctx)]
	if limiter == nil {
		limiter = rt.limiters[DefaultSlug]
	}

	if limiter != nil {
		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("error while waiting for rate limiters: %w", err)
		}
	}

	return rt.rt.RoundTrip(r)
}

func (rt *rateLimiterRoundTripper) Unwrap() http.RoundTripper { return rt.rt }
