package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"canvas/model"
)

type emailSender interface {
	SendNewsletterConfirmationEmail(ctx context.Context, to model.Email, token string) error
}

// SendNewsletterConfirmationEmail to a newsletter subscriber.
func SendNewsletterConfirmationEmail(r registry, es emailSender) {
	r.Register("confirmation_email", func(ctx context.Context, m model.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		token, ok := m["token"]
		if !ok {
			return errors.New("no token in message")
		}

		if err := es.SendNewsletterConfirmationEmail(ctx, model.Email(to), token); err != nil {
			return fmt.Errorf("error sending newsletter confirmation email: %w", err)
		}

		return nil
	})
}
