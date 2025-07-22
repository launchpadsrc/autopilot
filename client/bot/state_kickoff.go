package bot

import (
	"context"
	"io"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/autopilot"
)

// DONE: 1. Parse CV from document, adjust the pipeline and profile to consider the CV.
// DONE: 2. Considering the preferences, start "targeting" background task to look for the matching vacancies.
// TODO: 3. Mention the fact that these vacancies will be collected continuously until we form a long-list.
// TODO: 4. The user should be asked to provide a feedback for each sent vacancy.
// TODO: 5. Allow user to send their own vacancies, which will be added to the long-list.
// TODO: 6. Once the long-list of 30 vacancies is formed, move to the next step.
// TODO: 7. While the list is forming, the user can modify their profile, which will affect the targeting process.

func (b Bot) statKickoffHandler(c tele.Context) func(s *autopilot.StateKickoff) error {
	return func(s *autopilot.StateKickoff) error {
		ctx := context.Background()

		if err := c.Send(b.Text(c, "kickoff.profile", s)); err != nil {
			return err
		}
		
		// Process CV if it was not provided yet.
		if !s.HasResume() {
			doc := c.Message().Document
			if doc == nil {
				return nil
			}

			file, err := b.File(&doc.File)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			return s.AddResume(ctx, data)
		}

		// TODO: Profile verification here.

		return s.Complete(ctx)
	}
}
