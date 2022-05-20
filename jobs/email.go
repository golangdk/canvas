package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"canvas/model"
)

type newsletterconfirmationEmailSender interface {
	SendNewsletterConfirmationEmail(ctx context.Context, to model.Email, token string) error
}

// SendNewsletterConfirmationEmail to a newsletter subscriber.
func SendNewsletterConfirmationEmail(r registry, es newsletterconfirmationEmailSender) {
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

type newsletterWelcomeEmailSender interface {
	SendNewsletterWelcomeEmail(ctx context.Context, to model.Email, giftURL string) error
}

type giftCreator interface {
	CreateAndSaveNewsletterGift(ctx context.Context, name string) (string, error)
}

func SendNewsletterWelcomeEmail(r registry, es newsletterWelcomeEmailSender, gc giftCreator) {
	r.Register("welcome_email", func(ctx context.Context, m model.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		email := model.Email(to)
		giftURL, err := gc.CreateAndSaveNewsletterGift(ctx, email.Local())
		if err != nil {
			return fmt.Errorf("error creating welcome gift: %w", err)
		}

		if err := es.SendNewsletterWelcomeEmail(ctx, email, giftURL); err != nil {
			return fmt.Errorf("error sending newsletter welcome email: %w", err)
		}

		return nil
	})
}

type newsletterSender interface {
	SendNewsletterEmail(ctx context.Context, to model.Email, title, body string) error
}

type newsletterGetter interface {
	GetNewsletter(ctx context.Context, id string) (*model.Newsletter, error)
}

// SendNewsletter to a single subscriber by getting the content from the database using the id in the message.
func SendNewsletter(r registry, es newsletterSender, ng newsletterGetter) {
	r.Register("newsletter_email", func(ctx context.Context, m model.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		id, ok := m["id"]
		if !ok {
			return errors.New("no id in message")
		}

		n, err := ng.GetNewsletter(ctx, id)
		if err != nil {
			return fmt.Errorf("error getting newsletter: %w", err)
		}
		if n == nil {
			return fmt.Errorf("no newsletter with id %v", id)
		}

		if err := es.SendNewsletterEmail(ctx, model.Email(to), n.Title, n.Body); err != nil {
			return fmt.Errorf("error sending newsletter email: %w", err)
		}

		return nil
	})
}
