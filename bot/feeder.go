package bot

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/samber/lo"
	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/jobanalysis"
	"launchpad.icu/autopilot/database/sqlc"
	"launchpad.icu/autopilot/parsers"
	"launchpad.icu/autopilot/pkg/jsondump"
)

func (b Bot) goFeeder() {
	feeder := NewFeeder(b)
	logger := feeder.logger
	logger.Info("starting")

	go func() {
		for range time.Tick(10 * time.Second) {
			start := time.Now()
			if err := feeder.Execute(); err != nil {
				logger.Error("executing tick", "error", err)
			}
			logger.Debug("tick finished", "elapsed", time.Since(start))
		}
	}()
}

var feederParsers = map[string]parsers.Parser{
	"djinni.co":   parsers.NewDjinni(),
	"jobs.dou.ua": parsers.NewDou(),
}

// Feeder is a feed implementation that fetches job entries from various sources,
// analyzes them using AI, stores them in the DB, and sends them to a Telegram channel.
type Feeder struct {
	Bot
	logger *slog.Logger
	off    bool
}

func NewFeeder(b Bot) Feeder {
	return Feeder{
		Bot:    b,
		logger: slog.With("go", "feeder"),
		off:    os.Getenv("FEEDER_OFF") == "true",
	}
}

func (f Feeder) Execute() error {
	for source, parser := range feederParsers {
		logger := f.logger.With("parser", source)

		entries, err := f.uniqueEntries(parser)
		if err != nil {
			// Skip parsers that do not implement ParseFeed.
			if err.Error() == "not implemented" {
				continue
			}
			return err
		}

		if len(entries) > 0 {
			f.logger.Info("overviewing job entries", "count", len(entries))
		} else {
			continue
		}

		overviewer := jobanalysis.NewOverviewer(f.ai)
		for _, entry := range entries {
			logger := logger.With("entry", jsondump.Dump(entry))

			overview, err := overviewer.Overview(entry.Title, entry.Description)
			if err != nil {
				logger.Error("getting job overview", "error", err)
				return err
			}
			if len(overview.Hashtags) == 0 {
				logger.Error("no hashtags for job entry, skipping")
				continue
			}

			if err := f.insertJob(source, entry, overview); err != nil {
				return err
			}
			if err := f.sendToChannel(source, entry, overview); err != nil {
				// TODO: Implement queue instead?
				logger.Error("sending job entry", "error", err)
				time.Sleep(time.Second) // in case a flood wait occurred
			}
		}
	}

	return nil
}

func (f Feeder) uniqueEntries(parser parsers.Parser) ([]parsers.FeedEntry, error) {
	entries, err := parser.ParseFeed()
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

func (f Feeder) insertJob(source string, entry parsers.FeedEntry, overview jobanalysis.Overview) error {
	return f.db.InsertJob(context.Background(), sqlc.InsertJobParams{
		Source:      source,
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

func (f Feeder) sendToChannel(source string, entry parsers.FeedEntry, overview jobanalysis.Overview) error {
	text := f.TextLocale("ua", "feeder.job", struct {
		parsers.FeedEntry
		Parser   string
		Overview jobanalysis.Overview
	}{
		FeedEntry: entry,
		Parser:    source,
		Overview:  overview,
	})

	if f.off {
		f.logger.Warn("feeder is off, not sending message", "text", text)
		return nil
	}

	_, err := f.Send(f.ChatID("jobs_channel"), text, tele.NoPreview)
	return err
}
