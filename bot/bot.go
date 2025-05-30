package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/parsers"
)

type Bot struct {
	*tele.Bot

	ai *openai.Client

	parsers map[string]parsers.Parser
}

type Parsers = map[string]parsers.Parser

func New(ai *openai.Client, parsers Parsers) (*Bot, error) {
	lt, err := layout.New("bot.yml")
	if err != nil {
		return nil, err
	}

	b, err := tele.NewBot(lt.Settings())
	if err != nil {
		return nil, err
	}
	if err := b.SetCommands(lt.Commands()); err != nil {
		return nil, err
	}

	return &Bot{
		Bot:     b,
		ai:      ai,
		parsers: parsers,
	}, nil
}

func (b Bot) Start() {
	b.Handle("/keywords", b.onKeywords)
	b.Handle("/resume", b.onResume)
	b.Handle(tele.OnDocument, b.onResume)

	b.Bot.Start()
}

func (b Bot) sendHint(c tele.Context, hint string, v ...any) error {
	text := "💡 " + hint
	if len(v) > 0 {
		text += fmt.Sprintln(v...)
	}
	return c.Send(text)
}
