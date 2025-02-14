package chdb

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/yanakipre/bot/internal/metrics"
)

func sendMetrics(provider string, stats driver.Stats) {
	metrics.CHDBConnectionsOpen.WithLabelValues(provider, "max_open").
		Set(float64(stats.MaxOpenConns))
	metrics.CHDBConnectionsOpen.WithLabelValues(provider, "open").
		Set(float64(stats.Open))
	metrics.CHDBConnectionsOpen.WithLabelValues(provider, "max_idle").
		Set(float64(stats.MaxIdleConns))
	metrics.CHDBConnectionsOpen.WithLabelValues(provider, "idle").
		Set(float64(stats.Idle))
}
