package jobanalysis

import (
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/openaix"
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

func (jo Overviewer) Overview(title, description string) (Overview, error) {
	return openaix.Completion[Overview](jo.ai, "job_analysis.overview", openaix.Map{
		"Title":       title,
		"Description": description,
	})
}
