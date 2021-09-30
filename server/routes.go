package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"canvas/handlers"
)

func (s *Server) setupRoutes() {
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
}
