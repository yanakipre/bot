package status

import (
	"github.com/yanakipre/bot/internal/status/codes"
)

// Constructor helpers

// NewUnknown creates a new Status with the Unknown code and the given message.
func NewUnknown(message string) *Status {
	return New(codes.Unknown, message)
}

// NewCanceled creates a new Status with the Canceled code and the given message.
func NewCanceled(message string) *Status {
	return New(codes.Canceled, message)
}

// NewInvalidArgument creates a new Status with the InvalidArgument code and the given message.
func NewInvalidArgument(message string) *Status {
	return New(codes.InvalidArgument, message)
}

// NewDeadlineExceeded creates a new Status with the DeadlineExceeded code and the given message.
func NewDeadlineExceeded(message string) *Status {
	return New(codes.DeadlineExceeded, message)
}

// NewNotFound creates a new Status with the NotFound code and the given message.
func NewNotFound(message string) *Status {
	return New(codes.NotFound, message)
}

// NewAlreadyExists creates a new Status with the AlreadyExists code and the given message.
func NewAlreadyExists(message string) *Status {
	return New(codes.AlreadyExists, message)
}

// NewPermissionDenied creates a new Status with the PermissionDenied code and the given message.
func NewPermissionDenied(message string) *Status {
	return New(codes.PermissionDenied, message)
}

// NewTooManyRequests creates a new Status with the TooManyRequests code and the given message.
func NewTooManyRequests(message string) *Status {
	return New(codes.TooManyRequests, message)
}

// NewFailedPrecondition creates a new Status with the FailedPrecondition code and the given message.
func NewFailedPrecondition(message string) *Status {
	return New(codes.FailedPrecondition, message)
}

// NewAborted creates a new Status with the Aborted code and the given message.
func NewAborted(message string) *Status {
	return New(codes.Aborted, message)
}

// NewUnimplemented creates a new Status with the Unimplemented code and the given message.
func NewUnimplemented(message string) *Status {
	return New(codes.Unimplemented, message)
}

// NewInternal creates a new Status with the Internal code and the given message.
func NewInternal(message string) *Status {
	return New(codes.Internal, message)
}

// NewUnavailable creates a new Status with the Unavailable code and the given message.
func NewUnavailable(message string) *Status {
	return New(codes.Unavailable, message)
}

// NewUnauthenticated creates a new Status with the Unauthenticated code and the given message.
func NewUnauthenticated(message string) *Status {
	return New(codes.Unauthenticated, message)
}

// NewUnprocessable creates a new Status with the Unprocessable code and the given message.
func NewUnprocessable(message string) *Status {
	return New(codes.Unprocessable, message)
}

// NewNotAcceptable creates a new Status with the NotAcceptable code and the given message.
func NewNotAcceptable(message string) *Status {
	return New(codes.NotAcceptable, message)
}

// NewLocked creates a new Status with the Locked code and the given message.
func NewLocked(message string) *Status {
	return New(codes.Locked, message)
}

// Status helpers

// IsUnknown returns true if the status code is `codes.Unknown`.
func (s *Status) IsUnknown() bool {
	return s.code == codes.Unknown
}

// IsCanceled returns true if the status code is `codes.Canceled`.
func (s *Status) IsCanceled() bool {
	return s.code == codes.Canceled
}

// IsInvalidArgument returns true if the status code is `codes.InvalidArgument`.
func (s *Status) IsInvalidArgument() bool {
	return s.code == codes.InvalidArgument
}

// IsDeadlineExceeded returns true if the status code is `codes.DeadlineExceeded`.
func (s *Status) IsDeadlineExceeded() bool {
	return s.code == codes.DeadlineExceeded
}

// IsNotFound returns true if the status code is `codes.NotFound`.
func (s *Status) IsNotFound() bool {
	return s.code == codes.NotFound
}

// IsAlreadyExists returns true if the status code is `codes.AlreadyExists`.
func (s *Status) IsAlreadyExists() bool {
	return s.code == codes.AlreadyExists
}

// IsPermissionDenied returns true if the status code is `codes.PermissionDenied`.
func (s *Status) IsPermissionDenied() bool {
	return s.code == codes.PermissionDenied
}

// IsTooManyRequests returns true if the status code is `codes.TooManyRequests`.
func (s *Status) IsTooManyRequests() bool {
	return s.code == codes.TooManyRequests
}

// IsFailedPrecondition returns true if the status code is `codes.FailedPrecondition`.
func (s *Status) IsFailedPrecondition() bool {
	return s.code == codes.FailedPrecondition
}

// IsAborted returns true if the status code is `codes.Aborted`.
func (s *Status) IsAborted() bool {
	return s.code == codes.Aborted
}

// IsUnimplemented returns true if the status code is `codes.Unimplemented`.
func (s *Status) IsUnimplemented() bool {
	return s.code == codes.Unimplemented
}

// IsInternal returns true if the status code is `codes.Internal`.
func (s *Status) IsInternal() bool {
	return s.code == codes.Internal
}

// IsUnavailable returns true if the status code is `codes.Unavailable`.
func (s *Status) IsUnavailable() bool {
	return s.code == codes.Unavailable
}

// IsUnauthenticated returns true if the status code is `codes.Unauthenticated`.
func (s *Status) IsUnauthenticated() bool {
	return s.code == codes.Unauthenticated
}

// IsUnprocessable returns true if the status code is `codes.Unprocessable`.
func (s *Status) IsUnprocessable() bool {
	return s.code == codes.Unprocessable
}

// IsNotAcceptable returns true if the status code is `codes.NotAcceptable`.
func (s *Status) IsNotAcceptable() bool {
	return s.code == codes.NotAcceptable
}

// IsLocked returns true if the status code is `codes.Locked`.
func (s *Status) IsLocked() bool {
	return s.code == codes.Locked
}
