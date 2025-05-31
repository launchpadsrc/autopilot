package htmlstrip

import (
	"strings"

	"golang.org/x/net/html"
)

// Strip returns the plain-text content of an HTML fragment.
// If parsing fails we fall back to the original string.
func Strip(src string) string {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return src
	}

	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Convert '&amp;' → '&', '&#8217;' → '’', ...
	return strings.TrimSpace(html.UnescapeString(b.String()))
}
