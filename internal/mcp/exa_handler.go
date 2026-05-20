package mcp

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Mayveskii/Mimic/internal/rtk"
	"github.com/Mayveskii/Mimic/internal/tool/exa"
)

func debugf(format string, args ...interface{}) {
	f, _ := os.OpenFile("/tmp/mimic_exa_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] "+format+"\n", append([]interface{}{time.Now().Format("15:04:05.000")}, args...)...)
		f.Close()
	}
}

// ExaHandler routes EXA_SEARCH, EXA_FETCH, and MIMIC_RESEARCH to the Exa API.
// Invariant: if exa client is disabled (no API key), all tools return error with clear message.
// Source: Mayveskii/exa-mcp-server (exa_web_search, exa_web_fetch, exa_rate_limit_management behaviors).
type ExaHandler struct {
	client *exa.Client
}

// NewExaHandler creates an Exa handler. If cfg.Disabled(), returns nil-safe handler.
func NewExaHandler(cfg exa.Config) *ExaHandler {
	return &ExaHandler{
		client: exa.NewClient(cfg),
	}
}

func (h *ExaHandler) disabled() map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": "Exa integration is disabled: EXA_API_KEY not set. Configure environment variable and restart."},
		},
		"isError": true,
	}
}

// HandleExaSearch processes EXA_SEARCH tool call.
// Arguments: query (string, required), numResults (int, optional), type (string, optional).
func (h *ExaHandler) HandleExaSearch(args map[string]interface{}) map[string]interface{} {
	if h.client == nil {
		return h.disabled()
	}
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return mcpError("EXA_SEARCH requires 'query' (string)")
	}
	numResults := 10
	if n, ok := args["numResults"].(float64); ok {
		numResults = int(n)
	} else if n, ok := args["numResults"].(int); ok {
		numResults = n
	}
	searchType := "auto"
	if t, ok := args["type"].(string); ok && t != "" {
		searchType = t
	}
	resp, err := h.client.Search(query, numResults, searchType)
	if err != nil {
		return mcpError(fmt.Sprintf("Exa search error: %v", err))
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Exa search: %q (type=%s, results=%d)\n\n", query, searchType, len(resp.Results)))
	for i, r := range resp.Results {
		b.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n", i+1, r.Title, r.URL))
		if len(r.Highlights) > 0 {
			for _, hl := range r.Highlights {
				if len(hl) > 300 {
					hl = hl[:297] + "..."
				}
				b.WriteString(fmt.Sprintf("   Highlight: %s\n", hl))
			}
		} else if r.Text != "" {
			text := r.Text
			if len(text) > 300 {
				text = text[:297] + "..."
			}
			b.WriteString(fmt.Sprintf("   Text: %s\n", text))
		}
		b.WriteString("\n")
	}
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": b.String()},
		},
	}
}

// HandleExaFetch processes EXA_FETCH tool call.
// Arguments: urls ([]string, required), maxChars (int, optional).
func (h *ExaHandler) HandleExaFetch(args map[string]interface{}) map[string]interface{} {
	if h.client == nil {
		return h.disabled()
	}
	urlsRaw, ok := args["urls"].([]interface{})
	if !ok || len(urlsRaw) == 0 {
		return mcpError("EXA_FETCH requires 'urls' ([]string)")
	}
	var urls []string
	for _, u := range urlsRaw {
		if s, ok := u.(string); ok {
			urls = append(urls, s)
		}
	}
	if len(urls) == 0 {
		return mcpError("EXA_FETCH: no valid URLs in 'urls' array")
	}
	maxChars := 0
	if m, ok := args["maxChars"].(float64); ok {
		maxChars = int(m)
	} else if m, ok := args["maxChars"].(int); ok {
		maxChars = m
	}
	resp, err := h.client.Fetch(urls, maxChars)
	if err != nil {
		return mcpError(fmt.Sprintf("Exa fetch error: %v", err))
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Exa fetch: %d URL(s)\n\n", len(urls)))
	for i, r := range resp.Results {
		b.WriteString(fmt.Sprintf("--- Result %d ---\nURL: %s\nTitle: %s\n", i+1, r.URL, r.Title))
		if r.Text != "" {
			b.WriteString(fmt.Sprintf("Content:\n%s\n", r.Text))
		}
		b.WriteString("\n")
	}
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": b.String()},
		},
	}
}

