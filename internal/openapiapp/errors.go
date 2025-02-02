package openapiapp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/yanakipe/bot/internal/semerr"
)

func WrapWithSemantic(err error) *semerr.Error {
	if e := semerr.AsSemanticError(err); e == nil {
		return MapNonSemantic(err)
	} else {
		return e
	}
}

func MapNonSemantic(err error) *semerr.Error {
	if errors.Is(err, context.Canceled) {
		return semerr.WrapWithCancelled(err, "cancelled")
	} else if parsed := (*ogenerrors.DecodeParamsError)(nil); errors.As(err, &parsed) {
		return semerr.WrapWithInvalidInput(err, fmt.Sprintf("invalid input: %s", parsed.Err))
	} else if parsed := (*ogenerrors.DecodeRequestError)(nil); errors.As(err, &parsed) {
		return semerr.WrapWithInvalidInput(err, fmt.Sprintf("invalid request: %s", parsed.Err))
	} else if parsed := (*ogenerrors.SecurityError)(nil); errors.As(err, &parsed) {
		return semerr.WrapWithAuthentication(err, parsed.Error())
	}

	// Turn net.DNSError into a cancelled error
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/lookup.go#L343
	// https://github.com/golang/go/blob/36cd880878a9804489557c29fa768647d665fbe0/src/net/net.go#L421
	if dnsErr := asDNSError(err); dnsErr != nil && strings.Contains(dnsErr.Err, "operation was canceled") {
		return semerr.WrapWithCancelled(err, "dns lookup was cancelled")
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
