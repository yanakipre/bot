package redis

import (
	"context"
	"fmt"

	"github.com/yanakipre/bot/internal/logger"
)

type Log struct{}

func (l Log) Printf(ctx context.Context, format string, v ...any) {
	logger.Info(ctx, fmt.Sprintf(format, v...))
}
