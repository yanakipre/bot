package rdb

import (
	"database/sql"

	"github.com/yanakipre/bot/internal/metrics"
)

func sendMetrics(stats sql.DBStats, destinationType string) {
	metrics.RDBConnectionsOpen.WithLabelValues("total", destinationType).
		Set(float64(stats.OpenConnections))
	metrics.RDBConnectionsOpen.WithLabelValues("in_use", destinationType).
		Set(float64(stats.InUse))
	metrics.RDBConnectionsOpen.WithLabelValues("idle", destinationType).
		Set(float64(stats.Idle))
	metrics.RDBConnectionsWaitTotal.WithLabelValues(destinationType).
		Set(float64(stats.WaitCount))
	metrics.RDBConnectionsWaitSec.WithLabelValues(destinationType).
		Set(stats.WaitDuration.Seconds())
}
