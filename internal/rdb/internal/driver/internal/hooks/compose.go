package hooks

import (
	"context"

	"github.com/yanakipre/bot/internal/rdb/internal/driver/internal/wrapper"
)

func Compose(hooks ...wrapper.Hook) wrapper.Hook {
	return composed(hooks)
}

type composed []wrapper.Hook

func (c composed) Before(ctx context.Context) context.Context {
	for _, hook := range c {
		ctx = hook.Before(ctx)
	}
	return ctx
}

func (c composed) After(ctx context.Context, err error) {
	for _, hook := range c {
		hook.After(ctx, err)
	}
}
