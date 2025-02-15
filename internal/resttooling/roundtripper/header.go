package roundtripper

import (
	"context"
	"net/http"
)

func HeadersRoundTripper(
	rt http.RoundTripper,
	setHeaders func(ctx context.Context, r http.Request) (map[string]string, error),
) http.RoundTripper {
	return &headersRoundTripper{rt: rt, setHeaders: setHeaders}
}

type headersRoundTripper struct {
	rt         http.RoundTripper
	setHeaders func(ctx context.Context, r http.Request) (map[string]string, error)
}

// RoundTrip implements http.RoundTripper
// Does not error out when fails to log request or response
func (rt *headersRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	headers, err := rt.setHeaders(ctx, *r)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		r.Header.Set(key, value)
	}
	return rt.rt.RoundTrip(r)
}

func (rt *headersRoundTripper) Unwrap() http.RoundTripper { return rt.rt }
