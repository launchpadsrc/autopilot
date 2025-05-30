package parsers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	parsers := map[string]struct {
		Parser
		target string
	}{
		"djinni": {NewDjinni(), "https://djinni.co/jobs/736521-c-net-engineer/"},
		"dou":    {NewDou(), "https://jobs.dou.ua/companies/skelar/vacancies/309770/"},
	}

	for name, p := range parsers {
		t.Run(name, func(t *testing.T) {
			job, err := p.ParseJob(p.target)
			if err != nil {
				t.Fatalf("failed to parse job: %v", err)
			}

			data, _ := json.MarshalIndent(job, "", "  ")
			t.Logf("%s job parsed successfully: %s", name, string(data))

			require.Equal(t, p.target, job.URL)
			require.NotEmpty(t, job.Title)
			require.NotEmpty(t, job.Description)
			require.NotEmpty(t, job.DatePosted)
			require.NotEmpty(t, job.LocationType)
		})
	}
}
