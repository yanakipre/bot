package logger

import "go.uber.org/zap/zapcore"

// This package translates zap levels to unified format preferred for finding error levels across
// multiple apps.

func CapitalLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	level := ""
	switch l {
	case zapcore.ErrorLevel:
		level = "ERR"
	case zapcore.WarnLevel:
		level = "WARNING"
	default:
		level = l.CapitalString()
	}
	enc.AppendString(level)
}

func LowercaseLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	level := ""
	switch l {
	case zapcore.ErrorLevel:
		level = "err"
	case zapcore.WarnLevel:
		level = "warning"
	default:
		level = l.String()
	}
	enc.AppendString(level)
}
