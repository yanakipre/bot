package closer

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

func getType(myvar any) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return fmt.Sprintf("%q from %s", t.Elem().Name(), t.Elem().PkgPath())
	} else {
		return fmt.Sprintf("%q from %s", t.Name(), t.PkgPath())
	}
}

func closeImpl(ctx context.Context, closer io.Closer, verbose bool) {
	startedAt := time.Now()
	ctx = logger.WithName(logger.WithFields(ctx, zap.String("closer_name", getType(closer))), "shutdown")
	if verbose {
		logger.Info(ctx, "closing dependency")
	}
	if err := closer.Close(); err != nil {
		logger.Warn(
			ctx,
			"could not close",
			zap.Duration("took", time.Since(startedAt)),
			zap.Error(err),
		)
		return
	}
	if verbose {
		logger.Info(
			ctx,
			"closed dependency",
			zap.Duration("time_to_close", time.Since(startedAt)),
		)
	}
}

// Close is a convinience wrapper to close readers, writers.
func Close(ctx context.Context, closer io.Closer) {
	closeImpl(ctx, closer, false)
}

// CloseVerbose is a convinience wrapper to close readers, writers.
// Use CloseVerbose on non-hot-paths only. Otherwise you risk wasting too much logs.
func CloseVerbose(ctx context.Context, closer io.Closer) {
	closeImpl(ctx, closer, true)
}
