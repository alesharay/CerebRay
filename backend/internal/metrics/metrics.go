package metrics

import "github.com/prometheus/client_golang/prometheus"

// Metrics holds all Prometheus metrics for the application.
type Metrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge
	AITokensTotal        *prometheus.CounterVec
	AIRequestDuration    *prometheus.HistogramVec
	AIRequestsTotal      *prometheus.CounterVec
}

// New creates and registers all application metrics.
func New() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cerebray_http_requests_total",
			Help: "Total HTTP requests by method, route, and status code.",
		}, []string{"method", "route", "status_code"}),

		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cerebray_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds by method and route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),

		HTTPRequestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cerebray_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed.",
		}),

		AITokensTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cerebray_ai_tokens_total",
			Help: "Total AI tokens consumed by type and model.",
		}, []string{"type", "model"}),

		AIRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cerebray_ai_request_duration_seconds",
			Help:    "AI request duration in seconds by model.",
			Buckets: []float64{1, 2, 5, 10, 20, 30, 60, 120},
		}, []string{"model"}),

		AIRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cerebray_ai_requests_total",
			Help: "Total AI requests by model and status.",
		}, []string{"model", "status"}),
	}

	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.AITokensTotal,
		m.AIRequestDuration,
		m.AIRequestsTotal,
	)

	return m
}
