package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Middleware is an alias for a function that takes and returns a handler.
type Middleware = func(http.Handler) http.Handler

// AddMetrics constructs middleware to record request metrics.
func AddMetrics(registry *prometheus.Registry) Middleware {
	requests := promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_requests_total",
		Help: "The total number of HTTP requests.",
	}, []string{"method", "path", "code"})

	requestDurations := promauto.With(registry).NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_http_request_duration_seconds",
		Help:    "HTTP request durations.",
		Buckets: []float64{.005, .01, .05, .1, .5, 1},
	}, []string{"code"})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, 1)
			before := time.Now()
			next.ServeHTTP(ww, r)
			duration := time.Since(before)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			code := strconv.Itoa(status)
			requests.WithLabelValues(r.Method, r.URL.Path, code).Inc()
			requestDurations.WithLabelValues(code).Observe(duration.Seconds())
		})
	}
}

// Metrics exposes all Prometheus metrics from the given registry.
func Metrics(mux chi.Router, registry *prometheus.Registry) {
	mux.Get("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)
}
