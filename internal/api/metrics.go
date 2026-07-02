package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// serverMetrics holds the Prometheus metric descriptors for an API server instance.
type serverMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// newServerMetrics creates and registers a set of server metrics using the
// provided Prometheus registerer. Use prometheus.DefaultRegisterer when a
// custom registry is not required.
func newServerMetrics(reg prometheus.Registerer) *serverMetrics {
	factory := promauto.With(reg)
	return &serverMetrics{
		requestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "packllama",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests by method, path, and status.",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "packllama",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency distribution in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

// metricsHandler returns the Prometheus exposition handler for the given registry.
func metricsHandler(reg prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

// withMetrics is a middleware that records Prometheus counters and histograms
// for every request. Pass a nil m to make the middleware a no-op.
func withMetrics(m *serverMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)
			path := metricsPathLabel(r)
			m.requestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
			m.requestsTotal.WithLabelValues(
				r.Method,
				path,
				strconv.Itoa(rec.statusCode),
			).Inc()
		})
	}
}

func metricsPathLabel(r *http.Request) string {
	if r.Pattern == "" {
		return r.URL.Path
	}
	_, path, ok := strings.Cut(r.Pattern, " ")
	if !ok || path == "" {
		return r.Pattern
	}
	return path
}
