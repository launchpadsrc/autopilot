package bot

import (
	"net/url"
	"os"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/parsers"
)

func (b Bot) onDebugSetProxy(c tele.Context) error {
	proxy := c.Args()[0]

	// Validate.
	if _, err := url.ParseRequestURI(proxy); err != nil {
		return err
	}

	// Run tests.
	for _, parser := range b.parsers {
		_, err := parsers.WithProxy(parser, proxy).ParseFeed()
		if err != nil {
			return err
		}
	}

	// Update globally.
	if err := os.Setenv("FEEDER_PROXY", proxy); err != nil {
		return err
	}

	return c.Send("true")
}
