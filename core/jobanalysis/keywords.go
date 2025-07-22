package jobanalysis

import (
	"github.com/sashabaranov/go-openai"

	openaix2 "launchpad.icu/autopilot/internal/openaix"
)

type Keyword struct {
	Term  string  `json:"term"`
	Score float64 `json:"score"`
}

type KeywordsExtractor struct {
	ai *openai.Client
}

func NewKeywordsExtractor(ai *openai.Client) KeywordsExtractor {
	return KeywordsExtractor{ai: ai}
}

// Extract extracts K keywords from a list of job descriptions.
func (ke KeywordsExtractor) Extract(k int, jds []string) ([]Keyword, error) {
	return openaix2.Completion[[]Keyword](ke.ai, "job_analysis.keywords_extractor", openaix2.Map{
		"K":      k,
		"JobAds": jds,
	})
}
