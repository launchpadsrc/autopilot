package bot

import (
	"context"
	"strconv"

	tele "gopkg.in/telebot.v4"
)

func (b Bot) onDebugTargeting(c tele.Context) (err error) {
	var (
		ctx  = context.Background()
		user = b.User(c)
		args = c.Args()
	)

	if len(args) > 0 {
		userID, _ := strconv.ParseInt(args[0], 10, 64)

		user, err = b.ap.User(context.Background(), userID)
		if err != nil {
			return err
		}
	}

	jobs, err := b.ap.TargetJobs(ctx, user)
	if err != nil {
		return err
	}

	for i, job := range jobs {
		if i == 5 {
			break
		}
		err := c.Send(
			b.Text(c, "targeting.job", job),
			b.Markup(c, "targeting.job", job),
			tele.NoPreview,
		)
		if err != nil {
			return err
		}
	}

	return b.SendJSON(c, jobs)
}
