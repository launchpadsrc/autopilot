package bot

import (
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/autopilot"
	"launchpad.icu/autopilot/core/targeting"
)

func (b Bot) onTargetingJob(user *autopilot.User, job targeting.Job) error {
	_, err := b.Send(
		tele.ChatID(user.ID),
		b.TextLocale("ua", "targeting.job", job),
		b.MarkupLocale("ua", "targeting.job", job),
		tele.NoPreview,
	)
	if err != nil {
		logger.Error(
			"sending targeted job",
			"user", user.ID,
			"job", job.ID,
			"error", err,
		)
	}
	return nil
}
