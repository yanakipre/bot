package semerr

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	myErr  = errors.New("first error")
	myErr2 = errors.New("second error")
)

func TestErrCollector(t *testing.T) {
	var e ErrCollector
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		e.Join(myErr)
	}()
	go func() {
		defer wg.Done()
		e.Join(myErr2)
	}()

	wg.Wait()
	err := e.Resolve()
	require.ErrorIs(t, err, myErr)
	require.ErrorIs(t, err, myErr2)
}

func TestErrCollectorRecover(t *testing.T) {
	var e ErrCollector
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer e.Recover()

		panic("hello")
	}()

	wg.Wait()
	err := e.Resolve()
	require.Contains(t, err.Error(), "hello")
	require.ErrorIs(t, err, ErrRecoveredPanic)
}
