package jobs_test

import (
	"context"
	"errors"
	"fmt"
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
	err     error
	to      model.Email
	giftURL string
}

func (m *mockWelcomeEmailer) SendNewsletterWelcomeEmail(ctx context.Context, to model.Email, giftURL string) error {
	m.to = to
	m.giftURL = giftURL
	return m.err
}

type mockGiftCreator struct{}

func (m *mockGiftCreator) CreateAndSaveNewsletterGift(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("https://example.com/gift-for-%v.png", name), nil
}

func TestSendNewsletterWelcomeEmail(t *testing.T) {
	r := testRegistry{}

	t.Run("passes the recipient email and the gift url to the email sender", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockWelcomeEmailer{}
		gc := &mockGiftCreator{}
		jobs.SendNewsletterWelcomeEmail(r, emailer, gc)

		job, ok := r["welcome_email"]
		is.True(ok)

		err := job(context.Background(), model.Message{"email": "you@example.com"})
		is.NoErr(err)

		is.Equal("you@example.com", emailer.to.String())
		is.Equal("https://example.com/gift-for-you.png", emailer.giftURL)
	})

	t.Run("errors on email sending failure", func(t *testing.T) {
		is := is.New(t)

		emailer := &mockWelcomeEmailer{err: errors.New("email server down")}
		gc := &mockGiftCreator{}
		jobs.SendNewsletterWelcomeEmail(r, emailer, gc)
		job := r["welcome_email"]

		err := job(context.Background(), model.Message{"email": "you@example.com"})
		is.True(err != nil)
	})
}
