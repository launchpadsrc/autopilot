package bot

import (
	"context"
	"io"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
)

func (b Bot) onStateKickoff(c tele.Context, result *launchpad.ResultOf[launchpad.UserProfile]) error {
	user, state := b.User(c), b.State(c)

	// Process CV if it was not provided yet.
	if len(user.Resume) == 0 {
		if err := b.updateUserResume(c); err != nil {
			return err
		}
	}

	if len(result.Value.Problems) > 0 {
		return c.Send(result.Response)
	}

	if err := user.UpdateProfile(context.Background(), result.Value); err != nil {
		return err
	}

	return state.Transition()

	// DONE: 1. Parse CV from document, adjust the pipeline and profile to consider the CV.
	// DONE: 2. Considering the preferences, start "targeting" background task to look for the matching vacancies.
	// TODO: 3. Mention the fact that these vacancies will be collected continuously until we form a long-list.
	// TODO: 4. The user should be asked to provide a feedback for each sent vacancy. (human-in-the-loop)?
	// TODO: 5. Allow user to send their own vacancies, which will be added to the long-list.
	// TODO: 6. Once the long-list of 30 vacancies is formed, move to the next step.
	// TODO: 7. While the list is forming, the user can modify their profile, which will affect the targeting process.
}

func (b Bot) updateUserResume(c tele.Context) error {
	doc := c.Message().Document
	if doc == nil {
		return c.Send(b.Text(c, "kickoff.add_resume"))
	}

	file, err := b.File(&doc.File)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	resume, err := cvschema.NewParser(b.ai).Parse(data)
	if err != nil {
		return err
	}

	return b.User(c).UpdateResume(context.Background(), resume, data)
}
