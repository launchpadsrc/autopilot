package bot

import (
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v4"

	"launchpad.icu/autopilot/core/jobanalysis"
	"launchpad.icu/autopilot/parsers"
)

func (b Bot) goFeeder() {
	logger := slog.With("go", "feeder")
	logger.Info("starting")

	tick := func() error {
		for key, parser := range b.parsers {
			logger := logger.With("parser", key)

			entries, err := parser.ParseFeed()
			if err != nil {
				if err.Error() != "not implemented" {
					return err
				}
				continue
			}

			stored, err := b.cache.StoreFeed(key, entries)
			if err != nil {
				return err
			}

			if len(stored) > 0 {
				logger.Info("added job entries", "stored", len(stored))
			}

			for _, fe := range stored {
				overview, err := jobanalysis.NewOverviewer(b.ai).Overview(fe.Title, fe.Description)
				if err != nil {
					return err
				}
				if len(overview.Hashtags) == 0 {
					continue
				}

				entry := struct {
					parsers.FeedEntry
					Parser   string
					Overview jobanalysis.Overview
				}{
					FeedEntry: fe,
					Parser:    key,
					Overview:  overview,
				}

				_, err = b.Send(
					b.ChatID("jobs_channel"),
					b.TextLocale("ua", "feeder.job", entry),
					tele.NoPreview, tele.ModeHTML,
				)
				if err != nil {
					// TODO: put job entry into queue if error happened to try sending it later
					logger.Error("sending job entry", "error", err, "entry", fe)
					time.Sleep(time.Second) // in case a flood wait occurred
				}
			}
		}

		return nil
	}

	go func() {
		for range time.Tick(time.Minute) {
			start := time.Now()
			if err := tick(); err != nil {
				logger.Error("executing tick", "error", err)
			}
			logger.Debug("tick finished", "duration", time.Since(start))
		}
	}()
}
