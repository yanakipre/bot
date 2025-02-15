package roundtripper

import (
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

func MockRoundTripper(handler http.Handler) http.RoundTripper {
	return &mockRoundTripper{handler: handler}
}

type mockRoundTripper struct {
	handler http.Handler
}

// RoundTrip implements http.RoundTripper
func (rt *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	logger.Debug(ctx, "mocking", zap.Stringer(r.Method, r.URL))

	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		rt.handler.ServeHTTP(w, r)
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return w.Result(), nil
}
