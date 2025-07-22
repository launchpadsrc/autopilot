package autopilot

import (
	"context"

	"launchpad.icu/autopilot/core/targeting"
	"launchpad.icu/autopilot/internal/database"
)

// TargetJobs returns a list of unique jobs that match the user's profile and resume.
// Returns an error if the user's profile or resume is missing.
func (ap Autopilot) TargetJobs(ctx context.Context, user *User) ([]targeting.Job, error) {
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
