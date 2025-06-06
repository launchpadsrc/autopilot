package launchpad

import "launchpad.icu/autopilot/pkg/openaix"

type (
	Profile struct {
		Roles             []string    `json:"role"`
		Stack             []StackItem `json:"stack"`
		Motivation        string      `json:"motivation"`
		English           string      `json:"english"`
		WeeklyHours       int         `json:"weekly_hours"`
		Salary            Salary      `json:"salary"`
		Assets            Assets      `json:"assets"`
		Problems          []Problem   `json:"problems"`
		Observations      []string    `json:"observations"`       // free-text notes, one fact per item
		AssistantResponse string      `json:"assistant_response"` // response from the AI assistant
	}

	StackItem struct {
		Tech  string `json:"tech"`
		Level int    `json:"level"`
	}

	Salary struct {
		Range    string `json:"range"`
		Currency string `json:"currency"`
	}

	Assets struct {
		Github   *string  `json:"github"`
		Projects []string `json:"projects"`
		CvLink   *string  `json:"cv_link"`
	}

	Problem struct {
		Problem string `json:"problem"`
		Reason  string `json:"reason"`
	}
)

func (s *SmartSteps) Kickoff01UserProfile(answers string) (Profile, error) {
	if s.chat == nil {
		s.chat = openaix.Chat[Profile](s.ai, "launchpad_steps.01_kickoff")
	}
	return s.chat.Completion(answers)
}
