package autopilot

import (
	"context"
	"encoding/json"
	"time"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/internal/jsondump"
)

// User represents a user in Autopilot.
type User struct {
	CreatedAt time.Time
	ID        int64
	State     *launchpad.State
	Profile   launchpad.UserProfile
	Resume    cvschema.Resume

	ap Autopilot
}

// User returns a user by their ID.
func (ap *Autopilot) User(ctx context.Context, userID int64) (*User, error) {
	user, err := ap.db.User(ctx, userID)
	if err != nil {
		return nil, err
	}

	state, err := launchpad.LoadState(ap.ai, user.State, user.StateDump)
	if err != nil {
		return nil, err
	}

	var profile launchpad.UserProfile
	if err := json.Unmarshal(user.Profile, &profile); err != nil {
		return nil, err
	}

	var resume cvschema.Resume
	if err := json.Unmarshal(user.Resume, &resume); err != nil {
		return nil, err
	}

	return &User{
		CreatedAt: user.CreatedAt,
		ID:        user.ID,
		State:     state,
		Profile:   profile,
		Resume:    resume,
		ap:        ap,
	}, nil
}

// CreateUserIfNotExists creates a new user in the database if they do not exist.
func (ap *Autopilot) CreateUserIfNotExists(ctx context.Context, userID int64) error {
	exists, err := ap.db.UserExists(ctx, userID)
	if err != nil || exists {
		return err
	}
	if err := ap.db.InsertUser(ctx, userID); err != nil {
		return err
	}
	return ap.db.UpdateUserState(ctx, database.UpdateUserStateParams{
		ID:    userID,
		State: launchpad.StateKickoff,
	})
}

// DumpState dumps the current state and updates it in the database.
func (u *User) DumpState(ctx context.Context) error {
	dump, err := u.State.Dump()
	if err != nil {
		return err
	}
	return u.ap.db.UpdateUserState(ctx, database.UpdateUserStateParams{
		ID:        u.ID,
		State:     u.State.CurrentName(),
		StateDump: dump,
	})
}

func (u *User) updateProfile(ctx context.Context, profile launchpad.UserProfile) error {
	u.Profile = profile
	return u.ap.db.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
		ID:      u.ID,
		Profile: jsondump.DumpBytes(profile),
	})
}

func (u *User) updateResume(ctx context.Context, resume cvschema.Resume, file []byte) error {
	u.Resume = resume
	return u.ap.db.UpdateUserResume(ctx, database.UpdateUserResumeParams{
		ID:         u.ID,
		Resume:     jsondump.DumpBytes(resume),
		ResumeFile: file,
	})
}
