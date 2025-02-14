package openapiapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ogen-go/ogen/ogenerrors"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/codeerr"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/resttooling"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/sentrytooling"
	"github.com/yanakipre/bot/internal/status"
	"github.com/yanakipre/bot/internal/status/networkerrs"
)

// GeneralErrorStatusCode is implemented by our error structs generated with ogen.
type GeneralErrorStatusCode[T any] interface {
	GetStatusCode() int
	GetResponse() T
}

// NotFound return an HTTP handler that serves the standard "this route does not exist" error response.
func NotFound(h ogenerrors.ErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(r.Context(), w, r, semerr.NotFound("this route does not exist"))
	}
}

type ptrImplementsMarshaler[T any] interface {
	*T
	json.Marshaler
}

// ErrorHandler allows presenters to send errors in the expected format to the client.
func ErrorHandler[E GeneralErrorStatusCode[T], T any, _ ptrImplementsMarshaler[T]](
	newError func(context.Context, error) E,
) ogenerrors.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, appErr error) {
		generalError := newError(ctx, appErr)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(generalError.GetStatusCode())

		// We need to use MarshalJSON directly, see https://github.com/yanakipre/bot/pull/11970.
		// It's safe to cast to json.Marshaler, because ptrImplementsMarshaler[T] proves that *T implements it.
		response := generalError.GetResponse()
		marshal, err := any(&response).(json.Marshaler).MarshalJSON()
		if err != nil {
			logger.Error(
				ctx,
				clouderr.WrapWithFields(
					fmt.Errorf("cannot marshal err to JSON: %w", err),
					zap.NamedError("app_err", appErr),
				),
			)
		}

		_, err = w.Write(marshal)
		if err != nil {
			logger.Error(ctx, fmt.Errorf("cannot write response: %w", err))
		}
	}
}

// PresentError returns the status code, error code (if any),
// and message that should be presented in the API error response.
// Different APIs use different error structs, but they can leverage
// this function to return error information in a consistent manner.
func PresentError(ctx context.Context, err error) (statusCode int, errorCode codeerr.ErrorCode, message string) {
	var semanticErr *semerr.Error
	if codeErr := codeerr.AsCodeErr(err); codeErr != nil {
		errorCode = codeErr.GetCode()
		semanticErr = wrapWithSemantic(codeErr.RawError())
	} else {
		semanticErr = wrapWithSemantic(err)
	}

	statusCode, ok := resttooling.CodeFromSemanticError(semanticErr)
	if !ok {
		// Should never happen.
		panic("cannot get statusCode from semantic error")
	}

	// We use sentry in presenter. Because we have a single entry point for errors here.
	// It's just for convenience.
	sentrytooling.Report(ctx, semanticErr)
	resttooling.SetErrorInContext(ctx, semanticErr)

	return statusCode, errorCode, payloadFromSemantic(semanticErr)
}

// PresentStatus is similar to PresentError, but returns the error information as a status.Status.
func PresentStatus(ctx context.Context, err error) (st *status.Status, errorCode codeerr.ErrorCode) {
	st = status.FromError(err)

	// Enrich the Status we got with a user-facing message if it doesn't have one.
	st.EnrichWithUserFacingMessage()

	err = st.Error()
	if codeErr := codeerr.AsCodeErr(err); codeErr != nil {
		errorCode = codeErr.GetCode()
	}

	// We use sentry in presenter. Because we have a single entry point for errors here.
	// It's just for convenience.
	sentrytooling.Report(ctx, err)
	resttooling.SetErrorInContext(ctx, err)

	return st, errorCode
}

// Use message provided by semantic error (msg in the `semerr.WrapWithXXX(err, msg)`)
// as payload for GeneralError to do not leak internal errors to client.
func payloadFromSemantic(err *semerr.Error) string {
	return err.MessageWithFields()
}

func wrapWithSemantic(err error) *semerr.Error {
	if e := semerr.AsSemanticError(err); e == nil {
		return mapNonSemantic(err)
	} else {
		return e
	}
}

func mapNonSemantic(err error) *semerr.Error {
	if errors.Is(err, context.Canceled) {
		return semerr.WrapWithCanceled(err, "canceled")
	} else if errors.As(err, new(*ogenerrors.DecodeParamsError)) {
		//nolint:nolintlint
		//nolint:errlint
		return semerr.WrapWithInvalidInput(err, "invalid input: "+err.Error())
	} else if errors.As(err, new(*ogenerrors.DecodeRequestError)) {
		//nolint:nolintlint
		//nolint:errlint
		return semerr.WrapWithInvalidInput(err, "invalid request: "+err.Error())
	} else if parsed := (*ogenerrors.SecurityError)(nil); errors.As(err, &parsed) {
		return semerr.WrapWithAuthentication(err, parsed.Error())
	} else if parsed := semerrFromPGDriver(err); parsed != nil {
		return parsed
	} else if networkerrs.IsNetworkErrTranslatesToUnavailable(err) {
		return semerr.WrapWithUnavailable(err, "please try again later")
	}

	// Turn net.DNSError into a canceled error
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/lookup.go#L343
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/net.go#L421
	if dnsErr := asDNSError(err); dnsErr != nil && strings.Contains(dnsErr.Err, "operation was canceled") {
		return semerr.WrapWithCanceled(err, "dns lookup was canceled")
	}

	return semerr.WrapWithInternal(err, "unknown internal server error")
}

func asDNSError(err error) *net.DNSError {
	var target *net.DNSError
	if !errors.As(err, &target) {
		return nil
	}
	return target
}

// semerrFromPGDriver
// If it's an error from the pg driver, we try to map it to a semantic error.
func semerrFromPGDriver(err error) *semerr.Error {
	pgerr := &pgconn.PgError{}
	if !errors.As(err, &pgerr) {
		return nil
	}
	switch pgerr.Code {
	case pgerrcode.ConnectionFailure,
		pgerrcode.ConnectionException,
		pgerrcode.TooManyConnections,
		pgerrcode.CannotConnectNow:
		// transitional error
		return semerr.WrapWithUnavailable(err, "please try again later")
	default:
		return nil
	}
}
