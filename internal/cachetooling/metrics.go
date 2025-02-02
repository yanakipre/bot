package cachetooling

import (
	"time"

	"github.com/yanakipe/bot/internal/metrics"
	"github.com/yanakipe/bot/internal/scheduletooling"
)

var _ scheduletooling.MetricsCollector = &CacheMetricsCollector{}

// CacheMetricsCollector uses metrics package to report metrics.
type CacheMetricsCollector struct{}

func (w *CacheMetricsCollector) JobStarted(name string) (finished func()) {
	metrics.CloudBackgroundCacheJobSkips.WithLabelValues(name).Inc()
	metrics.CloudBackgroundCacheJobRuns.WithLabelValues(name).Inc()

	jobStartTime := time.Now()
	return func() {
		metrics.CloudBackgroundCacheJobsRunning.WithLabelValues(name).Dec()
		metrics.CloudBackgroundCacheJobDurationSec.WithLabelValues(name).
			Observe(time.Since(jobStartTime).Seconds())
	}
}

func (w *CacheMetricsCollector) SkippedJob(name string) {
	metrics.CloudBackgroundCacheJobSkips.WithLabelValues(name).Inc()
}
