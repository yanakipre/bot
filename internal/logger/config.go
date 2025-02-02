package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type Format string

const (
	FormatJSON    Format = "json"
	FormatConsole Format = "console"
	sinkSTDOUT           = "stdout"
)

type FilterType string

// NameFilter to filter out logger and children.
//
// Example:
//
// 1. for LoggerName "background_worker.clickhouse_consumption"
// * "background_worker.clickhouse_consumption.http" will be filtered out
// * "background_worker.clickhouse_consumption" will be filtered out
// * "background_worker" will stay, NOT filtered out
//
// 2. for Level "INFO"
// * "INFO", "DEBUG" entries will be filtered out
// * "WARN" will stay.
type NameFilter struct {
	// Level defines the max level to filter out.
	// Everything above will stay.
	Level string `yaml:"level"`
	// parsedLevel will be filled in by calling Validate of FilterConfig.
	parsedLevel zapcore.Level
	// LoggerName
	LoggerName string `yaml:"logger_name"`
}

// ExactSubnameFilter to filter out strings by name separated by dot.
// for "bar" given:
// * "b.bar.a" will filter out record.
// * "b.barFOO.a" will leave record be.
type ExactSubnameFilter struct {
	LoggerName string `yaml:"logger_name"`
}

type FilterConfig struct {
	FullNameFilter     []NameFilter         `yaml:"by_logger_name"`
	ExactSubnameFilter []ExactSubnameFilter `yaml:"by_exact_name"`
}

func (c *FilterConfig) validate() error {
	for i := range c.FullNameFilter {
		level, err := zapcore.ParseLevel(c.FullNameFilter[i].Level)
		if err != nil {
			return fmt.Errorf(
				"cannot parse FullNameFilter of %q: %w",
				c.FullNameFilter[i].Level,
				err,
			)
		}
		c.FullNameFilter[i].parsedLevel = level
	}
	return nil
}

func DefaultFilterConfig() FilterConfig {
	return FilterConfig{}
}

type Config struct {
	Sink     string `yaml:"sink"`
	LogLevel string `yaml:"log_level"`
	Format   Format `yaml:"log_format"`
	// Filters allow filter out some log lines based on conditions
	Filters FilterConfig
}

func (c *Config) Validate() error {
	if err := c.Filters.validate(); err != nil {
		return fmt.Errorf("cannot validate filters: %w", err)
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		Sink:     sinkSTDOUT,
		LogLevel: "DEBUG",
		Format:   FormatJSON,
		Filters:  DefaultFilterConfig(),
	}
}
