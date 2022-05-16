package views

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/outline"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

// Page with a title, head, and a basic body layout.
func Page(title, path string, body ...g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "en",
		Head: []g.Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
		},
		Body: []g.Node{
			Navbar(path),
			Container(true,
				Prose(g.Group(body)),
			),
		},
	})
}

func Navbar(path string) g.Node {
	return Nav(Class("bg-white shadow"),
		Container(false,
			Div(Class("flex items-center space-x-4 h-16"),
				Div(Class("flex-shrink-0"), outline.Globe(Class("h-6 w-6"))),
				NavbarLink("/", "Home", path),
			),
		),
	)
}

func NavbarLink(path, text, currentPath string) g.Node {
	active := path == currentPath
	return A(Href(path), g.Text(text),
		c.Classes{
			"text-lg font-medium hover:text-indigo-900": true,
			"text-indigo-700": active,
			"text-indigo-500": !active,
		},
	)
}

func Container(padY bool, children ...g.Node) g.Node {
	return Div(
		c.Classes{
			"max-w-7xl mx-auto px-4 sm:px-6 lg:px-8": true,
			"py-4 sm:py-6 lg:py-8":                   padY,
		},
		g.Group(children),
	)
}

func Prose(children ...g.Node) g.Node {
	return Div(Class("prose lg:prose-lg xl:prose-xl prose-indigo"), g.Group(children))
}
