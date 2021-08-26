package server

import (
	"canvas/handlers"
)

func (s *Server) setupRoutes() {
	handlers.Health(s.mux, s.database)

	handlers.FrontPage(s.mux)
	handlers.NewsletterSignup(s.mux, s.database, s.queue)
	handlers.NewsletterThanks(s.mux)
}
