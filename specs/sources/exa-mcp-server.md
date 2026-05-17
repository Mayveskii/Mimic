```yaml
repo: Mayveskii/exa-mcp-server
url: https://github.com/Mayveskii/exa-mcp-server
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of exa-mcp-server. MCP server providing web search and content
  fetching tools via Exa API.

advantages:
  - id: exa_web_search
    what: MCP tool for web search — query → POST /search → results with highlights
    evidence: "src/tools/ — web_search_exa tool definition, src/api/ — /search endpoint"

  - id: exa_web_fetch
    what: MCP tool for content fetching — URLs → POST /contents → markdown content
    evidence: "src/tools/ — web_fetch_exa tool definition, src/api/ — /contents endpoint"

  - id: exa_tool_pattern
    what: MCP tool registration pattern: define tool → validate input → execute → return result
    evidence: "src/index.ts — tool registration with zod schema validation"

applications:
  - advantage_id: exa_web_search
    implemented_in: internal/tool/tools_search.go
    mechanism: "MCP tool: websearch(query) → HTTP POST /search → results → agent"
    invariant: "Search results limited to 10 items, timeout 30s"
    status: planned

  - advantage_id: exa_web_fetch
    implemented_in: internal/tool/tools_fetch.go
    mechanism: "MCP tool: webfetch(urls) → HTTP POST /contents → markdown → agent"
    invariant: "Content size ≤ 1MB per URL, timeout 60s"
    status: planned

  - advantage_id: exa_tool_pattern
    implemented_in: internal/tool/registry.go
    mechanism: "Tool registration: define(id, description, zod schema, execute) → validate → execute → ExecuteResult"
    invariant: "Every tool must have zod schema. Invalid input → validation error before execution."
    status: planned

control:
  - advantage_id: exa_web_search
    verification: "Integration test: search query → verify results returned with highlights"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_web_fetch
    verification: "Integration test: fetch URL → verify markdown content returned"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never

  - advantage_id: exa_tool_pattern
    verification: "Unit test: register tool without schema → verify rejected"
    update_trigger: "Re-analyze when exa-mcp-server releases new version"
    last_verified: never
```
