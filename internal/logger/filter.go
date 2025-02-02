package logger

import (
	"strings"

	"go.uber.org/zap/zapcore"
	"k8s.io/utils/strings/slices"
	"moul.io/zapfilter"
)

// logFilterSink produces log filtering sink.
func logFilterSink(config FilterConfig) zapfilter.FilterFunc {
	return func(entry zapcore.Entry, fields []zapcore.Field) bool {
		if len(config.FullNameFilter) > 0 {
			if !filterLoggerNameByNameStartswith(config.FullNameFilter, entry) {
				return false
			}
		}
		if len(config.ExactSubnameFilter) > 0 {
			if !filterLoggerNameByExactSubNameStartswith(
				config.ExactSubnameFilter,
				entry.LoggerName,
			) {
				return false
			}
		}
		return true
	}
}

// filterLoggerNameByNameStartswith leaves log line if name starts with name passed.
func filterLoggerNameByNameStartswith(filters []NameFilter, entry zapcore.Entry) bool {
	for _, f := range filters {
		if entry.Level > f.parsedLevel {
			// not matches the level
			continue
		}
		if isStringExactOrChild(f.LoggerName, entry.LoggerName) {
			// filter record out
			return false
		}
	}
	// leave record be
	return true
}

// filterLoggerNameByExactSubNameStartswith leaves log line if part of the logger
// name split by dot contains the specified LoggerName.
func filterLoggerNameByExactSubNameStartswith(
	filters []ExactSubnameFilter,
	loggerName string,
) bool {
	parts := strings.Split(loggerName, ".")
	for _, f := range filters {
		if slices.Contains(parts, f.LoggerName) {
			// filter record out
			return false
		}
	}
	// leave record be
	return true
}

// isStringExactOrChild returns true if candidate logger name should be filtered out by root name.
func isStringExactOrChild(root, candidate string) bool {
	if len(candidate) < len(root) {
		return false
	}
	if candidate == root {
		return true
	}
	if len(candidate) < len(root)+1 {
		return false
	}
	child := root + "."

	return candidate[:len(child)] == child
}
