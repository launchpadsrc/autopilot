package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/bot"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
}

func main() {
	ai := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatalln("failed to connect to openai:", err)
	}

	b, err := bot.New(ai)
	if err != nil {
		log.Fatalln("failed to create bot:", err)
	}

	b.Start()
}
