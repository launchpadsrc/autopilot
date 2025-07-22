package autopilot

import (
	"github.com/sashabaranov/go-openai"

	"launchpad.icu/autopilot/internal/database"
)

// Config has required dependencies for Autopilot.
type Config struct {
	DB *database.DB
	AI *openai.Client
}

// Autopilot provides automation layer for Launchpad.
type Autopilot struct {
	db *database.DB
	ai *openai.Client
}

// New returns Autopilot.
func New(c Config) *Autopilot {
	return &Autopilot{
		db: c.DB,
		ai: c.AI,
	}
}
