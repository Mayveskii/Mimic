// Package embed provides a Go client for the Mimic embedding service.
// The service runs sentence-transformers/all-MiniLM-L6-v2 model and
// converts text to int8[384] embeddings compatible with embryo mesh.
package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultEndpoint = "http://localhost:1137"

// Client is an HTTP client for the embedding service.
type Client struct {
	Endpoint string
	client   *http.Client
}

// NewClient creates an embed client. Uses DefaultEndpoint if empty.
func NewClient(endpoint string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		Endpoint: endpoint,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// EmbedResponse from the Python service.
type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
	Dim       int       `json:"dim"`
	LatencyMs float64   `json:"latency_ms"`
}

// Int8Response from the Python service.
type Int8Response struct {
	Int8      []int8  `json:"int8"`
	Dim       int     `json:"dim"`
	LatencyMs float64 `json:"latency_ms"`
}

// Embed returns float32[384] embedding for text.
func (c *Client) Embed(text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]string{"text": text})
	resp, err := c.client.Post(c.Endpoint+"/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embed failed: %s: %s", resp.Status, string(data))
	}

	var result EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("embed decode: %w", err)
	}
	return result.Embedding, nil
}

// EmbedInt8 returns int8[384] embedding for text (one-shot).
func (c *Client) EmbedInt8(text string) ([384]int8, error) {
	body, _ := json.Marshal(map[string]string{"text": text})
	resp, err := c.client.Post(c.Endpoint+"/embed_int8", "application/json", bytes.NewReader(body))
	if err != nil {
		return [384]int8{}, fmt.Errorf("embed_int8 request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return [384]int8{}, fmt.Errorf("embed_int8 failed: %s: %s", resp.Status, string(data))
	}

	var result Int8Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return [384]int8{}, fmt.Errorf("embed_int8 decode: %w", err)
	}

	var arr [384]int8
	for i, v := range result.Int8 {
		if i >= 384 {
			break
		}
		arr[i] = v
	}
	return arr, nil
}

// Health checks if embedding service is alive.
func (c *Client) Health() bool {
	resp, err := c.client.Get(c.Endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
