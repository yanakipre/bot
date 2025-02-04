package chdb

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/yanakipre/bot/internal/metrics"
)

func sendMetrics(stats driver.Stats) {
	metrics.CHDBConnectionsOpen.WithLabelValues("max_open").
		Set(float64(stats.MaxOpenConns))
	metrics.CHDBConnectionsOpen.WithLabelValues("open").
		Set(float64(stats.Open))
	metrics.CHDBConnectionsOpen.WithLabelValues("max_idle").
		Set(float64(stats.MaxIdleConns))
	metrics.CHDBConnectionsOpen.WithLabelValues("idle").
		Set(float64(stats.Idle))
}
