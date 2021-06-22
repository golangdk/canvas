package views

import (
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func NewsletterThanksPage(path string) g.Node {
	return Page(
		"Thanks for signing up!",
		path,
		H1(g.Text(`Thanks for signing up!`)),
		P(g.Raw(`Now check your inbox (or spam folder) for a confirmation link. 😊`)),
	)
}
