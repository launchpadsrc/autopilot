package cvschema

import (
	"bytes"
	"log/slog"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/pkg/openaix"
	"launchpad.icu/autopilot/pkg/poppler"
)

type Parser struct {
	ai *openai.Client
}

func NewParser(ai *openai.Client) Parser {
	return Parser{ai: ai}
}

// Parse parses a PDF resume and returns a structured Resume object.
func (p Parser) Parse(pdf []byte) (*Resume, error) {
	content, err := poppler.ToHTML(bytes.NewReader(pdf))
	if err != nil {
		return nil, err
	}
	slog.Debug("converted pdf to html", "content", content)
	return openaix.Completion[*Resume](p.ai, "cv_schema", content)
}
