package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxkey string

const (
	loggerContextKey ctxkey = "zlogger"
)

// WithFields lets you put log fields in context, that will be seen in logs.
// Nested calls preserve parent log fields.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	l := FromContext(ctx)
	if l == nil {
		panic("logger has not been configured")
	}
	l = l.With(fields...)
	return context.WithValue(ctx, loggerContextKey, l)
}

// WithName puts a named logger into context.
// The names are "aggregated" in a sense, that:
//
//	 logger.WithName(
//	   logger.WithName(ctx, "foo"),
//		  "bar"
//		  )
//
// will give you "for.bar"
func WithName(ctx context.Context, name string) context.Context {
	l := FromContext(ctx).Named(name)
	return context.WithValue(ctx, loggerContextKey, l)
}

func WithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// FromContext returns logger from context if set. Otherwise, returns global `global` logger.
// In both cases returned logger is populated with `trace_id` & `span_id`.
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerContextKey).(Logger); ok {
		return logger
	}
	return global
}
