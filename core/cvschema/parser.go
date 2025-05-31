package cvschema

import (
	"bytes"
	"io"
	"math"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/prompts"
	"launchpad.icu/autopilot/pkg/simpleopenai"
)

type Parser struct {
	ai *openai.Client
}

func NewParser(ai *openai.Client) Parser {
	return Parser{ai: ai}
}

// Parse parses a PDF resume and returns a structured Resume object.
func (p Parser) Parse(pdf []byte) (*Resume, error) {
	content, err := p.readPDF(bytes.NewReader(pdf))
	if err != nil {
		return nil, err
	}

	const (
		key = "cv_schema"
	)
	prompt := simpleopenai.CompletionRequestPrompt{
		System: prompts.System(key),
		User:   prompts.User(key, content),
	}
	return simpleopenai.Completion[*Resume](p.ai, simpleopenai.CompletionRequest{
		Model:       prompts.Model(key),
		Prompt:      prompt,
		Temperature: math.SmallestNonzeroFloat32,
	})
}

func (p Parser) readPDF(raw *bytes.Reader) (string, error) {
	reader, err := pdf.NewReader(raw, raw.Size())
	if err != nil {
		return "", err
	}
	b, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	text, err := io.ReadAll(b)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(text)), nil
}
