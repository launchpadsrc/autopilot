package bot

import (
	"context"
	"time"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/database/sqlc"
)

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

func (b Bot) onReset(c tele.Context) error {
	state := b.State(c)
	state.FSM.SetState(launchpad.StateKickoff)
	state.Clear()

	return b.db.ResetUser(context.Background(), sqlc.ResetUserParams{
		ID:    c.Sender().ID,
		State: launchpad.StateKickoff,
	})
}
