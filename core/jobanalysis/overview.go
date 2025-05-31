package jobanalysis

import (
	"math"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/prompts"
	"launchpad.icu/autopilot/pkg/simpleopenai"
)

type Overview struct {
	Company     string   `json:"company"`
	RoleName    string   `json:"role_name"`
	Seniority   string   `json:"seniority"`
	Description string   `json:"overview"`
	Hashtags    []string `json:"hashtags"`
}

type Overviewer struct {
	ai *openai.Client
}

func NewOverviewer(ai *openai.Client) Overviewer {
	return Overviewer{ai: ai}
}

func (jo Overviewer) Overview(title, desc string) (Overview, error) {
	const (
		key = "job_analysis.overview"
	)
	prompt := simpleopenai.CompletionRequestPrompt{
		System: prompts.System(key),
		User:   prompts.User(key, prompts.Map{"Title": title, "Description": desc}),
	}
	return simpleopenai.Completion[Overview](jo.ai, simpleopenai.CompletionRequest{
		Model:       prompts.Model(key),
		Prompt:      prompt,
		Temperature: math.SmallestNonzeroFloat32,
	})
}
