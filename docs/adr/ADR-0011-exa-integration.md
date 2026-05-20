# ADR-0011: Exa Web Search Integration for Mimic MCP Server

## Status
Proposed

## Context

Mimic needs web search capability to help AI agents research production patterns, library documentation, and best practices. After evaluating multiple search APIs (SerpAPI, Tavily, Exa), Exa was selected because:

1. **Structured content extraction**: Returns clean markdown + highlights (not raw HTML)
2. **Content API**: Can fetch full page text by URL (critical for code examples)
3. **Research types**: `auto` (~1s) vs `deep` (4-15s) for different use cases
4. **Token-efficient**: Highlights mode reduces LLM context by ~10x vs full text

## Decision

We will integrate Exa API as three MCP tools:

### 1. EXA_SEARCH
- **Purpose**: Find relevant URLs for a query
- **When to use**: Need to discover resources, compare approaches, find documentation
- **Parameters**: `query` (required), `num_results` (1-100, default 10), `type` (auto|instant|fast|deep-lite|deep|deep-reasoning)
- **Returns**: Title, URL, highlights (if configured)
- **Timeout**: 30s default

### 2. EXA_FETCH  
- **Purpose**: Get full content from known URLs
- **When to use**: Already have URL(s), need code/text extraction
- **Parameters**: `urls` []string (max 100), `max_characters` (0=unlimited)
- **Returns**: Full text, title, published date per URL
- **Timeout**: 15s (with DisableKeepAlives to prevent connection stall)
- **Cache policy**: Default cache-first (no `max_age_hours` = use cache if available)

### 3. MIMIC_RESEARCH (deep/shallow)
- **Purpose**: One-shot research pipeline: search → fetch top results → compress
- **When to use**: Need comprehensive overview of a topic
- **Parameters**: `topic` (required), `depth` (shallow=urls only, deep=fetch+compress)
- **Returns**: Structured summary with sources
- **Limitation**: Deep mode fetches only 1 URL to stay under 30s timeout

## Authentication

Exa accepts **both** auth headers (verified by direct API test):
- `x-api-key: <key>` (canonical, per docs)
- `Authorization: Bearer <key>` (also works)

Mimic uses `x-api-key` (canonical).

## JSON Schema

Exa API uses `snake_case` for request fields (verified by direct API test):
- `num_results` (not `numResults`)
- `max_characters` (not `maxCharacters`)  
- `num_sentences` (not `numSentences`)
- `max_age_hours` (not `maxAgeHours`)

Response uses camelCase: `requestId`, `publishedDate`.

## Retry Policy

| Status Code | Behavior |
|-------------|----------|
| 200-299 | Success |
| 400-499 | **No retry** (client error: bad request, auth, payment) |
| 429 | Retry with exponential backoff |
| 500+ | Retry with exponential backoff (max 3 attempts) |

## Consequences

### Positive
- Agents can research production patterns without leaving IDE
- Search + Fetch combo covers 95% of research use cases
- Caching (default) keeps latency <1s for most queries

### Negative  
- Deep research limited to 1 URL fetched (to stay under timeout)
- Exa is paid API ($7/1k searches)
- No offline fallback (unlike mesh query)

## References
- Exa Docs: https://exa.ai/docs
- Test results: `/tmp/mimic_exa_debug.log`
