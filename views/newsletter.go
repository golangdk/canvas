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

func NewsletterConfirmPage(path string, token string) g.Node {
	return Page(
		"Confirm your newsletter subscription",
		path,
		H1(g.Text(`Confirm your newsletter subscription`)),
		P(g.Text(`Press the big button below to confirm your subscription.`)),
		FormEl(Action("/newsletter/confirm"), Method("post"),
			Input(Type("hidden"), Name("token"), Value(token)),
			Button(Type("submit"), g.Text("Confirm"),
				Class("inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 flex-none")),
		),
	)
}

func NewsletterConfirmedPage(path string) g.Node {
	return Page(
		"Newsletter subscription confirmed",
		path,
		H1(g.Text(`Newsletter subscription confirmed`)),
		P(g.Textf(`You will now receive the newsletter. 😎`)),
	)
}
