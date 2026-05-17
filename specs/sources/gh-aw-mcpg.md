```yaml
repo: Mayveskii/gh-aw-mcpg
url: https://github.com/Mayveskii/gh-aw-mcpg
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of gh-aw-mcpg. MCP Gateway implementing spec 1.9.0 with routed/unified/proxy
  modes, DIFC security, WASM guards, circuit breaker, and OAuth 2.0 PKCE.

advantages:
  - id: mcpg_routed_transport
    what: Routed MCP transport: /mcp/{serverID} — route to specific backend
    evidence: "src/ — routed mode implementation with server ID routing"

  - id: mcpg_unified_transport
    what: Unified MCP transport: /mcp — single endpoint for all backends
    evidence: "src/ — unified mode implementation"

  - id: mcpg_difc_security
    what: DIFC 6-phase security pipeline: label agent → label resource → check → execute → label response → filter
    evidence: "src/guards/ — DIFC implementation with 6 phases"

  - id: mcpg_circuit_breaker
    what: Circuit breaker per backend — prevent cascade failure when one backend is down
    evidence: "src/ — circuit breaker implementation"

  - id: mcpg_oauth_pkce
    what: OAuth 2.0 with PKCE for MCP client authorization
    evidence: "src/auth/ — OAuth PKCE implementation"

  - id: mcpg_wasm_guards
    what: WASM guards for sandboxed access control policies
    evidence: "src/guards/ — WASM guard implementation"

applications:
  - advantage_id: mcpg_routed_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "HTTP handler: /mcp/{serverID} routes to registered backend's JSON-RPC"
    invariant: "Unknown serverID → 404. Route table matches behavior-sources.yaml entries."
    status: planned

  - advantage_id: mcpg_unified_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "HTTP handler: /mcp dispatches to all backends, aggregates responses"
    invariant: "Unified mode returns results from all healthy backends."
    status: planned

  - advantage_id: mcpg_difc_security
    implemented_in: internal/orchestrator/security.go
    mechanism: "6-phase DIFC pipeline before every execution: label→check→execute→label→filter"
    invariant: "Information flows only from ≥ clearance to ≤ clearance."
    status: planned

  - advantage_id: mcpg_circuit_breaker
    implemented_in: internal/mcp/circuit_breaker.go
    mechanism: "Per-backend circuit breaker: 3 failures → open → half-open after 30s → close if success"
    invariant: "Circuit open → requests fail fast, no cascade."
    status: planned

  - advantage_id: mcpg_oauth_pkce
    implemented_in: internal/mcp/auth.go
    mechanism: "OAuth 2.0 PKCE flow for MCP client authentication"
    invariant: "No unauthenticated MCP tool calls in production mode."
    status: planned

  - advantage_id: mcpg_wasm_guards
    implemented_in: internal/orchestrator/security.go
    mechanism: "WASM sandbox for policy evaluation — isolated, no host access"
    invariant: "Policy evaluation in WASM cannot access host filesystem or network."
    status: future

control:
  - advantage_id: mcpg_routed_transport
    verification: "Integration test: route to known serverID → verify correct backend responds"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_unified_transport
    verification: "Integration test: unified request → verify all healthy backends respond"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_difc_security
    verification: "Unit test: low-clearance agent → verify blocked from high-classification resource"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_circuit_breaker
    verification: "Unit test: 3 failures → verify circuit opens, wait → verify half-open → verify close on success"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_oauth_pkce
    verification: "Integration test: PKCE flow → verify token obtained → verify authenticated call succeeds"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_wasm_guards
    verification: "Unit test: WASM policy denies access → verify blocked; allows → verify passes"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never
```
