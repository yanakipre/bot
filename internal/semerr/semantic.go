package semerr

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// Semantic meaning of the error
type Semantic int

// Known error meanings
const (
	SemanticUnknown Semantic = iota
	SemanticNotImplemented
	SemanticUnavailable
	SemanticInvalidInput
	SemanticTimeout
	SemanticInternal
	SemanticAuthentication
	SemanticForbidden
	SemanticFailedPrecondition
	SemanticNotFound
	SemanticAlreadyExists
	SemanticCancelled
	SemanticUnprocessable
	SemanticNotAcceptable
	SemanticResourceLocked
	SemanticTooManyRequests
	SemanticSkipError
	SemanticPartialSuccess
)

// Error with semantic info
type Error struct {
	stack *stack

	Message  string
	Semantic Semantic
	Details  any

	err error
}

func (e *Error) StackTrace() StackTrace {
	if e.stack == nil {
		return StackTrace{}
	}
	// We do not return the newest Frame, because it's an error creation.
	return e.stack.StackTrace()[1:]
}

// Sentry supports https://github.com/pkg/errors for stacktracing,
// https://docs.sentry.io/platforms/go/usage/#capturing-errors
// We're using same interface.
type stackTracer interface {
	StackTrace() StackTrace
}

var _ stackTracer = &Error{}

func errFmt(text string, args ...any) error {
	return xerrors.Errorf(text, args...) // only xerrors return type as we need
}

func newError(s Semantic, text string) *Error {
	return &Error{
		Semantic: s,
		Message:  text,
		err:      errFmt(text),
		stack:    callers(),
	}
}

// newErrorf constructs error with formatting
func newErrorf(s Semantic, format string, a ...any) *Error {
	return &Error{
		Semantic: s,
		Message:  fmt.Sprintf(format, a...),
		err:      errFmt(format, a...),
		stack:    callers(),
	}
}

func wrapError(s Semantic, err error, text string) *Error {
	return &Error{
		Semantic: s,
		Message:  text,
		err:      errFmt(text+": %w", err),
		stack:    callers(),
	}
}

func wrapErrorf(s Semantic, err error, format string, a ...any) *Error {
	args := a
	args = append(args, err)
	return &Error{
		Semantic: s,
		Message:  fmt.Sprintf(format, a...),
		err:      errFmt(format+": %w", args...),
		stack:    callers(),
	}
}

// Error implements error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s", e)
}

func (e *Error) Format(s fmt.State, v rune) {
	e.err.(fmt.Formatter).Format(s, v)
}

func (e *Error) Is(err error) bool {
	return err == e.err
}

func (e *Error) As(target any) bool {
	return errors.As(e.err, target)
}

// Unwrap implements Wrapper interface
func (e *Error) Unwrap() error {
	return e.err
}

// IsSemanticError returns true if target error is specified semantic error
func IsSemanticError(err error, semantic Semantic) bool {
	target := AsSemanticError(err)
	if target == nil {
		return false
	}

	return target.Semantic == semantic
}

// AsSemanticError returns semantic error if there is one
func AsSemanticError(err error) *Error {
	var target *Error
	if !errors.As(err, &target) {
		return nil
	}

	return target
}

// unknown constructs Unknown error
//
//nolint:deadcode,unused
func unknown(text string) *Error {
	return newError(SemanticUnknown, text)
}

// unknownf constructs Unknown error with formatting
//
//nolint:deadcode,unused
func unknownf(format string, a ...any) *Error {
	return newErrorf(SemanticUnknown, format, a...)
}

// wrapWithUnknown constructs Unknown error which wraps provided error
func wrapWithUnknown(err error, text string) *Error {
	return wrapError(SemanticUnknown, err, text)
}

// wrapWithCancelled constructs Cancelled error which wraps provided error
func WrapWithCancelled(err error, text string) *Error {
	return wrapError(SemanticCancelled, err, text)
}

// wrapWithUnknownf constructs Unknown error with formatting which wraps provided error
//
//nolint:deadcode,unused
func wrapWithUnknownf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticUnknown, err, format, a...)
}

