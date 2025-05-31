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
func Get(path string, vars ...any) (model, system, user string, err error) {
	if global == nil {
		err = fmt.Errorf("prompts: global instance is not initialized")
		return
	}
	return global.Get(path, vars)
}

// Model retrieves only the model of the prompts, ignoring the errors.
func Model(path string) string {
	if global == nil {
		return ""
	}
	model, _, _, _ := global.Get(path)
	return model
}

// System retrieves the system prompt, ignoring the errors.
func System(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	_, system, _, _ := global.Get(path, vars...)
	return system
}

// User retrieves the user prompt, ignoring the errors.
func User(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	_, _, user, _ := global.Get(path, vars...)
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
	v map[string]any               // raw
	c map[string]map[string]string // cache
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

	return &Prompts{
		v: prompts,
		c: make(map[string]map[string]string),
	}, nil
}

func (p *Prompts) Get(path string, vars ...any) (model, system, user string, err error) {
	if len(vars) == 0 {
		vars = []any{nil}
	}

	cached, ok := p.c[path]
	if !ok {
		cached, err = p.updateCache(path)
		if err != nil {
			return
		}
	}

	system, err = interpolateTemplate(cached["system"], vars[0])
	if err != nil {
		return
	}

	user, err = interpolateTemplate(cached["user"], vars[0])
	if err != nil {
		return
	}

	model = cached["model"]
	return
}

func (p *Prompts) updateCache(path string) (map[string]string, error) {
	final := any(p.v)
	for _, part := range strings.Split(path, ".") {
		if m, ok := final.(map[string]any); ok {
			final = m[part]
		} else {
			return nil, fmt.Errorf("prompts: invalid path: %s", path)
		}
	}

	prompts, ok := final.(map[string]any)
	if !ok || len(prompts) == 0 {
		return nil, fmt.Errorf("prompts: invalid prompt structure at path: %s", path)
	}

	cached := make(map[string]string, len(prompts))
	for k, v := range prompts {
		cached[k], ok = v.(string)
		if !ok {
			return nil, fmt.Errorf("prompts: invalid prompt value type at path %s", path)
		}
	}

	return cached, nil
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
