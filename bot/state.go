package bot

import (
	"context"
	"strings"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/database"
)

func (b Bot) User(c tele.Context) *database.User {
	user, ok := c.Get("user").(*database.User)
	if !ok {
		panic("wtf: user not found in context")
	}
	return user
}

func (b Bot) State(c tele.Context) *launchpad.State {
	state, ok := c.Get("state").(*launchpad.State)
	if !ok {
		panic("wtf: state not found in context")
	}
	return state
}

func (b Bot) withUserState(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		exists, _ := b.db.UserExists(context.Background(), c.Sender().ID)
		if !exists && strings.HasPrefix(c.Text(), "/start") {
			// First time user, we need to create a user record.
			if err := next(c); err != nil {
				return err
			}
		}

		user, err := b.db.User(context.Background(), c.Sender().ID)
		if err != nil {
			return err
		}

		state, err := launchpad.LoadState(b.ai, user.State, user.StateDump)
		if err != nil {
			return err
		}

		c.Set("user", user)
		c.Set("state", state)

		if exists {
			if err := next(c); err != nil {
				return err
			}
		}

		return b.updateUserState(c)
	}
}

func (b Bot) onChat(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	state := b.State(c)
	stepName, _ := state.Current()

	result, err := state.Execute(c.Text())
	if err != nil {
		return err
	}

	switch stepName {
	case launchpad.StateKickoff:
		expected := launchpad.NewResultOf[launchpad.UserProfile](result)
		return b.onStateKickoff(c, expected)
	}

	return nil
}

func (b Bot) sendCourseStep(c tele.Context, key string) error {
	return c.Send(
		b.Text(c, "course."+key),
		tele.NoPreview,
	)

}

func (b Bot) updateUserState(c tele.Context) error {
	user, state := b.User(c), b.State(c)

	dump, err := state.Dump()
	if err != nil {
		return err
	}

	current, _ := state.Current()
	return user.UpdateState(context.Background(), current, dump)
}