// isUnknown returns true if target error is Unknown semantic error
//
//nolint:deadcode,unused
func isUnknown(err error) bool {
	return IsSemanticError(err, SemanticUnknown)
}

// Internal constructs Internal error
func Internal(text string) *Error {
	return newError(SemanticInternal, text)
}

// Internalf constructs Internal error with formatting
func Internalf(format string, a ...any) *Error {
	return newErrorf(SemanticInternal, format, a...)
}

// InternalWithFields constructs Internal error with fields
func InternalWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticInternal, WithFields(text, fields...), "internal")
}

// WrapWithInternal constructs Internal error which wraps provided error
func WrapWithInternal(err error, text string) *Error {
	return wrapError(SemanticInternal, err, text)
}

// WrapWithInternal constructs Internal error with formatting which wraps provided
// error
func WrapWithInternalf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticInternal, err, format, a...)
}

// NotImplemented constructs NotImplemented error
func NotImplemented(text string) *Error {
	return newError(SemanticNotImplemented, text)
}

// NotImplementedf constructs NotImplemented error with formatting
func NotImplementedf(format string, a ...any) *Error {
	return newErrorf(SemanticNotImplemented, format, a...)
}

// NotImplementedWithFields constructs NotImplemented error with fields
func NotImplementedWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotImplemented, WithFields(text, fields...), "not implemented")
}

// WrapWithNotImplemented constructs NotImplemented error which wraps provided error
func WrapWithNotImplemented(err error, text string) *Error {
	return wrapError(SemanticNotImplemented, err, text)
}

// WrapWithNotImplementedf constructs NotImplemented error with formatting which wraps provided
// error
func WrapWithNotImplementedf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticNotImplemented, err, format, a...)
}

// IsNotImplemented returns true if target error is NotImplemented semantic error
func IsNotImplemented(err error) bool {
	return IsSemanticError(err, SemanticNotImplemented)
}

// Unavailable constructs Unavailable error
func Unavailable(text string) *Error {
	return newError(SemanticUnavailable, text)
}

// Unavailablef constructs Unavailable error with formatting
func Unavailablef(format string, a ...any) *Error {
	return newErrorf(SemanticUnavailable, format, a...)
}

// UnavailableWithFields constructs Unavailable error with fields
func UnavailableWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticUnavailable, WithFields(text, fields...), "unavailable")
}

// WrapWithUnavailable constructs Unavailable error which wraps provided error
func WrapWithUnavailable(err error, text string) *Error {
	return wrapError(SemanticUnavailable, err, text)
}

// Timeout constructs Timeout error
func Timeout(text string) *Error {
	return newError(SemanticTimeout, text)
}

func Timeoutf(format string, a ...any) *Error {
	return newErrorf(SemanticTimeout, format, a...)
}

// TimeoutWithFields constructs Unavailable error with fields
func TimeoutWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticTimeout, WithFields(text, fields...), "timeout")
}

// WrapWithTimeout constructs Timeout error which wraps provided error
func WrapWithTimeout(err error, text string) *Error {
	return wrapError(SemanticTimeout, err, text)
}

func IsTimeout(err error) bool {
	return IsSemanticError(err, SemanticTimeout)
}

func ResourceLocked(text string) *Error {
	return newError(SemanticResourceLocked, text)
}

// WrapWithResourceLocked constructs ResourceLocked error which wraps provided error
func WrapWithResourceLocked(err error, text string) *Error {
	return wrapError(SemanticResourceLocked, err, text)
}

func IsResourceLocked(err error) bool {
	return IsSemanticError(err, SemanticResourceLocked)
}

// IsUnavailable returns true if target error is Unavailable semantic error
func IsUnavailable(err error) bool {
	return IsSemanticError(err, SemanticUnavailable)
}

// IsInternal returns true if target error is Internal semantic error
func IsInternal(err error) bool {
	return IsSemanticError(err, SemanticInternal)
}

// InvalidInput constructs InvalidInput error
func InvalidInput(text string) *Error {
	return newError(SemanticInvalidInput, text)
}

