package bot

import (
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/autopilot"
)

func (b Bot) onChat(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	actions := autopilot.StateActions{
		Kickoff: b.statKickoffHandler(c),
	}

	sm, err := b.ap.StateMachine(b.User(c), actions)
	if err != nil {
		return err
	}

	// TODO: Empty input is not handled.
	// TODO: In case a manual interaction is needed, there is no adequate control over it.
	// TODO: For example, a resume cannot be added if the message's caption is empty.

	return sm.Entrypoint(c.Text())
}
