// Package server contains everything for setting up and running the HTTP server.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"canvas/messaging"
	"canvas/storage"
)

type Server struct {
	address         string
	adminPassword   string
	database        *storage.Database
	log             *zap.Logger
	metricsPassword string
	metrics         *prometheus.Registry
	mux             chi.Router
	queue           *messaging.Queue
	server          *http.Server
}

type Options struct {
	AdminPassword   string
	Database        *storage.Database
	Host            string
	Log             *zap.Logger
	MetricsPassword string
	Metrics         *prometheus.Registry
	Port            int
	Queue           *messaging.Queue
}

func New(opts Options) *Server {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	if opts.Metrics == nil {
		opts.Metrics = prometheus.NewRegistry()
	}

	address := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	mux := chi.NewMux()
	return &Server{
		address:         address,
		adminPassword:   opts.AdminPassword,
		database:        opts.Database,
		log:             opts.Log,
		metricsPassword: opts.MetricsPassword,
		metrics:         opts.Metrics,
		mux:             mux,
		queue:           opts.Queue,
		server: &http.Server{
			Addr:              address,
			Handler:           mux,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       5 * time.Second,
		},
	}
}

// Start the Server by setting up routes and listening for HTTP requests on the given address.
func (s *Server) Start() error {
	s.setupRoutes()

	s.log.Info("Starting", zap.String("address", s.address))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error starting server: %w", err)
	}
	return nil
}

// Stop the Server gracefully within the timeout.
func (s *Server) Stop() error {
	s.log.Info("Stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error stopping server: %w", err)
	}

	return nil
}
