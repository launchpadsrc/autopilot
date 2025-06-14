package launchpad

import (
	"errors"
	"slices"
	"sync"
)

// List of Launchpad states.
const (
	StateKickoff   = "kickoff"
	StateTargeting = "targeting"
)

type FSM struct {
	mu      sync.Mutex
	current int
	states  []string
}

// NewFSM initializes a finite state machine for the Launchpad roadmap.
func NewFSM() *FSM {
	return &FSM{
		states: []string{
			StateKickoff,
			StateTargeting,
		},
	}
}

func (f *FSM) SetState(state string) {
	f.current = slices.Index(f.states, state)
}

func (f *FSM) Current() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.states[f.current]
}

func (f *FSM) Transition() error {
	if f.current+1 >= len(f.states) {
		return errors.New("no more transitions available")
	}

	f.mu.Lock()
	f.current++
	f.mu.Unlock()
	return nil
}
