package prompts

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Get retrieves and interpolates a prompt by its path with the given variables.
func Get(path string, vars ...any) (string, string, error) {
	if global == nil {
		return "", "", fmt.Errorf("prompts: global instance is not initialized")
	}
	if len(vars) == 0 {
		vars = []any{nil}
	}
	return global.Get(path, vars)
}

// System retrieves the system prompt only, ignoring the errors.
func System(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	system, _, _ := global.Get(path, vars)
	return system
}

// User retrieves the user prompt only, ignoring the errors.
func User(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	_, user, _ := global.Get(path, vars)
	return user
}

var global *Prompts

func init() {
	const promptsFile = "prompts.yml"
	if _, err := os.Stat(promptsFile); err != nil {
		return
	}

	p, err := New(promptsFile)
	if err != nil {
		panic(err)
	}

	global = p
}

type Map = map[string]any

type Prompts struct {
	v map[string]any
}

func New(filePath string) (*Prompts, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("prompts: failed to read file: %w", err)
	}

	var prompts map[string]any
	if err := yaml.Unmarshal(data, &prompts); err != nil {
		return nil, fmt.Errorf("prompts: failed to parse YAML: %w", err)
	}

	return &Prompts{v: prompts}, nil
}

func (p *Prompts) Get(path string, vars any) (string, string, error) {
	final := any(p.v)
	for _, part := range strings.Split(path, ".") {
		if m, ok := final.(map[string]any); ok {
			final = m[part]
		} else {
			return "", "", fmt.Errorf("prompts: invalid path: %s", path)
		}
	}

	prompts, ok := final.(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("prompts: invalid prompt structure at path: %s", path)
	}

	system, err := interpolateTemplate(prompts["system"].(string), vars)
	if err != nil {
		return "", "", err
	}

	user, err := interpolateTemplate(prompts["user"].(string), vars)
	if err != nil {
		return "", "", err
	}

	return system, user, nil
}

func interpolateTemplate(tmpl string, vars any) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("prompts: failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("prompts: failed to execute template: %w", err)
	}

	return buf.String(), nil
}
