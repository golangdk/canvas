package handlers_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"
	"github.com/prometheus/client_golang/prometheus"

	"canvas/handlers"
)

func TestAddMetrics(t *testing.T) {
	t.Run("adds counter and histogram metrics", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		registry := prometheus.NewRegistry()
		mux.Use(handlers.AddMetrics(registry))
		handlers.Metrics(mux, registry)
		mux.Get("/exists", func(w http.ResponseWriter, r *http.Request) {})

		code, _, _ := makeGetRequest(mux, "/exists")
		is.Equal(http.StatusOK, code)
		code, _, _ = makeGetRequest(mux, "/doesnotexist")
		is.Equal(http.StatusNotFound, code)

		code, _, body := makeGetRequest(mux, "/metrics")
		is.Equal(http.StatusOK, code)

		is.True(strings.Contains(body, `app_http_requests_total{code="200",method="GET",path="/exists"} 1`))
		is.True(strings.Contains(body, `app_http_requests_total{code="404",method="GET",path="/doesnotexist"} 1`))

		is.True(strings.Contains(body, `app_http_request_duration_seconds_bucket{code="200",le="+Inf"} 1`))
		is.True(strings.Contains(body, `app_http_request_duration_seconds_bucket{code="404",le="+Inf"} 1`))
	})
}
