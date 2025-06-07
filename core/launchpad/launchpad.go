package launchpad

import (
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/openaix"
)

type SmartSteps struct {
	ai   *openai.Client
	chat *openaix.ChatContext[Profile]
}

func NewSmartSteps(ai *openai.Client) SmartSteps {
	return SmartSteps{ai: ai}
}
