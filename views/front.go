package views

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/solid"
	. "github.com/maragudk/gomponents/html"
)

func FrontPage() g.Node {
	return Page(
		"Canvas",
		"/",
		H1(g.Text(`Solutions to problems.`)),
		P(g.Raw(`Do you have problems? We also had problems.`)),
		P(g.Raw(`Then we created the <em>canvas</em> app, and now we don't! 😬`)),

		H2(g.Text(`Do you want to know more?`)),
		P(g.Text(`Sign up to our newsletter below.`)),

		FormEl(Action("/newsletter/signup"), Method("post"), Class("flex items-center max-w-md"),
			Label(For("email"), Class("sr-only"), g.Text("Email")),
			Div(Class("relative rounded-md shadow-sm flex-grow"),
				Div(Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
					solid.Mail(Class("h-5 w-5 text-gray-400")),
				),
				Input(Type("email"), Name("email"), ID("email"), AutoComplete("email"), Required(), Placeholder("me@example.com"), TabIndex("1"),
					Class("focus:ring-gray-500 focus:border-gray-500 block w-full pl-10 text-sm border-gray-300 rounded-md")),
			),
			Button(Type("submit"), g.Text("Sign up"),
				Class("ml-3 inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 flex-none")),
		),
	)
}
