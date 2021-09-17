package jobs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matryer/is"

	"canvas/jobs"
	"canvas/model"
)

type mockConfirmationEmailer struct {
	err   error
	to    model.Email
	token string
}

func (m *mockConfirmationEmailer) SendNewsletterConfirmationEmail(ctx context.Context, to model.Email, token string) error {
	m.to = to
	m.token = token
	return m.err
}

func TestSendConfirmationEmail(t *testing.T) {
	t.Run("passes the recipient email and token to the email sender", func(t *testing.T) {
		is := is.New(t)

		r := testRegistry{}
		emailer := &mockConfirmationEmailer{}
		jobs.SendNewsletterConfirmationEmail(r, emailer)

		fn, ok := r["confirmation_email"]
		is.True(ok)

		err := fn(context.Background(), model.Message{"email": "you@example.com", "token": "123"})
		is.NoErr(err)

		is.Equal("you@example.com", emailer.to.String())
		is.Equal("123", emailer.token)
	})

	t.Run("errors on email sending failure", func(t *testing.T) {
		is := is.New(t)

		r := testRegistry{}
		emailer := &mockConfirmationEmailer{err: errors.New("wire is cut")}
		jobs.SendNewsletterConfirmationEmail(r, emailer)
		fn := r["confirmation_email"]

		err := fn(context.Background(), model.Message{"email": "you@example.com", "token": "123"})
		is.True(err != nil)
	})
}
