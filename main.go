package main

import (
	"context"
	"log"
	"log/slog"

	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/pkg/aifactory"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	ai := aifactory.Client()
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatalln("failed to connect to openai:", err)
	}

	b, err := bot.New(ai)
	if err != nil {
		log.Fatalln("failed to create bot:", err)
	}

	b.Start()
}
