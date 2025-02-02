package roundtripper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
)

func LoggingRoundTripper(rt http.RoundTripper, clientName string, logBody bool) http.RoundTripper {
	return &loggingTripper{rt: rt, clientName: clientName, logBody: logBody}
}

type loggingTripper struct {
	rt         http.RoundTripper
	clientName string
	logBody    bool
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

	// Read and store request body
	var reqBody []byte
	if hrt.logBody && r.Body != nil {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Warn(ctx, "failed to read request body", zap.Error(err))
		} else {
			body := io.NopCloser(bytes.NewReader(payload))
			r.Body = body
			reqBody = payload
		}
	}

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

	// Read and store response body
	var resBody []byte
	if hrt.logBody && resp.Body != nil {
		resBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Warn(ctx, "failed to read response body", zap.Error(err))
		} else {
			body := io.NopCloser(bytes.NewReader(resBytes))
			resp.Body = body
			resBody = resBytes
		}
	}

	// Log request and response bodies
	if hrt.logBody {
		logFields = append(logFields,
			zap.ByteString("req_body", reqBody),
			zap.ByteString("res_body", resBody),
		)
	}

	logger.Debug(ctx, fmt.Sprintf("outgoing request finished with %q", resp.Status), logFields...)

	return resp, nil
}

func (hrt *loggingTripper) Unwrap() http.RoundTripper { return hrt.rt }
