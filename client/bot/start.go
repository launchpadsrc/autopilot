package bot

import (
	"context"
	"time"

	tele "gopkg.in/telebot.v4"
)

func (b Bot) onStart(c tele.Context) error {
	defer b.WithNotify(c, tele.Typing)()

	var (
		ctx    = context.Background()
		userID = c.Sender().ID
	)

	if err := b.ap.CreateUserIfNotExists(ctx, userID); err != nil {
		return err
	}

	if err := c.Send(b.Text(c, "welcome")); err != nil {
		return err
	}

	return b.AfterFunc(2*time.Second, func() error {
		return b.sendCourseStep(c, "01_kickoff")
	})
}

func (b Bot) sendCourseStep(c tele.Context, key string) error {
	return c.Send(
		b.Text(c, "course."+key),
		tele.NoPreview,
	)
}
