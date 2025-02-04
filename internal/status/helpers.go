package status

import (
	"fmt"

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

// NewUnknownf creates a new Status with the Unknown code, the given format, and the given arguments.
func NewUnknownf(format string, args ...any) *Status {
	return Newf(codes.Unknown, format, args...)
}

// NewCanceledf creates a new Status with the Canceled code, the given format, and the given arguments.
func NewCanceledf(format string, args ...any) *Status {
	return Newf(codes.Canceled, format, args...)
}

// NewInvalidArgumentf creates a new Status with the InvalidArgument code, the given format, and the given arguments.
func NewInvalidArgumentf(format string, args ...any) *Status {
	return Newf(codes.InvalidArgument, format, args...)
}

// NewDeadlineExceededf creates a new Status with the DeadlineExceeded code, the given format, and the given arguments.
func NewDeadlineExceededf(format string, args ...any) *Status {
	return Newf(codes.DeadlineExceeded, format, args...)
}

// NewNotFoundf creates a new Status with the NotFound code, the given format, and the given arguments.
func NewNotFoundf(format string, args ...any) *Status {
	return Newf(codes.NotFound, format, args...)
}

// NewAlreadyExistsf creates a new Status with the AlreadyExists code, the given format, and the given arguments.
func NewAlreadyExistsf(format string, args ...any) *Status {
	return Newf(codes.AlreadyExists, format, args...)
}

// NewPermissionDeniedf creates a new Status with the PermissionDenied code, the given format, and the given arguments.
func NewPermissionDeniedf(format string, args ...any) *Status {
	return Newf(codes.PermissionDenied, format, args...)
}

// NewTooManyRequestsf creates a new Status with the TooManyRequests code, the given format,
// and the given arguments.
func NewTooManyRequestsf(format string, args ...any) *Status {
	return Newf(codes.TooManyRequests, format, args...)
}

// NewFailedPreconditionf creates a new Status with the FailedPrecondition code, the given format,
// and the given arguments.
func NewFailedPreconditionf(format string, args ...any) *Status {
	return Newf(codes.FailedPrecondition, format, args...)
}

// NewAbortedf creates a new Status with the Aborted code, the given format, and the given arguments.
func NewAbortedf(format string, args ...any) *Status {
	return Newf(codes.Aborted, format, args...)
}

// NewUnimplementedf creates a new Status with the Unimplemented code, the given format, and the given arguments.
func NewUnimplementedf(format string, args ...any) *Status {
	return Newf(codes.Unimplemented, format, args...)
}

// NewInternalf creates a new Status with the Internal code, the given format, and the given arguments.
func NewInternalf(format string, args ...any) *Status {
	return Newf(codes.Internal, format, args...)
}

// NewUnavailablef creates a new Status with the Unavailable code, the given format, and the given arguments.
func NewUnavailablef(format string, args ...any) *Status {
	return Newf(codes.Unavailable, format, args...)
}

// NewUnauthenticatedf creates a new Status with the Unauthenticated code, the given format, and the given arguments.
func NewUnauthenticatedf(format string, args ...any) *Status {
	return Newf(codes.Unauthenticated, format, args...)
}

// NewUnprocessablef creates a new Status with the Unprocessable code, the given format, and the given arguments.
func NewUnprocessablef(format string, args ...any) *Status {
	return Newf(codes.Unprocessable, format, args...)
}

// NewNotAcceptablef creates a new Status with the NotAcceptable code, the given format, and the given arguments.
func NewNotAcceptablef(format string, args ...any) *Status {
	return Newf(codes.NotAcceptable, format, args...)
}

// NewLockedf creates a new Status with the Locked code, the given format, and the given arguments.
func NewLockedf(format string, args ...any) *Status {
	return Newf(codes.Locked, format, args...)
}

func wrapf(code codes.Code, err error, format string, args ...any) *Status {
	args = append(args, err)
	return &Status{
		code: code,
		err:  fmt.Errorf(format+": %w", args...),
	}
}

// WrapAsUnknown wraps the given error with the Unknown code and the given message.
func WrapAsUnknown(err error, message string) *Status {
	return wrapf(codes.Unknown, err, message)
}

// WrapAsUnknownf wraps the given error with the Unknown code,
// formatting the message with the given format and arguments.
func WrapAsUnknownf(err error, format string, args ...any) *Status {
	return wrapf(codes.Unknown, err, format, args...)
}

// WrapAsCanceled wraps the given error with the Canceled code and the given message.
func WrapAsCanceled(err error, message string) *Status {
	return wrapf(codes.Canceled, err, message)
}

// WrapAsCanceledf wraps the given error with the Canceled code,
// formatting the message with the given format and arguments.
func WrapAsCanceledf(err error, format string, args ...any) *Status {
	return wrapf(codes.Canceled, err, format, args...)
}

// WrapAsInvalidArgument wraps the given error with the InvalidArgument code and the given message.
func WrapAsInvalidArgument(err error, message string) *Status {
	return wrapf(codes.InvalidArgument, err, message)
}

// WrapAsInvalidArgumentf wraps the given error with the InvalidArgument code,
// formatting the message with the given format and arguments.
func WrapAsInvalidArgumentf(err error, format string, args ...any) *Status {
	return wrapf(codes.InvalidArgument, err, format, args...)
}

// WrapAsDeadlineExceeded wraps the given error with the DeadlineExceeded code and the given message.
func WrapAsDeadlineExceeded(err error, message string) *Status {
	return wrapf(codes.DeadlineExceeded, err, message)
}

// WrapAsDeadlineExceededf wraps the given error with the DeadlineExceeded code,
// formatting the message with the given format and arguments.
func WrapAsDeadlineExceededf(err error, format string, args ...any) *Status {
	return wrapf(codes.DeadlineExceeded, err, format, args...)
}

// WrapAsNotFound wraps the given error with the NotFound code and the given message.
func WrapAsNotFound(err error, message string) *Status {
	return wrapf(codes.NotFound, err, message)
}

// WrapAsNotFoundf wraps the given error with the NotFound code,
// formatting the message with the given format and arguments.
func WrapAsNotFoundf(err error, format string, args ...any) *Status {
	return wrapf(codes.NotFound, err, format, args...)
}

// WrapAsAlreadyExists wraps the given error with the AlreadyExists code and the given message.
func WrapAsAlreadyExists(err error, message string) *Status {
	return wrapf(codes.AlreadyExists, err, message)
}

// WrapAsAlreadyExistsf wraps the given error with the AlreadyExists code,
// formatting the message with the given format and arguments.
func WrapAsAlreadyExistsf(err error, format string, args ...any) *Status {
	return wrapf(codes.AlreadyExists, err, format, args...)
}

// WrapAsPermissionDenied wraps the given error with the PermissionDenied code and the given message.
func WrapAsPermissionDenied(err error, message string) *Status {
	return wrapf(codes.PermissionDenied, err, message)
}

// WrapAsPermissionDeniedf wraps the given error with the PermissionDenied code,
// formatting the message with the given format and arguments.
func WrapAsPermissionDeniedf(err error, format string, args ...any) *Status {
	return wrapf(codes.PermissionDenied, err, format, args...)
}

// WrapAsTooManyRequests wraps the given error with the TooManyRequests code and the given message.
func WrapAsTooManyRequests(err error, message string) *Status {
	return wrapf(codes.TooManyRequests, err, message)
}

// WrapAsTooManyRequestsf wraps the given error with the TooManyRequests code,
// formatting the message with the given format and arguments.
func WrapAsTooManyRequestsf(err error, format string, args ...any) *Status {
	return wrapf(codes.TooManyRequests, err, format, args...)
}

// WrapAsFailedPrecondition wraps the given error with the FailedPrecondition code and the given message.
func WrapAsFailedPrecondition(err error, message string) *Status {
	return wrapf(codes.FailedPrecondition, err, message)
}

// WrapAsFailedPreconditionf wraps the given error with the FailedPrecondition code,
// formatting the message with the given format and arguments.
func WrapAsFailedPreconditionf(err error, format string, args ...any) *Status {
	return wrapf(codes.FailedPrecondition, err, format, args...)
}

// WrapAsAborted wraps the given error with the Aborted code and the given message.
func WrapAsAborted(err error, message string) *Status {
	return wrapf(codes.Aborted, err, message)
}

// WrapAsAbortedf wraps the given error with the Aborted code,
// formatting the message with the given format and arguments.
func WrapAsAbortedf(err error, format string, args ...any) *Status {
	return wrapf(codes.Aborted, err, format, args...)
}

// WrapAsUnimplemented wraps the given error with the Unimplemented code and the given message.
func WrapAsUnimplemented(err error, message string) *Status {
	return wrapf(codes.Unimplemented, err, message)
}

// WrapAsUnimplementedf wraps the given error with the Unimplemented code,
// formatting the message with the given format and arguments.
func WrapAsUnimplementedf(err error, format string, args ...any) *Status {
	return wrapf(codes.Unimplemented, err, format, args...)
}

// WrapAsInternal wraps the given error with the Internal code and the given message.
func WrapAsInternal(err error, message string) *Status {
	return wrapf(codes.Internal, err, message)
}

// WrapAsInternalf wraps the given error with the Internal code,
// formatting the message with the given format and arguments.
func WrapAsInternalf(err error, format string, args ...any) *Status {
	return wrapf(codes.Internal, err, format, args...)
}

// WrapAsUnavailable wraps the given error with the Unavailable code and the given message.
func WrapAsUnavailable(err error, message string) *Status {
	return wrapf(codes.Unavailable, err, message)
}

// WrapAsUnavailablef wraps the given error with the Unavailable code,
// formatting the message with the given format and arguments.
func WrapAsUnavailablef(err error, format string, args ...any) *Status {
	return wrapf(codes.Unavailable, err, format, args...)
}

// WrapAsUnauthenticated wraps the given error with the Unauthenticated code and the given message.
func WrapAsUnauthenticated(err error, message string) *Status {
	return wrapf(codes.Unauthenticated, err, message)
}

// WrapAsUnauthenticatedf wraps the given error with the Unauthenticated code,
// formatting the message with the given format and arguments.
func WrapAsUnauthenticatedf(err error, format string, args ...any) *Status {
	return wrapf(codes.Unauthenticated, err, format, args...)
}

// WrapAsUnprocessable wraps the given error with the Unprocessable code and the given message.
func WrapAsUnprocessable(err error, message string) *Status {
	return wrapf(codes.Unprocessable, err, message)
}

// WrapAsUnprocessablef wraps the given error with the Unprocessable code,
// formatting the message with the given format and arguments.
func WrapAsUnprocessablef(err error, format string, args ...any) *Status {
	return wrapf(codes.Unprocessable, err, format, args...)
}

// WrapAsNotAcceptable wraps the given error with the NotAcceptable code and the given message.
func WrapAsNotAcceptable(err error, message string) *Status {
	return wrapf(codes.NotAcceptable, err, message)
}

// WrapAsNotAcceptablef wraps the given error with the NotAcceptable code,
// formatting the message with the given format and arguments.
func WrapAsNotAcceptablef(err error, format string, args ...any) *Status {
	return wrapf(codes.NotAcceptable, err, format, args...)
}

// WrapAsLocked wraps the given error with the Locked code and the given message.
func WrapAsLocked(err error, message string) *Status {
	return wrapf(codes.Locked, err, message)
}

// WrapAsLockedf wraps the given error with the Locked code,
// formatting the message with the given format and arguments.
func WrapAsLockedf(err error, format string, args ...any) *Status {
	return wrapf(codes.Locked, err, format, args...)
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
