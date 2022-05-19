package storage_test

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"canvas/integrationtest"
	"canvas/model"
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

func TestDatabase_GetSubscribers(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("gets newsletter subscribers", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		for _, e := range []model.Email{"foo@example.com", "bar@example.com"} {
			token, err := db.SignupForNewsletter(context.Background(), e)
			is.NoErr(err)
			_, err = db.ConfirmNewsletterSignup(context.Background(), token)
			is.NoErr(err)
		}

		subscribers, err := db.GetSubscribers(context.Background())
		is.NoErr(err)
		is.Equal(len(subscribers), 2)
		is.Equal(subscribers[0].String(), "foo@example.com")
		is.Equal(subscribers[1].String(), "bar@example.com")
	})
}

func TestDatabase_CreateNewsletter(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("creates a newsletter with a title and body, and returns the newsletter", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		n, err := db.CreateNewsletter(context.Background(), "Welcome to Canvas!", "This is Canvas.")
		is.NoErr(err)
		is.Equal(len(n.ID), 36)
		is.Equal(n.Title, "Welcome to Canvas!")
		is.Equal(n.Body, "This is Canvas.")
	})
}

func TestDatabase_GetNewsletters(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("gets all newsletters in reverse chronological order", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		_, err := db.CreateNewsletter(context.Background(), "First", "")
		is.NoErr(err)

		_, err = db.CreateNewsletter(context.Background(), "Second", "")
		is.NoErr(err)

		newsletters, err := db.GetNewsletters(context.Background())
		is.NoErr(err)

		is.Equal(len(newsletters), 2)
		is.Equal(newsletters[0].Title, "Second")
		is.Equal(newsletters[1].Title, "First")
	})
}

func TestDatabase_GetNewsletter(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("gets a single newsletter, or nil if no such id", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		n1, err := db.CreateNewsletter(context.Background(), "", "")
		is.NoErr(err)

		n2, err := db.GetNewsletter(context.Background(), n1.ID)
		is.NoErr(err)
		is.True(n2 != nil)
		is.Equal(n1.ID, n2.ID)

		n3, err := db.GetNewsletter(context.Background(), "b1c0cae5-28d1-4982-935e-08ba511cb466")
		is.NoErr(err)
		is.True(n3 == nil)
	})
}

func TestDatabase_UpdateNewsletter(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("updates newsletter content and updated timestamp", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		n1, err := db.CreateNewsletter(context.Background(), "", "")
		is.NoErr(err)

		n2, err := db.UpdateNewsletter(context.Background(), n1.ID, "Welcome", "Hi.")
		is.NoErr(err)
		is.Equal(n1.ID, n2.ID)
		is.Equal(n2.Title, "Welcome")
		is.Equal(n2.Body, "Hi.")
		is.True(n2.Updated.After(n2.Created))
	})
}

func TestDatabase_DeleteNewsletter(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("deletes a newsletter", func(t *testing.T) {
		is := is.New(t)
		db, cleanup := integrationtest.CreateDatabase()
		defer cleanup()

		n1, err := db.CreateNewsletter(context.Background(), "", "")
		is.NoErr(err)

		n2, err := db.DeleteNewsletter(context.Background(), n1.ID)
		is.NoErr(err)
		is.Equal(n1.ID, n2.ID)

		n3, err := db.GetNewsletter(context.Background(), n1.ID)
		is.NoErr(err)
		is.True(n3 == nil)

		n4, err := db.DeleteNewsletter(context.Background(), n1.ID)
		is.NoErr(err)
		is.True(n4 == nil)
	})
}
