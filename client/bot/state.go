package bot

import (
	"context"
	"strings"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/autopilot"
)

func (b Bot) User(c tele.Context) *autopilot.User {
	user, ok := c.Get("user").(*autopilot.User)
	if !ok {
		panic("wtf: user not found in context")
	}
	return user
}

func (b Bot) withUser(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		// Ignore the start command.
		if strings.HasPrefix(c.Text(), "/start") {
			return next(c)
		}

		var (
			ctx = context.Background()
		)

		user, err := b.ap.User(ctx, c.Sender().ID)
		if err != nil {
			return err
		}

		c.Set("user", user)
		if err := next(c); err != nil {
			return err
		}

		return user.DumpState(ctx)
	}
}
