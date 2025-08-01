package bot

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"text/template"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/layout"
	"gopkg.in/telebot.v4/middleware"

	"launchpad.icu/autopilot/autopilot"
	"launchpad.icu/autopilot/internal/htmlstrip"
)

var logger = slog.With("go", "bot")

type Config struct {
	Autopilot *autopilot.Autopilot
}

type Bot struct {
	*layout.Layout
	*tele.Bot
	ap *autopilot.Autopilot
}

func New(c Config) (*Bot, error) {
	lt, err := layout.New("bot.yml", templateFuncs)
	if err != nil {
		return nil, err
	}

	pref := lt.Settings()
	pref.OnError = func(err error, c tele.Context) {
		logger.Error("global handler", "error", err)
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}
	if err := b.SetCommands(lt.Commands()); err != nil {
		return nil, err
	}

	return &Bot{
		Layout: lt,
		Bot:    b,
		ap:     c.Autopilot,
	}, nil
}

func (b Bot) Start() {
	logger.Info("starting")

	b.Use(b.withRecover)
	b.Use(b.withError)
	b.Use(b.withUser)

	b.Use(b.Layout.Middleware("ua"))

	b.Handle("/start", b.onStart)
	b.Handle(tele.OnText, b.onChat)
	b.Handle(tele.OnDocument, b.onChat)

	debug := b.Group()
	debug.Use(middleware.Whitelist(b.Int64("admin_id")))
	debug.Handle("/_setproxy", b.onDebugSetProxy)
	debug.Handle("/_targeting", b.onDebugTargeting)

	b.ap.On(autopilot.Callbacks{
		OnFeederJob:    b.onFeederUniqueJob,
		OnTargetingJob: b.onTargetingJob,
	})

	go b.ap.StartFeeder(time.Minute)
	go b.ap.StartTargeting(time.Minute)

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
	"title":     strings.Title,
}

func (b Bot) withRecover(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("recovered from panic", "error", r)
				b.sendDebug(c, fmt.Sprintf("%s\n\n%s", r, debug.Stack()))
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
