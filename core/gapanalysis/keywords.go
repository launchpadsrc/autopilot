package gapanalysis

import (
	"math"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/prompts"
	"launchpad.icu/autopilot/pkg/simpleopenai"
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
	const (
		key   = "gap_analysis.keywords_extractor"
		model = openai.GPT4oMini
	)
	prompt := simpleopenai.CompletionRequestPrompt{
		System: prompts.System(key),
		User:   prompts.User(key, prompts.Map{"K": k, "JobAds": jds}),
	}
	return simpleopenai.Completion[[]Keyword](ke.ai, simpleopenai.CompletionRequest{
		Model:       model,
		Prompt:      prompt,
		Temperature: math.SmallestNonzeroFloat32,
		MaxTokens:   k*12 + 20, // ~12 tokens/entry + overhead
	})
}
