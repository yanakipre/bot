package recoverytooling

import (
	"context"
	"fmt"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
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
					if rvr := recover(); rvr != nil {
						restart = true
						err, ok := rvr.(error)
						if !ok {
							logger.Error(
								ctx,
								fmt.Sprintf(
									"recovering from panic without an error: %s",
									debug.Stack(),
								),
								zap.Any("panic", rvr),
							)
							return
						} else {
							logger.Error(
								ctx,
								fmt.Sprintf("recovering from panic: %s", debug.Stack()),
								zap.Error(err),
							)
						}
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
					if rvr := recover(); rvr != nil {
						err, ok := rvr.(error)
						if !ok {
							logger.Error(
								ctx,
								fmt.Sprintf(
									"recovering from panic without an error: %s",
									debug.Stack(),
								),
								zap.Any("panic", rvr),
							)
							return
						}
						logger.Error(
							ctx,
							fmt.Sprintf("recovering from panic: %s", debug.Stack()),
							zap.Error(err),
						)
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
	lg = lg.Named("suppress_panic")
	defer func() {
		if rvr := recover(); rvr != nil {
			lg.Error("got panic, will not repeat this automatically", zap.Any("panic", rvr))
		}
		exec()
	}()
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
					if rvr := recover(); rvr != nil {
						success = false

						err, ok := rvr.(error)
						if !ok {
							logger.Error(
								ctx,
								fmt.Sprintf(
									"recovering from panic without an error: %s",
									debug.Stack(),
								),
								zap.Any("panic", rvr),
							)
							return
						}
						logger.Error(
							ctx,
							fmt.Sprintf("recovering from panic: %s", debug.Stack()),
							zap.Error(err),
						)
					}
				}()
				err := f()
				if err != nil {
					logger.Error(ctx, "error while executing function, retrying", zap.Error(err))
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
