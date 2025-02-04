package roundtripper

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/yanakipre/bot/internal/metrics"
)

type metricEvent struct {
	standID  string
	duration time.Duration
	uri      string
	method   string
	code     string
	service  string
}

type MetricEvent struct {
	// StandID is what differs one installation of a service from another.
	// for example, a region name is a nice StandID for region-wide installations.
	StandID string
	Service string
	Uri     string
}

// MetricsRoundTripper
// uriSlug allows to group URI in any way by returning non-empty string,
// for example one can pass prepared slug to request context before sending a request.
func MetricsRoundTripper(
	rt http.RoundTripper,
	eventDescription func(ctx context.Context) (e MetricEvent),
) http.RoundTripper {
	return &metricsRoundTripper{rt: rt, eventDescription: eventDescription}
}

type metricsRoundTripper struct {
	// serviceName string
	rt               http.RoundTripper
	eventDescription func(ctx context.Context) (e MetricEvent)
}

func setMetrics(e metricEvent) {
	metrics.HTTPEgressDuration.
		WithLabelValues(e.uri, e.method, e.code, e.service, e.standID).
		Observe(e.duration.Seconds())
	metrics.HTTPEgressTotal.
		WithLabelValues(e.uri, e.method, e.code, e.service, e.standID).
		Inc()
}

// RoundTrip implements http.RoundTripper
// Does not error out when fails to log request or response
func (rt *metricsRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	started := time.Now()
	response, err := rt.rt.RoundTrip(r)
	e := rt.eventDescription(ctx)
	event := metricEvent{
		uri:     e.Uri,
		method:  r.Method,
		service: e.Service,
		standID: e.StandID,
	}
	if event.uri == "" {
		// allow clients that don't have URI handling logic still report unique URI's
		event.uri = r.RequestURI
	}

	if response == nil {
		event.code = "network-err" // special case, when no response received
	} else {
		event.code = strconv.Itoa(response.StatusCode)
	}
	event.duration = time.Since(started)
	setMetrics(event)
	return response, err
}

func (rt *metricsRoundTripper) Unwrap() http.RoundTripper { return rt.rt }
