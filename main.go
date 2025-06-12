package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/database"
	"launchpad.icu/autopilot/pkg/aifactory"
	"launchpad.icu/autopilot/pkg/openaix"
)

const (
	migrations = "database/migrations"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	db, err := database.Open(context.Background(), os.Getenv("DB_URI"))
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}
	if err := db.Migrate(migrations); err != nil {
		log.Fatalln("failed to migrate database:", err)
	}

	ai := aifactory.Client()
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatalln("failed to connect to openai client:", err)
	}

	if err := openaix.Read("ai.yml"); err != nil {
		log.Fatalln("failed to initialize openaix:", err)
	}

	b, err := bot.New(bot.Config{DB: db, AI: ai})
	if err != nil {
		log.Fatalln("failed to initialize bot:", err)
	}

	b.Start()
}
