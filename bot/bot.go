package bot

import (
	"fmt"
	"log/slog"
	"text/template"

	"github.com/sashabaranov/go-openai"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"
	"gopkg.in/telebot.v4/middleware"

	"launchpad.icu/autopilot/database"
	"launchpad.icu/autopilot/parsers"
	"launchpad.icu/autopilot/pkg/htmlstrip"
)

type Config struct {
	DB      *database.DB
	AI      *openai.Client
	Parsers map[string]parsers.Parser
}

type Bot struct {
	*layout.Layout
	*tele.Bot

	db      *database.DB
	ai      *openai.Client
	parsers map[string]parsers.Parser
}

func New(c Config) (*Bot, error) {
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

	return &Bot{
		Layout:  lt,
		Bot:     b,
		db:      c.DB,
		ai:      c.AI,
		parsers: c.Parsers,
	}, nil
}

func (b Bot) Start() {
	slog.Info("starting", "go", "bot")

	b.Use(b.withRecover)
	b.Use(b.withError)
	b.Use(b.Layout.Middleware("ua"))
	b.Use(b.withUserState)

	b.Handle("/start", b.onStart)
	b.Handle("/reset", b.onReset)
	b.Handle(tele.OnText, b.onChat)
	b.Handle(tele.OnDocument, b.onChat)

	debug := b.Group()
	debug.Use(middleware.Whitelist(b.Int64("admin_id")))
	debug.Handle("/_setproxy", b.onDebugSetProxy)

	b.Bot.Start()
}

func (b Bot) sendHint(c tele.Context, v ...any) error {
	return c.Send("💡 " + fmt.Sprintln(v...))
}

func (b Bot) sendDebug(c tele.Context, v any) {
	b.Send(
		b.ChatID("admin_id"),
		fmt.Sprintf("```debug-%d\n%s```", c.Sender().ID, v),
		tele.ModeMarkdownV2,
	)
}

var templateFuncs = template.FuncMap{
	"htmlstrip": htmlstrip.Strip,
}

func (b Bot) withRecover(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic", "error", r)
				b.sendDebug(c, r)
			}
		}()
		return next(c)
	}
}

func (b Bot) withError(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		err := next(c)
		if err != nil {
			b.sendDebug(c, err)
		}
		return err
	}
}
