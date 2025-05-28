package main

import (
	"context"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/bot"
)

func main() {
	ai := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	if _, err := ai.ListModels(context.Background()); err != nil { // ping
		log.Fatal("failed to connect to openai:", err)
	}

	b, err := bot.New(ai)
	if err != nil {
		log.Fatal("failed to create telegram bot", err)
	}

	b.Start()
}
