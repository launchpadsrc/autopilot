package launchpad

import "github.com/sashabaranov/go-openai"

type SmartSteps struct {
	ai *openai.Client
}

func NewSmartSteps(ai *openai.Client) SmartSteps {
	return SmartSteps{ai: ai}
}
