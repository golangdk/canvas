package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"canvas/model"
)

// SignupForNewsletter with the given email. Returns a token used for confirming the email address.
func (d *Database) SignupForNewsletter(ctx context.Context, email model.Email) (string, error) {
	token, err := createSecret()
	if err != nil {
		return "", err
	}
	query := `
		insert into newsletter_subscribers (email, token)
		values ($1, $2)
		on conflict (email) do update set
			token = excluded.token,
			updated = now()`
	_, err = d.DB.ExecContext(ctx, query, email, token)
	return token, err
}

// ConfirmNewsletterSignup with the given token. Returns the associated email if matched.
func (d *Database) ConfirmNewsletterSignup(ctx context.Context, token string) (*model.Email, error) {
	var email model.Email
	query := `
		update newsletter_subscribers
		set confirmed = true
		where token = $1
		returning email`
	err := d.DB.GetContext(ctx, &email, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &email, nil
}

func createSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", secret), nil
}

// GetSubscribers that are both confirmed and active.
func (d *Database) GetSubscribers(ctx context.Context) ([]model.Email, error) {
	var subscribers []model.Email
	query := `
		select email from newsletter_subscribers
		where confirmed and active
		order by created`
	err := d.DB.SelectContext(ctx, &subscribers, query)
	return subscribers, err
}

// CreateNewsletter with a title and body text, returning the newsletter.
func (d *Database) CreateNewsletter(ctx context.Context, title, body string) (model.Newsletter, error) {
	var n model.Newsletter
	query := `
		insert into newsletters (title, body)
		values ($1, $2)
		returning *`
	err := d.DB.GetContext(ctx, &n, query, title, body)
	return n, err
}

// GetNewsletters ordered reverse-chronologically.
func (d *Database) GetNewsletters(ctx context.Context) ([]model.Newsletter, error) {
	var newsletters []model.Newsletter
	query := `
		select * from newsletters
		order by created desc`
	err := d.DB.SelectContext(ctx, &newsletters, query)
	return newsletters, err
}

// GetNewsletter by id, or nil if the given ID does not exist.
func (d *Database) GetNewsletter(ctx context.Context, id string) (*model.Newsletter, error) {
	var n model.Newsletter
	query := `
		select * from newsletters
		where id = $1`
	if err := d.DB.GetContext(ctx, &n, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// UpdateNewsletter by ID, returning the newsletter if the ID exists. If not, return nil.
func (d *Database) UpdateNewsletter(ctx context.Context, id, title, body string) (*model.Newsletter, error) {
	var n model.Newsletter
	query := `
		update newsletters
		set
		    title = $1,
		    body = $2,
		    updated = now()
		where id = $3
		returning *`
	if err := d.DB.GetContext(ctx, &n, query, title, body, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// DeleteNewsletter by ID, returning the newsletter if the ID exists. If not, return nil.
func (d *Database) DeleteNewsletter(ctx context.Context, id string) (*model.Newsletter, error) {
	var n model.Newsletter
	query := `
		delete from newsletters
		where id = $1
		returning *`
	err := d.DB.GetContext(ctx, &n, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// SearchNewsletters with the given search query, and order by relevancy.
// Note that the searchtext should match the search index from the table exactly, or the search will not use the index.
// See https://www.postgresql.org/docs/current/textsearch.html
func (d *Database) SearchNewsletters(ctx context.Context, searchQuery string) ([]model.Newsletter, error) {
	var newsletters []model.Newsletter
	query := `
		select id, title, body, created, updated
		from newsletters,
			websearch_to_tsquery('english', $1) searchquery,
			to_tsvector('english', title || ' ' || body) searchtext
		where searchquery @@ searchtext
		order by ts_rank(searchtext, searchquery) desc`
	err := d.DB.SelectContext(ctx, &newsletters, query, searchQuery)
	return newsletters, err
}
