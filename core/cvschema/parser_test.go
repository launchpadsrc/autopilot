package cvschema_test

import (
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

	require.NoError(t, err)
	assert.Equal(t, "Charles McTurland", resume.Basics.Name)
	assert.Equal(t, "Software Engineer", resume.Basics.Label)
	assert.Equal(t, "cmcturland@email.com", resume.Basics.Email)
	assert.Equal(t, "New York", resume.Basics.Location.City)
	assert.Equal(t, "NY", resume.Basics.Location.Region)
	assert.NotEmpty(t, resume.Skills, 2)

	assert.Len(t, resume.Education, 1)
	assert.Equal(t, "University of Pittsburgh", resume.Education[0].Institution)
	assert.Equal(t, "Computer Science", resume.Education[0].Area)
	assert.Contains(t, resume.Education[0].StartDate, "2008-09")
	assert.Equal(t, resume.Education[0].EndDate, "2012-04")

	assert.Len(t, resume.Work, 3)     // TODO: check specifics
	assert.Len(t, resume.Projects, 1) // TODO: check specifics
}
