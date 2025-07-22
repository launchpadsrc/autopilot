package background

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/jobanalysis"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/internal/jsondump"
	"launchpad.icu/autopilot/parsers"
)

func (bg Background) Feeder(t Task) Func {
	if os.Getenv("FEEDER_OFF") == "true" {
		return nil
	}

	return func(ctx context.Context) error {
		// A proxy can be modified in runtime.
		proxy := os.Getenv("FEEDER_PROXY")
		if proxy != "" {
			t.logger.Debug("proxy set", "proxy", proxy)
		}

		var errs errgroup.Group
		for source, parser := range bg.parsers {
			errs.Go((feeder{
				Background: bg,
				source:     source,
				parser:     parsers.WithProxy(parser, proxy),
				logger:     t.logger.With("parser", source),
			}).execute)
		}
		return errs.Wait()
	}
}

type feeder struct {
	Background
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
		if err := f.sendToChannel(entry, overview); err != nil {
			// TODO: Implement queue instead?
			logger.Error("sending job entry", "error", err)
			time.Sleep(time.Second) // in case a flood wait occurred
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

func (f feeder) sendToChannel(entry parsers.FeedEntry, overview jobanalysis.Overview) error {
	text := f.b.TextLocale("ua", "feeder.job", struct {
		parsers.FeedEntry
		Overview jobanalysis.Overview
		Source   string
	}{
		FeedEntry: entry,
		Overview:  overview,
		Source:    f.source,
	})

	_, err := f.b.Send(f.b.ChatID("jobs_channel"), text, tele.NoPreview)
	return err
}
