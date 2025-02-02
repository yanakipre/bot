package resttooling

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/semerr"
)

type (
	ErrorKeyType string
)

const (
	UriParseErrorMsg   = "could not parse uri request"
	JsonParseErrorMsg  = "could not parse json request"
	QueryParseErrorMsg = "could not parse query request"

	errorKey ErrorKeyType = "error"

	// Not a real HTTP status code, because when context is canceled, connection
	// is closed by the client, so technically it's not a response from the server.
	// Yet, it's widely used in different places to distinguish it from other errors,
	// like internal server error (500).
	// See, for example Nginx documentation:
	// NGX_HTTP_CLIENT_CLOSED_REQUEST - 499
	// https://www.nginx.com/resources/wiki/extending/api/http/
	//
	// We will have a separate panel in Grafana to monitor these errors, as a lot of
	// them could be a symptom of LB and/or ingress misconfiguration.
	contextCanceledStatusCode int = 499
)

func CodeFromSemanticError(err error) (int, bool) {
	e := semerr.AsSemanticError(err)
	if e == nil {
		return 0, false
	}

	// Do not report context.Canceled as 500 (default for internal errors).
	// It's client closed the connection, so technically it's not a response
	// from the server at all.
	if errors.Is(err, context.Canceled) {
		return contextCanceledStatusCode, true
	} else {
		return SemanticErrorToHTTP(e.Semantic), true
	}
}

func HTTPToSemantic(s int) func(err error, t string) *semerr.Error {
	switch s {
	case http.StatusServiceUnavailable:
		return semerr.WrapWithUnavailable
	case http.StatusRequestTimeout:
		return semerr.WrapWithTimeout
	case http.StatusNotAcceptable:
		return semerr.WrapWithNotAcceptable
	case http.StatusUnprocessableEntity:
		return semerr.WrapWithUnprocessable
	case http.StatusLocked:
		return semerr.WrapWithResourceLocked
	case http.StatusConflict:
		return semerr.WrapWithAlreadyExists
	case http.StatusPreconditionFailed:
		return semerr.WrapWithFailedPrecondition
	case http.StatusNotFound:
		return semerr.WrapWithNotFound
	case http.StatusForbidden:
		return semerr.WrapWithForbidden
	case http.StatusUnauthorized:
		return semerr.WrapWithAuthentication
	case http.StatusBadRequest:
		return semerr.WrapWithInvalidInput
	}
	return semerr.WrapWithInternal
}

func SemanticErrorToHTTP(s semerr.Semantic) int {
	switch s {
	case semerr.SemanticInvalidInput:
		return http.StatusBadRequest
	case semerr.SemanticAuthentication:
		return http.StatusUnauthorized
	case semerr.SemanticForbidden:
		// We send 404 on Authorization errors to
		// prevent from listing resources without access
		// by looking at 403 status code.
		return http.StatusNotFound
	case semerr.SemanticNotFound:
		return http.StatusNotFound
	case semerr.SemanticFailedPrecondition:
		return http.StatusPreconditionFailed
	case semerr.SemanticAlreadyExists:
		return http.StatusConflict
	case semerr.SemanticNotImplemented:
		// We don't want to wake up in the middle of the night
		// because client wants some unimplemented functionality.
		return http.StatusBadRequest
	case semerr.SemanticResourceLocked:
		return http.StatusLocked
	case semerr.SemanticTooManyRequests:
		return http.StatusTooManyRequests
	case semerr.SemanticUnprocessable:
		return http.StatusUnprocessableEntity
	case semerr.SemanticNotAcceptable:
		return http.StatusNotAcceptable
	case semerr.SemanticTimeout:
		return http.StatusRequestTimeout
	case semerr.SemanticUnavailable:
		return http.StatusServiceUnavailable
	case semerr.SemanticSkipError:
		return http.StatusOK
	case semerr.SemanticPartialSuccess:
		return http.StatusMultiStatus
	case semerr.SemanticUnknown:
		fallthrough
	case semerr.SemanticInternal:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

func LogAppError(ctx context.Context, err error) {
	fields := make([]zap.Field, 0, 2)
	fields = append(fields, zap.Error(err))
	semanticErr := semerr.AsSemanticError(err)
	var message string
	if semanticErr == nil {
		semanticErr = semerr.WrapWithInternal(err, "")
	} else {
		// we got a semerr with trace, let's add it
		fields = append(fields,
			zap.String("trace_id", fmt.Sprintf("%+v", semanticErr.StackTrace())))
	}

	if errors.Is(err, context.Canceled) {
		message = "incoming request finished with cancellation"
		logger.Warn(ctx,
			message,
			fields...,
		)
		// return early,
		// we don't want it to be logged with error,
		// we don't want to fire up alerts on API errors.
		return
	} else {
		switch semanticErr.Semantic {
		case semerr.SemanticInternal:
			// we highlight all internal errors explicitly with separate message
			message = "incoming request finished with internal error"
		default:
			message = "incoming request finished with error"
		}
	}
	logger.Error(ctx,
		message,
		fields...,
	)
}

// WithError add error value to the context if it doesn't exist or replace it if it does.
// LoggingMiddleware uses the result in ErrorMustFromContext.
// WithError must be called at least once before using ErrorMustFromContext.
func WithError(ctx context.Context, err error) context.Context {
	ctxErrInterface := ctx.Value(errorKey)
	if ctxErrInterface == nil {
		return context.WithValue(ctx, errorKey, &err)
	} else if ctxErr, ok := ctxErrInterface.(*error); ok {
		*ctxErr = err
		return ctx
	} else {
		panic("error value has invalid type")
	}
}

func ErrorMustFromContext(ctx context.Context) error {
	ctxErrInterface := ctx.Value(errorKey)
	if ctxErrInterface != nil {
		if ctxErr, ok := ctxErrInterface.(*error); ok {
			return *ctxErr
		} else {
			panic("error key has invalid type")
		}
	} else {
		panic("error key not found in context")
	}
}
