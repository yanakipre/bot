package worker

import (
	"sync"
	"time"

	"github.com/yanakipre/bot/internal/scheduletooling"
)

var _ scheduletooling.MetricsCollector = &wellKnownMetricsCollector{}

var metricsRegistered = sync.Once{}

func NewWellKnownMetricsCollector() scheduletooling.MetricsCollector {
	metricsRegistered.Do(func() {
		registerMetrics()
	})
	return &wellKnownMetricsCollector{}
}

// wellKnownMetricsCollector uses metrics package to report metrics.
type wellKnownMetricsCollector struct{}

func (w *wellKnownMetricsCollector) JobStarted(name string) (finished func()) {
	CloudBackgroundJobsRunning.WithLabelValues(name).Inc()
	CloudBackgroundJobRuns.WithLabelValues(name).Inc()

	jobStartTime := time.Now()
	return func() {
		CloudBackgroundJobsRunning.WithLabelValues(name).Dec()
		CloudBackgroundJobDurationSec.WithLabelValues(name).
			Observe(time.Since(jobStartTime).Seconds())
	}
}

func (w *wellKnownMetricsCollector) SkippedJob(name string) {
	CloudBackgroundJobSkips.WithLabelValues(name).Inc()
}
