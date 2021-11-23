package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"canvas/handlers"
)

func (s *Server) setupRoutes() {
	s.mux.Use(handlers.AddMetrics(s.metrics))

	handlers.Health(s.mux, s.database)

	handlers.FrontPage(s.mux)
	handlers.NewsletterSignup(s.mux, s.database, s.queue)
	handlers.NewsletterThanks(s.mux)
	handlers.NewsletterConfirm(s.mux, s.database, s.queue)
	handlers.NewsletterConfirmed(s.mux)

	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("canvas", map[string]string{"admin": s.adminPassword}))

		handlers.MigrateTo(r, s.database)
		handlers.MigrateUp(r, s.database)
	})

	metricsAuth := middleware.BasicAuth("metrics", map[string]string{"prometheus": s.metricsPassword})
	handlers.Metrics(s.mux.With(metricsAuth), s.metrics)
}
