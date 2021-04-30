package server

import (
	"canvas/handlers"
)

func (s *Server) setupRoutes() {
	handlers.Health(s.mux)
}
