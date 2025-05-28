package simpleopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var logger = slog.With("package", "simpleopenai")

type (
	CompletionRequest struct {
		Model       string
		Temperature float32
		MaxTokens   int
		Prompt      CompletionRequestPrompt
	}

	CompletionRequestPrompt struct {
		System string `json:"system"`
		User   string `json:"user"`
	}
)

// Completion provides a simpler API to OpenAI's chat completion endpoint.
func Completion[T any](ai *openai.Client, r CompletionRequest) (v T, _ error) {
	req := openai.ChatCompletionRequest{
		Model:       r.Model,
		Temperature: r.Temperature,
		MaxTokens:   r.MaxTokens,
	}

	if r.Prompt.System != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: r.Prompt.System,
		})
	}
	if r.Prompt.User != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: r.Prompt.User,
		})
	}

	res, err := ai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return v, err
	}
	if len(res.Choices) == 0 {
		return v, errors.New("simpleopenai: empty choices in response")
	}

	content := res.Choices[0].Message.Content

	logger := logger.With("id", res.ID, "model", res.Model)
	logger.Debug("completion raw response", "content", content)

	// If the type T is a string, return the content directly.
	// Otherwise, unmarshal the content into the provided type T.
	if _, ok := any(v).(string); ok {
		return any(content).(T), nil
	}

	// Normalize the content to remove any code block formatting.
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	normalized := []byte(content)
	if !json.Valid(normalized) {
		return v, errors.New("simpleopenai: invalid json in completion response")
	}

	compact := new(bytes.Buffer)
	if err := json.Compact(compact, normalized); err != nil {
		return v, fmt.Errorf("simpleopenai: %w", err)
	}

	logger.Debug("completion compact response", "content", compact.String())
	if err := json.NewDecoder(compact).Decode(&v); err != nil {
		return v, fmt.Errorf("simpleopenai: %w", err)
	}

	return v, nil
}
