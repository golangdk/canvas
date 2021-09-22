package storage_test

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"canvas/integrationtest"
)

func TestDatabase_SignupForNewsletter(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("signs up", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		expectedToken, err := db.SignupForNewsletter(context.Background(), "me@example.com")
		is.NoErr(err)
		is.Equal(64, len(expectedToken))

		var email, token string
		err = db.DB.QueryRow(`select email, token from newsletter_subscribers`).Scan(&email, &token)
		is.NoErr(err)
		is.Equal("me@example.com", email)
		is.Equal(expectedToken, token)

		expectedToken2, err := db.SignupForNewsletter(context.Background(), "me@example.com")
		is.NoErr(err)
		is.True(expectedToken != expectedToken2)

		err = db.DB.QueryRow(`select email, token from newsletter_subscribers`).Scan(&email, &token)
		is.NoErr(err)
		is.Equal("me@example.com", email)
		is.Equal(expectedToken2, token)
	})
}

func TestDatabase_ConfirmNewsletterSignup(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("confirms subscriber from the token and returns the associated email address", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		token, err := db.SignupForNewsletter(context.Background(), "me@example.com")
		is.NoErr(err)

		var confirmed bool
		err = db.DB.Get(&confirmed, `select confirmed from newsletter_subscribers where token = $1`, token)
		is.NoErr(err)
		is.True(!confirmed)

		email, err := db.ConfirmNewsletterSignup(context.Background(), token)
		is.NoErr(err)
		is.Equal("me@example.com", email.String())

		err = db.DB.Get(&confirmed, `select confirmed from newsletter_subscribers where token = $1`, token)
		is.NoErr(err)
		is.True(confirmed)
	})

	t.Run("returns nil if no such token", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		_, err := db.SignupForNewsletter(context.Background(), "me@example.com")
		is.NoErr(err)

		email, err := db.ConfirmNewsletterSignup(context.Background(), "notmytoken")
		is.NoErr(err)
		is.True(email == nil)
	})
}
