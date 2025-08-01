package bot

import (
	"time"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/autopilot"
)

func (b Bot) onFeederUniqueJob(job autopilot.FeederJob) error {
	_, err := b.Send(
		b.ChatID("jobs_channel"),
		b.TextLocale("ua", "feeder.job", job),
		tele.NoPreview,
	)
	if err != nil {
		// TODO: Implement queue instead?
		logger.Error("sending job entry", "error", err)
		time.Sleep(time.Second) // in case a flood wait occurred
	}
	return nil
}
