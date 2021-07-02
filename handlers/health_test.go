package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"

	"canvas/handlers"
)

type pingerMock struct {
	err error
}

func (p *pingerMock) Ping(ctx context.Context) error {
	return p.err
}

func TestHealth(t *testing.T) {
	t.Run("returns 200", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		handlers.Health(mux, &pingerMock{})
		code, _, _ := makeGetRequest(mux, "/health")
		is.Equal(http.StatusOK, code)
	})

	t.Run("returns 502 if the database cannot be pinged", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		handlers.Health(mux, &pingerMock{err: errors.New("oh no")})
		code, _, _ := makeGetRequest(mux, "/health")
		is.Equal(http.StatusBadGateway, code)
	})
}

// makeGetRequest and returns the status code, response headers, and the body.
func makeGetRequest(handler http.Handler, target string) (int, http.Header, string) {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	result := res.Result()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	return result.StatusCode, result.Header, string(bodyBytes)
}
