package status

import (
	"errors"
	"net"
	"strings"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/yanakipe/bot/internal/semerr"
	"github.com/yanakipe/bot/internal/status/codes"
)

// FromError tries to create a new Status from an error, converting it if possible.
// If the error is not one this package is aware of, it will be wrapped with an Unknown code.
func FromError(err error) *Status {
	var sterr *Error
	if ok := errors.As(err, &sterr); ok {
		return sterr.Status()
	}

	var serr *semerr.Error
	if ok := errors.As(err, &serr); ok {
		return fromSemerr(serr)
	}

	var oerr ogenerrors.Error
	if ok := errors.As(err, &oerr); ok {
		return fromOgenerr(oerr)
	}

	// Turn net.DNSError into a cancelled error
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/lookup.go#L343
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/net.go#L421
	if st := fromDNSErr(err); st != nil {
		return st
	}

	return &Status{
		code: codes.Unknown,
		err:  err,
	}
}

func fromSemerr(err *semerr.Error) *Status {
	code := func(in semerr.Semantic) codes.Code {
		switch in {
		case semerr.SemanticAlreadyExists:
			return codes.AlreadyExists

		case semerr.SemanticAuthentication:
			return codes.Unauthenticated

		case semerr.SemanticCancelled:
			return codes.Canceled

		case semerr.SemanticFailedPrecondition:
			return codes.FailedPrecondition

		case semerr.SemanticForbidden:
			return codes.PermissionDenied

		case semerr.SemanticInternal:
			return codes.Internal

		case semerr.SemanticInvalidInput:
			return codes.InvalidArgument

		case semerr.SemanticNotFound:
			return codes.NotFound

		case semerr.SemanticNotImplemented:
			return codes.Unimplemented

		case semerr.SemanticResourceLocked:
			return codes.Locked

		case semerr.SemanticTooManyRequests:
			return codes.TooManyRequests

		case semerr.SemanticTimeout:
			return codes.DeadlineExceeded

		case semerr.SemanticUnavailable:
			return codes.Unavailable

		case semerr.SemanticUnprocessable:
			return codes.Unprocessable

		case semerr.SemanticNotAcceptable:
			return codes.NotAcceptable

		default:
			return codes.Unknown
		}
	}(err.Semantic)

	return &Status{
		code: code,
		err:  err,
	}
}

func fromOgenerr(err ogenerrors.Error) *Status {
	code := func(in ogenerrors.Error) codes.Code {
		switch in.(type) {
		case *ogenerrors.SecurityError:
			return codes.Unauthenticated

		case *ogenerrors.DecodeParamsError:
			return codes.InvalidArgument

		case *ogenerrors.DecodeRequestError:
			return codes.InvalidArgument

		default:
			return codes.Unknown
		}
	}(err)

	return &Status{
		code: code,
		err:  err,
	}
}

func fromDNSErr(err error) *Status {
	var dnserr *net.DNSError
	if !errors.As(err, &dnserr) {
		return nil
	}

	if !strings.Contains(dnserr.Err, "operation was canceled") {
		return nil
	}

	return WrapAsCanceled(err, "dns lookup was canceled")
}
