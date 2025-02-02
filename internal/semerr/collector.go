package semerr

import (
	"errors"
	"sync"
)

var ErrRecoveredPanic = errors.New("recovered panic")

// ErrCollector allows using `errors.Join` across several go routines.
// If you are not using go routines prefer using `errors.Join` directly.
type ErrCollector struct {
	err  error
	lock sync.Mutex
}

// Join combines errors into one and stores it in ErrCollector
func (e *ErrCollector) Join(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.err = errors.Join(e.err, err)
}

// Resolve returns the underlying error
func (e *ErrCollector) Resolve() error {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.err
}

// Recover handles the panic err and joins it into this ErrCollector
func (e *ErrCollector) Recover() {
	if r := recover(); r != nil {
		err := UnwrapPanic(r)
		e.Join(err)
		e.Join(ErrRecoveredPanic)
	}
}
