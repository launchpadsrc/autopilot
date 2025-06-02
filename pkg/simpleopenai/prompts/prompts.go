package prompts

// TODO: merge with autopilot/pkg/simpleopenai

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"

	"launchpad.icu/autopilot/pkg/simpleopenai"
)

// Get retrieves and interpolates a prompt by its path with the given variables.
func Get(path string, vars ...any) (prompt Prompt, err error) {
	if global == nil {
		err = fmt.Errorf("prompts: global instance is not initialized")
		return
	}
	return global.Get(path, vars...)
}

// Model retrieves only the model of the prompts, ignoring the errors.
func Model(path string) string {
	if global == nil {
		return ""
	}
	prompt, _ := global.Get(path)
	return prompt.Model
}

// System retrieves the system prompt, ignoring the errors.
func System(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	prompt, _ := global.Get(path, vars...)
	return prompt.System
}

// User retrieves the user prompt, ignoring the errors.
func User(path string, vars ...any) string {
	if global == nil {
		return ""
	}
	prompt, _ := global.Get(path, vars...)
	return prompt.User
}

// JSON retrieves the JSON schema if such specified.
func JSON(path string) *simpleopenai.CompletionResponseSchema {
	if global == nil {
		return nil
	}
	prompt, _ := global.Get(path)
	return prompt.JSON
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

type (
	Map = map[string]any

	Prompt struct {
		Model  string `mapstructure:"model"`
		System string `mapstructure:"system"`
		User   string `mapstructure:"user"`

		JSON *simpleopenai.CompletionResponseSchema `mapstructure:"json,omitempty"`
	}
)

type Prompts struct {
	v *viper.Viper
	// v map[string]any               // raw
	// c map[string]map[string]string // cache
}

func New(filePath string) (*Prompts, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Prompts{
		v: v,
		// v: prompts,
		// c: make(map[string]map[string]string),
	}, nil
}

func (p *Prompts) Get(path string, vars ...any) (prompt Prompt, err error) {
	if len(vars) == 0 {
		vars = []any{nil}
	}

	if err = p.v.UnmarshalKey(path, &prompt); err != nil {
		return
	}

	prompt.System, err = interpolateTemplate(prompt.System, vars[0])
	if err != nil {
		return
	}

	prompt.User, err = interpolateTemplate(prompt.User, vars[0])
	if err != nil {
		return
	}

	return
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
