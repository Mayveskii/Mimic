package model

import (
	"context"
	"errors"
	"testing"

	"github.com/Mayveskii/Mimic/internal/config"
)

type cascadeMockProvider struct {
	responses map[string]*CallResult
	errors    map[string]error
	callCount int
}

func (m *cascadeMockProvider) Name() string { return "mock" }

func (m *cascadeMockProvider) Call(ctx context.Context, req ChatRequest) (*CallResult, error) {
	m.callCount++
	if err, ok := m.errors[req.Model]; ok {
		return nil, err
	}
	if res, ok := m.responses[req.Model]; ok {
		return res, nil
	}
	return nil, errors.New("unexpected model")
}

func choiceWithContent(content string, finish string) ChatResponse {
	return ChatResponse{
		Choices: []struct {
			Index        int     `json:"index"`
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		}{
			{Message: Message{Content: content}, FinishReason: finish},
		},
	}
}

func TestCascade_LocalConfident(t *testing.T) {
	provider := &cascadeMockProvider{
		responses: map[string]*CallResult{
			"local": {Response: choiceWithContent("direct answer", "stop")},
		},
	}
	cascade := NewCascade(provider, config.ModelProfiles{Local: "local", Medium: "medium", Top: "top"})

	res, err := cascade.Run(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ModelID != "local" {
		t.Fatalf("expected local, got %s", res.ModelID)
	}
	if res.Escalated {
		t.Fatal("expected no escalation")
	}
	if provider.callCount != 1 {
		t.Fatalf("expected 1 call, got %d", provider.callCount)
	}
}

func TestCascade_EscalatesOnLowConfidence(t *testing.T) {
	provider := &cascadeMockProvider{
		responses: map[string]*CallResult{
			"local":  {Response: choiceWithContent("", "length")},
			"medium": {Response: choiceWithContent("better answer", "stop")},
		},
	}
	cascade := NewCascade(provider, config.ModelProfiles{Local: "local", Medium: "medium", Top: "top"})

	res, err := cascade.Run(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ModelID != "medium" {
		t.Fatalf("expected medium, got %s", res.ModelID)
	}
	if !res.Escalated {
		t.Fatal("expected escalation")
	}
	if provider.callCount != 2 {
		t.Fatalf("expected 2 calls, got %d", provider.callCount)
	}
}

func TestCascade_EscalatesToTop(t *testing.T) {
	provider := &cascadeMockProvider{
		responses: map[string]*CallResult{
			"local":  {Response: choiceWithContent("", "length")},
			"medium": {Response: choiceWithContent("", "length")},
			"top":    {Response: choiceWithContent("final answer", "stop")},
		},
	}
	cascade := NewCascade(provider, config.ModelProfiles{Local: "local", Medium: "medium", Top: "top"})

	res, err := cascade.Run(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ModelID != "top" {
		t.Fatalf("expected top, got %s", res.ModelID)
	}
	if provider.callCount != 3 {
		t.Fatalf("expected 3 calls, got %d", provider.callCount)
	}
}

func TestCascade_ToolCallsAreConfident(t *testing.T) {
	provider := &cascadeMockProvider{
		responses: map[string]*CallResult{
			"local": {
				Response: ChatResponse{
					Choices: []struct {
						Index        int     `json:"index"`
						Message      Message `json:"message"`
						FinishReason string  `json:"finish_reason"`
					}{
						{
							Message: Message{
								ToolCalls: []ToolCall{
									{ID: "1", Type: "function", Function: struct {
										Name      string `json:"name"`
										Arguments string `json:"arguments"`
									}{Name: "fix", Arguments: "{}"}},
								},
							},
							FinishReason: "tool_calls",
						},
					},
				},
			},
		},
	}
	cascade := NewCascade(provider, config.ModelProfiles{Local: "local", Medium: "medium", Top: "top"})

	res, err := cascade.Run(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ModelID != "local" {
		t.Fatalf("expected local, got %s", res.ModelID)
	}
	if len(res.Call.Response.Choices[0].Message.ToolCalls) != 1 {
		t.Fatal("expected tool call to be preserved")
	}
}

func TestCascade_AllTiersFail(t *testing.T) {
	provider := &cascadeMockProvider{
		errors: map[string]error{
			"local":  errors.New("local down"),
			"medium": errors.New("medium down"),
			"top":    errors.New("top down"),
		},
	}
	cascade := NewCascade(provider, config.ModelProfiles{Local: "local", Medium: "medium", Top: "top"})

	_, err := cascade.Run(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if provider.callCount != 3 {
		t.Fatalf("expected 3 calls, got %d", provider.callCount)
	}
}