// HandleMimicResearch processes MIMIC_RESEARCH tool call.
// Tier 2: search → fetch top N → RTK compress → return structured summary.
// Arguments: topic (string, required), depth (string, optional: shallow|deep).
func (h *ExaHandler) HandleMimicResearch(args map[string]interface{}) map[string]interface{} {
	debugf("HandleMimicResearch START args=%v", args)
	if h.client == nil {
		debugf("HandleMimicResearch client=nil -> disabled")
		return h.disabled()
	}
	topic, ok := args["topic"].(string)
	if !ok || topic == "" {
		debugf("HandleMimicResearch empty topic")
		return mcpError("MIMIC_RESEARCH requires 'topic' (string)")
	}
	depth := "shallow"
	if d, ok := args["depth"].(string); ok && (d == "shallow" || d == "deep") {
		depth = d
	}
	debugf("HandleMimicResearch topic=%q depth=%s", topic, depth)

	// Step 1: search
	debugf("HandleMimicResearch calling Search...")
	searchResp, err := h.client.Search(topic, 5, "auto")
	if err != nil {
		debugf("HandleMimicResearch Search error: %v", err)
		return mcpError(fmt.Sprintf("Research search error: %v", err))
	}
	debugf("HandleMimicResearch Search OK results=%d", len(searchResp.Results))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Mimic Research: %q (depth=%s)\n\n", topic, depth))
	b.WriteString(fmt.Sprintf("Found %d search results.\n\n", len(searchResp.Results)))

	if depth == "shallow" {
		for i, r := range searchResp.Results {
			b.WriteString(fmt.Sprintf("%d. %s\n   %s\n", i+1, r.Title, r.URL))
			if r.Text != "" {
				text := r.Text
				if len(text) > 300 {
					text = text[:297] + "..."
				}
				b.WriteString(fmt.Sprintf("   %s\n", text))
			}
			b.WriteString("\n")
		}
		debugf("HandleMimicResearch shallow building response len=%d", b.Len())
		result := map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": b.String()},
			},
		}
		debugf("HandleMimicResearch shallow returning")
		return result
	}

	// Step 2: deep — fetch top URL only (keep under opencode 30s timeout)
	var urls []string
	for _, r := range searchResp.Results {
		if r.URL != "" {
			urls = append(urls, r.URL)
		}
	}
	if len(urls) > 1 {
		urls = urls[:1]
	}

	debugf("HandleMimicResearch fetching urls=%v", urls)
	fetchResp, err := h.client.Fetch(urls, 50)
	debugf("HandleMimicResearch fetch done err=%v", err)
	if err != nil {
		b.WriteString(fmt.Sprintf("\nFetch failed for top results: %v\n", err))
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": b.String()},
			},
		}
	}

	// Step 3: compress each result with RTK
	for i, r := range fetchResp.Results {
		b.WriteString(fmt.Sprintf("--- Source %d ---\nURL: %s\nTitle: %s\n", i+1, r.URL, r.Title))
		if r.Text != "" {
			compressed := rtk.Compress(r.Text, rtk.ContentText, rtk.AggressiveConfig())
			b.WriteString(fmt.Sprintf("Compressed content (%d → %d chars):\n%s\n", len(r.Text), len(compressed), compressed))
		}
		b.WriteString("\n")
	}

	debugf("HandleMimicResearch deep returning")
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": b.String()},
		},
	}
}

func mcpError(msg string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": msg},
		},
		"isError": true,
	}
}
