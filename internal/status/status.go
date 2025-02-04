package status

import (
	"errors"
	"fmt"

	"github.com/yanakipre/bot/internal/status/codes"
	"github.com/yanakipre/bot/internal/status/details"
	"github.com/yanakipre/bot/internal/status/details/reason"
)

// Status represents the status code, error message, and details.
type Status struct {
	code    codes.Code
	err     error
	details details.ErrorDetails
}

// New creates a new Status with the given code and message.
func New(code codes.Code, message string) *Status {
	return &Status{
		code: code,
		err:  errors.New(message),
	}
}

// Newf creates a new Status with the given code and formatted message.
func Newf(code codes.Code, format string, args ...any) *Status {
	return &Status{
		code: code,
		err:  fmt.Errorf(format, args...),
	}
}

// Code returns the status code.
func (s *Status) Code() codes.Code {
	return s.code
}

// Message returns the error message of the underlying error.
func (s *Status) Message() string {
	return s.err.Error()
}

// Details returns the details of the status.
func (s *Status) Details() details.ErrorDetails {
	return s.details
}

// WithDetails sets the details of the status.
func (s *Status) WithDetails(details details.ErrorDetails) *Status {
	s.details = details

	return s
}

// WithErrorInfo sets the ErrorInfo detail of the status.
func (s *Status) WithErrorInfo(ei details.ErrorInfo) *Status {
	s.details.ErrorInfo = &ei

	return s
}

// WithReason sets the `Reason` field of the `ErrorInfo` detail of the status.
func (s *Status) WithReason(reason reason.Reason) *Status {
	if s.details.ErrorInfo == nil {
		s.details.ErrorInfo = &details.ErrorInfo{}
	}

	s.details.ErrorInfo.Reason = reason

	return s
}

// WithRetryDelay sets the RetryInfo detail of the status.
func (s *Status) WithRetryDelay(ms int) *Status {
	s.details.RetryInfo = &details.RetryInfo{
		RetryDelayMs: ms,
	}

	return s
}

// WithUserFacingMessage sets the UserFacingMessage detail of the status.
// Prefer to have a mapping of Reason -> UserFacingMessage in the presenter layer if possible,
// instead of scattering UserFacingMessages throughout the codebase.
// This makes it easier for the documentation team to review those,
// and serves as a clear list of error reasons we want to expose directly to users.
func (s *Status) WithUserFacingMessage(ufm string) *Status {
	s.details.UserFacingMessage = &details.UserFacingMessage{
		Message: ufm,
	}

	return s
}

// WithUserFacingMessagef sets the UserFacingMessage detail of the status with a formatted message.
// Prefer to have a mapping of Reason -> UserFacingMessage in the presenter layer if possible,
// instead of scattering UserFacingMessages throughout the codebase.
// This makes it easier for the documentation team to review those,
// and serves as a clear list of error reasons we want to expose directly to users.
func (s *Status) WithUserFacingMessagef(format string, args ...any) *Status {
	return s.WithUserFacingMessage(fmt.Sprintf(format, args...))
}

// Error wraps a Status in an Error.
func (s *Status) Error() error {
	return &Error{s: s}
}

// Error wraps a pointer to Status, and implements the error interface.
type Error struct {
	s *Status
}

// Error satisfies the `error` interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.s.code, e.s.Message())
}

// Status returns the underlying Status of an Error.
func (e *Error) Status() *Status {
	return e.s
}

// Unwrap returns the underlying error of the wrapped Status, allowing `errors.Is` and `errors.As` to work.
func (e *Error) Unwrap() error {
	return e.s.err
}
