package launchpad

import (
	"github.com/looplab/fsm"
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/wrap"
)

// stepNewFunc is a Step constructor.
type stepNewFunc = func(*State) Step

// stepsNew maps step names to their constructors.
var stepsNew = map[string]stepNewFunc{
	StateKickoff: NewKickoffStep,
}

// Step defines step actions.
type Step interface {
	Execute(string) (*Result, error)
}

// Result represents the result of a step execution.
type Result struct {
	// Wrapped is a generic wrapper for the step's output.
	// To access the actual data, use wrap.Unwrap[T](result.Wrapped).
	Wrapped any
	// Problems indicates if there were any issues during the step execution.
	// A presence of problems typically means the state should not be transitioned.
	Problems bool
	// Response is the message to be sent back to the user.
	Response string
}

// NewResult creates a new Result with the provided value wrapped.
func NewResult[T any](v T) *Result {
	return &Result{Wrapped: wrap.Wrap[T](v)}
}

// State represents a finite state machine for managing the launchpad steps of the user.
type State struct {
	ai    *openai.Client
	fsm   *fsm.FSM
	steps map[string]Step
}

// NewState initializes a new State.
func NewState(ai *openai.Client) *State {
	s := &State{
		ai:    ai,
		fsm:   NewFSM(),
		steps: make(map[string]Step),
	}
	for name, newFunc := range stepsNew {
		s.steps[name] = newFunc(s)
	}
	return s
}

// Execute runs the current step with the provided input.
func (s *State) Execute(input string) (*Result, error) {
	_, step := s.Current()
	return step.Execute(input)
}

// Current returns the current state name and the corresponding step.
func (s *State) Current() (string, Step) {
	state := s.fsm.Current()
	return state, s.steps[state]
}
