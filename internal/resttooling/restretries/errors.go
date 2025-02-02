package restretries

import (
	"errors"
)

var errMustStop = errors.New("request must not be retried")

func IsPermanentError(err error) bool { return errors.Is(err, errMustStop) }

func NewPermanentError(err error) error { return errors.Join(err, errMustStop) }
