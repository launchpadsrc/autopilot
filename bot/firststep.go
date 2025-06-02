package bot

import (
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/launchpad"
)

func (b Bot) onFirstStep(c tele.Context) error {
	answers := c.Text()

	profile, err := launchpad.NewSmartSteps(b.ai).Kickoff01(answers)
	if err != nil {
		return err
	}

	return b.SendJSON(c, profile)
}
