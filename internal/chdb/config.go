package chdb

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/secret"
)

type Config struct {
	Enabled                bool                     `yaml:"enabled"`
	Addr                   string                   `yaml:"addr"`
	Database               string                   `yaml:"database"`
	Username               string                   `yaml:"username"`
	Password               secret.String            `yaml:"password"`
	MaxOpenConns           int                      `yaml:"max_open_conns"`
	MaxIdleConns           int                      `yaml:"max_idle_conns"`
	CollectMetricsInterval encodingtooling.Duration `yaml:"collect_metrics_interval"`
	UseTls                 bool                     `yaml:"use_tls"`
	S3DumpConfig           S3DumpConfig             `yaml:"s3_dump_config"`
	// settings
	MaxPeriodsPerQuery int    `yaml:"max_periods_per_query"`
	MaxExecutionTime   uint64 `yaml:"max_execution_time"` // max execution time limit per query, in seconds.
	MaxQuerySize       uint64 `yaml:"max_query_size"`
	MaxResultRows      uint64 `yaml:"max_result_rows"`
	MaxResultBytes     uint64 `yaml:"max_result_bytes"`
}

func DefaultConfig() Config {
	cfg := Config{
		Addr:                   "10.30.42.61:9000",
		Database:               "yanakipre",
		Username:               "default",
		MaxOpenConns:           10,
		MaxIdleConns:           10,
		CollectMetricsInterval: encodingtooling.Duration{Duration: time.Minute},
		UseTls:                 false,
		MaxQuerySize:           10 * 1024 * 1024, // 10 MiB
		MaxExecutionTime:       60,
		MaxResultRows:          1000000,
		MaxResultBytes:         10 * 1024 * 1024, // 10 MiB
		// 10000 of periods (uuid) is about 350 KiB. max_query_size is set to 10 MiB just in case.
		MaxPeriodsPerQuery: 10000,
		S3DumpConfig:       DefaultS3DumpConfig(),
	}
	cfg.Password.FromEnv("CLICKHOUSE_PASSWORD")
	return cfg
}

type S3DumpConfig struct {
	S3Endpoint        string        `yaml:"s3_endpoint"`
	S3Bucket          string        `yaml:"s3_bucket"`
	S3AccessKeyID     secret.String `yaml:"s3_access_key_id"`
	S3SecretAccessKey secret.String `yaml:"s3_secret_access_key"`
}

func DefaultS3DumpConfig() S3DumpConfig {
	cfg := S3DumpConfig{
		S3Endpoint: "http://aws.yanakipre.local:4566",
		S3Bucket:   "enriched-usage-events-aggregates",
	}
	cfg.S3AccessKeyID.FromEnv("CLICKHOUSE_S3_ORB_EVENTS_ACCESS_KEY_ID")
	cfg.S3SecretAccessKey.FromEnv("CLICKHOUSE_S3_ORB_EVENTS_SECRET_ACCESS_KEY")

	return cfg
}
