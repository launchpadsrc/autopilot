package main

import (
	"context"
	"log"
	"log/slog"

	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/pkg/aifactory"
	"launchpad.icu/autopilot/pkg/openaix"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	ai := aifactory.Client()
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatalln("failed to connect to openai client:", err)
	}

	if err := openaix.Read("ai.yml"); err != nil {
		log.Fatalln("failed to initialize openaix:", err)
	}

	b, err := bot.New(ai)
	if err != nil {
		log.Fatalln("failed to initialize bot:", err)
	}

	b.Start()
}
