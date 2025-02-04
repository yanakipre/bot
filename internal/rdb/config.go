package rdb

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/semerr"
)

type Config struct {
	// SearchPath
	//
	// Allows one to specify a schema to search objects in when executing queries.
	// Empty string means "use database and role defaults".
	SearchPath string `yaml:"search_path"              json:"search_path"`
	// DatabaseType
	//
	// An indicator which database type is used
	DatabaseType    DatabaseType             `yaml:"database_type"            json:"database_type"`
	DSN             secret.String            `yaml:"database_url"             json:"database_url"`
	MaxConnLifetime encodingtooling.Duration `yaml:"max_conn_lifetime"        json:"max_conn_lifetime"`
	MaxOpenConns    int                      `yaml:"max_open_conns"           json:"max_open_conns"`
	// CollectMetricsInterval
	//
	// How often to report database metrics.
	CollectMetricsInterval encodingtooling.Duration `yaml:"collect_metrics_interval" json:"collect_metrics_interval"` //nolint:lll
}

func (c *Config) CheckAndSetDefaults() error {
	if c.DSN.Unmask() == "" {
		return semerr.InvalidInput("DSN is required")
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		DSN: secret.NewString(
			"postgres://postgres:password@localhost:5432/postgres",
		),
		MaxConnLifetime:        encodingtooling.Duration{Duration: time.Hour},
		MaxOpenConns:           50,
		CollectMetricsInterval: encodingtooling.Duration{Duration: time.Second * 5},
	}
}

type DatabaseType string

const (
	driverName = "pgx"
)
