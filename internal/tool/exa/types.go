package exa

// SearchRequest mirrors Exa POST /search API (2026).
// Invariant: numResults <= 100, query len >= 1.
// Reference: https://exa.ai/docs/reference/search-api-guide
type SearchRequest struct {
	Query         string        `json:"query"`
	Type          string        `json:"type,omitempty"`       // auto | instant | fast | deep-lite | deep | deep-reasoning
	NumResults    int           `json:"numResults,omitempty"` // max 100
	Contents      *ContentsSpec `json:"contents,omitempty"`   // highlights, summary, text
	Category      string        `json:"category,omitempty"`   // company | people | research_paper | news | personal_site | financial_report
	UseAutoprompt bool          `json:"useAutoprompt,omitempty"`
	OutputSchema  interface{}   `json:"output_schema,omitempty"` // structured JSON extraction
}

// ContentsSpec configures what content to return per result.
type ContentsSpec struct {
	Highlights *HighlightsSpec `json:"highlights,omitempty"` // 10x token efficient extracts
	Text       *TextSpec       `json:"text,omitempty"`       // full text
	Summary    *SummarySpec    `json:"summary,omitempty"`    // AI-generated summary
}

type HighlightsSpec struct {
	NumSentences     int `json:"num_sentences,omitempty"` // default varies by plan
	HighlightsPerURL int `json:"highlights_per_url,omitempty"`
}

type TextSpec struct {
	MaxCharacters int  `json:"max_characters,omitempty"`
	IncludeHTML   bool `json:"include_html,omitempty"`
}

type SummarySpec struct {
	Query string `json:"query,omitempty"` // what to summarize about
}

// SearchResponse mirrors Exa /search response.
type SearchResponse struct {
	Results            []SearchResult `json:"results"`
	RequestID          string         `json:"requestId,omitempty"`
	AutopromptString   string         `json:"autopromptString,omitempty"`
	ResolvedSearchType string         `json:"resolvedSearchType,omitempty"`
	CostDollars        *CostBreakdown `json:"costDollars,omitempty"`
	SearchTime         float64        `json:"searchTime,omitempty"`
}

type CostBreakdown struct {
	Total  float64            `json:"total,omitempty"`
	Search map[string]float64 `json:"search,omitempty"`
}

// SearchResult is a single result from Exa search.
type SearchResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	ID            string  `json:"id"`
	PublishedDate string  `json:"publishedDate,omitempty"`
	Author        string  `json:"author,omitempty"`
	Score         float64 `json:"score,omitempty"`
	Favicon       string  `json:"favicon,omitempty"`
	// Content fields only present if Contents spec was requested
	Highlights []string `json:"highlights,omitempty"`
	Text       string   `json:"text,omitempty"`
	Summary    string   `json:"summary,omitempty"`
}

// ContentsRequest mirrors Exa POST /contents API (2026).
// Invariant: len(urls or ids) <= 100.
// Reference: https://exa.ai/docs/reference/contents-api
type ContentsRequest struct {
	IDs           []string      `json:"ids,omitempty"`           // Exa result IDs from search
	URLs          []string      `json:"urls,omitempty"`          // direct URLs
	Contents      *ContentsSpec `json:"contents,omitempty"`      // what to return
	MaxAgeHours   int           `json:"max_age_hours,omitempty"` // -1=cache only, 0=always live, 24=24h cache
	Subpages      int           `json:"subpages,omitempty"`
	SubpageTarget []string      `json:"subpage_target,omitempty"`
}

// ContentsResponse mirrors Exa /contents response.
type ContentsResponse struct {
	Results   []ContentResult `json:"results"`
	RequestID string          `json:"requestId,omitempty"`
}

// ContentResult is a single fetched content.
type ContentResult struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Title         string   `json:"title,omitempty"`
	Text          string   `json:"text,omitempty"`
	Highlights    []string `json:"highlights,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	Author        string   `json:"author,omitempty"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	Score         float64  `json:"score,omitempty"`
}

// Config holds Exa client configuration from environment.
// Source: Mayveskii/exa-mcp-server (rate_limit_management + web_search + web_fetch behaviors).
type Config struct {
	APIKey             string
	BaseURL            string
	MaxResults         int
	TimeoutMs          int
	RetryMax           int
	RetryBackoffBaseMs int
}
