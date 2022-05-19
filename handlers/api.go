package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"canvas/model"
)

type newsletterCreator interface {
	CreateNewsletter(ctx context.Context, title, body string) (model.Newsletter, error)
	GetSubscribers(ctx context.Context) ([]model.Email, error)
}

type createNewsletterRequest struct {
	Title string
	Body  string
}

type createNewsletterResponse struct {
	Newsletter model.Newsletter
}

func (createNewsletterResponse) StatusCode() int {
	return http.StatusCreated
}

func CreateNewsletter(mux chi.Router, log *zap.Logger, nc newsletterCreator, q sender) {
	mux.Post("/newsletters", createAPIHandler(
		func(ctx context.Context, r createNewsletterRequest) (*createNewsletterResponse, error) {
			if r.Title == "" {
				return nil, invalidTitleError
			}
			if r.Body == "" {
				return nil, invalidBodyError
			}

			n, err := nc.CreateNewsletter(ctx, r.Title, r.Body)
			if err != nil {
				log.Info("Error creating newsletter", zap.Error(err))
				return nil, serverError{err}
			}

			subscribers, err := nc.GetSubscribers(ctx)
			if err != nil {
				log.Info("Error getting subscribers", zap.Error(err))
				return nil, serverError{err}
			}

			for _, s := range subscribers {
				err = q.Send(ctx, model.Message{
					"job":   "newsletter_email",
					"id":    n.ID,
					"email": s.String(),
				})
				if err != nil {
					log.Info("Error sending newsletter email message",
						zap.Error(err), zap.String("id", n.ID), zap.Stringer("email", s))
				}
			}

			return &createNewsletterResponse{Newsletter: n}, nil
		}))
}

type newsletterGetter interface {
	GetNewsletters(ctx context.Context) ([]model.Newsletter, error)
}

type getNewslettersResponse struct {
	Newsletters []model.Newsletter
}

func GetNewsletters(mux chi.Router, log *zap.Logger, ng newsletterGetter) {
	mux.Get("/newsletters", createAPIHandler(
		func(ctx context.Context, _ any) (*getNewslettersResponse, error) {
			newsletters, err := ng.GetNewsletters(ctx)
			if err != nil {
				log.Info("Error getting newsletters", zap.Error(err))
				return nil, serverError{err}
			}

			return &getNewslettersResponse{jsonSlice(newsletters)}, nil
		}))
}

type newsletterUpdater interface {
	UpdateNewsletter(ctx context.Context, id, title, body string) (*model.Newsletter, error)
}

type updateNewsletterRequest struct {
	ID    model.UUID
	Title string
	Body  string
}

type updateNewsletterResponse struct {
	Newsletter model.Newsletter
}

func UpdateNewsletter(mux chi.Router, log *zap.Logger, nu newsletterUpdater) {
	mux.Put("/newsletters", createAPIHandler(
		func(ctx context.Context, r updateNewsletterRequest) (*updateNewsletterResponse, error) {
			if !r.ID.IsValid() {
				return nil, invalidUUIDError
			}
			if r.Title == "" {
				return nil, invalidTitleError
			}
			if r.Body == "" {
				return nil, invalidBodyError
			}

			n, err := nu.UpdateNewsletter(ctx, r.ID.String(), r.Title, r.Body)
			if err != nil {
				log.Info("Error updating newsletter", zap.Error(err))
				return nil, serverError{err}
			}
			if n == nil {
				return nil, idNotFoundError
			}

			return &updateNewsletterResponse{Newsletter: *n}, nil
		}))
}

type newsletterDeleter interface {
	DeleteNewsletter(ctx context.Context, id string) (*model.Newsletter, error)
}

type deleteNewsletterRequest struct {
	ID model.UUID
}

type deleteNewsletterResponse struct{}

