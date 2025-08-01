package autopilot

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"launchpad.icu/autopilot/core/jobanalysis"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/internal/jsondump"
	"launchpad.icu/autopilot/parsers"
)

// FeederJob represents a job entry processed by the feeder.
type FeederJob struct {
	parsers.FeedEntry
	ParserName string
	Overview   jobanalysis.Overview
}

// StartFeeder starts the background Feeder task that periodically fetches job entries.
func (ap *Autopilot) StartFeeder(d time.Duration) {
	if off, _ := strconv.ParseBool(os.Getenv("FEEDER_OFF")); off {
		return
	}
	go ap.startBackground(ap.feederTask, cmp.Or(d, time.Minute))
}

func (ap *Autopilot) feederTask(t bgTask) bgFunc {
	return func(ctx context.Context) error {
		// A proxy can be modified in runtime.
		proxy := os.Getenv("FEEDER_PROXY")
		if proxy != "" {
			t.logger.Debug("proxy set", "proxy", proxy)
		}

		var errs errgroup.Group
		for source, parser := range ap.parsers {
			errs.Go((feeder{
				Autopilot: ap,
				source:    source,
				parser:    parsers.WithProxy(parser, proxy),
				logger:    t.logger.With("parser", source),
			}).execute)
		}
		return errs.Wait()
	}
}

type feeder struct {
	*Autopilot
	source string
	parser parsers.Parser
	logger *slog.Logger
}

func (f feeder) execute() error {
	entries, err := f.uniqueEntries()
	if err != nil {
		// Skip parsers that do not implement ParseFeed.
		if err.Error() == "not implemented" {
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	f.logger.Info("overviewing job entries", "count", len(entries))

	overviewer := jobanalysis.NewOverviewer(f.ai)
	for _, entry := range entries {
		logger := f.logger.With("entry", jsondump.Dump(entry))

		overview, err := overviewer.Overview(entry.Title, entry.Description)
		if err != nil {
			logger.Error("getting job overview", "error", err)
			return err
		}
		if len(overview.Hashtags) == 0 {
			logger.Error("no hashtags for job entry, skipping")
			continue
		}

		if err := f.insertJob(entry, overview); err != nil {
			return err
		}

		if f.callbacks.OnFeederJob != nil {
			if err := f.callbacks.OnFeederJob(FeederJob{
				ParserName: f.source,
				FeedEntry:  entry,
				Overview:   overview,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f feeder) uniqueEntries() ([]parsers.FeedEntry, error) {
	entries, err := f.parser.ParseFeed()
	if err != nil {
		return nil, err
	}

	ids := lo.Map(entries, func(e parsers.FeedEntry, _ int) string {
		return e.ID
	})

	// Returns IDs that already exist in the DB.
	ids, err = f.db.JobsExist(context.Background(), ids)
	if err != nil {
		return nil, err
	}

	// Filter entries that already exist in the DB.
	return lo.Filter(entries, func(e parsers.FeedEntry, _ int) bool {
		return !slices.Contains(ids, e.ID)
	}), nil
}

func (f feeder) insertJob(entry parsers.FeedEntry, overview jobanalysis.Overview) error {
	return f.db.InsertJob(context.Background(), database.InsertJobParams{
		Source:      f.source,
		ID:          entry.ID,
		PublishedAt: entry.Published,
		Link:        entry.Link,
		Title:       entry.Title,
		Description: entry.Description,
		CompanyAI:   overview.Company,
		RoleAI:      overview.RoleName,
		SeniorityAI: overview.Seniority,
		OverviewAI:  overview.Description,
		HashtagsAI:  overview.Hashtags,
	})
}
