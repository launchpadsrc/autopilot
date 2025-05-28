package bot

import (
	"github.com/sashabaranov/go-openai"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"
)

type Bot struct {
	*tele.Bot
	ai *openai.Client
}

func New(ai *openai.Client) (*Bot, error) {
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
		Bot: b,
		ai:  ai,
	}, nil
}

func (b Bot) Start() {
	b.Handle("/resume", b.onResume)
	b.Handle(tele.OnDocument, b.onResume)

	b.Bot.Start()
}
