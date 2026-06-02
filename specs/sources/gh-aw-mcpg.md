```yaml
repo: Mayveskii/gh-aw-mcpg
url: https://github.com/Mayveskii/gh-aw-mcpg
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of gh-aw-mcpg. MCP Gateway implementing spec 1.9.0 with routed/unified/proxy
  modes, DIFC 6-phase security pipeline, WASM guards for sandboxed policy evaluation,
  per-backend circuit breaker, and OAuth 2.0 PKCE for client authorization.

advantages:
  - id: mcpg_routed_transport
    what: Routed MCP transport: HTTP /mcp/{serverID} → JSON-RPC dispatch to specific registered backend; serverID lookup in route table
    evidence: "src/ — routed mode: /mcp/:serverId endpoint with serverId → backend mapping from config"

  - id: mcpg_unified_transport
    what: Unified MCP transport: HTTP /mcp → fan-out to all healthy backends → aggregate responses; single endpoint for multi-backend queries
    evidence: "src/ — unified mode: /mcp endpoint dispatches to all backends, aggregates JSON-RPC responses"

  - id: mcpg_difc_security
    what: DIFC (Decentralized Information Flow Control) 6-phase security pipeline: label_agent → label_resource → check_clearance → execute → label_response → filter_output; enforces information flow constraints
    evidence: "src/guards/ — DIFC implementation with 6 phases; clearance level comparison before execution; response filtering after execution"

  - id: mcpg_circuit_breaker
    what: Per-backend circuit breaker with 3 states (closed→open→half-open): consecutive failures trigger open state, half-open probes after timeout, success closes circuit
    evidence: "src/ — CircuitBreaker class with failure counter, state transitions, half-open probe logic"

  - id: mcpg_oauth_pkce
    what: OAuth 2.0 with PKCE (Proof Key for Code Exchange): code_verifier + code_challenge S256; prevents authorization code interception
    evidence: "src/auth/ — OAuth PKCE implementation: generate code_verifier, compute code_challenge = S256(verifier), token exchange with code_verifier"

  - id: mcpg_wasm_guards
    what: WASM guards for sandboxed access control policy evaluation: policies compiled to WASM → executed in sandboxed runtime → no host filesystem/network access
    evidence: "src/guards/ — WASM guard implementation with sandboxed WASM runtime; policy evaluation isolated from host"

applications:
  - advantage_id: mcpg_routed_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "HTTP handler: /mcp/{serverID} → lookup serverID in route table → forward JSON-RPC to backend → return response; unknown serverID → 404"
    invariant: "Unknown serverID → 404 with diagnostic message. Route table matches behavior-sources.yaml entries. No wildcard routing."
    status: planned

  - advantage_id: mcpg_unified_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "HTTP handler: /mcp → fan-out JSON-RPC to all registered healthy backends → wait for responses (or timeout) → aggregate into single response"
    invariant: "Unified mode returns results from all healthy backends. Unhealthy backends skipped (circuit open). Per-backend timeout 30s."
    status: planned

  - advantage_id: mcpg_difc_security
    implemented_in: internal/orchestrator/security.go
    mechanism: "6-phase DIFC: (1)label_agent with clearance level (2)label_resource with classification level (3)check clearance≥classification (4)execute if allowed (5)label_response with output classification (6)filter_output by agent clearance"
    invariant: "Information flows only from ≥ clearance to ≤ clearance. Phase 3 denial = operation never executed. No bypass of clearance check."
    status: planned

  - advantage_id: mcpg_circuit_breaker
    implemented_in: internal/mcp/circuit_breaker.go
    mechanism: "Per-backend state machine: closed(failure_count++ on error, ≥3→open) → open(fail fast, after 30s→half-open) → half-open(probe with 1 request, success→closed, failure→open)"
    invariant: "Circuit open → requests fail fast, no backend call, no cascade. Half-open allows exactly 1 probe. State transitions logged."
    status: planned

  - advantage_id: mcpg_oauth_pkce
    implemented_in: internal/mcp/auth.go
    mechanism: "OAuth 2.0 PKCE: generate code_verifier (32 random bytes) → code_challenge = BASE64URL(SHA256(verifier)) → auth URL with challenge → token exchange with verifier"
    invariant: "No unauthenticated MCP tool calls in production mode. code_verifier never transmitted in auth URL. S256 mandatory, not plain."
    status: planned

  - advantage_id: mcpg_wasm_guards
    implemented_in: internal/orchestrator/security.go
    mechanism: "Policy compiled to WASM → instantiate in sandboxed runtime → evaluate policy(input_context) → allow/deny → no host access during evaluation"
    invariant: "Policy evaluation in WASM cannot access host filesystem, network, or environment. Evaluation timeout 100ms. Deny on timeout."
    status: future

control:
  - advantage_id: mcpg_routed_transport
    verification: "Integration test: route to known serverID → verify correct backend responds; unknown serverID → verify 404"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_unified_transport
    verification: "Integration test: unified request with 3 healthy + 1 open-circuit backends → verify 3 responses aggregated; verify circuit-open backend skipped"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_difc_security
    verification: "Unit test: low-clearance agent → verify blocked from high-classification resource; matching clearance → verify allowed; response filtering → verify classified fields removed"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_circuit_breaker
    verification: "Unit test: 3 consecutive failures → verify circuit opens; wait 30s → verify half-open; success in half-open → verify closed; failure in half-open → verify re-opened"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_oauth_pkce
    verification: "Integration test: PKCE flow → verify code_challenge = S256(verifier); token exchange → verify authenticated call succeeds; missing verifier → verify rejection"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never

  - advantage_id: mcpg_wasm_guards
    verification: "Unit test: WASM policy denies access → verify blocked; allows → verify passes; WASM attempts filesystem access → verify sandbox prevents"
    update_trigger: "Re-analyze when gh-aw-mcpg releases new version"
    last_verified: never
```
