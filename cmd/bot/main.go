package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/parsers"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	ai := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatal("failed to connect to openai:", err)
	}

	parsers := bot.Parsers{
		"djinni.co":   parsers.NewDjinni(),
		"jobs.dou.ua": parsers.NewDou(),
	}

	b, err := bot.New(ai, parsers)
	if err != nil {
		log.Fatal("failed to create telegram bot", err)
	}

	b.Start()
}
