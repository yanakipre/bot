package roundtripper

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

func LoggingRoundTripper(rt http.RoundTripper, clientName string) http.RoundTripper {
	return &loggingTripper{rt: rt, clientName: clientName}
}

type loggingTripper struct {
	rt         http.RoundTripper
	clientName string
}

// RoundTrip implements http.RoundTripper
// Does not error out when fails to log request or response
func (hrt *loggingTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	path := "unknown"
	if r.URL != nil {
		path = r.URL.Path
	}
	ctx := logger.WithFields(logger.WithName(logger.WithName(r.Context(), "http"), hrt.clientName),
		zap.String("egress_method", r.Method),
		zap.String("egress_path", path),
		zap.String("client_name", hrt.clientName),
	)

	startTime := time.Now().UTC()
	resp, err := hrt.rt.RoundTrip(r.WithContext(ctx))
	duration := time.Since(startTime)

	if err != nil {
		logger.Debug(ctx, "outgoing request resulted in error",
			zap.Error(err),
			zap.Int64("egress_duration_ms", duration.Milliseconds()),
		)
		return nil, err
	}

	// max 3 fields
	logFields := make([]zap.Field, 0, 3)
	// log location
	switch resp.StatusCode {
	case http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
		http.StatusFound,
		http.StatusMovedPermanently:
		logFields = append(logFields, zap.String("location", resp.Header.Get("Location")))
	}
	// log default fields
	logFields = append(logFields,
		zap.String("http_code", resp.Status),
		zap.Int64("egress_duration_ms", duration.Milliseconds()),
	)

	logger.Debug(ctx, fmt.Sprintf("outgoing request finished with %q", resp.Status), logFields...)

	return resp, nil
}

func (hrt *loggingTripper) Unwrap() http.RoundTripper { return hrt.rt }
