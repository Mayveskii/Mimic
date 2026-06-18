package failover

import (
	"errors"
	"testing"
)

func TestClassifier_Classify(t *testing.T) {
	c := NewClassifier()

	tests := []struct {
		name  string
		err   string
		want  Action
		reset bool
	}{
		{"nil error", "", Retry, false},
		{"rate limit", "429 rate limit exceeded", Retry, false},
		{"timeout", "connection timeout", Retry, false},
		{"invalid args", "invalid arguments from model", Fallback, false},
		{"unknown tool", "unknown tool name", Fallback, false},
		{"permission denied", "permission denied", Abort, false},
		{"budget exhausted", "budget exhausted: tokens=100/100", Abort, false},
		{"hallucination", "hallucination detected", Fallback, false},
		{"generic", "something went wrong", Retry, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reset {
				c.Reset()
			}
			var err error
			if tt.err != "" {
				err = errors.New(tt.err)
			}
			got := c.Classify(err)
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestClassifier_MaxRetries(t *testing.T) {
	c := NewClassifier()
	c.MaxRetries = 2

	// First two rate limits → Retry
	if c.Classify(errors.New("rate limit")) != Retry {
		t.Fatal("expected retry on 1st")
	}
	if c.Classify(errors.New("rate limit")) != Retry {
		t.Fatal("expected retry on 2nd")
	}
	// Third rate limit → Fallback (max retries exceeded)
	if c.Classify(errors.New("rate limit")) != Fallback {
		t.Fatal("expected fallback on 3rd")
	}

	// Reset and retry again
	c.Reset()
	if c.Classify(errors.New("rate limit")) != Retry {
		t.Fatal("expected retry after reset")
	}
}

func TestClassifier_RetryDelay(t *testing.T) {
	c := NewClassifier()
	c.BackoffMs = 100

	if c.RetryDelay() != 0 {
		t.Fatalf("expected delay=0 before retries, got %d", c.RetryDelay())
	}

	c.Classify(errors.New("timeout"))
	if c.RetryDelay() != 100 {
		t.Fatalf("expected delay=100 after 1st retry, got %d", c.RetryDelay())
	}

	c.Classify(errors.New("timeout"))
	if c.RetryDelay() != 200 {
		t.Fatalf("expected delay=200 after 2nd retry, got %d", c.RetryDelay())
	}
}
