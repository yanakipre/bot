package roundtripper

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/resttooling/requestid"
)

func RequestIDRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &requestIDRoundTripper{
		rt: rt,
	}
}

type requestIDRoundTripper struct {
	rt http.RoundTripper
}

// RoundTrip implements http.RoundTripper
// Does not error out when fails to log request or response
func (zrt *requestIDRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	reqID := requestid.FromContextOrNew(ctx)
	ctx = requestid.WithRequestID(ctx, reqID)
	ctx = logger.WithFields(ctx,
		zap.String("egress_request_id", reqID),
	)
	r.Header.Set(requestid.HeaderName, reqID)

	resp, err := zrt.rt.RoundTrip(r.WithContext(ctx))
	return resp, err
}

func (zrt *requestIDRoundTripper) Unwrap() http.RoundTripper { return zrt.rt }
