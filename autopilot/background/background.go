// Package background implements Launchpad background tasks.
package background

import (
	"context"
	"log/slog"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/client/bot"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/parsers"
)

type Config struct {
	Bot     *bot.Bot
	DB      *database.DB
	AI      *openai.Client
	Parsers map[string]parsers.Parser
}

type Background struct {
	b       *bot.Bot
	db      *database.DB
	ai      *openai.Client
	parsers map[string]parsers.Parser
	logger  *slog.Logger
}

func New(c Config) Background {
	return Background{
		b:       c.Bot,
		db:      c.DB,
		ai:      c.AI,
		parsers: c.Parsers,
		logger:  slog.Default().With("go", "background"),
	}
}

type (
	Task struct{ logger *slog.Logger }
	Func = func(context.Context) error
)

func (bg Background) Start() {
	go bg.start(bg.Feeder, time.Minute)
	go bg.start(bg.Targeting, time.Minute)

	bg.b.Start()
}

func (bg Background) start(taskFunc func(Task) Func, d time.Duration) {
	t := Task{
		logger: bg.logger.With("task", funcName(taskFunc)),
	}

	f := taskFunc(t)
	if f == nil {
		t.logger.Warn("task is disabled")
		return
	}

	for {
		t.logger.Info("triggered")

		elapsed := measure(func() {
			if err := f(context.Background()); err != nil { // TODO: ctx
				t.logger.Error("failed", "error", err)
			}
		})

		t.logger.Info("finished", "elapsed", elapsed.String())

		time.Sleep(d)
	}
}

func funcName(f any) string {
	n := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	return strings.TrimSuffix(n[strings.LastIndex(n, ".")+1:], "-fm")
}

func measure(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