func DeleteNewsletter(mux chi.Router, log *zap.Logger, nd newsletterDeleter) {
	mux.Delete("/newsletters", createAPIHandler(
		func(ctx context.Context, r deleteNewsletterRequest) (*deleteNewsletterResponse, error) {
			if !r.ID.IsValid() {
				return nil, invalidUUIDError
			}

			n, err := nd.DeleteNewsletter(ctx, r.ID.String())
			if err != nil {
				log.Info("Error deleting newsletter", zap.Error(err))
				return nil, serverError{err}
			}
			if n == nil {
				return nil, idNotFoundError
			}

			return &deleteNewsletterResponse{}, nil
		}))
}

var (
	invalidUUIDError  = fieldValueError{Name: "ID", Explanation: "must be a valid UUID"}
	invalidTitleError = fieldValueError{Name: "Title", Explanation: "cannot be empty"}
	invalidBodyError  = fieldValueError{Name: "Body", Explanation: "cannot be empty"}
	idNotFoundError   = clientError{Message: "no such ID", Code: http.StatusNotFound}
)

// fieldValueError is used for invalid field values in requests, identified by a name.
type fieldValueError struct {
	Name        string
	Explanation string
}

func (e fieldValueError) Error() string {
	message := fmt.Sprintf("invalid value for field '%v'", e.Name)
	if e.Explanation != "" {
		message += ", " + e.Explanation
	}
	return message
}

func (fieldValueError) StatusCode() int {
	return http.StatusBadRequest
}

type clientError struct {
	Message string
	Code    int
}

func (e clientError) Error() string {
	return e.Message
}

func (e clientError) StatusCode() int {
	return e.Code
}

// serverError is used for errors not caused by the client.
type serverError struct {
	Err error
}

func (e serverError) Error() string {
	return "server error: " + e.Err.Error()
}

func (serverError) StatusCode() int {
	return http.StatusBadGateway
}

// statusCodeGiver is something that can give a status code.
type statusCodeGiver interface {
	StatusCode() int
}

// createAPIHandler takes a callback function which receives the unmarshalled JSON request body struct as a parameter,
// and which should return a response body struct to be marshalled to JSON.
//
// If there is no request body (like in most GET requests), the type of the request body struct should just be "any".
//
// If there is a request body, but it cannot be parsed into the request body struct, HTTP 400 Bad Request is returned,
// along with an error response.
//
// If the response body struct satisfied statusCodeGiver, its status code is used in the HTTP response,
// except if there's an error.
//
// If the callback returns any error, it is translated into an HTTP status code and an error response is returned
// instead of the response body struct. If the error satisfies the statusCodeGiver interface,
// that status code is used. Otherwise, HTTP 500 Internal Server Error is used.
func createAPIHandler[Req, Res any](cb func(context.Context, Req) (Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req
		br := bufio.NewReader(r.Body)
		if _, err := br.Peek(1); err == nil {
			dec := json.NewDecoder(br)
			dec.DisallowUnknownFields()

			if err := dec.Decode(&req); err != nil {
				err = fmt.Errorf("malformed JSON in request body: %w", err)
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, errorResponse{jsonError{err}})
				return
			}
		}

		res, err := cb(r.Context(), req)
		if err != nil {
			if err, ok := err.(statusCodeGiver); ok {
				w.WriteHeader(err.StatusCode())
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			writeJSON(w, errorResponse{jsonError{err}})
			return
		}

		// The any() wrapper around responseBody is to avoid this error:
		// "cannot use type assertion on type parameter value responseBody (variable of type Res constrained by any)"
		// See https://github.com/golang/go/issues/45380#issuecomment-1014950980
		if res, ok := any(res).(statusCodeGiver); ok {
			w.WriteHeader(res.StatusCode())
		} else {
			w.WriteHeader(http.StatusOK)
		}

		writeJSON(w, res)
	}
}

// writeJSON panics on errors, because any error returned is basically a bug that should be handled at development time.
func writeJSON(w io.Writer, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(err)
	}
}

// errorResponse just wraps an error.
type errorResponse struct {
	Error jsonError
}

// jsonError is a wrapper around an error which makes sure the error message is serialized properly.
type jsonError struct {
	Err error
}

func (e jsonError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Err.Error())
}

// jsonSlice turns a nil-slice into an empty slice, since otherwise the slice would be serialized as "null" in JSON.
func jsonSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
