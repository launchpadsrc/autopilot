package autopilot

import (
	"context"
	"errors"

	"github.com/samber/lo"

	"launchpad.icu/autopilot/core/launchpad"
)

type (
	// StateMachine handles the state execution and transition.
	StateMachine struct {
		ap      *Autopilot
		user    *User
		actions StateActions
	}

	// StateActions holds all the actions that can be performed in the state machine.
	StateActions struct {
		Kickoff StateAction[*StateKickoff]
	}

	// StateHandler requires a Complete method that finalizes the state interaction.
	StateHandler interface {
		Complete(context.Context) error
	}

	// StateAction is a wrapper for state handler function.
	StateAction[T StateHandler] = func(s T) error
)

// StateMachine returns a new StateMachine instance for the given user.
// Always register all the state actions.
func (ap *Autopilot) StateMachine(user *User, actions StateActions) (*StateMachine, error) {
	if !actions.Valid() {
		return nil, errors.New("autopilot: invalid state actions")
	}
	return &StateMachine{ap: ap, user: user, actions: actions}, nil
}

// Valid returns true if all actions are defined.
func (sa StateActions) Valid() bool {
	return !lo.Contains([]any{sa.Kickoff}, nil)
}

// State returns the current state of the user.
func (sm *StateMachine) State() *launchpad.State {
	return sm.user.State
}

// StateName returns the name of the current state.
func (sm *StateMachine) StateName() string {
	return sm.State().CurrentName()
}

// Entrypoint is the main entry point for the state machine.
// It executes the current state and calls the state handler.
// If the state handler signalizes a completion, it will transition to the next state.
func (sm *StateMachine) Entrypoint(input string) error {
	result, err := sm.user.State.Execute(input)
	if err != nil {
		return err
	}

	switch sm.StateName() {
	case launchpad.StateKickoff:
		result := launchpad.NewResultOf[StateKickoffResult](result)
		return sm.actions.Kickoff(sm.NewStateKickoff(result))
	}

	return nil
}

func (sm *StateMachine) complete(f func() error) error {
	if err := f(); err != nil {
		return err
	}
	return sm.user.State.Transition()
}
