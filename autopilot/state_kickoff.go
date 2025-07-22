package autopilot

import (
	"context"
	"errors"
	"reflect"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
)

type StateKickoffResult = launchpad.UserProfile

type StateKickoff struct {
	*launchpad.ResultOf[StateKickoffResult]
	sm *StateMachine
}

func (sm *StateMachine) NewStateKickoff(r *launchpad.ResultOf[StateKickoffResult]) *StateKickoff {
	return &StateKickoff{ResultOf: r, sm: sm}
}

// HasResume returns true if the resume is defined.
func (s *StateKickoff) HasResume() bool {
	// The resume has zero value if it was not added yet.
	return !reflect.ValueOf(s.sm.user.Resume).IsZero()
}

// AddResume processes the CV data and updates the resume.
// Required to be able to complete the state.
func (s *StateKickoff) AddResume(ctx context.Context, data []byte) error {
	resume, err := cvschema.NewParser(s.sm.ap.ai).Parse(data)
	if err != nil {
		return err
	}
	return s.sm.user.updateResume(ctx, *resume, data)
}

// Complete should only be called if resume was added.
func (s *StateKickoff) Complete(ctx context.Context) error {
	if !s.HasResume() {
		return errors.New("autopilot/state_kickoff: no resume added")
	}
	return s.sm.complete(func() error {
		return s.sm.user.updateProfile(ctx, s.Value)
	})
}
