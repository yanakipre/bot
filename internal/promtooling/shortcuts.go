package promtooling

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const DefaultNamespace = ""

// Register registers the provided Collector with the DefaultRegisterer.
//
// Register is a shortcut for DefaultRegisterer.Register(c). See there for more
// details.
func Register(c prometheus.Collector) error {
	return prometheus.Register(c)
}

// MustRegister registers the provided Collectors with the DefaultRegisterer and
// panics if any error occurs.
//
// MustRegister is a shortcut for DefaultRegisterer.MustRegister(cs...). See
// there for more details.
func MustRegister(cs ...prometheus.Collector) {
	prometheus.MustRegister(cs...)
}

// Unregister removes the registration of the provided Collector from the
// DefaultRegisterer.
//
// Unregister is a shortcut for DefaultRegisterer.Unregister(c). See there for
// more details.
func Unregister(c prometheus.Collector) bool {
	return prometheus.Unregister(c)
}

// Handler returns an http.Handler for the prometheus.DefaultGatherer, using
// default HandlerOpts, i.e. it reports the first error as an HTTP error, it has
// no error logging, and it applies compression if requested by the client.
//
// The returned http.Handler is already instrumented using the
// InstrumentMetricHandler function and the prometheus.DefaultRegisterer. If you
// create multiple http.Handlers by separate calls of the Handler function, the
// metrics used for instrumentation will be shared between them, providing
// global scrape counts.
//
// This function is meant to cover the bulk of basic use cases. If you are doing
// anything that requires more customization (including using a non-default
// Gatherer, different instrumentation, and non-default HandlerOpts), use the
// HandlerFor function. See there for details.
func Handler() http.Handler {
	return promhttp.Handler()
}

// NewCounter creates a new Counter.
func NewCounter(name, help string) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
	})
}

// NewCounterVec creates a new CounterVec partitioned by the given label names.
func NewCounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
	}, labelNames)
}

// NewGauge creates a new Gauge.
func NewGauge(name, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
	})
}

// NewGaugeVec creates a new GaugeVec partitioned by the given label names.
func NewGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
	}, labelNames)
}

// NewHistogram creates a new Histogram.
func NewHistogram(name, help string, buckets []float64) prometheus.Histogram {
	return prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	})
}

// NewHistogramVec creates a new HistogramVec partitioned by the given label names.
func NewHistogramVec(
	name, help string,
	buckets []float64,
	labelNames []string,
) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: DefaultNamespace,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}, labelNames)
}

// NewSummary creates a new Summary.
func NewSummary(name, help string, objectives map[float64]float64) prometheus.Summary {
	return prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:  DefaultNamespace,
		Name:       name,
		Help:       help,
		Objectives: objectives,
		MaxAge:     prometheus.DefMaxAge,
	})
}

// NewSummaryVec creates a new SummaryVec partitioned by the given label names.
func NewSummaryVec(
	name, help string,
	objectives map[float64]float64,
	labelNames []string,
) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  DefaultNamespace,
		Name:       name,
		Help:       help,
		Objectives: objectives,
		MaxAge:     prometheus.DefMaxAge,
	}, labelNames)
}

// MustRegisterCounter creates and registers a new Counter.
// Panics if any error occurs.
func MustRegisterCounter(name, help string) prometheus.Counter {
	c := NewCounter(name, help)
	MustRegister(c)
	return c
}

// MustRegisterCounterVec creates and registers a new CounterVec.
// Panics if any error occurs.
func MustRegisterCounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	c := NewCounterVec(name, help, labelNames)
	MustRegister(c)
	return c
}

// MustRegisterGauge creates and registers a new Gauge.
// Panics if any error occurs.
func MustRegisterGauge(name, help string) prometheus.Gauge {
	c := NewGauge(name, help)
	MustRegister(c)
	return c
}

// MustRegisterGaugeVec creates and registers a new GaugeVec.
// Panics if any error occurs.
func MustRegisterGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec {
	c := NewGaugeVec(name, help, labelNames)
	MustRegister(c)
	return c
}

// MustRegisterHistogram creates and registers a new Histogram.
// Panics if any error occurs.
func MustRegisterHistogram(name, help string, buckets []float64) prometheus.Histogram {
	c := NewHistogram(name, help, buckets)
	MustRegister(c)
	return c
}

// MustRegisterHistogramVec creates and registers a new HistogramVec.
// Panics if any error occurs.
func MustRegisterHistogramVec(
	name, help string,
	buckets []float64,
	labelNames []string,
) *prometheus.HistogramVec {
	c := NewHistogramVec(name, help, buckets, labelNames)
	MustRegister(c)
	return c
}

// MustRegisterSummary creates and registers a new Summary.
// Panics if any error occurs.
func MustRegisterSummary(name, help string, objectives map[float64]float64) prometheus.Summary {
	c := NewSummary(name, help, objectives)
	MustRegister(c)
	return c
}

// MustRegisterSummaryVec creates and registers a new SummaryVec.
// Panics if any error occurs.
func MustRegisterSummaryVec(
	name, help string,
	objectives map[float64]float64,
	labelNames []string,
) *prometheus.SummaryVec {
	c := NewSummaryVec(name, help, objectives, labelNames)
	MustRegister(c)
	return c
}
