package metricsapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/yanakipe/bot/internal/openapiapp"
	"github.com/yanakipe/bot/internal/promtooling"
	"github.com/yanakipe/bot/internal/resttooling"
)

func NewMetricsHandler(logSubjField resttooling.SubjectIdentityLogFieldFunc) http.Handler {
	noUserId := errors.New("user id not in context")
	return openapiapp.Wrap(
		promtooling.Handler(),
		resttooling.LoggingMiddleware(
			"metrics-api",
			func(r *http.Request) resttooling.UrlMethod { return resttooling.UrlMethod{} },
			func(ctx context.Context) (string, error) { return "", noUserId },
			logSubjField,
		),
		resttooling.TracingMiddleware("metrics-api"),
		resttooling.SentryMiddleware(),
		resttooling.RecoveryMiddleware(
			func(w http.ResponseWriter, r *http.Request, appErr error) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("500 internal server error"))
			},
		),
	)
}
