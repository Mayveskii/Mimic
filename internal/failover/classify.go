package failover

import (
	"strings"
)

// Action is the recommended response to an error.
type Action int

const (
	Retry Action = iota
	Fallback
	Abort
)

func (a Action) String() string {
	switch a {
	case Retry:
		return "retry"
	case Fallback:
		return "fallback"
	case Abort:
		return "abort"
	default:
		return "unknown"
	}
}

// Classifier determines the appropriate action for an error.
// Behavior source: hermes-agent error classifier.
type Classifier struct {
	RetryCount int
	MaxRetries int
	BackoffMs  int
}

// NewClassifier creates a classifier with default limits.
func NewClassifier() *Classifier {
	return &Classifier{
		MaxRetries: 3,
		BackoffMs:  1000,
	}
}

// Classify analyzes an error message and returns the recommended action.
func (c *Classifier) Classify(err error) Action {
	if err == nil {
		return Retry // no error → continue
	}

	msg := strings.ToLower(err.Error())

	// Abort: unrecoverable errors
	if containsAny(msg, []string{
		"permission denied", "unauthorized", "auth failed",
		"invalid api key", "budget exhausted", "max turns reached",
	}) {
		return Abort
	}

	// Retry: transient errors
	if containsAny(msg, []string{
		"rate limit", "429", "too many requests", "timeout",
		"connection refused", "temporary failure", "eof",
		"gateway timeout", "503", "502",
	}) {
		if c.RetryCount < c.MaxRetries {
			c.RetryCount++
			return Retry
		}
		return Fallback
	}

	// Fallback: hallucination / bad input
	if containsAny(msg, []string{
		"invalid arguments", "unknown tool", "hallucination",
		"tool not found", "bad request", "invalid opcode",
	}) {
		return Fallback
	}

	// Default: retry once, then fallback
	if c.RetryCount < c.MaxRetries {
		c.RetryCount++
		return Retry
	}
	return Fallback
}

// Reset clears the retry counter.
func (c *Classifier) Reset() {
	c.RetryCount = 0
}

// RetryDelay returns the current backoff duration.
func (c *Classifier) RetryDelay() int {
	return c.BackoffMs * c.RetryCount
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
