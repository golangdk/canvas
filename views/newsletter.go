package views

import (
	"net/url"
	"strings"

	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/solid"
	. "github.com/maragudk/gomponents/html"

	"canvas/model"
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

func NewslettersPage(path string, newsletters []model.Newsletter, search string) g.Node {
	return Page(
		"Newsletters",
		path,
		H1(g.Text(`Newsletters`)),
		P(Class("lead"),
			g.Text("This is our newsletter archive. Click the link beneath the title to read the newsletter."),
		),

		FormEl(Action("/newsletters"), Method("get"), Class("flex items-center max-w-md"),
			Label(For("search"), Class("sr-only"), g.Text("Search")),
			Div(Class("relative rounded-md shadow-sm flex-grow"),
				Div(Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
					solid.Search(Class("h-5 w-5 text-gray-400")),
				),
				Input(Type("text"), Name("search"), ID("search"), Value(search), TabIndex("1"),
					Class("focus:ring-gray-500 focus:border-gray-500 block w-full pl-10 text-sm border-gray-300 rounded-md")),
			),
		),

		g.If(search != "",
			P(g.Textf(`%v search result%v for your search on `, len(newsletters), plural(len(newsletters))), Strong(g.Text(search))),
		),

		g.Group(g.Map(newsletters, func(n model.Newsletter) g.Node {
			return NewsletterSummary(n, search)
		})),
	)
}

func plural(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}

const timeFormat = "Monday January 2 2006 at 15:04:05 MST"

func NewsletterSummary(n model.Newsletter, search string) g.Node {
	return Div(
		H2(g.Text(n.Title)),
		P(g.Textf("From %v.", n.Created.Format(timeFormat))),
		P(A(Href("/newsletters?id="+n.ID+"&search="+url.QueryEscape(search)), g.Textf("Read “%v”.", n.Title))),
	)
}

func NewsletterPage(path string, n model.Newsletter, search string) g.Node {
	paragraphs := strings.Split(n.Body, "\n\n")
	return Page(
		n.Title,
		path,
		H1(g.Text(n.Title)),
		P(
			g.Textf("Published %v.", n.Created.Format(timeFormat)),
			g.If(n.Updated.After(n.Created),
				g.Textf(" Last updated %v.", n.Updated.Format(timeFormat)),
			),
		),
		g.Group(g.Map(paragraphs, func(p string) g.Node {
			return P(g.Text(p))
		})),
		g.If(search == "",
			P(A(Href("/newsletters"), g.Text("Go back to the overview."))),
		),
		g.If(search != "",
			P(A(Href("/newsletters?search="+url.QueryEscape(search)),
				g.Text("Go back to your search for "), Em(g.Text(search)), g.Text("."))),
		),
	)
}
