package simpleopenai_test

import (
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"

	"launchpad.icu/autopilot/pkg/simpleopenai"
)

var ai = openai.NewClient(os.Getenv("OPENAI_API_KEY"))

func TestCompletion_ReturnsString_WhenValidResponse(t *testing.T) {
	request := simpleopenai.CompletionRequest{
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		Prompt: simpleopenai.CompletionRequestPrompt{
			System: "You are a helpful assistant.",
			User:   "What is the capital of France?",
		},
	}

	response, err := simpleopenai.Completion[string](ai, request)

	require.NoError(t, err)
	require.NotEmpty(t, response)
	require.Contains(t, response, "Paris")
}

func TestCompletion_UnmarshalsJSONResponse_WhenValidJSON(t *testing.T) {
	request := simpleopenai.CompletionRequest{
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		Prompt: simpleopenai.CompletionRequestPrompt{
			System: "You are a helpful assistant.",
			User:   "Provide a JSON object with a key 'example' and value 'test'.",
		},
	}

	response, err := simpleopenai.Completion[map[string]string](ai, request)

	require.NoError(t, err)
	require.Equal(t, "test", response["example"])
}
