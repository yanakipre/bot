package logger

import (
	"fmt"
	"log/slog"

	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
)

func SlogHandlerFrom(lg Logger) slog.Handler {
	logger, ok := lg.(*zap.Logger)
	if !ok {
		panic(fmt.Sprintf("expected *zap.Logger, got %T", lg))
	}
	return slogzap.Option{
		Level:  slog.LevelDebug,
		Logger: logger,
	}.NewZapHandler()
}
