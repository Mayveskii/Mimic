package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/config"
)

func TestGonkaGateProvider_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected auth header, got %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "qwen/qwen3-235b" {
			t.Errorf("expected model qwen/qwen3-235b, got %s", req.Model)
		}

		resp := ChatResponse{
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Message:      Message{Role: "assistant", Content: "pong"},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GONKAGATE_API_KEY", "test-key")
	p, err := NewGonkaGateProvider(server.URL, "GONKAGATE_API_KEY", 5000, 1)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	res, err := p.Call(context.Background(), ChatRequest{
		Model:    "qwen/qwen3-235b",
		Messages: []Message{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.Response.Choices[0].Message.Content != "pong" {
		t.Errorf("expected pong, got %s", res.Response.Choices[0].Message.Content)
	}
	if res.TokensIn != 10 || res.TokensOut != 5 {
		t.Errorf("unexpected token counts: in=%d out=%d", res.TokensIn, res.TokensOut)
	}
}

func TestGonkaGateProvider_Call_ToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Message: Message{
						Role: "assistant",
						ToolCalls: []ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: struct {
									Name      string `json:"name"`
									Arguments string `json:"arguments"`
								}{Name: "SYS_FILE_READ", Arguments: `{"path":"/tmp/test"}`},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GONKAGATE_API_KEY", "test-key")
	p, err := NewGonkaGateProvider(server.URL, "GONKAGATE_API_KEY", 5000, 1)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	res, err := p.Call(context.Background(), ChatRequest{
		Model:    "qwen/qwen3-235b",
		Messages: []Message{{Role: "user", Content: "read file"}},
		Tools: []ToolSchema{{
			Type: "function",
			Function: map[string]interface{}{
				"name":        "SYS_FILE_READ",
				"description": "read file",
				"parameters":  map[string]interface{}{"type": "object"},
			},
		}},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if len(res.Response.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(res.Response.Choices[0].Message.ToolCalls))
	}
	if res.Response.Choices[0].Message.ToolCalls[0].Function.Name != "SYS_FILE_READ" {
		t.Errorf("expected SYS_FILE_READ, got %s", res.Response.Choices[0].Message.ToolCalls[0].Function.Name)
	}
}

func TestGonkaGateProvider_MissingKey(t *testing.T) {
	os.Unsetenv("GONKAGATE_API_KEY")
	_, err := NewGonkaGateProvider("", "GONKAGATE_API_KEY", 1000, 1)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCascade_Run(t *testing.T) {
	provider := &mockProvider{responses: []*CallResult{
		{
			Response: ChatResponse{
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					FinishReason string  `json:"finish_reason"`
				}{{
					Message:      Message{Content: ""},
					FinishReason: "stop",
				}},
			},
		},
		{
			Response: ChatResponse{
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					FinishReason string  `json:"finish_reason"`
				}{{
					Message:      Message{Content: "answer"},
					FinishReason: "stop",
				}},
			},
		},
	}}

	cfg := configForTest()
	cascade := NewCascade(provider, cfg)
	res, err := cascade.Run(context.Background(), []Message{{Role: "user", Content: "test"}}, nil)
	if err != nil {
		t.Fatalf("cascade: %v", err)
	}
	if !res.Escalated {
		t.Error("expected escalation")
	}
	if res.ModelID != "moonshotai/kimi-k2.6" {
		t.Errorf("expected medium model, got %s", res.ModelID)
	}
}

type mockProvider struct {
	responses []*CallResult
	calls     int
}

func (m *mockProvider) Call(ctx context.Context, req ChatRequest) (*CallResult, error) {
	if m.calls >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses")
	}
	res := m.responses[m.calls]
	m.calls++
	return res, nil
}

func (m *mockProvider) Name() string { return "mock" }

func configForTest() config.ModelProfiles {
	return config.ModelProfiles{
		Local:  "qwen/qwen3-235b",
		Medium: "moonshotai/kimi-k2.6",
		Top:    "minimaxai/minimax-m2.7",
		Cascade: config.CascadeConfig{
			ConfidenceThreshold: 0.85,
			MaxEscalations:      2,
		},
	}
}
