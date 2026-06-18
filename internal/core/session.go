package core

import (
	"time"

	"github.com/Mayveskii/Mimic/internal/config"
)

// SessionContext holds all state for a single model session.
// It lives in a shared package to avoid import cycles between session and orchestrator.
type SessionContext struct {
	ID           string
	ModelID      string
	RepoPath     string
	WorktreePath string
	BaseSHA      string
	Budget       Budget
	Config       *config.RepoConfig
	StartTime    time.Time
}

// Budget is the minimal budget interface needed by the orchestrator.
type Budget interface {
	Consume(tokens int, seconds int) bool
	Exhausted() bool
	String() string
	TokenBudget() int
	TimeBudget() int
}
