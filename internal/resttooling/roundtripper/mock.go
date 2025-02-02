package roundtripper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
)

func MockRoundTripper(handler http.Handler) http.RoundTripper {
	return &mockRoundTripper{handler: handler}
}

type mockRoundTripper struct {
	handler http.Handler
}

// RoundTrip implements http.RoundTripper
func (rt *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	logger.Debug(r.Context(), "mocking", zap.String(r.Method, r.URL.String()))

	w := newResWriter(r)
	rt.handler.ServeHTTP(w, r)

	return &w.res, nil
}

// resWriter implements http.ResponseWriter interface
type resWriter struct {
	res http.Response
}

func newResWriter(req *http.Request) *resWriter {
	w := resWriter{
		res: http.Response{
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Request:    req,
		},
	}
	return &w
}

func (w *resWriter) Header() http.Header { return w.res.Header }

func (w *resWriter) WriteHeader(code int) {
	w.res.StatusCode = code
	w.res.Status = fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func (w *resWriter) Write(bs []byte) (int, error) {
	if w.res.StatusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	w.res.Body = io.NopCloser(bytes.NewBuffer(bs))
	w.res.ContentLength = int64(len(bs))
	return len(bs), nil
}

// SendJson writes json response. This is a helper function when writing mock round tripper.
func SendJson(ctx context.Context, w http.ResponseWriter, body any) {
	bytes, err := json.Marshal(&body)
	if err != nil {
		logger.Error(ctx, "failed to marshal body:", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(bytes); err != nil {
		logger.Error(ctx, "failed to write response:", zap.Error(err))
	}
}
