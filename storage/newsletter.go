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
