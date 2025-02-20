package resttooling

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/resttooling/requestid"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/sentrytooling"
)

type ctxKey string

const (
	TraceIDKey ctxKey = "trace_id"
)

type SubjectIdentityLogFieldFunc func(identity string) zap.Field

var SubjectIdentityAsUserID SubjectIdentityLogFieldFunc = func(identity string) zap.Field {
	return zap.String("user_id", identity)
}

var SubjectIdentityAsAccountID SubjectIdentityLogFieldFunc = func(identity string) zap.Field {
	return zap.String("account_id", identity)
}

// LoggingMiddleware logs every incoming request and sets required context
// Beware of order. Should be at topmost position to log everything.
func LoggingMiddleware(
	appName string,
	routeF RouteNameFunc,
	getSubjectID func(ctx context.Context) (string, error),
	logSubjField SubjectIdentityLogFieldFunc,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := requestid.RequestIDFromRequestOrNew(r)

			// grab this value now in case some middleware modifies the r.URL
			httpPath := r.URL.Path

			ctx := r.Context()

			sentryCtx := sentrytooling.FromContext(ctx)
			sentryCtx["Request ID"] = reqID
			sentryCtx["HTTP Method"] = r.Method
			sentryCtx["HTTP Path"] = httpPath

			defaultLoggerFields := []zap.Field{
				zap.String("http_meth", r.Method),
				zap.String("http_path", httpPath),
				zap.String("route", routeF(r).OperationID),
				// request id is passing through services
				zap.String("request_id", reqID),
				// trace_id is unique for each call. The same key is used in periodic jobs.
				zap.String(string(TraceIDKey), shortuuid.New()),
			}

			idsFromHttpPath := getIdsFromHttpPath(httpPath)
			loggerFields := make([]zap.Field, 0, len(defaultLoggerFields)+len(idsFromHttpPath))

			loggerFields = append(loggerFields, defaultLoggerFields...)
			for key, id := range idsFromHttpPath {
				sentryCtx[key] = id
				loggerFields = append(loggerFields, zap.String(key, id))
			}

			ctx = logger.WithFields(logger.WithName(ctx, appName), loggerFields...)

			// Preinitialize error pointer so it could be set by errors handler
			// and then logged below after the request is finished.
			*r = *r.WithContext(WithErrorPlaceholder(ctx))

			defer func() {
				if rvr := recover(); rvr != nil {
					err := semerr.UnwrapPanic(rvr)

					// we don't log http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					if !errors.Is(err, http.ErrAbortHandler) {
						logger.Panic(r.Context(), rvr)
						logger.Error(
							r.Context(),
							clouderr.WrapWithFields(
								fmt.Errorf("incoming request finished with error: %w", err),
								zap.Int("status", http.StatusInternalServerError),
							),
						)
					}
					panic(rvr)
				}
			}()

			logger.Debug(r.Context(), "incoming request")

			// save status code
			responseWriter := newLoggingResponseWriter(w)
			// we return request id back to user
			// to ease debugging problems
			responseWriter.Header().Set(requestid.ReturnRequestIDHeader, reqID)

			next.ServeHTTP(responseWriter, r)

			ctx = r.Context()
			statusCode := responseWriter.StatusCode
			zapFields := []zap.Field{
				zap.Int("status", statusCode),
				zap.String("ip", r.RemoteAddr),
				zap.String("referer", r.Referer()),
				zap.String("user_agent", r.UserAgent()),
			}

			subjectID, err := getSubjectID(ctx)
			if err == nil {
				zapFields = append(zapFields, logSubjField(subjectID))
			}
			ctx = logger.WithFields(ctx, zapFields...)

			err = ErrorMustFromContext(ctx)
			if err != nil {
				LogAppError(ctx, err)
			} else {
				logger.Info(ctx, "incoming request finished successfully")
			}
		})
	}
}

// loggingResponseWriter saves response writer
type loggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

var httpPathKeyToIdKey = map[string]string{
	"projects":  "project_id",
	"branches":  "branch_id",
	"endpoints": "endpoint_id",
	"computes":  "compute_id",
}

func getIdsFromHttpPath(httpPath string) map[string]string {
	ids := make(map[string]string, len(httpPathKeyToIdKey))
	parts := strings.Split(httpPath, "/")
	for pathKey, idKey := range httpPathKeyToIdKey {
		id, ok := getIdFromHttpPathParts(parts, pathKey)
		if ok {
			ids[idKey] = id
		}
	}
	return ids
}

func getIdFromHttpPathParts(httpPathParts []string, key string) (id string, ok bool) {
	keyIdx := lo.IndexOf(httpPathParts, key)
	if keyIdx == -1 || keyIdx == len(httpPathParts)-1 {
		return "", false
	}
	return httpPathParts[keyIdx+1], true
}