// InvalidInputf constructs InvalidInput error with formatting
func InvalidInputf(format string, a ...any) *Error {
	return newErrorf(SemanticInvalidInput, format, a...)
}

// InvalidInputWithFields constructs InvalidInput error with fields
func InvalidInputWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticInvalidInput, WithFields(text, fields...), "invalid input")
}

// WrapWithInvalidInput constructs InvalidInput error which wraps provided error
func WrapWithInvalidInput(err error, text string) *Error {
	return wrapError(SemanticInvalidInput, err, text)
}

// WrapWithInvalidInputf constructs InvalidInput error with formatting which wraps provided error
func WrapWithInvalidInputf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticInvalidInput, err, format, a...)
}

// IsInvalidInput returns true if target error is InvalidInput semantic error
func IsInvalidInput(err error) bool {
	return IsSemanticError(err, SemanticInvalidInput)
}

// Authentication constructs Authentication error
func Authentication(text string) *Error {
	return newError(SemanticAuthentication, text)
}

// Authenticationf constructs Authentication error with formatting
func Authenticationf(format string, a ...any) *Error {
	return newErrorf(SemanticAuthentication, format, a...)
}

// AuthenticationWithFields constructs Authentication error with fields
func AuthenticationWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticAuthentication, WithFields(text, fields...), "authentication")
}

// WrapWithAuthentication constructs Authentication error which wraps provided error
func WrapWithAuthentication(err error, text string) *Error {
	return wrapError(SemanticAuthentication, err, text)
}

// WrapWithAuthenticationf constructs Authentication error with formatting which wraps provided
// error
func WrapWithAuthenticationf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticAuthentication, err, format, a...)
}

// IsAuthentication returns true if target error is Authentication semantic error
func IsAuthentication(err error) bool {
	return IsSemanticError(err, SemanticAuthentication)
}

// Forbidden constructs Forbidden error
func Forbidden(text string) *Error {
	return newError(SemanticForbidden, text)
}

// Forbiddenf constructs Forbidden error with formatting
func Forbiddenf(format string, a ...any) *Error {
	return newErrorf(SemanticForbidden, format, a...)
}

// ForbiddenWithFields constructs Forbidden error with fields
func ForbiddenWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticForbidden, WithFields(text, fields...), "forbidden")
}

// WrapWithForbidden constructs Forbidden error which wraps provided error
func WrapWithForbidden(err error, text string) *Error {
	return wrapError(SemanticForbidden, err, text)
}

// WrapWithForbiddenf constructs Forbidden error with formatting which wraps provided error
func WrapWithForbiddenf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticForbidden, err, format, a...)
}

// IsForbidden returns true if target error is Forbidden semantic error
func IsForbidden(err error) bool {
	return IsSemanticError(err, SemanticForbidden)
}

// FailedPrecondition constructs FailedPrecondition error
func FailedPrecondition(text string) *Error {
	return newError(SemanticFailedPrecondition, text)
}

// FailedPreconditionf constructs FailedPrecondition error with formatting
func FailedPreconditionf(format string, a ...any) *Error {
	return newErrorf(SemanticFailedPrecondition, format, a...)
}

// FailedPreconditionWithFields constructs FailedPrecondition error with fields
func FailedPreconditionWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticFailedPrecondition, WithFields(text, fields...), "failed precondition")
}

// WrapWithFailedPrecondition constructs FailedPrecondition error which wraps provided error
func WrapWithFailedPrecondition(err error, text string) *Error {
	return wrapError(SemanticFailedPrecondition, err, text)
}

// WrapWithFailedPreconditionf constructs FailedPrecondition error
// with formatting which wraps provided error
func WrapWithFailedPreconditionf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticFailedPrecondition, err, format, a...)
}

// IsFailedPrecondition returns true if target error is FailedPrecondition semantic error
func IsFailedPrecondition(err error) bool {
	return IsSemanticError(err, SemanticFailedPrecondition)
}

