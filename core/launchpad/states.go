package launchpad

import "github.com/looplab/fsm"

// List of Launchpad states.
const (
	StateKickoff   = "kickoff"
	StateTargeting = "targeting"
)

var fsmEvents = fsm.Events{
	{Name: StateTargeting, Src: []string{StateKickoff}, Dst: ""},
}

// NewFSM initializes a finite state machine for the Launchpad roadmap.
func NewFSM() *fsm.FSM {
	return fsm.NewFSM(StateKickoff, fsmEvents, nil)
}
