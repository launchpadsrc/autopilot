package autopilot

import (
	"cmp"
	"context"
	"time"

	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/core/targeting"
	"launchpad.icu/autopilot/internal/database"
)

// StartTargeting starts the background Targeting task that processes users with state "targeting".
// The relevant unique jobs will be added to the user's current longlist.
func (ap *Autopilot) StartTargeting(d time.Duration) {
	go ap.startBackground(ap.targetingTask, cmp.Or(d, time.Minute))
}

// TargetJobs returns a list of unique jobs that match the user's profile and resume.
// Returns an error if the user's profile or resume is missing.
func (ap *Autopilot) TargetJobs(ctx context.Context, user *User) ([]targeting.Job, error) {
	const (
		// How many jobs to fetch from DB.
		uniqueJobsLimit = 10000
		// The higher the score, the more relevant the job is.
		targetingMinScore = 5
	)

	jobs, err := ap.db.UniqueJobs(ctx, database.UniqueJobsParams{
		UserID: user.ID,
		Limit:  uniqueJobsLimit,
	})
	if err != nil {
		return nil, err
	}

	return targeting.Find(targeting.FindParams{
		Profile:  user.Profile,
		Resume:   user.Resume,
		Jobs:     jobs,
		MinScore: targetingMinScore,
	})
}

func (ap *Autopilot) targetingTask(t bgTask) bgFunc {
	return func(ctx context.Context) error {
		users, err := ap.db.UsersByState(ctx, launchpad.StateTargeting)
		if err != nil {
			return err
		}

		t.logger.Info("targeting", "users", len(users))

		for _, u := range users {
			user, err := ap.User(ctx, u.ID)
			if err != nil {
				return err
			}

			if err := ap.targeting(ctx, user); err != nil {
				return err
			}
		}

		return nil
	}
}

func (ap *Autopilot) targeting(ctx context.Context, user *User) error {
	targeted, err := ap.TargetJobs(ctx, user)
	if err != nil {
		return err
	}

	for _, job := range targeted {
		if err := ap.db.UpsertUserJob(ctx, database.UpsertUserJobParams{
			UserID:   user.ID,
			JobID:    job.ID,
			Feedback: database.UserJobFeedbackScored,
		}); err != nil {
			return err
		}

		if ap.callbacks.OnTargetingJob != nil {
			if err := ap.callbacks.OnTargetingJob(user, job); err != nil {
				return err
			}
		}
	}

	return nil
}
