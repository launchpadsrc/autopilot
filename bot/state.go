package bot

import (
	"context"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/database"
	"launchpad.icu/autopilot/database/sqlc"
)

func (b Bot) forwardStep(c tele.Context, key string) error {
	const (
		channel = -1002533811868
	)
	return c.Forward(&tele.StoredMessage{
		MessageID: b.String("steps." + key),
		ChatID:    channel,
	})
}

func (b Bot) mwState(next tele.HandlerFunc) tele.HandlerFunc {
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

		var previous string
		if exists {
			previous, _ = state.Current()
			if err := next(c); err != nil {
				return err
			}
		}

		current, _ := state.Current()
		if previous != current {
			return b.updateUserState(c)
		}

		return nil
	}
}

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

func (b Bot) updateUserState(c tele.Context) error {
	user, state := b.User(c), b.State(c)

	dump, err := state.Dump()
	if err != nil {
		return err
	}

	current, _ := state.Current()
	return user.UpdateState(context.Background(), current, dump)
}

func (b Bot) onStart(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()
	var (
		ctx    = context.Background()
		userID = c.Sender().ID
	)

	exists, _ := b.db.UserExists(ctx, userID)
	if !exists {
		if err := b.db.InsertUser(ctx, userID); err != nil {
			return err
		}
		if err := b.db.UpdateUserState(ctx, sqlc.UpdateUserStateParams{
			ID:    userID,
			State: launchpad.StateKickoff,
		}); err != nil {
			return err
		}
	}

	if err := c.Send(b.Text(c, "welcome")); err != nil {
		return err
	}

	return b.AfterFunc(2*time.Second, func() error {
		return b.forwardStep(c, "01_kickoff")
	})
}

func (b Bot) onChat(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	state := b.State(c)
	stepName, _ := state.Current()

	result, err := state.Execute(c.Text())
	if err != nil {
		return err
	}

	if result.Problems {
		return c.Send(result.Response)
	}

	switch stepName {
	case launchpad.StateKickoff:
		expected := launchpad.NewResultOf[launchpad.UserProfile](result)
		return b.onStateKickoff(c, expected)
	}

	return nil
}

func (b Bot) onStateKickoff(c tele.Context, result *launchpad.ResultOf[launchpad.UserProfile]) error {
	user, state := b.User(c), b.State(c)

	if err := user.UpdateProfile(context.Background(), result.Value); err != nil {
		return err
	}

	return state.Transition()

	// TODO: 1. Ask user to verify their profile before changing state.
	// DONE: 2. Considering the preferences, start "targeting" background task to look for the matching vacancies.
	// TODO: 3. Mention the fact that these vacancies will be collected continuously until we form a long-list.
	// TODO: 4. The user should be asked to provide a feedback for each sent vacancy.
	// TODO: 5. Allow user to send their own vacancies, which will be added to the long-list.
	// TODO: 6. Once the long-list of 30 vacancies is formed, move to the next step.
}