// NotFound constructs NotFound error
func NotFound(text string) *Error {
	return newError(SemanticNotFound, text)
}

// NotFoundf constructs NotFound error with formatting
func NotFoundf(format string, a ...any) *Error {
	return newErrorf(SemanticNotFound, format, a...)
}

// NotFoundWithFields constructs NotFound error with fields
func NotFoundWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotFound, WithFields(text, fields...), "not found")
}

// WrapWithNotFound constructs NotFound error which wraps provided error
func WrapWithNotFound(err error, text string) *Error {
	return wrapError(SemanticNotFound, err, text)
}

// WrapWithNotFoundf constructs NotFound error with formatting which wraps provided error
func WrapWithNotFoundf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticNotFound, err, format, a...)
}

// IsNotFound returns true if target error is NotFound semantic error
func IsNotFound(err error) bool {
	return IsSemanticError(err, SemanticNotFound)
}

// AlreadyExists constructs AlreadyExists error
func AlreadyExists(text string) *Error {
	return newError(SemanticAlreadyExists, text)
}

// AlreadyExistsf constructs AlreadyExists error with formatting
func AlreadyExistsf(format string, a ...any) *Error {
	return newErrorf(SemanticAlreadyExists, format, a...)
}

// AlreadyExistsWithFields constructs AlreadyExists error with fields
func AlreadyExistsWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticAlreadyExists, WithFields(text, fields...), "already exists")
}

// WrapWithAlreadyExists constructs AlreadyExists error which wraps provided error
func WrapWithAlreadyExists(err error, text string) *Error {
	return wrapError(SemanticAlreadyExists, err, text)
}

// WrapWithAlreadyExistsf constructs AlreadyExists error with formatting which wraps provided error
func WrapWithAlreadyExistsf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticAlreadyExists, err, format, a...)
}

// IsAlreadyExists returns true if target error is AlreadyExists semantic error
func IsAlreadyExists(err error) bool {
	return IsSemanticError(err, SemanticAlreadyExists)
}

// WhitelistErrors provided in argument, convert others into unknown errors
func WhitelistErrors(err error, semantics ...Semantic) error {
	se := AsSemanticError(err)
	if se == nil {
		return err
	}

	for _, sem := range semantics {
		if se.Semantic == sem {
			return err
		}
	}

	// Do not return any semantic error beyond those specified above
	return wrapWithUnknown(err, "unknown")
}

// Unprocessable constructs Unprocessable error
func Unprocessable(text string) *Error {
	return newError(SemanticUnprocessable, text)
}

// Unprocessablef constructs Unprocessable error with formatting
func Unprocessablef(format string, a ...any) *Error {
	return newErrorf(SemanticUnprocessable, format, a...)
}

// UnprocessableWithFields constructs Unprocessable error with fields
func UnprocessableWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticUnprocessable, WithFields(text, fields...), "unprocessable")
}

// WrapWithUnprocessable constructs Unprocessable error which wraps provided error
func WrapWithUnprocessable(err error, text string) *Error {
	return wrapError(SemanticUnprocessable, err, text)
}

// WrapWithUnprocessablef constructs Unprocessable error with formatting which wraps provided error
func WrapWithUnprocessablef(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticUnprocessable, err, format, a...)
}

// IsUnprocessable returns true if target error is Unprocessable semantic error
func IsUnprocessable(err error) bool {
	return IsSemanticError(err, SemanticUnprocessable)
}

// NotAcceptable constructs NotAcceptable error
func NotAcceptable(text string) *Error {
	return newError(SemanticNotAcceptable, text)
}

// NotAcceptablef constructs NotAcceptable error with formatting
func NotAcceptablef(format string, a ...any) *Error {
	return newErrorf(SemanticNotAcceptable, format, a...)
}

// NotAcceptableWithFields constructs NotAcceptable error with fields
func NotAcceptableWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotAcceptable, WithFields(text, fields...), "not acceptable")
}

// WrapWithNotAcceptable constructs NotAcceptable error which wraps provided error
func WrapWithNotAcceptable(err error, text string) *Error {
	return wrapError(SemanticNotAcceptable, err, text)
}

