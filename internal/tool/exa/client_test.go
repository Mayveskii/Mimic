package exa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewClientDisabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "empty API key",
			cfg: Config{
				APIKey:  "",
				BaseURL: "https://api.exa.ai",
			},
		},
		{
			name: "empty base URL",
			cfg: Config{
				APIKey:  "test-key",
				BaseURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.cfg)
			if client != nil {
				t.Fatal("expected nil client for disabled config")
			}
		})
	}
}

func TestSearchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/search" {
			t.Errorf("expected /search, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Bearer auth, got %s", auth)
		}

		resp := SearchResponse{
			Results: []SearchResult{
				{
					Title: "Go Programming Language",
					URL:   "https://golang.org",
					ID:    "go-123",
					Text:  "Go is an open source programming language.",
				},
				{
					Title: "Go Documentation",
					URL:   "https://go.dev/doc",
					ID:    "go-456",
					Text:  "Official Go documentation.",
				},
			},
			RequestID: "req-abc",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		MaxResults:         10,
		RetryMax:           2,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	result, err := client.Search("go programming", 2, "auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.RequestID != "req-abc" {
		t.Errorf("expected requestId req-abc, got %s", result.RequestID)
	}

	first := result.Results[0]
	if first.Title != "Go Programming Language" {
		t.Errorf("expected title 'Go Programming Language', got %s", first.Title)
	}
	if first.URL != "https://golang.org" {
		t.Errorf("expected URL https://golang.org, got %s", first.URL)
	}
	if first.Text != "Go is an open source programming language." {
		t.Errorf("expected text match, got %s", first.Text)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Query == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"query cannot be empty"}`))
			return
		}
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           0,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	_, err := client.Search("", 1, "auto")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !strings.Contains(err.Error(), "exa search failed") {
		t.Errorf("expected wrapped search error, got: %v", err)
	}
}

func TestFetchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contents" {
			t.Errorf("expected /contents, got %s", r.URL.Path)
		}

		var req ContentsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Contents != nil && req.Contents.Text != nil && req.Contents.Text.MaxCharacters != 1000 {
			t.Errorf("expected maxChars 1000, got %d", req.Contents.Text.MaxCharacters)
		}

		resp := ContentsResponse{
			Results: []ContentResult{
				{
					ID:    "content-1",
					URL:   "https://example.com/article",
					Title: "Sample Article",
					Text:  "# Heading\n\nArticle body here.",
				},
			},
			RequestID: "req-fetch-1",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           1,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	result, err := client.Fetch([]string{"https://example.com/article"}, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}

	item := result.Results[0]
	if item.URL != "https://example.com/article" {
		t.Errorf("expected URL https://example.com/article, got %s", item.URL)
	}
	if item.Title != "Sample Article" {
		t.Errorf("expected title 'Sample Article', got %s", item.Title)
	}
	if item.Text != "# Heading\n\nArticle body here." {
		t.Errorf("expected text match, got %s", item.Text)
	}
}

func TestFetchEmptyURLs(t *testing.T) {
	cfg := Config{
		APIKey:  "test-key",
		BaseURL: "http://localhost",
	}
	client := NewClient(cfg)

	result, err := client.Fetch([]string{}, 0)
	if err != nil {
		t.Fatalf("unexpected error for empty URLs: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Results))
	}
}

func TestPostRateLimitBackoffSuccess(t *testing.T) {
	var mu sync.Mutex
	reqCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		count := reqCount
		reqCount++

		if count < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"title":"OK","url":"https://ok.com","id":"ok"}]}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           2,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	start := time.Now()
	result, err := client.Search("test", 1, "auto")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if result == nil || len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %+v", result)
	}

	mu.Lock()
	finalCount := reqCount
	mu.Unlock()
	if finalCount != 3 {
		t.Fatalf("expected 3 requests (2x 429 + 1 success), got %d", finalCount)
	}

	// With RetryBackoffBaseMs=1, sleeps are ~1ms and ~2ms, so total > 2ms
	if elapsed < 2*time.Millisecond {
		t.Logf("warning: backoff may not have occurred (elapsed %v)", elapsed)
	}
}

func TestPostRateLimitMaxRetriesExceeded(t *testing.T) {
	var mu sync.Mutex
	reqCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		reqCount++
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           2,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	start := time.Now()
	_, err := client.Search("test", 1, "auto")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error after max retries exceeded")
	}
	if !strings.Contains(err.Error(), "exa request failed") {
		t.Errorf("expected wrapped request error, got: %v", err)
	}

	mu.Lock()
	finalCount := reqCount
	mu.Unlock()
	if finalCount != 3 {
		t.Fatalf("expected 3 requests for RetryMax=2, got %d", finalCount)
	}

	if elapsed < 2*time.Millisecond {
		t.Logf("warning: backoff may not have occurred (elapsed %v)", elapsed)
	}
}

func TestPostServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           1,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	_, err := client.post("/test", map[string]string{"key": "value"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "exa API error 500") {
		t.Errorf("expected API error in message, got: %v", err)
	}
}

func TestPostRequestBody(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"received":true}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:             "test-key",
		BaseURL:            server.URL,
		RetryMax:           0,
		RetryBackoffBaseMs: 1,
		TimeoutMs:          5000,
	}
	client := NewClient(cfg)

	body, err := client.post("/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), `"received":true`) {
		t.Fatalf("unexpected response body: %s", string(body))
	}
	if receivedBody == nil {
		t.Fatal("expected server to receive body")
	}
	if receivedBody["key"] != "value" {
		t.Errorf("expected body key=value, got %+v", receivedBody)
	}
}

func TestPostURLWithTrailingSlash(t *testing.T) {
	var reqPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath = r.URL.Path
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := Config{
		APIKey:    "test-key",
		BaseURL:   server.URL + "/",
		RetryMax:  0,
		TimeoutMs: 5000,
	}
	client := NewClient(cfg)
	_, err := client.post("/search", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqPath != "/search" {
		t.Errorf("expected path /search, got %s", reqPath)
	}
}
