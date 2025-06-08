package launchpad

import (
	"dario.cat/mergo"

	"launchpad.icu/autopilot/pkg/openaix"
)

// KickoffStep is the first step in the Launchpad roadmap.
// It collects user profile information to tailor the experience.
// Conversation history and context are maintained.
type KickoffStep struct {
	state   *State
	chat    *openaix.ChatContext[UserProfile]
	profile *UserProfile
}

func NewKickoffStep(state *State) Step {
	return &KickoffStep{state: state}
}

type (
	UserProfile struct {
		Roles             []string             `json:"roles"`
		Stack             []UserProfileStack   `json:"stack"`
		Motivation        string               `json:"motivation"`
		English           string               `json:"english"`
		WeeklyHours       int                  `json:"weekly_hours"`
		Salary            UserProfileSalary    `json:"salary"`
		Assets            UserProfileAssets    `json:"assets"`
		Problems          []UserProfileProblem `json:"problems"`
		Observations      []string             `json:"observations"`
		AssistantResponse string               `json:"assistant_response"`
	}

	UserProfileStack struct {
		Tech  string `json:"tech"`
		Level int    `json:"level"`
	}

	UserProfileSalary struct {
		Range    string `json:"range"`
		Currency string `json:"currency"`
	}

	UserProfileAssets struct {
		Github   *string  `json:"github"`
		Projects []string `json:"projects"`
		CvLink   *string  `json:"cv_link"`
	}

	UserProfileProblem struct {
		Problem string `json:"problem"`
		Reason  string `json:"reason"`
	}
)

func (s *KickoffStep) Execute(input string) (*Result, error) {
	profile, err := s.UserProfile(input)
	if err != nil {
		return nil, err
	}

	if s.profile == nil {
		s.profile = &profile
	} else {
		// Merge profiles in case LLM fails to consider chat history.
		if err := mergo.Merge(s.profile, profile); err != nil {
			return nil, err
		}
	}

	result := NewResult(profile)
	if len(profile.Problems) == 0 {
		return result, nil
	}

	// Indicate about problems needed to be solved.
	result.Problems = true
	result.Response = profile.AssistantResponse
	return result, nil
}

func (s *KickoffStep) UserProfile(answers string) (UserProfile, error) {
	if s.chat == nil {
		s.chat = openaix.Chat[UserProfile](s.state.ai, "launchpad.01_kickoff")
	}
	return s.chat.Completion(answers)
}
