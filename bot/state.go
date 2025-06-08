package bot

import (
	"time"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/pkg/wrap"
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

var (
	// FIXME: Use a proper storage for user states.
	states = make(map[int64]*launchpad.State)
)

func (b Bot) onStart(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	if err := c.Send(b.Text(c, "welcome")); err != nil {
		return err
	}

	// Initialize user's state for the first time.
	states[c.Sender().ID] = launchpad.NewState(b.ai)

	return b.AfterFunc(2*time.Second, func() error {
		return b.forwardStep(c, "01_kickoff")
	})
}

func (b Bot) onChat(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	state, ok := states[c.Sender().ID]
	if !ok {
		return b.sendHint(c, "No state")
	}

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
		profile, _ := wrap.Unwrap[launchpad.UserProfile](result.Wrapped)
		// 1. Ask user to verify their profile, store it if they confirm.
		// 2. Forward to user the next step, which asks about five target vacancies they like.
		// 3. Considering the preferences, start "targeting" background task to look for the matching vacancies.
		// 4. Mention the fact that these vacancies will be collected continuously until we form a long-list.
		// 5. The user will be asked to provide a feedback for each sent vacancy.
		// 6. Allow user to send their own vacancies, which will be added to the long-list.
		// 7. Once the long-list of 30 vacancies is formed, move to the next step.
		return b.SendJSON(c, profile)
	}

	return nil

}
