package parsers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"launchpad.icu/autopilot/pkg/jsondump"
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

			t.Logf("%s job parsed successfully: %s", name, jsondump.Dump(job))

			require.Equal(t, p.target, job.URL)
			require.NotEmpty(t, job.Title)
			require.NotEmpty(t, job.Description)
			require.NotEmpty(t, job.DatePosted)
			require.NotEmpty(t, job.LocationType)
		})
	}
}
