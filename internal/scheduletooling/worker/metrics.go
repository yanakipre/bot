package worker

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanakipre/bot/internal/promtooling"
)

func registerMetrics() {
	var coll []prometheus.Collector
	coll = append(coll, cloudBackgroundJobMetrics()...)
	promtooling.MustRegister(coll...)
}

var (
	//
	// Background jobs
	//
	CloudBackgroundJobSkips = promtooling.NewCounterVec(
		"cloud_background_job_skips",
		"Number of skipped background job, per job type",
		[]string{"background_job_name"},
	)
	CloudBackgroundJobRuns = promtooling.NewCounterVec(
		"cloud_background_job_runs",
		"Number of background job runs, per job type",
		[]string{"background_job_name"},
	)
	CloudBackgroundJobDurationSec = promtooling.NewHistogramVec(
		"cloud_background_job_duration_sec",
		"Histogram of background job duration since start, per job type",
		[]float64{
			0.5, 1, 2, 5, 7, 10, 15, 20, 30, 45, 60, 90, 120, 180, 240,
			300, 600, 900, 1200, 1800, 2400, 3000, 3600,
		},
		[]string{"background_job_name"},
	)
	CloudBackgroundJobsRunning = promtooling.NewGaugeVec(
		"cloud_background_jobs_in_progress",
		"Number of in progress background jobs, per job type",
		[]string{"background_job_name"},
	)
)

func cloudBackgroundJobMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		CloudBackgroundJobSkips,
		CloudBackgroundJobRuns,
		CloudBackgroundJobDurationSec,
		CloudBackgroundJobsRunning,
	}
}
