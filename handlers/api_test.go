package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matryer/is"
	"go.uber.org/zap"

	"canvas/handlers"
	"canvas/model"
)

var now = time.Date(2022, 5, 17, 12, 0, 0, 0, time.UTC)

type newsletterCreatorMock struct {
	title string
	body  string
}

func (n *newsletterCreatorMock) CreateNewsletter(ctx context.Context, title, body string) (model.Newsletter, error) {
	n.title = title
	n.body = body
	return model.Newsletter{
		ID:      "123",
		Title:   title,
		Body:    body,
		Created: now,
		Updated: now,
	}, nil
}

func (n *newsletterCreatorMock) GetSubscribers(ctx context.Context) ([]model.Email, error) {
	return []model.Email{
		"foo@example.com",
		"bar@example.com",
	}, nil
}

type multiMessageSenderMock struct {
	messages []model.Message
}

func (m *multiMessageSenderMock) Send(ctx context.Context, message model.Message) error {
	m.messages = append(m.messages, message)
	return nil
}

type erroringNewsletterCreatorMock struct{}

func (erroringNewsletterCreatorMock) CreateNewsletter(ctx context.Context, title, body string) (model.Newsletter, error) {
	return model.Newsletter{}, errors.New("doesn't work")
}

func (erroringNewsletterCreatorMock) GetSubscribers(ctx context.Context) ([]model.Email, error) {
	return nil, errors.New("doesn't work")
}

