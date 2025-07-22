package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"go.etcd.io/bbolt"

	"launchpad.icu/autopilot/core/jobanalysis"
	"launchpad.icu/autopilot/internal/aifactory"
	"launchpad.icu/autopilot/internal/bboltx"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/internal/openaix"
)

func main() {
	ai := aifactory.Client()
	if err := openaix.Read("ai.yml"); err != nil {
		log.Fatal(err)
	}

	cache, err := bbolt.Open("cache.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(context.Background(), os.Getenv("DB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	for _, source := range []string{"djinni.co", "jobs.dou.ua"} {
		bucket := bboltx.NewBucket[FeedEntry](cache, "jobs").Bucket(source)

		slog.Info("migrating", "source", source, "count", bucket.Count())

		err := bucket.Walk(func(entry FeedEntry) error {
			overview, err := jobanalysis.NewOverviewer(ai).Overview(entry.Title, entry.Description)
			if err != nil {
				return err
			}

			published, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700 ", entry.Published)
			if err != nil {
				return err
			}

			return db.InsertJob(context.Background(), database.InsertJobParams{
				ID:          entry.ID,
				Source:      source,
				PublishedAt: &published,
				Link:        entry.Link,
				Title:       entry.Title,
				Description: entry.Description,
				CompanyAI:   overview.Company,
				RoleAI:      overview.RoleName,
				SeniorityAI: overview.Seniority,
				OverviewAI:  overview.Description,
				HashtagsAI:  overview.Hashtags,
			})
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

type FeedEntry struct {
	ID          string
	Link        string
	Title       string
	Description string
	Published   string
}

func (fe FeedEntry) BoltID() string {
	return fe.ID
}
