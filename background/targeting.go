package background

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/samber/lo"
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/database/sqlc"
)

func (bg Background) Targeting(t Task) func(ctx context.Context) error {
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

	params := sqlc.ScoredJobsParams{
		UserID:         user.ID,
		Hashtags:       profile.StackTags(),
		RolePatterns:   profile.RolePatterns(),
		ResumeKeywords: resume.Keywords(),
		Limit:          5,
	}

	slog.Debug(
		"targeting params",
		"user_id", user.ID,
		"hashtags", params.Hashtags,
		"role_patterns", params.RolePatterns,
		"keywords", params.ResumeKeywords,
	)

	jobs, err := bg.db.ScoredJobs(ctx, params)
	if err != nil || len(jobs) == 0 {
		return err
	}

	jobs = lo.Filter(jobs, func(job sqlc.ScoredJobsRow, _ int) bool {
		return job.Score >= 3.0 // TODO: const
	})

	for _, job := range jobs {
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
