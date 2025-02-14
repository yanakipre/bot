package status

import (
	"context"
	"errors"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/status/codes"
	"github.com/yanakipre/bot/internal/status/networkerrs"
	"net"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ogen-go/ogen/ogenerrors"
)

// FromError tries to create a new Status from an error, converting it if possible.
// If the error is not one this package is aware of, it will be wrapped with an Unknown code.
func FromError(err error) *Status {
	// If there's a Status somewhere down the chain, use it.
	var sterr *Error
	if ok := errors.As(err, &sterr); ok {
		return sterr.Status()
	}

	// Otherwise, try getting the code from some of the known error types that might be anywhere in the chain,
	// but wrap the original error.
	code := func() codes.Code {
		if ok, code := codeFromContext(err); ok {
			return code
		}

		if ok, code := codeFromDNSErr(err); ok {
			return code
		}

		if ok, code := codeFromSemerr(err); ok {
			return code
		}

		if ok, code := codeFromOgenerr(err); ok {
			return code
		}

		if ok, code := codeFromPGDriver(err); ok {
			return code
		}

		if ok, code := codeFromNetworkErr(err); ok {
			return code
		}

		return codes.Unknown
	}()

	return &Status{
		code: code,
		err:  err,
	}
}

// codeFromPGDriver
// Check if the error is a Postgres driver error
// if yes, then connection errors are treated as temporary and should be retried by the client.
func codeFromPGDriver(err error) (bool, codes.Code) {
	pgerr := &pgconn.PgError{}
	if !errors.As(err, &pgerr) {
		return false, codes.Unknown
	}
	switch pgerr.Code {
	case pgerrcode.ConnectionFailure, pgerrcode.ConnectionException:
		// transitional error
		return true, codes.Unavailable
	case pgerrcode.TooManyConnections:
		// https://www.metisdata.io/knowledgebase/errors/postgresql-53300
		// transitional error
		return true, codes.Unavailable
	case pgerrcode.CannotConnectNow:
		// https://www.metisdata.io/knowledgebase/errors/postgresql-57p03
		// transitional error
		return true, codes.Unavailable
	default:
		return true, codes.Unknown
	}
}

// codeFromNetworkErr
// Check if the error is a network error
// if yes, then temporary connection errors to other services are treated as temporary
// and should be retried by the client.
func codeFromNetworkErr(err error) (bool, codes.Code) {
	if networkerrs.IsNetworkErrTranslatesToUnavailable(err) {
		return true, codes.Unavailable
	}
	return false, codes.Unknown
}

func codeFromSemerr(err error) (bool, codes.Code) {
	var serr *semerr.Error
	if ok := errors.As(err, &serr); !ok {
		return false, codes.Unknown
	}

	switch serr.Semantic {
	case semerr.SemanticAlreadyExists:
		return true, codes.AlreadyExists

	case semerr.SemanticAuthentication:
		return true, codes.Unauthenticated

	case semerr.SemanticCanceled:
		return true, codes.Canceled

	case semerr.SemanticFailedPrecondition:
		return true, codes.FailedPrecondition

	case semerr.SemanticForbidden:
		return true, codes.PermissionDenied

	case semerr.SemanticInternal:
		return true, codes.Internal

	case semerr.SemanticInvalidInput:
		return true, codes.InvalidArgument

	case semerr.SemanticNotFound:
		return true, codes.NotFound

	case semerr.SemanticNotImplemented:
		return true, codes.Unimplemented

	case semerr.SemanticResourceLocked:
		return true, codes.Locked

	case semerr.SemanticTooManyRequests:
		return true, codes.TooManyRequests

	case semerr.SemanticTimeout:
		return true, codes.DeadlineExceeded

	case semerr.SemanticUnavailable:
		return true, codes.Unavailable

	case semerr.SemanticUnprocessable:
		return true, codes.Unprocessable

	case semerr.SemanticNotAcceptable:
		return true, codes.NotAcceptable

	default:
		return false, codes.Unknown
	}
}

func codeFromOgenerr(err error) (bool, codes.Code) {
	var oerr ogenerrors.Error
	if ok := errors.As(err, &oerr); !ok {
		return false, codes.Unknown
	}

	switch oerr.(type) {
	case *ogenerrors.SecurityError:
		return true, codes.Unauthenticated

	case *ogenerrors.DecodeParamsError:
		return true, codes.InvalidArgument

	case *ogenerrors.DecodeRequestError:
		return true, codes.InvalidArgument

	default:
		return false, codes.Unknown
	}
}

func codeFromDNSErr(err error) (bool, codes.Code) {
	// Turn net.DNSError into a canceled error
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/lookup.go#L343
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/net.go#L421

	var dnserr *net.DNSError
	if !errors.As(err, &dnserr) {
		return false, codes.Unknown
	}

	if !strings.Contains(dnserr.Err, "operation was canceled") {
		return false, codes.Unknown
	}

	return true, codes.Canceled
}

func codeFromContext(err error) (bool, codes.Code) {
	if errors.Is(err, context.Canceled) {
		return true, codes.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true, codes.DeadlineExceeded
	}

	return false, codes.Unknown
}
