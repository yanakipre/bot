// Package codes defines the canonical error codes used by the `Status` message in our APIs.
//
// This package is based on the canonical error codes defined in the gRPC project:
// https://github.com/grpc/grpc-go/blob/7c377708dc070bc6f4dfeedfb3c3f38c3410912b/codes/codes.go
//
// The original package is licensed under Apache License, Version 2.0, copyright 2014 gRPC authors.
package codes

import (
	"fmt"
	"net/http"
)

// A Code is a status code for use with the `status` package.
//
// Only the codes defined as constants in this package are valid codes, do not use other code values.
// Codes are internally held as uint32, they are transformed to strings for the API presentation layer.
// They are not directly compatible with gRPC codes.
type Code uint32

const (
	// Unknown error.
	//
	// Errors raised by APIs that do not return enough error information may be converted to this error.
	Unknown Code = iota

	// Canceled indicates the operation was canceled (typically by the caller).
	//
	// Note on the spelling: "cancelled" is the British spelling, and "canceled" is the American spelling.
	// Since we tend to use the American spelling in our code, we use "canceled" here, as does Go itself.
	Canceled

	// InvalidArgument indicates that the client specified an invalid argument.
	//
	// Note that this differs from FailedPrecondition.
	// It indicates arguments that are problematic regardless of the state of the system
	// (e.g., a malformed endpoint ID).
	InvalidArgument

	// DeadlineExceeded means some process timed out before it could complete.
	//
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully.
	// For example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	DeadlineExceeded

	// NotFound means some requested entity (e.g., project or endpoint) was not found.
	NotFound

	// AlreadyExists means an attempt to create an entity failed because one already exists.
	AlreadyExists

	// PermissionDenied indicates the caller does not have permission to execute the specified request.
	//
	// It must not be used for rejections caused by exhausting some resource, use TooManyRequests instead.
	// It must not be used if the caller cannot be identified - use Unauthenticated instead.
	PermissionDenied

	// TooManyRequests indicates that the caller is being rate-limited.
	// for quota errors, prefer Locked.
	TooManyRequests

	// FailedPrecondition indicates that a request was rejected because the
	// system is not in a state required for the operation's execution.
	FailedPrecondition

	// Aborted indicates the operation was aborted,
	// typically due to a concurrency issue like sequencer check failures, transaction aborts, etc.
	Aborted

	// Unimplemented indicates operation is not implemented or not supported/enabled in this service.
	Unimplemented

	// Internal errors.
	// Means some invariants expected by the underlying system have been broken.
	// If you see one of these errors, something is very broken.
	Internal

	// Unavailable indicates the service is currently unavailable.
	Unavailable

	// Unauthenticated indicates the request does not have valid authentication credentials for the operation.
	Unauthenticated

	// Unprocessable indicates that the request contains well-formed (i.e., syntactically correct),
	// but semantically incorrect data.
	// This should rarely be used - usually there's a better code for the situation.
	// Included mostly for compatibility with the `semerr` package.
	Unprocessable

	// NotAcceptable indicates that the server cannot fulfil the request due to the client's Accept headers.
	// Prefer to use a different error code if the request is rejected because of the state of the system or because
	// of the request parameters.
	// Included mostly for compatibility with the `semerr` package.
	NotAcceptable

	// Locked indicates that the resource is locked, possibly because of a concurrent running operation,
	// or because per-user quota was exhausted.
	Locked
)

// Code implements the fmt.Stringer interface
var _ fmt.Stringer = Code(0)

// String returns the canonical representation of a `Code`.
func (c Code) String() string {
	switch c {
	case Unknown:
		return "UNKNOWN"
	case Canceled:
		return "CANCELED"
	case InvalidArgument:
		return "INVALID_ARGUMENT"
	case DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case NotFound:
		return "NOT_FOUND"
	case AlreadyExists:
		return "ALREADY_EXISTS"
	case PermissionDenied:
		return "PERMISSION_DENIED"
	case TooManyRequests:
		return "TOO_MANY_REQUESTS"
	case FailedPrecondition:
		return "FAILED_PRECONDITION"
	case Aborted:
		return "ABORTED"
	case Unimplemented:
		return "UNIMPLEMENTED"
	case Internal:
		return "INTERNAL"
	case Unavailable:
		return "UNAVAILABLE"
	case Unauthenticated:
		return "UNAUTHENTICATED"
	case Unprocessable:
		return "UNPROCESSABLE"
	case NotAcceptable:
		return "NOT_ACCEPTABLE"
	case Locked:
		return "LOCKED"
	default:
		// Should not happen, users of this package are not supposed to use codes outside of the defined ones.
		return fmt.Sprintf("CODE(%d)", c)
	}
}

// HTTP returns the HTTP status code for a given `Code`.
func (c Code) HTTP() int {
	switch c {
	case Unknown:
		return http.StatusInternalServerError
	case Canceled:
		// TODO(mattpodraza): document that this is not user-visible?
		return 499
	case InvalidArgument:
		return http.StatusBadRequest
	case DeadlineExceeded:
		return http.StatusRequestTimeout
	case NotFound:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case PermissionDenied:
		// Just like semerr, we "hide" 403s from the user to prevent information leakage.
		// Additionally, some clients rely on the existing error-to-status-code mapping.
		// See https://github.com/yanakipre/bot/pull/13481#discussion_r1611086997 for more context.
		return http.StatusNotFound
	case TooManyRequests:
		return http.StatusTooManyRequests
	case FailedPrecondition:
		return http.StatusPreconditionFailed
	case Aborted:
		return http.StatusConflict
	case Unimplemented:
		return http.StatusNotImplemented
	case Internal:
		return http.StatusInternalServerError
	case Unavailable:
		return http.StatusServiceUnavailable
	case Unauthenticated:
		return http.StatusUnauthorized
	case Unprocessable:
		return http.StatusUnprocessableEntity
	case NotAcceptable:
		return http.StatusNotAcceptable
	case Locked:
		return http.StatusLocked
	default:
		// Should not happen, users of this package are not supposed to use codes outside of the defined ones.
		return http.StatusInternalServerError
	}
}
