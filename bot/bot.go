package bot

import (
	"fmt"
	"log/slog"
	"text/template"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"
	"gopkg.in/telebot.v4/middleware"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/bot/cache"
	"launchpad.icu/autopilot/parsers"
	"launchpad.icu/autopilot/pkg/htmlstrip"
)

type Bot struct {
	*layout.Layout
	*tele.Bot

	cache   *cache.Cache
	parsers map[string]parsers.Parser

	ai *openai.Client
}

func New(ai *openai.Client) (*Bot, error) {
	lt, err := layout.New("bot.yml", templateFuncs)
	if err != nil {
		return nil, err
	}

	pref := lt.Settings()
	pref.OnError = func(err error, c tele.Context) {
		slog.Error("global handler", "error", err)
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}
	if err := b.SetCommands(lt.Commands()); err != nil {
		return nil, err
	}

	cache, err := cache.New("cache.db")
	if err != nil {
		return nil, err
	}

	parsers := map[string]parsers.Parser{
		"djinni.co":   parsers.NewDjinni(),
		"jobs.dou.ua": parsers.NewDou(),
	}

	return &Bot{
		Layout:  lt,
		Bot:     b,
		cache:   cache,
		parsers: parsers,
		ai:      ai,
	}, nil
}

func (b Bot) Start() {
	slog.Info("starting", "go", "bot")

	b.Use(middleware.Recover())
	b.Use(b.Layout.Middleware("ua"))

	b.Handle("/start", b.onStart)
	b.Handle(tele.OnText, b.onChat)
	b.Handle("/keywords", b.onKeywords)
	b.Handle("/resume", b.onResume)
	b.Handle(tele.OnDocument, b.onResume)

	b.goFeeder()
	b.Bot.Start()
}

func (b Bot) sendHint(c tele.Context, hint string, v ...any) error {
	text := "💡 " + hint
	if len(v) > 0 {
		text += " " + fmt.Sprintln(v...)
	}
	return c.Send(text)
}

var templateFuncs = template.FuncMap{
	"htmlstrip": htmlstrip.Strip,
}
