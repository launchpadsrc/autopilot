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

	"launchpad.icu/autopilot/bot"
	"launchpad.icu/autopilot/database"
)

type Config struct {
	Bot *bot.Bot
	DB  *database.DB
	AI  *openai.Client
}

type Background struct {
	b      *bot.Bot
	db     *database.DB
	ai     *openai.Client
	logger *slog.Logger
}

func New(c Config) Background {
	return Background{
		b:      c.Bot,
		db:     c.DB,
		ai:     c.AI,
		logger: slog.Default().With("go", "background"),
	}
}

type Task struct {
	logger *slog.Logger
}

func (bg Background) Start() {
	go bg.start(bg.Targeting, time.Minute)

	bg.b.Start()
}

func (bg Background) start(f func(Task) func(context.Context) error, d time.Duration) {
	t := Task{
		logger: bg.logger.With("task", funcName(f)),
	}

	measure := func(f func()) time.Duration {
		start := time.Now()
		f()
		return time.Since(start)
	}

	for {
		t.logger.Info("triggered")

		elapsed := measure(func() {
			// TODO: ctx
			if err := f(t)(context.Background()); err != nil {
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
