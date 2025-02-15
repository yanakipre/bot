package resttooling

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/logger"
)

type dummySubjectKey struct{}

func SubjectToContext(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, dummySubjectKey{}, s)
}

func SubjectFromContext(ctx context.Context) (string, error) {
	s, ok := ctx.Value(dummySubjectKey{}).(string)
	if !ok {
		return "", errors.New("subject not in context")
	}
	return s, nil
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	// init logger package
	logger.SetNewGlobalLoggerOnce(logger.Config{
		Sink:     "stdout",
		LogLevel: "INFO",
		Format:   "console",
	})

	tt := []struct {
		name           string
		req            *http.Request
		handler        http.HandlerFunc
		subjIdLogField SubjectIdentityLogFieldFunc
		want           func(t *testing.T, buf *bytes.Buffer)
	}{
		{
			name: "logs ids that are found in the http path",
			req: httptest.NewRequest(
				http.MethodDelete,
				"/api/v2/projects/red-dream-62543137/branches/br-wandering-thunder-a78f1kz8/endpoints/ep-billowing-frost-a4u2nado/computes/a78f1kz8",
				nil,
			),
			subjIdLogField: SubjectIdentityAsUserID,
			want: func(t *testing.T, buf *bytes.Buffer) {
				dec := json.NewDecoder(buf)

				logged := map[string]any{}
				for dec.More() {

					err := dec.Decode(&logged)
					require.NoError(t, err)

					require.Equal(t, http.MethodDelete, logged["http_meth"].(string))
					require.Equal(
						t,
						"/api/v2/projects/red-dream-62543137/branches/br-wandering-thunder-a78f1kz8/endpoints/ep-billowing-frost-a4u2nado/computes/a78f1kz8",
						logged["http_path"].(string),
					)
					require.Equal(t, "red-dream-62543137", logged["project_id"].(string))
					require.Equal(t, "br-wandering-thunder-a78f1kz8", logged["branch_id"].(string))
					require.Equal(t, "ep-billowing-frost-a4u2nado", logged["endpoint_id"].(string))
					require.Equal(t, "a78f1kz8", logged["compute_id"].(string))
				}
			},
		},
		{
			name:           "logs when there are no ids in request http",
			req:            httptest.NewRequest(http.MethodGet, "/api/v2/projects", nil),
			subjIdLogField: SubjectIdentityAsUserID,
			want: func(t *testing.T, buf *bytes.Buffer) {
				dec := json.NewDecoder(buf)

				logged := map[string]any{}
				for dec.More() {

					err := dec.Decode(&logged)
					require.NoError(t, err)

					require.Equal(t, http.MethodGet, logged["http_meth"].(string))
					require.Equal(t, "/api/v2/projects", logged["http_path"].(string))
					require.Equal(t, nil, logged["project_id"])
					require.Equal(t, nil, logged["branch_id"])
					require.Equal(t, nil, logged["project_id"])
					require.Equal(t, nil, logged["endpoint_id"])
				}
			},
		},
		{
			name: "logs user_id",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/api/v2/projects", nil)
				ctx := SubjectToContext(r.Context(), uuid.MustParse("2b6e7c62-6d37-4a89-9330-82a92dc4343a").String())
				return r.WithContext(ctx)
			}(),
			subjIdLogField: SubjectIdentityAsUserID,
			want: func(t *testing.T, buf *bytes.Buffer) {
				dec := json.NewDecoder(buf)

				logged := map[string]any{}
				for dec.More() {

					err := dec.Decode(&logged)
					require.NoError(t, err)
				}
				require.Equal(t, "2b6e7c62-6d37-4a89-9330-82a92dc4343a", logged["user_id"].(string))
			},
		},
		{
			name: "logs subject identity when present",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/api/v2/projects", nil)
				ctx := SubjectToContext(r.Context(), uuid.MustParse("2b6e7c62-6d37-4a89-9330-82a92dc4343a").String())
				return r.WithContext(ctx)
			}(),
			subjIdLogField: SubjectIdentityAsUserID,
			want: func(t *testing.T, buf *bytes.Buffer) {
				dec := json.NewDecoder(buf)

				logged := map[string]any{}
				for dec.More() {

					err := dec.Decode(&logged)
					require.NoError(t, err)
				}
				require.Equal(t, "2b6e7c62-6d37-4a89-9330-82a92dc4343a", logged["user_id"].(string))
			},
		},
	}

	for _, tc := range tt {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			if tc.handler != nil {
				handler = tc.handler
			}

			urlMethodGetter := func(r *http.Request) UrlMethod {
				return UrlMethod{}
			}

			h := SentryMiddleware(urlMethodGetter)(
				LoggingMiddleware(
					"unit-test",
					urlMethodGetter,
					func(ctx context.Context) (string, error) {
						return SubjectFromContext(ctx)
					},
					tc.subjIdLogField,
				)(handler))

			buf := &bytes.Buffer{}

			h.ServeHTTP(
				httptest.NewRecorder(),
				tc.req.WithContext(logger.WithLogger(tc.req.Context(), logger.NewWithSink(nil, buf, nil))),
			)

			tc.want(t, buf)
		})
	}
}
