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
