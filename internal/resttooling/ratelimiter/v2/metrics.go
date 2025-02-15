package ratelimiter

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanakipre/bot/internal/promtooling"
)

var (
	// SeenRequestsTotal is a counter that tracks the total number of requests seen by the rate limiter.
	SeenRequestsTotal = promtooling.NewCounterVec(
		"ratelimiter_seen_requests_total",
		"Total number of requests seen by the rate limiter",
		[]string{"app_name", "pattern", "key"},
	)

	// RejectedRequestsTotal is a counter that tracks the total number of requests rejected by the rate limiter.
	RejectedRequestsTotal = promtooling.NewCounterVec(
		"ratelimiter_rejected_requests_total",
		"Total number of requests rejected by the rate limiter",
		[]string{"app_name", "pattern", "key"},
	)

	// WouldBeRejectedRequestsTotal is a counter that tracks the total number of requests that would be rejected
	// by the rate limiter if its mode was set to `enforcing`.
	WouldBeRejectedRequestsTotal = promtooling.NewCounterVec(
		"ratelimiter_would_be_rejected_requests_total",
		"Total number of requests that would be rejected by the rate limiter if it was enforcing its configuration",
		[]string{"app_name", "pattern", "key"},
	)
)

// Metrics returns the prometheus-style metrics that this package tracks.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		SeenRequestsTotal,
		RejectedRequestsTotal,
		WouldBeRejectedRequestsTotal,
	}
}
