package openaix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sashabaranov/go-openai"
)

var reflector = &jsonschema.Reflector{
	DoNotReference: true,
	ExpandedStruct: true,
	Anonymous:      true,
}

type ChatContext[T any] struct {
	ai      *openai.Client
	key     string
	History []openai.ChatCompletionMessage
}

type (
	CompletionRequest struct {
		Model       string                   `mapstructure:"model"`
		Temperature float32                  `mapstructure:"temperature"`
		MaxTokens   int                      `mapstructure:"max_tokens"`
		Prompts     CompletionRequestPrompts `mapstructure:"prompts"`
		JSON        *CompletionRequestJSON   `mapstructure:"json"`
	}

	CompletionRequestPrompts struct {
		System string `mapstructure:"system"`
		User   string `mapstructure:"user"`
	}

	CompletionRequestJSON struct {
		Name        string     `mapstructure:"name"`
		Description string     `mapstructure:"description"`
		Strict      bool       `mapstructure:"strict"`
		Reflect     bool       `mapstructure:"reflect"`
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

// Completion is deprecated.
func Completion[T any](ai *openai.Client, key string, vars ...any) (T, error) {
	return Chat[T](ai, key).Completion(vars...)
}

// Chat wraps CreateChatCompletion for a more convenient interface.
// If the type T is a string, the completion requests will return the string content as it is.
// Otherwise, it expects the content to be a valid JSON and unmarshals it into T.
// Preserves the chat history.
//
// TODO: Should we use the new Responses API instead? See: https://platform.openai.com/docs/api-reference/responses.
// TODO: Implement logging of requests.
func Chat[T any](ai *openai.Client, key string) *ChatContext[T] {
	return &ChatContext[T]{ai: ai, key: key}
}

func (cc *ChatContext[T]) Completion(vars ...any) (T, error) {
	historyBefore := slices.Clone(cc.History)
	v, err := cc.completion(vars...)
	if err != nil {
		cc.History = historyBefore
	}
	return v, err
}

func (cc *ChatContext[T]) completion(vars ...any) (v T, _ error) {
	r, err := configUnmarshalKey[CompletionRequest](cc.key)
	if err != nil {
		return v, err
	}
	if err := r.Prompts.Execute(vars...); err != nil {
		return v, err
	}
	if r.Prompts.User == "" {
		return v, errors.New("openaix: user prompt is empty")
	}

	req := openai.ChatCompletionRequest{
		Model:       r.Model,
		Temperature: r.Temperature,
		MaxTokens:   r.MaxTokens,
	}

	if r.JSON != nil {
		var schema json.Marshaler
		if r.JSON.Reflect {
			schema = reflector.Reflect(&v)
		} else {
			schema = r.JSON.Schema
		}

		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        r.JSON.Name,
				Description: r.JSON.Description,
				Strict:      r.JSON.Strict,
				Schema:      schema,
			},
		}
	}

	if r.Prompts.System != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: r.Prompts.System,
		})
	}
	if len(cc.History) > 0 {
		req.Messages = append(req.Messages, cc.History...)
	}

	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: r.Prompts.User,
	}

	// Add user message to the history.
	cc.History = append(cc.History, message)
	req.Messages = append(req.Messages, message)

	res, err := cc.ai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return v, err
	}
	if len(res.Choices) == 0 {
		return v, errors.New("openaix: empty choices in response")
	}

	message = res.Choices[0].Message
	content := message.Content

	// Add assistant response to the history.
	cc.History = append(cc.History, message)

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

	logger := logger.With("id", res.ID, "model", res.Model)
	logger.Debug("completion compact response", "content", compact.String())

	if err := json.NewDecoder(compact).Decode(&v); err != nil {
		return v, fmt.Errorf("openaix: %w", err)
	}

	return v, nil
}