// WrapWithNotAcceptablef constructs NotAcceptable error with formatting which wraps provided error
func WrapWithNotAcceptablef(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticNotAcceptable, err, format, a...)
}

// IsNotAcceptable returns true if target error is NotAcceptable semantic error
func IsNotAcceptable(err error) bool {
	return IsSemanticError(err, SemanticNotAcceptable)
}

// TooManyRequests constructs TooManyRequests error
func TooManyRequests(text string) *Error {
	return newError(SemanticTooManyRequests, text)
}

// TooManyRequestsf constructs TooManyRequests error with formatting
func TooManyRequestsf(format string, a ...any) *Error {
	return newErrorf(SemanticTooManyRequests, format, a...)
}

// TooManyRequestsWithFields constructs TooManyRequests error with fields
func TooManyRequestsWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticTooManyRequests, WithFields(text, fields...), "too many requests")
}

// WrapWithTooManyRequests constructs TooManyRequests error which wraps provided error
func WrapWithTooManyRequests(err error, text string) *Error {
	return wrapError(SemanticTooManyRequests, err, text)
}

// WrapWithTooManyRequestsf constructs TooManyRequests error with formatting which wraps provided
// error
func WrapWithTooManyRequestsf(err error, format string, a ...any) *Error {
	return wrapErrorf(SemanticTooManyRequests, err, format, a...)
}

// IsTooManyRequests returns true if target error is TooManyRequests semantic error
func IsTooManyRequests(err error) bool {
	return IsSemanticError(err, SemanticTooManyRequests)
}

func IsSkipError(err error) bool {
	return IsSemanticError(err, SemanticSkipError)
}

func WrapWithSkipError(err error) *Error {
	return wrapError(SemanticSkipError, err, err.Error())
}

// PartialSuccessf constructs PartialSuccess error with formatting
func PartialSuccessf(format string, a ...any) *Error {
	return newErrorf(SemanticPartialSuccess, format, a...)
}

// PartialSuccessWithFields constructs PartialSuccess error with fields
func PartialSuccessWithFields(text string, fields ...zap.Field) *Error {
	return wrapError(SemanticPartialSuccess, WithFields(text, fields...), "partial success")
}

// IsPartialSuccess returns true if target error is PartialSuccessf semantic error
func IsPartialSuccess(err error) bool {
	return IsSemanticError(err, SemanticPartialSuccess)
}

type errorWithFields struct {
	error
	fields []zap.Field
}

func (err errorWithFields) Unwrap() error {
	return err.error
}

func WithFields(text string, fields ...zap.Field) error {
	return WrapWithFields(errors.New(text), fields...)
}

func WrapWithFields(err error, fields ...zap.Field) error {
	if len(fields) == 0 {
		return err
	}
	return errorWithFields{
		error:  err,
		fields: fields,
	}
}

func UnwrapFields(err error) []zap.Field {
	fields := make(map[string]zap.Field)
	errs := []error{err}
	for len(errs) > 0 {
		e := errs[0]
		errs = errs[1:]
		if withFields, ok := e.(errorWithFields); ok {
			// deduplicate keys added at different call tree levels
			// first write wins, i.e. write at the lowest wrapping level
			for _, f := range withFields.fields {
				fields[f.Key] = f
			}
		}
		if uw, ok := e.(interface{ Unwrap() error }); ok {
			errs = append(errs, uw.Unwrap())
		}
		if uw, ok := e.(interface{ Unwrap() []error }); ok {
			errs = append(errs, uw.Unwrap()...)
		}
	}
	return lo.MapToSlice(fields, func(_ string, f zap.Field) zap.Field {
		return f
	})
}

func UnwrapPanic(p any) error {
	if p == nil {
		return Internal("panic")
	}
	err, ok := p.(error)
	if !ok {
		//nolint:nolintlint
		//nolint:errlint
		return WrapWithInternal(fmt.Errorf("%v", p), "panic")
	}
	return err
}
