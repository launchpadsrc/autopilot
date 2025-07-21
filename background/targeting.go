package background

import (
	"context"
	"encoding/json"
	"fmt"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/core/targeting"
	"launchpad.icu/autopilot/database/sqlc"
)

func (bg Background) Targeting(t Task) Func {
	return func(ctx context.Context) error {
		users, err := bg.db.UsersByState(ctx, launchpad.StateTargeting)
		if err != nil {
			return err
		}

		t.logger.Info("targeting", "users", len(users))

		for _, user := range users {
			if err := bg.targeting(ctx, user); err != nil {
				return err
			}
		}

		return nil
	}
}

func (bg Background) targeting(ctx context.Context, user sqlc.User) error {
	var profile launchpad.UserProfile
	if err := json.Unmarshal(user.Profile, &profile); err != nil {
		return err
	}

	var resume cvschema.Resume
	if err := json.Unmarshal(user.Resume, &resume); err != nil {
		return err
	}

	jobs, err := bg.db.UniqueJobs(ctx, sqlc.UniqueJobsParams{UserID: user.ID, Limit: 10000})
	if err != nil {
		return fmt.Errorf("unique jobs: %w", err)
	}

	targeted, err := targeting.Find(targeting.FindParams{
		Profile:  profile,
		Resume:   resume,
		Jobs:     jobs,
		MinScore: targeting.MinScore,
	})
	if err != nil {
		return err
	}

	for _, job := range targeted {
		if err := bg.db.UpsertUserJob(ctx, sqlc.UpsertUserJobParams{
			UserID:   user.ID,
			JobID:    job.ID,
			Feedback: sqlc.UserJobFeedbackScored,
		}); err != nil {
			return err
		}

		_, err := bg.b.Send(
			tele.ChatID(user.ID),
			bg.b.TextLocale("ua", "targeting.job", job),
			bg.b.MarkupLocale("ua", "targeting.job", job),
			tele.NoPreview,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
