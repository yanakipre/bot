package recoverytooling

import (
	"context"
	"fmt"
	"github.com/yanakipre/bot/internal/semerr"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

// RecoverOrStop either block and executes again if panic.
func RecoverOrStop(ctx context.Context, exec func()) {
	ctx = logger.WithName(ctx, "recover_or_stop")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			restart := false
			func() {
				defer func() {
					if p := recover(); p != nil {
						restart = true
						logger.Panic(ctx, p)
					}
				}()
				exec()
			}()
			if !restart {
				return
			}
		}
	}
}

// Loop is a helper function to run a loop with recovery from panics.
func Loop(ctx context.Context, loop func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			func() {
				defer func() {
					if p := recover(); p != nil {
						logger.Panic(ctx, p)
					}
				}()
				loop()
				logger.Info(ctx, "exited from recovery loop")
			}()
		}
	}
}

// SuppressPanic logs fact of a panic if it occurred.
func SuppressPanic(lg logger.Logger, exec func()) {
	defer func() {
		lg = lg.Named("suppress_panic")
		if p := recover(); p != nil {
			logger.Panic(logger.WithLogger(context.Background(), lg), p)
			lg.Error("panic suppressed, won't retry", zap.Error(semerr.UnwrapPanic(p)))
		}
	}()
	exec()
}

// DoUntilSuccess is a helper function to run a function until it succeeds.
func DoUntilSuccess(ctx context.Context, f func() error) {
	success := true
	for {
		select {
		case <-ctx.Done():
			return
		default:
			func() {
				defer func() {
					if p := recover(); p != nil {
						success = false
						logger.Panic(ctx, p)
					}
				}()
				err := f()
				if err != nil {
					logger.Error(ctx, fmt.Errorf("retrying error from function execution: %w", err))
					success = false
				}
			}()
			if success {
				return
			} else {
				success = true
			}
		}
	}
}
