```yaml
repo: Mayveskii/opencode-anomalyco-
url: https://github.com/Mayveskii/opencode-anomalyco-
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of opencode (anomalyco). AI coding assistant with 17 built-in tools,
  MCP client support (StreamableHTTP/SSE/Stdio), tool loop, and permission system.

advantages:
  - id: oc_tool_registry
    what: Tool.Def interface: id, description, parameters (zod), execute → ExecuteResult
    evidence: "packages/opencode/src/tool/ — Tool.Def type, 17 tool implementations"

  - id: oc_tool_loop
    what: while(true) generator loop: API call → stream response → execute tools → check continuation
    evidence: "packages/opencode/src/ — tool loop implementation in query.ts"

  - id: oc_mcp_transport
    what: MCP transport: StreamableHTTP + SSE + Stdio — three transport modes
    evidence: "packages/opencode/src/mcp/ — transport implementations"

  - id: oc_codesearch
    what: Code search via Exa MCP endpoint — semantic code search across repositories
    evidence: "packages/opencode/src/tool/ — codesearch tool"

  - id: oc_truncation
    what: Output truncation: 2000 lines / 50KB, overflow → temp file for offset/limit reading
    evidence: "packages/opencode/src/tool/ — truncation logic in tool output"

applications:
  - advantage_id: oc_tool_registry
    implemented_in: internal/tool/registry.go
    mechanism: "Go ToolDef struct: ID, Description, Parameters (schema), Execute func → ExecuteResult"
    invariant: "Every tool must have ID, Description, and parameter schema. No anonymous tools."
    status: planned

  - advantage_id: oc_tool_loop
    implemented_in: internal/orchestrator/loop.go
    mechanism: "while(has_pending_tools) → validate → execute → collect results → check continuation"
    invariant: "Loop terminates when: no more tools to call, budget exhausted, or maxTurns reached."
    status: planned

  - advantage_id: oc_mcp_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "Three transport modes: stdio (local), SSE (remote), StreamableHTTP (remote)"
    invariant: "At least one transport must be active. stdio is default."
    status: planned

  - advantage_id: oc_codesearch
    implemented_in: internal/tool/tools_codesearch.go
    mechanism: "MCP tool: codesearch(query) → Exa API → code results → agent"
    invariant: "Search results limited, timeout 30s."
    status: planned

  - advantage_id: oc_truncation
    implemented_in: internal/tool/truncation.go
    mechanism: "Output truncated at 2000 lines / 50KB, overflow written to temp file, agent reads with offset/limit"
    invariant: "No tool output exceeds 50KB in direct response. Overflow available via temp file."
    status: planned

control:
  - advantage_id: oc_tool_registry
    verification: "Unit test: register tool without ID → rejected; without schema → rejected"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_tool_loop
    verification: "Integration test: 5-turn loop → verify terminates when no more tools needed"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_mcp_transport
    verification: "Integration test: connect via stdio → tool call → verify response; connect via HTTP → verify"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_codesearch
    verification: "Integration test: search query → verify code results returned"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never

  - advantage_id: oc_truncation
    verification: "Unit test: generate 100KB output → verify truncated to 50KB + temp file available"
    update_trigger: "Re-analyze when opencode releases new version"
    last_verified: never
```
