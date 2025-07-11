package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"launchpad.icu/autopilot/api"
	"launchpad.icu/autopilot/background"
	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/database"
	"launchpad.icu/autopilot/parsers"
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

	parsers := map[string]parsers.Parser{
		"djinni.co":   parsers.NewDjinni(),
		"jobs.dou.ua": parsers.NewDou(),
	}

	b, err := bot.New(bot.Config{
		DB:      db,
		AI:      ai,
		Parsers: parsers,
	})
	if err != nil {
		log.Fatalln("failed to initialize bot:", err)
	}

	bg := background.New(background.Config{
		Bot:     b,
		DB:      db,
		AI:      ai,
		Parsers: parsers,
	})
	go bg.Start()

	s := api.New(api.Config{
		Addr: os.Getenv("SERVER_ADDR"),
		DB:   db,
	})
	if err := s.Start(); err != nil {
		log.Fatalln("failed to start server:", err)
	}
}
