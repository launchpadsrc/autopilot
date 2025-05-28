package cvschema_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"launchpad.icu/autopilot/core/cvschema"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestParser(t *testing.T) {
	pdf, err := os.ReadFile("testdata/resume.pdf")
	if err != nil {
		t.Fatalf("failed to read test PDF: %v", err)
	}

	parser := cvschema.NewParser(openai.NewClient(os.Getenv("OPENAI_API_KEY")))
	resume, err := parser.Parse(pdf)

	data, _ := json.MarshalIndent(resume, "", "  ")
	t.Logf("Parsed resume: %+v", string(data))

	require.NoError(t, err)
	assert.Equal(t, "Charles McTurland", resume.Basics.Name)
	assert.Equal(t, "Software Engineer", resume.Basics.Label)
	assert.Equal(t, "cmcturland@email.com", resume.Basics.Email)
	assert.Equal(t, "(123) 456-7890", resume.Basics.Phone)
	assert.Equal(t, "New York", resume.Basics.Location.City)
	assert.Equal(t, "NY", resume.Basics.Location.Region)

	assert.Len(t, resume.Education, 1)
	assert.Equal(t, "University of Pittsburgh", resume.Education[0].Institution)
	assert.Equal(t, "Computer Science", resume.Education[0].Area)
	assert.Contains(t, resume.Education[0].StartDate, "2008-09")
	assert.Equal(t, resume.Education[0].EndDate, "2012-04")

	assert.Len(t, resume.Work, 3)     // TODO: check specifics
	assert.Len(t, resume.Skills, 2)   // TODO: check specifics
	assert.Len(t, resume.Projects, 1) // TODO: check specifics
}
