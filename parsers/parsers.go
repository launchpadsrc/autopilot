package parsers

import (
	"log/slog"
	"net/http"
	"net/url"
)

type Parser interface {
	client() *http.Client
	Host() string
	ParseJob(url string) (*Job, error)
	ParseFeed() ([]FeedEntry, error)
}

type Job struct {
	URL            string
	Title          string
	Description    string
	DatePosted     string
	EmploymentType string
	Industry       string
	LocationType   string
	ValidThrough   string
}

func WithProxy[P Parser](parser P, proxy string) P {
	if proxy == "" {
		return parser
	}

	uri, err := url.Parse(proxy)
	if err != nil {
		slog.Warn(
			"proxy is not enabled",
			"error", err,
			"proxy", proxy,
			"parser", parser.Host(),
		)
		return parser
	}

	c := parser.client()
	c.Transport = &http.Transport{Proxy: http.ProxyURL(uri)}
	return parser
}
