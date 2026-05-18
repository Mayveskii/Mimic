package orchestrator

/*
loop.go — Multi-turn orchestrator loop for MCP tool chains.

Problem: LLMs (kimi, Claude, GPT) make ONE tool_call per API request.
They don't internally loop. For multi-step tasks, we need external coordination:

  User Request
      → Model → tool_call #1 → Execute → Result → Feed back to Model
      → Model → tool_call #2 → Execute → Result → Feed back to Model
      → ...repeat until final answer...

This loop handles the round-trip automatically.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/rtk"
)

// TurnResult captures one iteration of the loop
type TurnResult struct {
	Turn        int                    `json:"turn"`
	ToolName    string                 `json:"tool_name"`
	Arguments   map[string]interface{} `json:"arguments"`
	RawOutput   string                 `json:"raw_output"`
	Compressed  string                 `json:"compressed,omitempty"`
	LatencyMs   int64                  `json:"latency_ms"`
	Error       string                 `json:"error,omitempty"`
}

// ChainResult captures the complete multi-turn execution
type ChainResult struct {
	Turns       []TurnResult `json:"turns"`
	FinalAnswer string       `json:"final_answer"`
	TotalTimeMs int64        `json:"total_time_ms"`
	TotalTokens int          `json:"total_tokens"`
}

// ToolExecutor interface abstracts tool execution (MCP, local, remote)
type ToolExecutor interface {
	Execute(ctx context.Context, name string, args map[string]interface{}) (string, error)
}

// MCPExecutor executes tools via the local C-core MCP bridge
type MCPExecutor struct{}

func (e *MCPExecutor) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	pkt := cgo.Packet{Opcode: name, Args: args}
	result, err := cgo.ExecuteChain([]cgo.Packet{pkt}, 100000, 60000)
	if err != nil {
		return "", err
	}
	return result.Result, nil
}

// LoopConfig controls multi-turn behavior
type LoopConfig struct {
	MaxTurns        int   // Maximum iterations (safety)
	CompressThreshold int // Bytes; above this, apply RTK compression
	BudgetTokens    int   // Max total tokens before graceful stop
	BudgetTimeMs    int64 // Max total time
}

var DefaultLoopConfig = LoopConfig{
	MaxTurns:        10,
	CompressThreshold: 2048,
	BudgetTokens:    100000,
	BudgetTimeMs:    60000,
}

// ModelCaller abstracts the LLM API (OpenRouter, OpenAI, etc.)
type ModelCaller interface {
	Call(ctx context.Context, messages []Message, tools []ToolDef) (*ModelResponse, error)
}

// Message for LLM chat format
type Message struct {
	Role    string     `json:"role"`
	Content string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall from model response
type ToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function FunctionCall    `json:"function"`
}

// FunctionCall details
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ToolDef for registering available tools
type ToolDef struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

// ModelResponse from LLM API
type ModelResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
	Usage     struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// MultiTurnLoop executes a task with external multi-turn coordination.
// It keeps calling the model, executing tools, and feeding results back
// until the model provides a final answer or a limit is reached.
func MultiTurnLoop(
	ctx context.Context,
	caller ModelCaller,
	executor ToolExecutor,
	tools []ToolDef,
	initialPrompt string,
	cfg LoopConfig,
) (*ChainResult, error) {

	start := time.Now()
	chain := &ChainResult{Turns: make([]TurnResult, 0)}

	// Initial message from user
	messages := []Message{
		{Role: "system", Content: "You are an autonomous agent with access to tools. Use them to complete the user's request."},
		{Role: "user", Content: initialPrompt},
	}

	for turn := 1; turn <= cfg.MaxTurns; turn++ {
		turnStart := time.Now()

		// Call model
		resp, err := caller.Call(ctx, messages, tools)
		if err != nil {
			return nil, fmt.Errorf("turn %d model call failed: %w", turn, err)
		}
		chain.TotalTokens += resp.Usage.TotalTokens

		// Check if model wants to use tools
		if len(resp.ToolCalls) == 0 {
			// Final answer
			chain.FinalAnswer = resp.Content
			chain.TotalTimeMs = time.Since(start).Milliseconds()
			return chain, nil
		}

		// Execute each tool call (usually just 1, but handle multiple)
		for _, tc := range resp.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				// Invalid arguments from model — record and continue
				chain.Turns = append(chain.Turns, TurnResult{
					Turn:     turn,
					ToolName: tc.Function.Name,
					Error:    fmt.Sprintf("invalid JSON arguments: %v", err),
					LatencyMs: time.Since(turnStart).Milliseconds(),
				})
				continue
			}

			// Execute tool
			output, err := executor.Execute(ctx, tc.Function.Name, args)
			if err != nil {
				chain.Turns = append(chain.Turns, TurnResult{
					Turn:     turn,
					ToolName: tc.Function.Name,
					Arguments: args,
					Error:    err.Error(),
					LatencyMs: time.Since(turnStart).Milliseconds(),
				})
				// Feed error back to model
				messages = append(messages,
					Message{Role: "assistant", Content: "", ToolCalls: []ToolCall{tc}},
					Message{Role: "tool", Content: fmt.Sprintf("Error: %v", err)},
				)
				continue
			}

			// Compress if large
			compressed := ""
			if len(output) > cfg.CompressThreshold {
				compressed = rtk.Compress(output, rtk.ContentText, rtk.Config{MaxLines: 50, StripAnsi: true, CollapseBlanks: true})
				if len(compressed) < len(output) {
					output = compressed
				}
			}

			chain.Turns = append(chain.Turns, TurnResult{
				Turn:       turn,
				ToolName:   tc.Function.Name,
				Arguments:  args,
				RawOutput:  output,
				Compressed: compressed,
				LatencyMs:  time.Since(turnStart).Milliseconds(),
			})

			// Feed result back to model
			messages = append(messages,
				Message{Role: "assistant", Content: "", ToolCalls: []ToolCall{tc}},
				Message{Role: "tool", Content: output},
			)
		}

		// Budget checks
		if chain.TotalTokens >= cfg.BudgetTokens {
			chain.FinalAnswer = "[Budget exhausted: token limit reached]"
			chain.TotalTimeMs = time.Since(start).Milliseconds()
			return chain, nil
		}
		if time.Since(start).Milliseconds() >= cfg.BudgetTimeMs {
			chain.FinalAnswer = "[Budget exhausted: time limit reached]"
			chain.TotalTimeMs = time.Since(start).Milliseconds()
			return chain, nil
		}
	}

	// Max turns reached
	chain.FinalAnswer = "[Max turns reached without final answer]"
	chain.TotalTimeMs = time.Since(start).Milliseconds()
	return chain, nil
}

// ToMCPMessages converts internal Message format to OpenAI/MCP-compatible messages
func ToMCPMessages(messages []Message) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(messages))
	for _, m := range messages {
		msg := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if len(m.ToolCalls) > 0 {
			msg["tool_calls"] = m.ToolCalls
		}
		out = append(out, msg)
	}
	return out
}

// FromMCPResponse converts OpenRouter/OpenAI response to ModelResponse
func FromMCPResponse(data map[string]interface{}) *ModelResponse {
	resp := &ModelResponse{}
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		choice := choices[0].(map[string]interface{})
		if msg, ok := choice["message"].(map[string]interface{}); ok {
			resp.Content = msg["content"].(string)
			if tcs, ok := msg["tool_calls"].([]interface{}); ok {
				for _, tc := range tcs {
					m := tc.(map[string]interface{})
					resp.ToolCalls = append(resp.ToolCalls, ToolCall{
						ID:   m["id"].(string),
						Type: m["type"].(string),
						Function: FunctionCall{
							Name:      m["function"].(map[string]interface{})["name"].(string),
							Arguments: m["function"].(map[string]interface{})["arguments"].(string),
						},
					})
				}
			}
		}
	}
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		resp.Usage.PromptTokens = int(usage["prompt_tokens"].(float64))
		resp.Usage.CompletionTokens = int(usage["completion_tokens"].(float64))
		resp.Usage.TotalTokens = int(usage["total_tokens"].(float64))
	}
	return resp
}
