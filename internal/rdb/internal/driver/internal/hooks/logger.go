package hooks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/rdb/internal/driver/internal/wrapper"
	"github.com/yanakipe/bot/internal/rdb/internal/driver/sqldata"
)

// NewLogger create new LoggerHook.
func NewLogger(logLevel zapcore.Level, host, db string) wrapper.Hook {
	return &LoggerHook{
		host:     host,
		db:       db,
		logLevel: logLevel,
	}
}

// LoggerHook log queries with given log level.
type LoggerHook struct {
	host     string
	db       string
	logLevel zapcore.Level
}

func (l *LoggerHook) Before(ctx context.Context) context.Context {
	return contextWithStartTime(ctx, time.Now().UTC())
}

func (l *LoggerHook) After(ctx context.Context, err error) {
	data := sqldata.FromContext(ctx)
	logMsg := fmt.Sprint("rdb.sql:", data.Action)

	// preallocate because we're in a hot path, because resizing
	// this slice is probably worse than overallocating by the
	// size of one pointer.
	fields := make([]zap.Field, 0, 7)

	if data.Operation != "" {
		fields = append(fields, zap.String("sql_operation", data.Operation))
	} else if data.Stmt != "" {
		fields = append(fields, zap.String("sql_statement", data.Stmt))
	}
	if t := startTimeFromContext(ctx); t != nil {
		fields = append(fields, zap.Int64("duration_ms", time.Since(*t).Milliseconds()))
	}
	if l.db != "" {
		fields = append(fields, zap.String("database_name", l.db))
	}
	if l.host != "" {
		fields = append(fields, zap.String("db_host", l.host))
	}
	if len(data.Args) > 0 {
		args := make([]any, len(data.Args))
		for i := range data.Args {
			if v, ok := data.Args[i].(driver.NamedValue); ok {
				args[i] = v.Value
			} else {
				args[i] = data.Args[i]
			}
		}
		fields = append(fields, zap.Any("sql_args", args))
	}

	var isError bool

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		isError = true
		fields = append(fields, zap.Error(err))
	} else if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		// this is probably "deadline exceeded"/timeout
		isError = true
		fields = append(fields, zap.Error(ctx.Err()))
	}

	if isError {
		logger.FromContext(ctx).Error(logMsg, fields...)
	}
}
