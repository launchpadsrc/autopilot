package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/core/targeting"
	"launchpad.icu/autopilot/database"
	"launchpad.icu/autopilot/database/sqlc"
)

func (b Bot) onDebugTargeting(c tele.Context) (err error) {
	var user *database.User

	args := c.Args()
	if len(args) == 0 {
		user = b.User(c)
	} else {
		userID, _ := strconv.ParseInt(args[0], 10, 64)

		user, err = b.db.User(context.Background(), userID)
		if err != nil {
			return err
		}
	}

	var profile launchpad.UserProfile
	if err := json.Unmarshal(user.Profile, &profile); err != nil {
		return err
	}

	var resume cvschema.Resume
	if err := json.Unmarshal(user.Resume, &resume); err != nil {
		return err
	}

	jobs, err := b.db.UniqueJobs(context.Background(), sqlc.UniqueJobsParams{UserID: user.ID, Limit: 10000})
	if err != nil {
		return fmt.Errorf("unique jobs: %w", err)
	}

	targeted, err := targeting.Find(profile, resume, jobs)
	if err != nil {
		return err
	}

	_ = b.SendJSON(c, targeted)
	
	if len(targeted) > 5 {
		targeted = targeted[:5]
	}

	for _, job := range targeted {
		err := c.Send(
			b.Text(c, "targeting.job", job),
			b.Markup(c, "targeting.job", job),
			tele.NoPreview,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
