package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

var ctxKeyRequestID = ctxKey("request_id")

const HeaderName = "x-request-id"

// we return request id back to user
// to ease debugging problems
const ReturnRequestIDHeader = "X-Neon-Ret-Request-ID"

// NewReqID returns new random request ID
func NewReqID() string {
	return uuid.NewString()
}

// FromContext retrieves request ID from context
func FromContext(ctx context.Context) (string, bool) {
	rid, ok := ctx.Value(ctxKeyRequestID).(string)
	return rid, ok
}

// FromContextOrNew retrieves request ID from context or generates a NewReqID one
func FromContextOrNew(ctx context.Context) string {
	rid, ok := FromContext(ctx)
	if !ok {
		return NewReqID()
	}

	return rid
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, rid string) context.Context {
	// Do not replace value already stored in ctx
	if ctx.Value(ctxKeyRequestID) != nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKeyRequestID, rid)
}

// RequestIDFromRequestOrNew either returns request id that client sent,
// or creates synthetic request id and puts it into context.
// it can return shallow copy of original request.
func RequestIDFromRequestOrNew(r *http.Request) string {
	ctx := r.Context()
	// already set correctly
	if headerValue, ok := FromContext(ctx); ok {
		return headerValue
	}
	headerValue := r.Header.Get(HeaderName)
	if headerValue == "" {
		// not set by client, generate new.
		headerValue = NewReqID()
	}
	// remember what we collected.
	ctx = WithRequestID(ctx, headerValue)
	*r = *r.WithContext(ctx)

	return headerValue
}
