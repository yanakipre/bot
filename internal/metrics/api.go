package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanakipe/bot/internal/promtooling"
)

var (
	APIRequestsTotal = promtooling.NewCounterVec(
		"gin_request_total", // for compatibility with existing dashboards
		"Total number of requests served by the server",
		[]string{"app_name"},
	)
	APIRequestsTotalURI = promtooling.NewCounterVec(
		"gin_uri_request_total", // for compatibility with existing dashboards
		"Number of requests served by the server per URI",
		[]string{"app_name", "uri", "method", "code"},
	)
	APIRequestsDuration = promtooling.NewHistogramVec(
		"gin_request_duration", // for compatibility with existing dashboards
		"Histogram of request durations in seconds",
		[]float64{
			0.01,
			0.05,
			0.1,
			0.2,
			0.3,
			0.4,
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
		[]string{"app_name", "uri"},
	)
)

func APIMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		APIRequestsTotal,
		APIRequestsTotalURI,
		APIRequestsDuration,
	}
}
