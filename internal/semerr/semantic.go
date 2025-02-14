package semerr

import (
	"errors"
	"fmt"
	"github.com/yanakipre/bot/internal/clouderr"
	"strings"

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
	SemanticCanceled
	SemanticUnprocessable
	SemanticNotAcceptable
	SemanticResourceLocked
	SemanticTooManyRequests
	SemanticSkipError
	SemanticPartialSuccess
)

// Error with semantic info
type Error struct {
	Semantic Semantic
	err      error
	fields   []zap.Field
	message  string
	stack    *stack
}

func (e *Error) Fields() []zap.Field {
	return e.fields
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

func newError(s Semantic, text string, fields ...zap.Field) *Error {
	return &Error{
		Semantic: s,
		message:  text,
		fields:   fields,
		err:      errFmt(text),
		stack:    callers(),
	}
}

// newErrorf constructs error with formatting
//
//nolint:deadcode,unused
func newErrorf(s Semantic, format string, a ...any) *Error {
	return &Error{
		Semantic: s,
		message:  fmt.Sprintf(format, a...),
		err:      errFmt(format, a...),
		stack:    callers(),
	}
}

func wrapError(s Semantic, err error, text string, fields ...zap.Field) *Error {
	return &Error{
		Semantic: s,
		message:  text,
		fields:   fields,
		err:      errFmt(text+": %w", err),
		stack:    callers(),
	}
}

func wrapErrorf(s Semantic, err error, format string, a ...any) *Error {
	args := a
	args = append(args, err)
	return &Error{
		Semantic: s,
		message:  fmt.Sprintf(format, a...),
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

func (e *Error) Message() string {
	return e.message
}

func (e *Error) MessageWithFields() string {
	ff := clouderr.UnwrapFields(e)
	if len(ff) == 0 {
		return e.Message()
	}
	ss := lo.Map(ff, func(f zap.Field, _ int) string {
		return clouderr.FieldToString(f)
	})
	return fmt.Sprintf("%s; %s", e.Message(), strings.Join(ss, ", "))
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
//
//nolint:deadcode,unused
func wrapWithUnknown(err error, text string) *Error {
	return wrapError(SemanticUnknown, err, text)
}

// wrapWithCancelled constructs Canceled error which wraps provided error
func WrapWithCanceled(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticCanceled, err, text, fields...)
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
func Internal(text string, fields ...zap.Field) *Error {
	return newError(SemanticInternal, text, fields...)
}

// WrapWithInternal constructs Internal error which wraps provided error
func WrapWithInternal(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticInternal, err, text, fields...)
}

// NotImplemented constructs NotImplemented error
func NotImplemented(text string, fields ...zap.Field) *Error {
	return newError(SemanticNotImplemented, text, fields...)
}

// WrapWithNotImplemented constructs NotImplemented error which wraps provided error
func WrapWithNotImplemented(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotImplemented, err, text, fields...)
}

// IsNotImplemented returns true if target error is NotImplemented semantic error
func IsNotImplemented(err error) bool {
	return IsSemanticError(err, SemanticNotImplemented)
}

// Unavailable constructs Unavailable error
func Unavailable(text string, fields ...zap.Field) *Error {
	return newError(SemanticUnavailable, text, fields...)
}

// WrapWithUnavailable constructs Unavailable error which wraps provided error
func WrapWithUnavailable(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticUnavailable, err, text, fields...)
}

// WrapWithTimeout constructs Timeout error which wraps provided error
func WrapWithTimeout(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticTimeout, err, text, fields...)
}

// ResourceLocked constructs ResourceLocked error
func ResourceLocked(text string, fields ...zap.Field) *Error {
	return newError(SemanticResourceLocked, text, fields...)
}

// WrapWithResourceLocked constructs ResourceLocked error which wraps provided error
func WrapWithResourceLocked(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticResourceLocked, err, text, fields...)
}

// WrapWithTooManRequests constructs TooManyRequests error which wraps provided error
func WrapWithTooManyRequests(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticTooManyRequests, err, text, fields...)
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
func InvalidInput(text string, fields ...zap.Field) *Error {
	return newError(SemanticInvalidInput, text, fields...)
}

// WrapWithInvalidInput constructs InvalidInput error which wraps provided error
func WrapWithInvalidInput(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticInvalidInput, err, text, fields...)
}

// IsInvalidInput returns true if target error is InvalidInput semantic error
func IsInvalidInput(err error) bool {
	return IsSemanticError(err, SemanticInvalidInput)
}

