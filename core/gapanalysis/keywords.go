package gapanalysis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/prompts"
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
func (e KeywordsExtractor) Extract(k int, jds []string) ([]Keyword, error) {
	const (
		key   = "gap_analysis.keywords_extractor"
		model = "gpt-4o-mini"
	)

	req := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: 0,         // deterministic output
		MaxTokens:   k*12 + 20, // ~12 tokens/entry + overhead
	}

	system, user, err := prompts.Get(key, prompts.Map{"K": k, "JobAds": jds})
	if err != nil {
		return nil, err
	}

	req.Messages = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: user,
		},
	}

	res, err := e.ai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if len(res.Choices) == 0 {
		return nil, errors.New("openai: empty response")
	}

	var kws []Keyword
	data := []byte(res.Choices[0].Message.Content)
	return kws, json.Unmarshal(data, &kws)
}
