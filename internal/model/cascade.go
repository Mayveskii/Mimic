package model

import (
	"context"
	"fmt"

	"github.com/Mayveskii/Mimic/internal/config"
)

// Cascade implements the local -> medium -> top model escalation strategy.
type Cascade struct {
	provider  Provider
	profiles  config.ModelProfiles
	threshold float64
	maxSteps  int
}

// NewCascade creates a model cascade using the given provider and config.
func NewCascade(provider Provider, profiles config.ModelProfiles) *Cascade {
	threshold := profiles.Cascade.ConfidenceThreshold
	if threshold == 0 {
		threshold = 0.85
	}
	maxSteps := profiles.Cascade.MaxEscalations
	if maxSteps == 0 {
		maxSteps = 2
	}
	return &Cascade{
		provider:  provider,
		profiles:  profiles,
		threshold: threshold,
		maxSteps:  maxSteps,
	}
}

// CascadeResult is the outcome of a cascade run.
type CascadeResult struct {
	ModelID   string
	Call      *CallResult
	Escalated bool
	Attempts  int
}

// Run executes the intent, escalating from local to medium to top until a confident result is obtained.
func (c *Cascade) Run(ctx context.Context, messages []Message, tools []ToolSchema) (*CascadeResult, error) {
	order := []struct {
		name  string
		model string
	}{
		{"local", c.profiles.Local},
		{"medium", c.profiles.Medium},
		{"top", c.profiles.Top},
	}

	var lastErr error
	for i, tier := range order {
		if i > c.maxSteps {
			break
		}

		req := ChatRequest{
			Model:       tier.model,
			Messages:    messages,
			Tools:       tools,
			ToolChoice:  "auto",
			Temperature: 0.1,
		}

		res, err := c.provider.Call(ctx, req)
		if err != nil {
			lastErr = err
			continue
		}

		confident := c.isConfident(res)
		if confident || i == len(order)-1 {
			return &CascadeResult{
				ModelID:   tier.model,
				Call:      res,
				Escalated: i > 0,
				Attempts:  i + 1,
			}, nil
		}

		// Not confident and not last tier: escalate.
		lastErr = fmt.Errorf("%s tier not confident (finish=%s)", tier.name, finishReason(res))
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("model cascade exhausted all tiers")
}

func (c *Cascade) isConfident(res *CallResult) bool {
	if len(res.Response.Choices) == 0 {
		return false
	}
	choice := res.Response.Choices[0]
	// Tool calls are treated as confident execution intent.
	if len(choice.Message.ToolCalls) > 0 {
		return true
	}
	// Direct answer with content is confident.
	if choice.Message.Content != "" && choice.FinishReason != "length" {
		return true
	}
	return false
}

func finishReason(res *CallResult) string {
	if len(res.Response.Choices) == 0 {
		return "no_choices"
	}
	return res.Response.Choices[0].FinishReason
}

// Call implements model.Caller by running the cascade and returning the final result.
func (c *Cascade) Call(ctx context.Context, req ChatRequest) (*CallResult, error) {
	res, err := c.Run(ctx, req.Messages, req.Tools)
	if err != nil {
		return nil, err
	}
	return res.Call, nil
}
