package launchpad

import (
	"encoding/json"
	"strings"

	"dario.cat/mergo"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/openaix"
)

// KickoffStep is the first step in the Launchpad roadmap.
// It collects user profile information to tailor the experience.
// Conversation history and context are maintained.
type KickoffStep struct {
	state   *State
	chat    *openaix.ChatContext[ResultOf[UserProfile]]
	profile *UserProfile
}

func NewKickoffStep(state *State) Step {
	return &KickoffStep{
		state: state,
		chat:  openaix.Chat[ResultOf[UserProfile]](state.ai, "launchpad.01_kickoff"),
	}
}

type (
	UserProfile struct {
		Roles        []string             `json:"roles"`
		Stack        []UserProfileStack   `json:"stack"`
		Motivation   string               `json:"motivation"`
		English      string               `json:"english" jsonschema:"enum=A1,enum=A2,enum=B1,enum=B2,enum=C1,enum=C2"`
		WeeklyHours  int                  `json:"weekly_hours" jsonschema:"minimum=1"`
		Salary       UserProfileSalary    `json:"salary"`
		Assets       UserProfileAssets    `json:"assets"`
		Problems     []UserProfileProblem `json:"problems"`
		Observations []string             `json:"observations"`
	}

	UserProfileStack struct {
		Tech  string `json:"tech"`
		Level int    `json:"level" jsonschema:"minimum=0,maximum=5"`
	}

	UserProfileSalary struct {
		Range    string `json:"range"`
		Currency string `json:"currency" jsonschema:"enum=USD,enum=EUR,enum=UAH"`
	}

	UserProfileAssets struct {
		Github   string   `json:"github"`
		Linkedin string   `json:"linkedin"`
		Projects []string `json:"projects"`
	}

	UserProfileProblem struct {
		Problem string `json:"problem" jsonschema:"enum=MISSING_FIELD,enum=BAD_FORMAT,enum=AMBIGUOUS_VALUE"`
		Reason  string `json:"reason"`
	}
)

func (p UserProfile) StackTags() []string {
	return lo.Map(p.Stack, func(s UserProfileStack, _ int) string {
		return strings.ToLower(s.Tech)
	})
}

func (p UserProfile) RolePatterns() []string {
	return lo.Map(p.Roles, func(r string, _ int) string {
		return "%" + r + "%"
	})
}

func (s *KickoffStep) Execute(input string) (*Result, error) {
	result, err := s.chat.Completion(input)
	if err != nil {
		return nil, err
	}

	profile := &result.Value
	if s.profile == nil {
		s.profile = profile
	} else {
		// Merge profiles in case LLM fails to consider chat history.
		if err := mergo.Merge(s.profile, profile); err != nil {
			return nil, err
		}
	}

	return result.OfAny(), nil
}

type dumpedKickoffStep struct {
	Profile     *UserProfile                   `json:"profile,omitempty"`
	ChatHistory []openai.ChatCompletionMessage `json:"chat_history,omitempty"`
}

func (s *KickoffStep) Dump() (json.RawMessage, error) {
	return json.Marshal(dumpedKickoffStep{
		Profile:     s.profile,
		ChatHistory: s.chat.History,
	})
}

func (s *KickoffStep) Load(data json.RawMessage) error {
	var dumped dumpedKickoffStep
	if err := json.Unmarshal(data, &dumped); err != nil {
		return err
	}

	s.profile = dumped.Profile
	s.chat.History = dumped.ChatHistory
	return nil
}