func TestCreateNewsletter(t *testing.T) {
	t.Run("creates the newsletter with title and body and creates a job for each subscriber", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nc := &newsletterCreatorMock{}
		q := &multiMessageSenderMock{}
		handlers.CreateNewsletter(mux, zap.NewNop(), nc, q)

		code, _, body := makePostRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"Title":"Hey you", "Body":"Welcome to the newsletter."}`))
		is.Equal(code, http.StatusCreated)
		is.Equal(body, `{"Newsletter":{"ID":"123","Title":"Hey you","Body":"Welcome to the newsletter.",`+
			`"Created":"2022-05-17T12:00:00Z","Updated":"2022-05-17T12:00:00Z"}}`)

		is.Equal(nc.title, "Hey you")
		is.Equal(nc.body, "Welcome to the newsletter.")

		is.Equal(len(q.messages), 2)
		is.Equal(q.messages[0]["job"], "newsletter_email")
		is.Equal(q.messages[0]["id"], "123")
		is.Equal(q.messages[0]["email"], "foo@example.com")
		is.Equal(q.messages[1]["email"], "bar@example.com")
	})

	t.Run("errors on empty title or body", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nc := &newsletterCreatorMock{}
		q := &multiMessageSenderMock{}
		handlers.CreateNewsletter(mux, zap.NewNop(), nc, q)

		code, _, body := makePostRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"Title":"", "Body":"Welcome to the newsletter."}`))
		is.Equal(code, http.StatusBadRequest)
		is.Equal(body, `{"Error":"invalid value for field 'Title', cannot be empty"}`)

		code, _, body = makePostRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"Title":"Hey you", "Body":""}`))
		is.Equal(code, http.StatusBadRequest)
		is.Equal(body, `{"Error":"invalid value for field 'Body', cannot be empty"}`)
	})

	t.Run("errors on bad connection to database", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nc := &erroringNewsletterCreatorMock{}
		q := &multiMessageSenderMock{}
		handlers.CreateNewsletter(mux, zap.NewNop(), nc, q)

		code, _, body := makePostRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"Title":"Hey you", "Body":"Welcome to the newsletter."}`))
		is.Equal(code, http.StatusBadGateway)
		is.Equal(body, `{"Error":"server error: doesn't work"}`)
	})

	t.Run("errors on bad request JSON", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nc := &newsletterCreatorMock{}
		q := &multiMessageSenderMock{}
		handlers.CreateNewsletter(mux, zap.NewNop(), nc, q)

		code, _, body := makePostRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{`))
		is.Equal(code, http.StatusBadRequest)
		is.Equal(body, `{"Error":"malformed JSON in request body: unexpected EOF"}`)
	})
}

type newsletterGetterMock struct{}

func (n *newsletterGetterMock) GetNewsletters(ctx context.Context) ([]model.Newsletter, error) {
	return []model.Newsletter{
		{ID: "123", Title: "Second", Body: "Content.", Created: now, Updated: now},
		{ID: "234", Title: "First", Body: "Content.", Created: now, Updated: now},
	}, nil
}

func TestGetNewsletters(t *testing.T) {
	t.Run("gets newsletters", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		ng := &newsletterGetterMock{}
		handlers.GetNewsletters(mux, zap.NewNop(), ng)

		code, _, body := makeGetRequest(mux, "/newsletters")
		is.Equal(code, http.StatusOK)
		is.Equal(body, `{"Newsletters":[`+
			`{"ID":"123","Title":"Second","Body":"Content.",`+
			`"Created":"2022-05-17T12:00:00Z","Updated":"2022-05-17T12:00:00Z"},`+
			`{"ID":"234","Title":"First","Body":"Content.",`+
			`"Created":"2022-05-17T12:00:00Z","Updated":"2022-05-17T12:00:00Z"}`+
			`]}`)
	})
}

type newsletterUpdaterMock struct {
	id    string
	title string
	body  string
}

func (n *newsletterUpdaterMock) UpdateNewsletter(ctx context.Context, id, title, body string) (*model.Newsletter, error) {
	n.id = id
	n.title = title
	n.body = body
	return &model.Newsletter{
		ID:      id,
		Title:   title,
		Body:    body,
		Created: now,
		Updated: now,
	}, nil
}

func TestUpdateNewsletter(t *testing.T) {
	t.Run("updates an existing newsletter and returns the newsletter content", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nu := &newsletterUpdaterMock{}
		handlers.UpdateNewsletter(mux, zap.NewNop(), nu)

		code, _, body := makePutRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"ID":"ca4bee7f-2ea8-4bc7-8e4e-ed60298fe765", "Title":"Foo", "Body":"Bar"}`))
		is.Equal(code, http.StatusOK)
		is.Equal(body, `{"Newsletter":{"ID":"ca4bee7f-2ea8-4bc7-8e4e-ed60298fe765","Title":"Foo","Body":"Bar",`+
			`"Created":"2022-05-17T12:00:00Z","Updated":"2022-05-17T12:00:00Z"}}`)
		is.Equal(nu.id, "ca4bee7f-2ea8-4bc7-8e4e-ed60298fe765")
		is.Equal(nu.title, "Foo")
		is.Equal(nu.body, "Bar")
	})

	t.Run("errors if id is not a valid uuid", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nu := &newsletterUpdaterMock{}
		handlers.UpdateNewsletter(mux, zap.NewNop(), nu)

		code, _, body := makePutRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"ID":"123", "Title":"Foo", "Body":"Bar"}`))

		is.Equal(code, http.StatusBadRequest)
		is.Equal(body, `{"Error":"invalid value for field 'ID', must be a valid UUID"}`)
	})
}

type newsletterDeleterMock struct {
	id string
}

func (n *newsletterDeleterMock) DeleteNewsletter(ctx context.Context, id string) (*model.Newsletter, error) {
	n.id = id
	return &model.Newsletter{}, nil
}

func TestDeleteNewsletter(t *testing.T) {
	t.Run("deletes an existing newsletter", func(t *testing.T) {
		is := is.New(t)

		mux := chi.NewMux()
		nd := &newsletterDeleterMock{}
		handlers.DeleteNewsletter(mux, zap.NewNop(), nd)

		code, _, body := makeDeleteRequest(mux, "/newsletters", createJSONHeader(),
			strings.NewReader(`{"ID":"ca4bee7f-2ea8-4bc7-8e4e-ed60298fe765"}`))
		is.Equal(code, http.StatusOK)
		is.Equal(body, `{}`)
	})
}

// makeGetRequest and return the status code, response headers, and the body.
func makeGetRequest(h http.Handler, target string) (int, http.Header, string) {
	return makeRequest(h, http.MethodGet, target, nil, nil)
}

// makePostRequest and return the status code, response header, and the body.
func makePostRequest(h http.Handler, target string, header http.Header, body io.Reader) (int, http.Header, string) {
	return makeRequest(h, http.MethodPost, target, header, body)
}

// makePutRequest and return the status code, response header, and the body.
func makePutRequest(h http.Handler, target string, header http.Header, body io.Reader) (int, http.Header, string) {
	return makeRequest(h, http.MethodPut, target, header, body)
}

// makeDeleteRequest and return the status code, response header, and the body.
func makeDeleteRequest(h http.Handler, target string, header http.Header, body io.Reader) (int, http.Header, string) {
	return makeRequest(h, http.MethodDelete, target, header, body)
}

func makeRequest(h http.Handler, method, target string, header http.Header, body io.Reader) (int, http.Header, string) {
	req := httptest.NewRequest(method, target, body)
	if header != nil {
		req.Header = header
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	result := res.Result()
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	return result.StatusCode, result.Header, strings.TrimSpace(string(bodyBytes))
}

func createJSONHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	return header
}
