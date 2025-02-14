package worker

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/net/context"
)

var metricsRegistered = sync.Once{}

func NewWellKnownMetricsCollector() *wellKnownMetricsCollector {
	metricsRegistered.Do(func() {
		registerMetrics()
	})
	return &wellKnownMetricsCollector{}
}

// wellKnownMetricsCollector uses metrics package to report metrics.
type wellKnownMetricsCollector struct{}

const (
	statusUNHANDLED = "unhandled"
	statusPanic     = "panic"
	statusTimeout   = "timeout"
	statusCancelled = "canceled"
	statusSuccess   = "success"
)

func errToStatus(err error) string {
	switch {
	case err == nil:
		return statusSuccess
	case errors.Is(err, context.DeadlineExceeded):
		return statusTimeout
	case errors.Is(err, context.Canceled):
		return statusCancelled
	default:
		return statusUNHANDLED
	}
}

func (w *wellKnownMetricsCollector) JobStarted(name string) (finished func(error)) {
	CloudBackgroundJobsRunning.WithLabelValues(name).Inc()
	CloudBackgroundJobRuns.WithLabelValues(name).Inc()

	jobStartTime := time.Now()
	return func(err error) {
		if p := recover(); p != nil {
			CloudBackgroundJobStatus.WithLabelValues(name, statusPanic).Inc()
			panic(p) // pass it further for handling
		}
		CloudBackgroundJobStatus.WithLabelValues(name, errToStatus(err)).Inc()
		CloudBackgroundJobsRunning.WithLabelValues(name).Dec()
		CloudBackgroundJobDurationSec.WithLabelValues(name).
			Observe(time.Since(jobStartTime).Seconds())
	}
}

func (w *wellKnownMetricsCollector) SkippedJob(name string) {
	CloudBackgroundJobSkips.WithLabelValues(name).Inc()
}
