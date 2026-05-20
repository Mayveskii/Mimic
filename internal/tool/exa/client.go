package exa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an Exa API HTTP client with rate-limit backoff.
// Invariant: no more than 1 concurrent request per API key during backoff.
// Source: Mayveskii/exa-mcp-server (exa_rate_limit_management behavior).
type Client struct {
	cfg    Config
	client *http.Client
}

// NewClient creates an Exa client from Config.
// If cfg.Disabled(), returns nil.
func NewClient(cfg Config) *Client {
	if cfg.Disabled() {
		return nil
	}
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout(),
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

// Search performs an Exa web search.
// Invariant: numResults >= 1 && <= 100. query non-empty.
func (c *Client) Search(query string, numResults int, searchType string) (*SearchResponse, error) {
	if numResults < 1 || numResults > 100 {
		numResults = c.cfg.MaxResults
	}
	if searchType == "" {
		searchType = "auto"
	}
	reqBody := SearchRequest{
		Query:      query,
		Type:       searchType,
		NumResults: numResults,
		Contents: &ContentsSpec{
			Highlights: &HighlightsSpec{NumSentences: 1},
		},
	}
	respBody, err := c.post("/search", reqBody)
	if err != nil {
		return nil, fmt.Errorf("exa search failed: %w", err)
	}
	var resp SearchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("exa search decode failed: %w", err)
	}
	return &resp, nil
}

// Fetch retrieves markdown content for given URLs or Exa IDs.
// Invariant: len(urls) <= 100, each URL valid HTTP(S).
func (c *Client) Fetch(urls []string, maxChars int) (*ContentsResponse, error) {
	if len(urls) == 0 {
		return &ContentsResponse{Results: []ContentResult{}}, nil
	}
	if len(urls) > 100 {
		urls = urls[:100]
	}
	reqBody := ContentsRequest{
		URLs:        urls,
		MaxAgeHours: -1, // cache only, avoid livecrawl timeouts
	}
	if maxChars > 0 {
		reqBody.Contents = &ContentsSpec{
			Text: &TextSpec{MaxCharacters: maxChars, IncludeHTML: false},
		}
	}
	respBody, err := c.post("/contents", reqBody)
	if err != nil {
		return nil, fmt.Errorf("exa fetch failed: %w", err)
	}
	var resp ContentsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("exa fetch decode failed: %w", err)
	}
	return &resp, nil
}

// post performs a POST to Exa API with auth header and exponential backoff on 429.
// Invariant: retry count <= cfg.RetryMax, backoff = base * 2^attempt.
func (c *Client) post(path string, body interface{}) ([]byte, error) {
	url := c.cfg.BaseURL + path
	baseLen := len(c.cfg.BaseURL)
	if baseLen > 0 && c.cfg.BaseURL[baseLen-1] == '/' && len(path) > 0 && path[0] == '/' {
		url = c.cfg.BaseURL[:baseLen-1] + path
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.RetryMax; attempt++ {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.cfg.APIKey)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.cfg.RetryMax {
				sleepBackoff(attempt, c.cfg.RetryBackoffBaseMs)
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (429)")
			if attempt < c.cfg.RetryMax {
				sleepBackoff(attempt, c.cfg.RetryBackoffBaseMs)
				continue
			}
			break
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			if attempt < c.cfg.RetryMax {
				sleepBackoff(attempt, c.cfg.RetryBackoffBaseMs)
				continue
			}
			break
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		lastErr = fmt.Errorf("exa API error %d: %s", resp.StatusCode, string(respBody))
		// Fatal client errors: do not retry (auth, payment, bad request)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
		if resp.StatusCode >= 500 && attempt < c.cfg.RetryMax {
			sleepBackoff(attempt, c.cfg.RetryBackoffBaseMs)
			continue
		}
		break
	}
	return nil, fmt.Errorf("exa request failed after %d attempts: %w", c.cfg.RetryMax+1, lastErr)
}

func sleepBackoff(attempt, baseMs int) {
	delay := time.Duration(baseMs) * time.Millisecond * time.Duration(1<<attempt)
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	time.Sleep(delay)
}