// Authentication constructs Authentication error
func Authentication(text string, fields ...zap.Field) *Error {
	return newError(SemanticAuthentication, text, fields...)
}

// WrapWithAuthentication constructs Authentication error which wraps provided error
func WrapWithAuthentication(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticAuthentication, err, text, fields...)
}

// IsAuthentication returns true if target error is Authentication semantic error
func IsAuthentication(err error) bool {
	return IsSemanticError(err, SemanticAuthentication)
}

// Forbidden constructs Forbidden error
func Forbidden(text string, fields ...zap.Field) *Error {
	return newError(SemanticForbidden, text, fields...)
}

// WrapWithForbidden constructs Forbidden error which wraps provided error
func WrapWithForbidden(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticForbidden, err, text, fields...)
}

// IsForbidden returns true if target error is Forbidden semantic error
func IsForbidden(err error) bool {
	return IsSemanticError(err, SemanticForbidden)
}

// FailedPrecondition constructs FailedPrecondition error
func FailedPrecondition(text string, fields ...zap.Field) *Error {
	return newError(SemanticFailedPrecondition, text, fields...)
}

// WrapWithFailedPrecondition constructs FailedPrecondition error which wraps provided error
func WrapWithFailedPrecondition(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticFailedPrecondition, err, text, fields...)
}

// IsFailedPrecondition returns true if target error is FailedPrecondition semantic error
func IsFailedPrecondition(err error) bool {
	return IsSemanticError(err, SemanticFailedPrecondition)
}

// NotFound constructs NotFound error
func NotFound(text string, fields ...zap.Field) *Error {
	return newError(SemanticNotFound, text, fields...)
}

// WrapWithNotFound constructs NotFound error which wraps provided error
func WrapWithNotFound(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotFound, err, text, fields...)
}

// IsNotFound returns true if target error is NotFound semantic error
func IsNotFound(err error) bool {
	return IsSemanticError(err, SemanticNotFound)
}

// AlreadyExists constructs AlreadyExists error
func AlreadyExists(text string, fields ...zap.Field) *Error {
	return newError(SemanticAlreadyExists, text, fields...)
}

// WrapWithAlreadyExists constructs AlreadyExists error which wraps provided error
func WrapWithAlreadyExists(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticAlreadyExists, err, text, fields...)
}

// IsAlreadyExists returns true if target error is AlreadyExists semantic error
func IsAlreadyExists(err error) bool {
	return IsSemanticError(err, SemanticAlreadyExists)
}

// Unprocessable constructs Unprocessable error
func Unprocessable(text string, fields ...zap.Field) *Error {
	return newError(SemanticUnprocessable, text, fields...)
}

// WrapWithUnprocessable constructs Unprocessable error which wraps provided error
func WrapWithUnprocessable(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticUnprocessable, err, text, fields...)
}

// IsUnprocessable returns true if target error is Unprocessable semantic error
func IsUnprocessable(err error) bool {
	return IsSemanticError(err, SemanticUnprocessable)
}

// NotAcceptable constructs NotAcceptable error
func NotAcceptable(text string, fields ...zap.Field) *Error {
	return newError(SemanticNotAcceptable, text, fields...)
}

// WrapWithNotAcceptable constructs NotAcceptable error which wraps provided error
func WrapWithNotAcceptable(err error, text string, fields ...zap.Field) *Error {
	return wrapError(SemanticNotAcceptable, err, text, fields...)
}

// IsNotAcceptable returns true if target error is NotAcceptable semantic error
func IsNotAcceptable(err error) bool {
	return IsSemanticError(err, SemanticNotAcceptable)
}

// TooManyRequests constructs TooManyRequests error
func TooManyRequests(text string, fields ...zap.Field) *Error {
	return newError(SemanticTooManyRequests, text, fields...)
}

// IsTooManyRequests returns true if target error is TooManyRequests semantic error
func IsTooManyRequests(err error) bool {
	return IsSemanticError(err, SemanticTooManyRequests)
}

func WrapWithSkipError(err error, fields ...zap.Field) *Error {
	return wrapError(SemanticSkipError, err, err.Error(), fields...)
}

// PartialSuccess constructs PartialSuccess error
func PartialSuccess(text string, fields ...zap.Field) *Error {
	return newError(SemanticPartialSuccess, text, fields...)
}

// IsPartialSuccess returns true if target error is PartialSuccessf semantic error
func IsPartialSuccess(err error) bool {
	return IsSemanticError(err, SemanticPartialSuccess)
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
