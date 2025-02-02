package hooks

import (
	"context"
	"time"
)

type startTime int

var startTimeKey startTime

func contextWithStartTime(ctx context.Context, started time.Time) context.Context {
	return context.WithValue(ctx, startTimeKey, started)
}

func startTimeFromContext(ctx context.Context) *time.Time {
	v := ctx.Value(startTimeKey)
	if t, ok := v.(time.Time); ok {
		return &t
	}
	return nil
}
