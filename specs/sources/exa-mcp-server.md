```yaml
repo: Mayveskii/exa-mcp-server
url: https://github.com/Mayveskii/exa-mcp-server
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of exa-mcp-server. MCP server providing web search, content fetching,
  and content extraction tools via Exa API. Includes rate limit management
  with backoff, content size enforcement, and zod schema validation on all inputs.

advantages:
  - id: exa_web_search
    what: MCP tool for web search — query → POST /search → results with highlights, configurable numResults and type (auto/keyword/neural)
    evidence: "src/tools/ — web_search_exa tool definition with zod schema; src/api/ — /search endpoint with query parameters"

  - id: exa_web_fetch
    what: MCP tool for content fetching — URLs → POST /contents → markdown content with optional includeHtml and livecrawl options
    evidence: "src/tools/ — web_fetch_exa tool definition; src/api/ — /contents endpoint with URL array and format options"

  - id: exa_content_extraction
    what: Content extraction from search results: extract key paragraphs, entities, and summaries; configurable extraction depth per query
    evidence: "src/tools/ — content extraction parameters in search tool; src/api/ — /search with includes parameter for entity/summary extraction"

  - id: exa_rate_limit_management
    what: Rate limit management: per-request rate limit tracking with backoff on 429; configurable retry count and backoff multiplier; fail fast after max retries
    evidence: "src/api/ — HTTP client with rate limit headers parsing; retry logic with exponential backoff on 429 responses"

  - id: exa_tool_pattern
    what: MCP tool registration pattern: define tool with zod schema → validate input → execute API call → return structured ExecuteResult; every tool follows same contract
    evidence: "src/index.ts — tool registration with zod schema validation; all tools in src/tools/ follow identical pattern"

applications:
  - advantage_id: exa_web_search
    implemented_in: internal/tool/tools_search.go
    mechanism: "MCP tool: websearch(query, numResults, type) → HTTP POST /search with auth header → parse results → return to agent with highlights"
    invariant: "Search results limited to numResults (default 10, max 100). Timeout 30s. Auth required."
    status: planned

  - advantage_id: exa_web_fetch
    implemented_in: internal/tool/tools_fetch.go
    mechanism: "MCP tool: webfetch(urls, format) → HTTP POST /contents with URL array → parse markdown → return to agent"
    invariant: "Content size ≤ 1MB per URL. Timeout 60s per URL. URLs must be valid HTTP(S)."
    status: planned

  - advantage_id: exa_content_extraction
    implemented_in: internal/tool/tools_extract.go
    mechanism: "Search with includes={entities, summary} → extract key content from results → return structured extraction with entities + summary"
    invariant: "Extraction depth configurable (shallow/deep). Entities include type and relevance score. Summary ≤ 200 words."
    status: planned

  - advantage_id: exa_rate_limit_management
    implemented_in: internal/mcp/ratelimit.go
    mechanism: "Track X-RateLimit-Remaining header → if 0 → backoff for X-RateLimit-Reset seconds → retry with exponential backoff on 429 → fail after max_retries"
    invariant: "No more than 1 concurrent request per API key during backoff. Max 3 retries. Backoff = base_delay × 2^attempt."
    status: planned

  - advantage_id: exa_tool_pattern
    implemented_in: internal/tool/registry.go
    mechanism: "ToolDef: {ID, Description, ParameterSchema(zod), Execute} → register → on call: validate(params) → execute() → ExecuteResult{content, isError}"
    invariant: "Every tool must have zod schema. Invalid input → validation error before execution. isError=true on any execution failure."
    status: planned

control:
  - advantage_id: exa_web_search
    verification: "Integration test: search query → verify results with highlights returned; invalid auth → verify 401 error"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_web_fetch
    verification: "Integration test: fetch valid URL → verify markdown content; fetch invalid URL → verify error in ExecuteResult"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_content_extraction
    verification: "Integration test: search with extraction → verify entities and summary in results"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_rate_limit_management
    verification: "Unit test: mock 429 response → verify backoff and retry; 3 consecutive 429s → verify fail fast after max retries"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_tool_pattern
    verification: "Unit test: register tool without schema → verify rejected; call tool with invalid params → verify validation error"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never
```
