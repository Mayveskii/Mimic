// Package qdrant provides a lightweight HTTP client for Qdrant vector search.
// Used as fallback for large-scale queries when local mesh is insufficient.
package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const DefaultEndpoint = "http://localhost:6333"

// Client is a simple HTTP client for Qdrant REST API.
type Client struct {
	Endpoint   string
	Collection string
	httpClient *http.Client
}

// NewClient creates a qdrant client.
func NewClient(endpoint, collection string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	if collection == "" {
		collection = "binary_mesh_chunks"
	}
	return &Client{
		Endpoint:   endpoint,
		Collection: collection,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SearchResult is a single match from Qdrant.
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float64              `json:"vector,omitempty"`
	Version uint64                 `json:"version"`
}

// Search performs vector similarity search.
func (c *Client) Search(vector []float64, topK int, threshold float64) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}

	payload := map[string]interface{}{
		"vector":          vector,
		"limit":           topK,
		"with_payload":    true,
		"with_vector":     false,
		"score_threshold": threshold,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", c.Endpoint, c.Collection)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant search failed: %s", resp.Status)
	}

	var result struct {
		Result []SearchResult `json:"result"`
		Status string         `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Status != "ok" {
		return nil, fmt.Errorf("qdrant status: %s", result.Status)
	}
	return result.Result, nil
}

// Count returns total points in collection.
func (c *Client) Count() (int, error) {
	url := fmt.Sprintf("%s/collections/%s", c.Endpoint, c.Collection)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			PointsCount int `json:"points_count"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Result.PointsCount, nil
}

// Health checks if Qdrant is reachable.
func (c *Client) Health() bool {
	resp, err := c.httpClient.Get(c.Endpoint + "/healthz")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
