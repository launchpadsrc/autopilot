package database

import (
	"context"

	"launchpad.icu/autopilot/database/sqlc"
	"launchpad.icu/autopilot/pkg/jsondump"
)

type User struct {
	sqlc.User
	q *sqlc.Queries
}

func (db DB) User(ctx context.Context, id int64) (*User, error) {
	user, err := db.Queries.User(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{User: user, q: db.Queries}, nil
}

func (user User) UpdateState(ctx context.Context, state string, stateDump []byte) error {
	return user.q.UpdateUserState(ctx, sqlc.UpdateUserStateParams{
		ID:        user.ID,
		State:     state,
		StateDump: stateDump,
	})
}

func (user User) UpdateProfile(ctx context.Context, profile any) error {
	return user.q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:      user.ID,
		Profile: jsondump.DumpBytes(profile),
	})
}

func (user User) UpdateResume(ctx context.Context, resume any, file []byte) error {
	return user.q.UpdateUserResume(ctx, sqlc.UpdateUserResumeParams{
		ID:         user.ID,
		Resume:     jsondump.DumpBytes(resume),
		ResumeFile: file,
	})
}
