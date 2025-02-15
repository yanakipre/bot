package codeerr

import (
	"errors"
)

type Error struct {
	code ErrorCode
	err  error
}

func (e *Error) Is(target error) bool {
	return errors.Is(e.err, target)
}

func (e *Error) As(target any) bool {
	return errors.As(e.err, target)
}

// Unwrap implements Wrapper interface
func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) RawError() error {
	return e.err
}

func (e *Error) GetCodeString() string {
	return string(e.code)
}

func (e *Error) GetCode() ErrorCode {
	return e.code
}

func (e *Error) Error() string {
	return string(e.code)
}

var _ error = &Error{}

func AsCodeErr(err error) *Error {
	var target *Error
	if !errors.As(err, &target) {
		return nil
	}
	return target
}
