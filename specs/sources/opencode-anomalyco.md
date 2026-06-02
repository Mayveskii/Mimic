```yaml
repo: Mayveskii/opencode-anomalyco-
url: https://github.com/Mayveskii/opencode-anomalyco-
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of opencode (anomalyco). AI coding assistant with 17 built-in tools,
  MCP client support (StreamableHTTP/SSE/Stdio), tool loop, permission system
  with per-tool allow/deny, codesearch via Exa, output truncation with temp file
  overflow, and structured tool registry with zod schema validation.

advantages:
  - id: oc_tool_registry
    what: Tool.Def interface: id, description, parameters (zod schema), execute → ExecuteResult{content, isError}; 17 built-in tools + MCP tools
    evidence: "packages/opencode/src/tool/ — Tool.Def type definition; 17 tool implementations (bash, edit, read, glob, grep, webfetch, etc.)"

  - id: oc_tool_loop
    what: while(true) generator loop: API call → stream response → parse tool_use blocks → execute tools → check continuation condition → repeat or stop
    evidence: "packages/opencode/src/ — tool loop in query.ts: streamResponse → parseToolCalls → executeTools → check shouldContinue"

  - id: oc_mcp_transport
    what: MCP transport: StreamableHTTP (POST with SSE response) + SSE (server-sent events) + Stdio (stdin/stdout); auto-detect from URL or config
    evidence: "packages/opencode/src/mcp/ — transport implementations: streamable_http.ts, sse.ts, stdio.ts"

  - id: oc_codesearch_tool
    what: Code search via Exa MCP endpoint: semantic code search across repositories with query → ranked code results with relevance scores
    evidence: "packages/opencode/src/tool/ — codesearch tool using Exa API for semantic code search"

  - id: oc_output_truncation
    what: Output truncation: 2000 lines / 50KB hard limit; overflow written to temp file (/tmp/opencode); agent reads overflow with offset/limit parameters
    evidence: "packages/opencode/src/tool/ — truncation logic: if output > 50KB → write to /tmp/opencode, return first 2000 lines + temp file path"

  - id: oc_permission_system
    what: Per-tool permission system: allow/deny rules per tool ID; dangerous operations require explicit allow; session-scoped permission state
    evidence: "packages/opencode/src/ — permission system with per-tool allow/deny lists; session permission state tracking"

applications:
  - advantage_id: oc_tool_registry
    implemented_in: internal/tool/registry.go
    mechanism: "Go ToolDef struct: ID string, Description string, Parameters (JSON Schema), Execute func(ctx, params) → ExecuteResult; register() adds to tool map; lookup by ID"
    invariant: "Every tool must have ID, Description, and parameter schema. No anonymous tools. Duplicate ID → error at registration."
    status: planned

  - advantage_id: oc_tool_loop
    implemented_in: internal/orchestrator/loop.go
    mechanism: "while(has_pending_tools || continue_flag) → API call → stream response → parse tool_use → validate → execute → collect results → append to conversation → check continuation"
    invariant: "Loop terminates when: no more tools to call, budget exhausted, or maxTurns reached. Every tool result appended to conversation before next iteration."
    status: planned

  - advantage_id: oc_mcp_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "Three transport modes: stdio (stdin/stdout JSON-RPC), SSE (HTTP GET event stream + POST), StreamableHTTP (HTTP POST with SSE response); config selects mode"
    invariant: "At least one transport must be active per MCP server. stdio is default. Transport auto-detected from URL if not specified."
    status: planned

  - advantage_id: oc_codesearch_tool
    implemented_in: internal/tool/tools_codesearch.go
    mechanism: "MCP tool: codesearch(query, language, limit) → Exa API POST /search with code filter → parse results → return ranked code snippets with file/line info"
    invariant: "Search results limited to limit (default 10). Timeout 30s. Results include file path, line range, and relevance score."
    status: planned

  - advantage_id: oc_output_truncation
    implemented_in: internal/tool/truncation.go
    mechanism: "After tool execution: if output > 50KB or > 2000 lines → truncate, write full output to /tmp/opencode/{session_id}/{tool_id}.txt → return truncated output + temp file path"
    invariant: "No tool output exceeds 50KB in direct response. Overflow available via temp file path in response metadata. Temp files cleaned on session end."
    status: planned

  - advantage_id: oc_permission_system
    implemented_in: internal/orchestrator/permission.go
    mechanism: "Per-tool permission map: toolID → {allow, deny, ask}; before execution: check deny list → ask list → allow list; session tracks cumulative permission decisions"
    invariant: "Denied tool never executes. Ask requires user confirmation per session. Allow persists for session duration. No permission persists across sessions."
    status: planned

control:
  - advantage_id: oc_tool_registry
    verification: "Unit test: register tool without ID → rejected; without schema → rejected; duplicate ID → error; valid tool → registered"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_tool_loop
    verification: "Integration test: 5-turn loop with tools → verify terminates when no more tools needed; budget=1 → verify terminates after 1 iteration"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_mcp_transport
    verification: "Integration test: connect via stdio → tool call → verify response; connect via StreamableHTTP → verify; invalid transport config → verify error"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_codesearch_tool
    verification: "Integration test: search 'react hooks' with language=typescript → verify code results with file paths; empty query → verify validation error"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_output_truncation
    verification: "Unit test: generate 100KB output → verify truncated to 50KB + temp file created; verify temp file contains full output; verify temp file path in response metadata"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_permission_system
    verification: "Unit test: tool in deny list → verify blocked; tool in ask list → verify confirmation required; tool in allow list → verify auto-approved"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never
```
