package metricsapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/yanakipre/bot/internal/openapiapp"
	"github.com/yanakipre/bot/internal/promtooling"
	"github.com/yanakipre/bot/internal/resttooling"
)

func NewMetricsHandler(logSubjField resttooling.SubjectIdentityLogFieldFunc) http.Handler {
	noUserId := errors.New("user id not in context")
	urlMethodGetter := func(r *http.Request) resttooling.UrlMethod { return resttooling.UrlMethod{} }
	return openapiapp.Wrap(
		promtooling.Handler(),
		resttooling.LoggingMiddleware(
			"metrics-api",
			urlMethodGetter,
			func(ctx context.Context) (string, error) { return "", noUserId },
			logSubjField,
		),
		resttooling.TracingMiddleware("metrics-api"),
		resttooling.SentryMiddleware(urlMethodGetter),
		resttooling.RecoveryMiddleware(
			func(w http.ResponseWriter, r *http.Request, appErr error) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("500 internal server error"))
			},
		),
	)
}
