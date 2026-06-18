package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Message represents a single chat message in OpenAI-compatible format.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call requested by the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ToolSchema describes a tool available to the model.
type ToolSchema struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

// ChatRequest is the OpenAI-compatible chat completion request.
type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Tools          []ToolSchema    `json:"tools,omitempty"`
	ToolChoice     string          `json:"tool_choice,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ResponseFormat requests structured output from the model.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatResponse is the OpenAI-compatible chat completion response.
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents an error returned by the provider.
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error: %s (type=%s code=%s)", e.Message, e.Type, e.Code)
}

// CallResult captures the model response plus metadata.
type CallResult struct {
	Response      ChatResponse
	Latency       time.Duration
	TokensIn      int
	TokensOut     int
	RetryAttempts int
}

// Caller is the minimal interface for LLM callers.
type Caller interface {
	Call(ctx context.Context, req ChatRequest) (*CallResult, error)
}

// Provider is the interface for named LLM providers.
type Provider interface {
	Caller
	Name() string
}

// GonkaGateProvider implements Provider for the GonkaGate OpenAI-compatible API.
type GonkaGateProvider struct {
	endpoint   string
	apiKey     string
	timeout    time.Duration
	retryMax   int
	httpClient *http.Client
}

// NewGonkaGateProvider creates a GonkaGate provider from config values.
func NewGonkaGateProvider(endpoint, envKey string, timeoutMs, retryMax int) (*GonkaGateProvider, error) {
	if endpoint == "" {
		endpoint = "https://api.gonkagate.com/v1"
	}
	apiKey := os.Getenv(envKey)
	if apiKey == "" {
		apiKey = os.Getenv("GONKAGATE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GonkaGate API key not found in env var %s", envKey)
	}
	if timeoutMs <= 0 {
		timeoutMs = 120000
	}
	if retryMax <= 0 {
		retryMax = 3
	}

	return &GonkaGateProvider{
		endpoint:   endpoint,
		apiKey:     apiKey,
		timeout:    time.Duration(timeoutMs) * time.Millisecond,
		retryMax:   retryMax,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}, nil
}

// Name returns the provider name.
func (p *GonkaGateProvider) Name() string { return "gonkagate" }

// Call sends a chat completion request and returns the result.
func (p *GonkaGateProvider) Call(ctx context.Context, req ChatRequest) (*CallResult, error) {
	url := p.endpoint + "/chat/completions"
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	var result *CallResult

	for attempt := 0; attempt <= p.retryMax; attempt++ {
		start := time.Now()
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("HTTP-Referer", "https://github.com/Mayveskii/Mimic")
		httpReq.Header.Set("X-Title", "Mimic")

		resp, err := p.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			if attempt < p.retryMax {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("request failed after %d retries: %w", p.retryMax, lastErr)
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			if attempt < p.retryMax {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("read response failed after %d retries: %w", p.retryMax, lastErr)
		}

		latency := time.Since(start)

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
			if attempt < p.retryMax {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, lastErr
		}

		var chatResp ChatResponse
		if err := json.Unmarshal(data, &chatResp); err != nil {
			return nil, fmt.Errorf("decode response: %w (body=%s)", err, string(data))
		}

		if chatResp.Error != nil {
			// Do not retry 4xx client errors (auth, payment, bad request).
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, chatResp.Error
			}
			lastErr = chatResp.Error
			if attempt < p.retryMax {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, lastErr
		}

		result = &CallResult{
			Response:      chatResp,
			Latency:       latency,
			TokensIn:      chatResp.Usage.PromptTokens,
			TokensOut:     chatResp.Usage.CompletionTokens,
			RetryAttempts: attempt,
		}
		return result, nil
	}

	return nil, lastErr
}

func backoff(attempt int) time.Duration {
	base := time.Second
	max := 30 * time.Second
	d := base * time.Duration(1<<attempt)
	if d > max {
		return max
	}
	return d
}
