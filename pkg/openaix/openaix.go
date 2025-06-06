package openaix

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/spf13/viper"
)

type Map = map[string]any

var logger = slog.With("package", "openaix")

var config *viper.Viper

func Read(path string) error {
	config = viper.New()
	config.SetConfigFile(path)
	return config.ReadInConfig()
}

func configUnmarshalKey[T any](key string) (v T, _ error) {
	if config == nil {
		return v, errors.New("openaix: config is not read")
	}
	if err := config.UnmarshalKey(key, &v); err != nil {
		return v, fmt.Errorf("openaix: %w", err)
	}
	return v, nil
}

func interpolateTemplate(s string, v any) (string, error) {
	t, err := template.New("openaix").Parse(s)
	if err != nil {
		return "", fmt.Errorf("openaix: parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, v); err != nil {
		return "", fmt.Errorf("openaix: execute template: %w", err)
	}
	return buf.String(), nil
}
