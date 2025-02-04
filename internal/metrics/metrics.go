// Package metrics defines common metrics, shared between applications
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanakipre/bot/internal/promtooling"
)

func init() {
	registerMetrics()
}

func registerMetrics() {
	var coll []prometheus.Collector
	coll = append(coll, HTTPEgress()...)
	coll = append(coll, mbusMetrics()...)
	coll = append(coll, rdbMetrics()...)
	coll = append(coll, chdbMetrics()...)
	coll = append(coll, cacheMetrics()...)
	promtooling.MustRegister(coll...)
}

var (
	HTTPEgressDuration = promtooling.NewHistogramVec(
		"http_egress_duration",
		"Histogram of egress request durations in seconds",
		[]float64{
			0.01,
			0.05,
			0.1,
			0.2,
			0.3,
			0.5,
			0.8,
			1.0,
			1.2,
			1.5,
			2,
			3,
			5,
			7,
			10,
			15,
			20,
			30,
			60,
			120,
		},
		[]string{"uri", "method", "code", "http_client", "stand_id"},
	)
	HTTPEgressTotal = promtooling.NewCounterVec(
		"http_egress_total",
		"outgoing http requests status codes.",
		[]string{"uri", "method", "code", "http_client", "stand_id"},
	)
)

func HTTPEgress() []prometheus.Collector {
	return []prometheus.Collector{
		HTTPEgressDuration,
		HTTPEgressTotal,
	}
}

var (
	RDBConnectionsOpen = promtooling.NewGaugeVec(
		"rdb_connections_open",
		"Database connection pool status.",
		// total, in_use, idle + read_write, read_only
		[]string{"status", "type"},
	)
	RDBConnectionsWaitTotal = promtooling.NewGaugeVec(
		"rdb_connections_wait_total",
		"The total number of connections waited for.",
		[]string{"type"},
	)
	RDBConnectionsWaitSec = promtooling.NewGaugeVec(
		"rdb_connections_wait_sec",
		"The total time blocked waiting for a new connection.",
		[]string{"type"},
	)
)

func rdbMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		RDBConnectionsOpen,
		RDBConnectionsWaitTotal,
		RDBConnectionsWaitSec,
	}
}

var CHDBConnectionsOpen = promtooling.NewGaugeVec(
	"chdb_connections_open",
	"ClickHouse database connection pool status.",
	// max_open, max_idle, open, idle
	[]string{"status"},
)

func chdbMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		CHDBConnectionsOpen,
	}
}

var (
	//
	// Background cache jobs
	//
	CloudBackgroundCacheJobSkips = promtooling.NewCounterVec(
		"cloud_background_cache_job_skips",
		"Number of skipped background cache job, per job type",
		[]string{"background_cache_job_name"},
	)
	CloudBackgroundCacheJobRuns = promtooling.NewCounterVec(
		"cloud_background_cache_job_runs",
		"Number of background job cache runs, per job type",
		[]string{"background_cache_job_name"},
	)
	CloudBackgroundCacheJobDurationSec = promtooling.NewHistogramVec(
		"cloud_background_cache_job_duration_sec",
		"Histogram of background cache job duration since start, per job type",
		[]float64{
			0.5, 1, 2, 5, 7, 10, 15, 20, 30, 45, 60, 90, 120, 180, 240,
			300, 600, 900, 1200, 1800, 2400, 3000, 3600,
		},
		[]string{"background_cache_job_name"},
	)
	CloudBackgroundCacheJobsRunning = promtooling.NewGaugeVec(
		"cloud_background_cache_jobs_in_progress",
		"Number of in progress background cache jobs, per job type",
		[]string{"background_cache_job_name"},
	)
)

func cacheMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		CloudBackgroundCacheJobSkips,
		CloudBackgroundCacheJobRuns,
		CloudBackgroundCacheJobDurationSec,
		CloudBackgroundCacheJobsRunning,
	}
}
