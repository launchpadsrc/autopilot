package autopilot

import (
	"context"
	"log/slog"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type (
	bgTask struct{ logger *slog.Logger }
	bgFunc = func(context.Context) error
)

func (ap *Autopilot) startBackground(taskFunc func(bgTask) bgFunc, d time.Duration) {
	t := bgTask{
		logger: slog.With("go", "autopilot/background", "task", funcName(taskFunc)),
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
	n = n[strings.LastIndex(n, ".")+1:]
	return strings.Title(strings.TrimSuffix(n, "Task-fm"))
}

func measure(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}
