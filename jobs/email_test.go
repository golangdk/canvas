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
	r := testRegistry{}

	t.Run("passes the recipient email and token to the email sender", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockConfirmationEmailer{}
		jobs.SendNewsletterConfirmationEmail(r, emailer)

		job, ok := r["confirmation_email"]
		is.True(ok)

		err := job(context.Background(), model.Message{"email": "you@example.com", "token": "123"})
		is.NoErr(err)

		is.Equal("you@example.com", emailer.to.String())
		is.Equal("123", emailer.token)
	})

	t.Run("errors on email sending failure", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockConfirmationEmailer{err: errors.New("wire is cut")}
		jobs.SendNewsletterConfirmationEmail(r, emailer)
		job := r["confirmation_email"]

		err := job(context.Background(), model.Message{"email": "you@example.com", "token": "123"})
		is.True(err != nil)
	})
}

type mockWelcomeEmailer struct {
	err error
	to  model.Email
}

func (m *mockWelcomeEmailer) SendNewsletterWelcomeEmail(ctx context.Context, to model.Email) error {
	m.to = to
	return m.err
}

func TestSendNewsletterWelcomeEmail(t *testing.T) {
	r := testRegistry{}

	t.Run("passes the recipient email to the email sender", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockWelcomeEmailer{}
		jobs.SendNewsletterWelcomeEmail(r, emailer)

		job, ok := r["welcome_email"]
		is.True(ok)

		err := job(context.Background(), model.Message{"email": "you@example.com"})
		is.NoErr(err)

		is.Equal("you@example.com", emailer.to.String())
	})

	t.Run("errors on email sending failure", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockWelcomeEmailer{err: errors.New("email server down")}
		jobs.SendNewsletterWelcomeEmail(r, emailer)
		job := r["welcome_email"]

		err := job(context.Background(), model.Message{"email": "you@example.com"})
		is.True(err != nil)
	})
}
