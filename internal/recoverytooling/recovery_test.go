package recoverytooling

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/testtooling"
)

func TestRecovery_DoUntilSuccess(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	i := 0

	DoUntilSuccess(context.Background(), func() error {
		i++
		if i == 1 {
			panic("panic")
		}

		return nil
	})

	require.Equal(t, 2, i)
}

func TestRecovery_DoUntilSuccessWithError(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	i := 0

	DoUntilSuccess(context.Background(), func() error {
		i++
		if i == 1 {
			return errors.New("error")
		}

		return nil
	})

	require.Equal(t, 2, i)
}

func TestRecovery_Recovery(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	i := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	Loop(ctx, func() {
		i++
		if i%2 == 0 {
			panic("panic")
		}
	})
	cancel()

	require.Greater(t, i, 2)
}

func TestRecoverOrStop(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	i := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	RecoverOrStop(ctx, func() {
		i++
		if i == 0 {
			panic(errors.New("test-error"))
		}
		if i < 3 {
			panic("panic")
		}
	})
	cancel()

	require.Equal(t, 3, i)
}
