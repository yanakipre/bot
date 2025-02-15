package hooks

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/rdb/internal/driver/internal/wrapper"
	"github.com/yanakipre/bot/internal/rdb/internal/driver/sqldata"
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
	}
	// Log the query if we don't have the operation name that could be used to reference it,
	// or if the raw query turned out to be invalid.
	if data.Stmt != "" && (data.Operation == "" || isSyntaxError(err)) {
		fields = append(fields, zap.String("query", data.Stmt))
	}
	if t := startTimeFromContext(ctx); t != nil {
		fields = append(fields, zap.Int64("duration_ms", time.Since(lo.FromPtr(t)).Milliseconds()))
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
		fields = append(fields, logger.RawJSON("sql_args", args))
	}

	// the following code:
	// - logs successful queries with query context
	// - does not log context cancellations and timeouts
	// - logs unsuccessful queries with the error and query context.
	switch {
	case err == nil && ctx.Err() == nil:
		// log successful queries.
		// unsuccessful ones will be handled by the calling code
		// they will either be context cancellations or driver-specific errors
		switch l.logLevel {
		case zapcore.ErrorLevel:
			logger.FromContext(ctx).Warn(logMsg, fields...)
		case zapcore.WarnLevel:
			logger.FromContext(ctx).Warn(logMsg, fields...)
		case zapcore.InfoLevel:
			logger.FromContext(ctx).Info(logMsg, fields...)
		default:
			logger.FromContext(ctx).Debug(logMsg, fields...)
		}
	case ctx.Err() != nil:
		// do not log cancellations and timeouts
		return
	default:
		// logs unsuccessful queries with the error and query context.
		fields = append(fields, zap.Error(err))
		logger.FromContext(ctx).Error(logMsg, fields...)
	}
}

func isSyntaxError(err error) bool {
	var pgerr *pgconn.PgError
	return errors.As(err, &pgerr) && pgerr.Code == pgerrcode.SyntaxError
}
