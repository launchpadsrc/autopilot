package bot

import (
	"io"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/cvschema"
)

func (b Bot) onResume(c tele.Context) error {
	doc := c.Message().Document
	if doc == nil {
		return c.Send("💡 Attach a PDF resume.")
	}

	rc, err := b.File(&doc.File)
	if err != nil {
		return err
	}
	defer rc.Close()

	pdf, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	go c.Notify(tele.Typing)

	resume, err := cvschema.NewParser(b.ai).Parse(pdf)
	if err != nil {
		return err
	}

	return b.SendJSON(c, resume)
}
