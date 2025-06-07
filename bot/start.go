package bot

import (
	"time"

	"github.com/looplab/fsm"
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
)

const (
	StateKickoff   = "kickoff"
	StateTargeting = "targeting"
)

var (
	states = make(map[int64]*fsm.FSM)
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

func (b Bot) onStart(c tele.Context) error {
	// defer b.WithNotify(c, tele.Typing)()

	if err := c.Send(b.Text(c, "welcome")); err != nil {
		return err
	}

	states[c.Sender().ID] = fsm.NewFSM(StateKickoff, fsm.Events{
		{Name: StateTargeting, Src: []string{StateKickoff}, Dst: ""},
	}, nil)

	smartSteps = launchpad.NewSmartSteps(b.ai)

	return b.AfterFunc(1*time.Second, func() error {
		return b.forwardStep(c, "01_kickoff")
	})
}

var smartSteps launchpad.SmartSteps

func (b Bot) onChat(c tele.Context) error {
	// defer b.WithNotify(c, tele.Typing)()

	state, ok := states[c.Sender().ID]
	if !ok {
		return b.sendHint(c, "No state")
	}

	// 1. Store user profile, if answers are sufficient.
	// 2. Forward to user the next step, which asks about five target vacancies they like.
	// 3. Considering the preferences, start "targeting" background task to look for the matching vacancies.
	// 4. Mention the fact that these vacancies will be collected continuously until we form a long-list.
	// 5. The user will be asked to provide a feedback for each sent vacancy.
	// 6. Allow user to send their own vacancies, which will be added to the long-list.
	// 7. Once the long-list of 30 vacancies is formed, move to the next step.

	switch state.Current() {
	case StateKickoff:
		profile, err := smartSteps.Kickoff01UserProfile(c.Text())
		if err != nil {
			return err
		}

		if len(profile.Problems) > 0 {
			if err := c.Send(profile.AssistantResponse); err != nil {
				return err
			}
		} else {
			return b.SendJSON(c, profile)
		}
	}

	return nil
}
