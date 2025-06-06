package openaix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type (
	CompletionRequest struct {
		Model       string                   `mapstructure:"model"`
		Temperature float32                  `mapstructure:"temperature"`
		MaxTokens   int                      `mapstructure:"max_tokens"`
		Prompts     CompletionRequestPrompts `mapstructure:"prompts"`
		JSON        *CompletionResponseJSON  `mapstructure:"json"`
	}

	CompletionRequestPrompts struct {
		System string `mapstructure:"system"`
		User   string `mapstructure:"user"`
	}

	CompletionResponseJSON struct {
		Name        string     `mapstructure:"name"`
		Description string     `mapstructure:"description"`
		Strict      bool       `mapstructure:"strict"`
		Schema      JSONSchema `mapstructure:"schema"`
	}
)

// Execute interpolates the system and user prompts with the provided variables.
func (crp *CompletionRequestPrompts) Execute(vars ...any) (err error) {
	if len(vars) == 0 {
		return nil
	}
	crp.System, err = interpolateTemplate(crp.System, vars[0])
	if err != nil {
		return err
	}
	crp.User, err = interpolateTemplate(crp.User, vars[0])
	if err != nil {
		return err
	}
	return nil
}

// Completion wraps CreateChatCompletion for a more convenient interface.
// If the type T is a string, it returns the content directly.
// Otherwise, it expects the content to be a valid JSON and unmarshals it into T.
//
// TODO: Should we use the new Responses API instead?
// TODO: See: https://platform.openai.com/docs/api-reference/responses.
// TODO: Implement logging of requests.
func Completion[T any](ai *openai.Client, key string, vars ...any) (v T, _ error) {
	r, err := configUnmarshalKey[CompletionRequest](key)
	if err != nil {
		return v, err
	}
	if err := r.Prompts.Execute(vars...); err != nil {
		return v, err
	}

	req := openai.ChatCompletionRequest{
		Model:       r.Model,
		Temperature: r.Temperature,
		MaxTokens:   r.MaxTokens,
	}

	if r.JSON != nil {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        r.JSON.Name,
				Description: r.JSON.Description,
				Schema:      r.JSON.Schema,
				Strict:      r.JSON.Strict,
			},
		}
	}

	if r.Prompts.System != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: r.Prompts.System,
		})
	}
	if r.Prompts.User != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: r.Prompts.User,
		})
	}

	res, err := ai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return v, err
	}
	if len(res.Choices) == 0 {
		return v, errors.New("openaix: empty choices in response")
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
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	normalized := []byte(content)
	if !json.Valid(normalized) {
		return v, errors.New("openaix: invalid json in completion response")
	}

	compact := new(bytes.Buffer)
	if err := json.Compact(compact, normalized); err != nil {
		return v, fmt.Errorf("openaix: %w", err)
	}

	logger.Debug("completion compact response", "content", compact.String())
	if err := json.NewDecoder(compact).Decode(&v); err != nil {
		return v, fmt.Errorf("openaix: %w", err)
	}

	return v, nil
}
