package session

import "fmt"

// Budget tracks token and time consumption per session.
// Behavior source: hermes-agent iteration budget.
type Budget struct {
	MaxTokens      int
	MaxTimeSeconds int
	UsedTokens     int
	UsedTimeSeconds int
}

// NewBudget creates a budget from config values.
func NewBudget(maxTokens, maxTimeSeconds int) *Budget {
	return &Budget{
		MaxTokens:       maxTokens,
		MaxTimeSeconds:  maxTimeSeconds,
		UsedTokens:      0,
		UsedTimeSeconds: 0,
	}
}

// Consume checks if the requested tokens and time fit within remaining budget.
// Returns true if consumption is allowed and budget is updated.
func (b *Budget) Consume(tokens, timeSeconds int) bool {
	if b.UsedTokens+tokens > b.MaxTokens {
		return false
	}
	if b.UsedTimeSeconds+timeSeconds > b.MaxTimeSeconds {
		return false
	}
	b.UsedTokens += tokens
	b.UsedTimeSeconds += timeSeconds
	return true
}

// Remaining returns unused tokens and time.
func (b *Budget) Remaining() (tokens, timeSeconds int) {
	return b.MaxTokens - b.UsedTokens, b.MaxTimeSeconds - b.UsedTimeSeconds
}

// Exhausted returns true if no budget remains.
func (b *Budget) Exhausted() bool {
	return b.UsedTokens >= b.MaxTokens || b.UsedTimeSeconds >= b.MaxTimeSeconds
}

// String returns a human-readable budget summary.
func (b *Budget) String() string {
	return fmt.Sprintf("tokens=%d/%d time=%ds/%ds", b.UsedTokens, b.MaxTokens, b.UsedTimeSeconds, b.MaxTimeSeconds)
}
