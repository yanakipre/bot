package resttooling

import (
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/yanakipe/bot/internal/sentrytooling"
)

// SentryMiddleware initializes sentry usage fror HTTP stack.
func SentryMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = sentrytooling.InitCtx(ctx, "Neon HTTP")

			span := sentry.StartSpan(
				ctx,
				"http_handler_execution",
			)
			defer span.Finish()

			*r = *r.WithContext(ctx) // execute normal process.
			next.ServeHTTP(w, r)
		})
	}
}
