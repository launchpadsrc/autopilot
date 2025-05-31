package prompts

import (
	"os"
	"testing"
)

func TestNewParser(t *testing.T) {
	// Create a temporary YAML file for testing
	yamlContent := `
job_analysis:
  keywords_extractor:
    model: gpt-4o
    system: "System prompt with {{.K}} keywords."
    user: "User prompt for {{.JobAds}}."
`
	tmpFile, err := os.CreateTemp("", "prompts_test_*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test New
	parser, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if parser == nil {
		t.Fatalf("Prompts is nil")
	}
}

func TestGetPrompt(t *testing.T) {
	// Create a temporary YAML file for testing
	yamlContent := `
job_analysis:
  keywords_extractor:
    model: gpt-4o
    system: "System prompt with {{.K}} keywords."
    user: "User prompt for {{.JobAds}}."
`
	tmpFile, err := os.CreateTemp("", "prompts_test_*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Initialize parser
	parser, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Test Get
	vars := map[string]any{
		"K":      5,
		"JobAds": []string{"Job ad 1", "Job ad 2"},
	}

	model, systemPrompt, userPrompt, err := parser.Get("job_analysis.keywords_extractor", vars)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	expectedModel := "gpt-4o"
	expectedSystem := "System prompt with 5 keywords."
	expectedUser := "User prompt for [Job ad 1 Job ad 2]."

	if model != expectedModel {
		t.Errorf("Expected model: %s, got: %s", expectedModel, model)
	}

	if systemPrompt != expectedSystem {
		t.Errorf("Expected system prompt: %s, got: %s", expectedSystem, systemPrompt)
	}

	if userPrompt != expectedUser {
		t.Errorf("Expected user prompt: %s, got: %s", expectedUser, userPrompt)
	}
}

func TestGetPromptInvalidPath(t *testing.T) {
	// Create a temporary YAML file for testing
	yamlContent := `
job_analysis:
  keywords_extractor:
    model: gpt-4o
    system: "System prompt with {{.K}} keywords."
    user: "User prompt for {{.JobAds}}."
`
	tmpFile, err := os.CreateTemp("", "prompts_test_*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Initialize parser
	parser, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Test Get with an invalid path
	_, _, _, err = parser.Get("invalid.path", nil)
	if err == nil {
		t.Fatalf("Expected error for invalid path, got nil")
	}
}

func TestInterpolateTemplate(t *testing.T) {
	tmpl := "Hello, {{.Name}}!"
	vars := map[string]any{
		"Name": "World",
	}

	result, err := interpolateTemplate(tmpl, vars)
	if err != nil {
		t.Fatalf("InterpolateTemplate failed: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}
