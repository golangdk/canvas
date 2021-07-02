package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type pinger interface {
	Ping(ctx context.Context) error
}

func Health(mux chi.Router, p pinger) {
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := p.Ping(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	})
}
