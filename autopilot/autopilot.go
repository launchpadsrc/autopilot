package autopilot

import (
	"maps"

	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/core/targeting"
	"launchpad.icu/autopilot/internal/database"
	"launchpad.icu/autopilot/parsers"
)

// Config has required dependencies for Autopilot.
type Config struct {
	DB *database.DB
	AI *openai.Client
}

// Autopilot provides automation layer for Launchpad.
type Autopilot struct {
	db        *database.DB
	ai        *openai.Client
	parsers   map[string]parsers.Parser
	callbacks Callbacks
}

// Callbacks defines Autopilot callback functions usually used by background tasks.
type Callbacks struct {
	OnFeederJob    func(FeederJob) error
	OnTargetingJob func(*User, targeting.Job) error
}

// New returns Autopilot.
func New(c Config) *Autopilot {
	return &Autopilot{
		db:      c.DB,
		ai:      c.AI,
		parsers: maps.Clone(Parsers),
	}
}

// Parsers is a map of based parsers used for job feeds.
var Parsers = map[string]parsers.Parser{
	"djinni.co":   parsers.NewDjinni(),
	"jobs.dou.ua": parsers.NewDou(),
}

// On registers Autopilot callbacks.
func (ap *Autopilot) On(c Callbacks) {
	ap.callbacks = c
}
