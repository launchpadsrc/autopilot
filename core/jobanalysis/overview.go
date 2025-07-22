package jobanalysis

import (
	"github.com/sashabaranov/go-openai"

	openaix2 "launchpad.icu/autopilot/internal/openaix"
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
	return openaix2.Completion[Overview](jo.ai, "job_analysis.overview", openaix2.Map{
		"Title":       title,
		"Description": description,
	})
}
