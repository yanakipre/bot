package resttooling

import (
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/yanakipre/bot/internal/sentrytooling"
)

const contextRouteName = "Route Name"

// SentryMiddleware initializes sentry usage fror HTTP stack.
func SentryMiddleware(routeNameFunc RouteNameFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routeName := routeNameFunc(r).OperationID

			ctx := r.Context()
			ctx = sentrytooling.InitCtx(ctx, "Yanakipre HTTP")
			sCtx := sentrytooling.FromContext(ctx)
			sCtx[contextRouteName] = routeName

			span := sentry.StartSpan(
				ctx,
				"http_handler_execution",
				sentry.WithTransactionName(routeName),
			)
			defer span.Finish()

			*r = *r.WithContext(ctx) // execute normal process.
			next.ServeHTTP(w, r)
		})
	}
}
