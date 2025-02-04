package resttooling

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/metrics"
	"github.com/yanakipre/bot/internal/sentrytooling"
)

func MetricsMiddleware(
	appName string,
	routeNameFunc RouteNameFunc,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for _, stopUrl := range UrlStopList {
				if strings.HasPrefix(r.URL.Path, stopUrl) {
					next.ServeHTTP(w, r)
					return
				}
			}
			routeName := routeNameFunc(r).OperationID

			sentryCtx := sentrytooling.FromContext(ctx)
			sentryCtx["Route Name"] = routeName

			startTime := time.Now()

			sentryScope := sentry.GetHubFromContext(ctx).Scope()
			sentryScope.SetTransaction(routeName)

			*r = *r.WithContext(ctx)

			responseWriter := NewLoggingResponseWriter(w)

			// execute normal process.
			next.ServeHTTP(responseWriter, r)

			// after request
			latency := time.Since(startTime)

			*r = *r.WithContext(logger.WithFields(
				r.Context(), zap.Int64("ingress_duration_ms", latency.Milliseconds())))

			// set request total
			metrics.APIRequestsTotal.WithLabelValues(appName).Inc()
			// set uri request total
			metrics.APIRequestsTotalURI.WithLabelValues(
				appName, routeName, r.Method, strconv.Itoa(responseWriter.StatusCode)).
				Inc()
			// set request duration
			metrics.APIRequestsDuration.WithLabelValues(appName, routeName).
				Observe(latency.Seconds())
		})
	}
}
