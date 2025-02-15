package resttooling

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Add OpenTelemetry tracing to incoming requests
//
// Starts a new span for each incoming request. If the request contains tracing
// context, the new span is created as a child span of the caller.
//
// NEVER put TracingMiddleware after any middleware that manipulates logging ctx.
// It messes the context and changes the logging.
func TracingMiddleware(appName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := otelhttp.NewHandler(next, appName)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}
}
