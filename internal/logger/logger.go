package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/semerr"
	"io"
	"os"
	"runtime/debug"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/zapfilter"
)

var (
	// global logger instance.
	global       *zap.Logger
	defaultLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	globalConfig *Config
	closers      []func()
)

const DurationMsKey = "duration_ms"

// SetNewGlobalLoggerOnce configures logging in application once.
// Subsequent calls just complain to log with ERROR.
// This returns io.Writer that can be reused for integrations with other log engines.
func SetNewGlobalLoggerOnce(config Config) io.Writer {
	return SetNewGlobalLoggerOnceWithUnfiltered(config, nil)
}

func SetNewGlobalLoggerOnceWithUnfiltered(config Config, unfilteredSink io.Writer) io.Writer {
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
		fmt.Printf("cannot configure logging: %v", err)
		os.Exit(1)
	}
	defaultLevel.SetLevel(lvl)

	sink, err := sinkFromConfig(config)
	if err != nil {
		fmt.Printf("cannot configure log sink: %v", err)
		os.Exit(1)
	}
	// init global logger once.
	setLogger(
		New(
			defaultLevel,
			sink,
			unfilteredSink,
			zap.WithCaller(config.Caller),
			zap.AddCallerSkip(config.CallerSkip),
			zap.AddStacktrace(config.StackTraceLevel),
		))
	// traverse call depth for more useful log lines))
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
func New(level zapcore.LevelEnabler, sink io.Writer, unfilteredSink io.Writer, options ...zap.Option) *zap.Logger {
	if globalConfig.Format == FormatConsole {
		return NewDevelopmentConfig(level, sink, unfilteredSink, options...)
	}
	return NewWithSink(level, sink, unfilteredSink, options...)
}

func defaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
}

func newLoggerWithEncoder(
	level zapcore.LevelEnabler,
	sink io.Writer,
	unfilteredSink io.Writer,
	encoder zapcore.Encoder,
	options ...zap.Option,
) *zap.Logger {
	if level == nil {
		level = defaultLevel
	}
	cores := []zapcore.Core{
		zapfilter.NewFilteringCore(
			zapcore.NewCore(
				encoder,
				zapcore.AddSync(sink),
				level,
			),
			logFilterSink(globalConfig.Filters),
		),
	}
	if unfilteredSink != nil {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(unfilteredSink),
			level,
		))
	}
	if len(cores) == 1 {
		return zap.New(cores[0], options...)
	}
	return zap.New(zapcore.NewTee(cores...), options...)
}

func NewDevelopmentConfig(
	level zapcore.LevelEnabler,
	sink io.Writer,
	unfilteredSink io.Writer,
	options ...zap.Option,
) *zap.Logger {
	if level == nil {
		level = defaultLevel
	}
	encoderConfig := defaultEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return newLoggerWithEncoder(
		level,
		sink,
		unfilteredSink,
		zapcore.NewConsoleEncoder(encoderConfig),
		options...,
	)
}

func NewWithSink(
	level zapcore.LevelEnabler,
	sink io.Writer,
	unfilteredSink io.Writer,
	options ...zap.Option,
) *zap.Logger {
	if level == nil {
		level = defaultLevel
	}
	encoderConfig := defaultEncoderConfig()
	encoderConfig.EncodeLevel = CapitalLevelEncoder
	return newLoggerWithEncoder(
		level,
		sink,
		unfilteredSink,
		zapcore.NewJSONEncoder(encoderConfig),
		options...,
	)
}

// Level returns current global logger level
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// setLogger sets global used logger. This function is not thread-safe.
func setLogger(l *zap.Logger) {
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
	FromContext(ctx).Debug(msg, unwrapFields(args...)...)
}

func Info(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Info(msg, unwrapFields(args...)...)
}

func Warn(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Warn(msg, unwrapFields(args...)...)
}

func Error(ctx context.Context, err error) {
	FromContext(ctx).Error(err.Error(), unwrapErrorFields(err)...)
}

func Fatal(ctx context.Context, msg string, args ...zap.Field) {
	FromContext(ctx).Fatal(msg, unwrapFields(args...)...)
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

func Panic(ctx context.Context, p any, fields ...zap.Field) {
	err := semerr.UnwrapPanic(p)
	logPanicErr(FromContext(ctx), err, fields...)
}

func LogPanic(logger Logger, p any, fields ...zap.Field) {
	err := semerr.UnwrapPanic(p)
	logPanicErr(logger, err, fields...)
}

func logPanicErr(l Logger, err error, fields ...zap.Field) {
	ff := append(fields, unwrapErrorFields(err)...)
	ff = append(ff, zap.String("stack_trace", string(debug.Stack())))
	l.Error("app panicked and recovering", ff...)
}

// Typeof returns zap.Field with type of passed value
func Typeof(v any) zap.Field {
	//nolint:nolintlint
	//nolint:loglint
	return zap.String("type", fmt.Sprintf("%T", v))
}

// FloatFormatted returns zap.Field with f truncated to 2 digits after decimal point
func FloatFormatted(key string, f float64) zap.Field {
	//nolint:nolintlint
	//nolint:loglint
	return zap.String(key, fmt.Sprintf("%.2f", f))
}

// RawJSON returns zap.Field field with JSON string representation of a value for logging purposes
// If a marshaling error occures, an error will be logged instead of a formatted value
// USE WITH EXTRA CARE NOT TO EXPOSE SENSITIVE DATA
func RawJSON(k string, v any) zap.Field {
	if v == nil {
		return zap.String(k, "nil")
	}
	out, err := json.Marshal(v)
	if err != nil {
		//nolint:nolintlint
		//nolint:errlint
		return zap.NamedError("json_encoding_error", fmt.Errorf("cannot encode %T: %w", v, err))
	}
	//nolint:nolintlint
	//nolint:loglint
	return zap.ByteString(k, out)
}

func unwrapErrorFields(err error) []zap.Field {
	fmap := lo.Reduce(
		clouderr.UnwrapFields(err),
		func(agg map[string]zap.Field, f zap.Field, _ int) map[string]zap.Field {
			agg[f.Key] = f
			return agg
		}, map[string]zap.Field{})

	ff := make([]zap.Field, 0, len(fmap)+1)
	for _, f := range fmap {
		ff = append(ff, f)
	}
	// If there are named errors in the fields they might wrap their own fields
	ff = unwrapFields(ff...)
	// The actual error we log is the highest priority to put into err field
	ff = append(ff, zap.Error(err))

	return ff
}

func unwrapFields(args ...zap.Field) []zap.Field {
	ff := args
	for _, f := range args {
		if err, ok := f.Interface.(error); ok {
			ff = append(ff, clouderr.UnwrapFields(err)...)
		}
	}
	return ff
}
