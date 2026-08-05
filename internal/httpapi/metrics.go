package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type httpMetrics struct {
	registry         *prometheus.Registry
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newHTTPMetrics() *httpMetrics {
	metrics := &httpMetrics{
		registry: prometheus.NewRegistry(),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "golden_path",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed by the service.",
			},
			[]string{"method", "route", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "golden_path",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests processed by the service.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "golden_path",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed.",
			},
		),
	}

	metrics.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.requestsInFlight,
	)

	return metrics
}

func (m *httpMetrics) handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *httpMetrics) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		m.requestsInFlight.Inc()
		defer m.requestsInFlight.Dec()

		response := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(response, r)

		route := requestRoute(r)
		status := strconv.Itoa(response.statusCode)

		m.requestsTotal.WithLabelValues(r.Method, route, status).Inc()
		m.requestDuration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
	})
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func requestRoute(r *http.Request) string {
	if r.Pattern == "" {
		return "unmatched"
	}

	_, route, found := strings.Cut(r.Pattern, " ")
	if !found {
		route = r.Pattern
	}

	return strings.TrimSuffix(route, "{$}")
}
