package launchpad

import (
	"encoding/json"

	"github.com/sashabaranov/go-openai"
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

// Dumpable requires a step to be able to return its state as JSON.
type Dumpable interface {
	// Dump returns a JSON representation of the step's state.
	Dump() (json.RawMessage, error)
	// Load loads the step's state from JSON.
	Load(json.RawMessage) error
}

// Result represents the result of a step execution.
type Result struct {
	// Value is a generic wrapper for the step's output.
	Value any `json:"value"`
	// Response is the message to be sent back to the user.
	Response string `json:"assistant_response,omitempty"`
}

// ResultOf is a generic wrapper for the Result type.
type ResultOf[T any] struct {
	*Result
	Value T `json:"value"`
}

// NewResultOf creates a new ResultOf[T] with unwrapped value.
func NewResultOf[T any](r *Result) *ResultOf[T] {
	v, _ := r.Value.(T)
	return &ResultOf[T]{Result: r, Value: v}
}

func (r *ResultOf[T]) OfAny() *Result {
	return &Result{
		Value:    r.Value,
		Response: r.Response,
	}
}

// State represents a finite state machine for managing the launchpad steps of the user.
type State struct {
	ai    *openai.Client
	FSM   *FSM // FIXME: merge with State
	steps map[string]Step
}

// NewState initializes a new State.
func NewState(ai *openai.Client) *State {
	s := &State{
		ai:    ai,
		FSM:   NewFSM(),
		steps: make(map[string]Step),
	}
	s.Clear()
	return s
}

func (s *State) Clear() {
	s.steps = make(map[string]Step)
	for name, newFunc := range stepsNew {
		s.steps[name] = newFunc(s)
	}
}

// Execute runs the current step with the provided input.
func (s *State) Execute(input string) (*Result, error) {
	_, step := s.Current()
	return step.Execute(input)
}

func (s *State) Transition() error {
	return s.FSM.Transition()
}

// Current returns the current state name and the corresponding step.
func (s *State) Current() (string, Step) {
	state := s.FSM.Current()
	return state, s.steps[state] // FIXME: nil
}

// Dump dumps the state into JSON.
func (s *State) Dump() (json.RawMessage, error) {
	dump := make(map[string]json.RawMessage)
	for name, step := range s.steps {
		dumpable, ok := step.(Dumpable)
		if !ok {
			continue
		}
		data, err := dumpable.Dump()
		if err != nil {
			return nil, err
		}
		dump[name] = data
	}
	return json.Marshal(dump)
}

// LoadState loads the state from a JSON dump and sets the current state.
func LoadState(ai *openai.Client, current string, dump json.RawMessage) (*State, error) {
	var stepsDump map[string]json.RawMessage
	if len(dump) != 0 {
		if err := json.Unmarshal(dump, &stepsDump); err != nil {
			return nil, err
		}
	}

	state := NewState(ai)
	state.FSM.SetState(current)

	for name, step := range stepsDump {
		dumpable, ok := state.steps[name].(Dumpable)
		if !ok {
			continue
		}
		if err := dumpable.Load(step); err != nil {
			return nil, err
		}
	}

	return state, nil
}
