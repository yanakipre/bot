package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/zapfilter"
)

var (
	// global logger instance.
	global       Logger
	defaultLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	globalConfig *Config
	closers      []func()
)

const DurationMsKey = "duration_ms"

// SetNewGlobalLoggerOnce configures logging in application once.
// Subsequent calls just complain to log with ERROR.
// This returns io.Writer that can be reused for integrations with other log engines.
func SetNewGlobalLoggerOnce(config Config) io.Writer {
	if global != nil {
		// if this is not during tests - this is a source of severe errors.
		global.Error("you have already set global logger once")
		return nil
	}
	if globalConfig != nil {
		panic("set config once per application")
	}
	globalConfig = &config
	lvl := defaultLevel.Level()
	if err := lvl.Set(config.LogLevel); err != nil {
		fmt.Printf("cannot configure logging %v", err)
		os.Exit(1)
	}
	defaultLevel.SetLevel(lvl)

	sink, err := sinkFromConfig(config)
	if err != nil {
		fmt.Printf("cannot configure log sink: %v", err)
		os.Exit(1)
	}
	// init global logger once.
	setLogger(New(defaultLevel, sink))
	global.Debug("set global logger")
	return sink
}

// sinkFromConfig returns sink for the logger from config
func sinkFromConfig(config Config) (io.Writer, error) {
	writer, closer, err := zap.Open(config.Sink)
	closers = append(closers, closer)
	return writer, err
}

// Close and release resources used.
func Close() {
	global.Info("logger is closing, no more application logs after this line")
	for _, c := range closers {
		c()
	}
}

// New creates new *zap.SugaredLogger with standard EncoderConfig
// if lvl == nil, global AtomicLevel will be used
func New(level zapcore.LevelEnabler, sink io.Writer, options ...zap.Option) Logger {
	if globalConfig.Format == FormatConsole {
		return NewDevelopmentConfig(level, sink, options...)
	}
	return NewWithSink(level, sink, options...)
}

func NewDevelopmentConfig(
	level zapcore.LevelEnabler,
	sink io.Writer,
	options ...zap.Option,
) *zap.Logger {
	if level == nil {
		level = defaultLevel
	}
	return zap.New(
		zapfilter.NewFilteringCore(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
					TimeKey:        "ts",
					LevelKey:       "level",
					NameKey:        "logger",
					CallerKey:      "caller",
					MessageKey:     "message",
					StacktraceKey:  "stacktrace",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    zapcore.CapitalColorLevelEncoder,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
				zapcore.AddSync(sink),
				level,
			),
			logFilterSink(globalConfig.Filters),
		),
		options...,
	)
}

func NewWithSink(level zapcore.LevelEnabler, sink io.Writer, options ...zap.Option) *zap.Logger {
	if level == nil {
		level = defaultLevel
	}
	return zap.New(
		zapfilter.NewFilteringCore(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zapcore.EncoderConfig{
					TimeKey:        "ts",
					LevelKey:       "level",
					NameKey:        "logger",
					CallerKey:      "caller",
					MessageKey:     "message",
					StacktraceKey:  "stacktrace",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    CapitalLevelEncoder,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
				zapcore.AddSync(sink),
				level,
			),
			logFilterSink(globalConfig.Filters),
		),
		options...,
	)
}

// Level returns current global logger level
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// setLogger sets global used logger. This function is not thread-safe.
func setLogger(l Logger) {
	global = l
}

// SetNewGlobalLoggerQuietly sets global logger once per application and does not panic if already
// set.
func SetNewGlobalLoggerQuietly(cfg Config) {
	if global != nil {
		return
	}
	SetNewGlobalLoggerOnce(cfg)
}

func Debug(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Debug(msg, args...)
}

func Info(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Error(msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Fatal(msg, args...)
}

type Logger interface {
	Named(s string) *zap.Logger
	With(fields ...zap.Field) *zap.Logger
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}
