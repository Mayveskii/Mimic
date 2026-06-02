```yaml
repo: Mayveskii/go-service-template-rest
url: https://github.com/Mayveskii/go-service-template-rest
language: Go
status: partial
last_sync: "2025-05-17"

description: |
  Fork of msitarzewski/go-service-template-rest (98.7K stars). AI-native Go REST template with
  orchestrator-first workflow, OpenAPI-first codegen, koanf layered config, pgx/sqlc persistence,
  chi HTTP routing, OTel/Prometheus observability, testcontainers integration testing, 30+
  canonical agent skills, and distroless Docker deployment.

advantages:
  - id: gst_strict_interface
    what: OpenAPI-first → oapi-codegen generates StrictServerInterface with typed request/response objects per endpoint — maps directly to MCP tool interface
    evidence: "internal/api/openapi.gen.go — StrictServerInterface, PingRequestObject, PingResponseObject; api/openapi/service.yaml"

  - id: gst_probe_health
    what: Probe interface (Name() + Check(ctx)) for health/readiness — C-core subsystems implement Probe, Go orchestrator checks readiness without C internals
    evidence: "internal/app/health/service.go — Probe interface, Service.Check() iterates probes"

  - id: gst_koanf_config
    what: Config struct with koanf tags → defaults → YAML → env overrides + strict unknown-key validation + per-stage startup budgets
    evidence: "internal/config/types.go, config.go, defaults.go, validate.go, context_budget.go"

  - id: gst_middleware_stack
    what: RequestCorrelation → OTel → SecurityHeaders → AccessLog → RequestBodyLimit → RequestFramingGuard → Recover — directly reusable for MCP JSON-RPC handler
    evidence: "internal/infra/http/middleware.go, router.go"

  - id: gst_rfc7807_errors
    what: Problem Details JSON (type/title/status/detail/request_id) — structured errors map to MCP JSON-RPC error responses
    evidence: "internal/infra/http/problem.go — Problem struct, writeProblem()"

  - id: gst_repo_pattern
    what: Querier interface + sqlc-generated code + adapter-safe domain records — pattern for C-core result persistence
    evidence: "internal/infra/postgres/ping_history_repository.go — pingHistoryQuerier, mapPingHistoryRecord()"

  - id: gst_bootstrap_lifecycle
    what: Phased startup with budgets (config=10s, probe=15s, telemetry=2s, total=30s) + signal handling + graceful shutdown with drain
    evidence: "cmd/service/internal/bootstrap/run.go, shutdown.go, startup_*.go"

  - id: gst_drain_mode
    what: atomic.Bool draining flag → /health/ready returns 503 → LB stops traffic → existing requests drain
    evidence: "internal/app/health/service.go — StartDrain(), draining atomic.Bool"

  - id: gst_network_policy
    what: Startup-time egress allow/deny lists — control which external services process can contact
    evidence: "cmd/service/internal/bootstrap/network_policy.go, network_policy_enforcement.go"

  - id: gst_multi_agent_skills
    what: 30+ canonical skills + 16 subagent TOML definitions — AI-native development workflow for building Mimic itself
    evidence: ".agents/skills/, .codex/agents/, .claude/agents/"

applications:
  - advantage_id: gst_strict_interface
    implemented_in: internal/mcp/tools_gen.go
    mechanism: "Define MCP tools in schema → generate StrictToolInterface with typed request/response per tool → implement each tool method"
    invariant: "Every MCP tool has typed input/output. No untyped map[string]any in tool implementations."
    status: planned

  - advantage_id: gst_probe_health
    implemented_in: internal/app/health/service.go
    mechanism: "Implement Probe for C-core (calls core_ping via cgo), Postgres, Redis → register with health service → /health/ready checks all"
    invariant: "All critical subsystems registered as probes. Unhealthy probe → 503 response."
    status: planned

  - advantage_id: gst_koanf_config
    implemented_in: internal/config/*.go
    mechanism: "Add CoreConfig + MCPConfig to main Config struct → koanf loads defaults→YAML→env → strict validation rejects unknowns"
    invariant: "Config loads in <100ms. Unknown keys always rejected. CoreConfig validated before C-core init."
    status: planned

  - advantage_id: gst_middleware_stack
    implemented_in: internal/mcp/middleware.go
    mechanism: "Adapt middleware stack for JSON-RPC: RequestCorrelation→OTel→BodyLimit→Recover→AccessLog wraps MCP handler"
    invariant: "Every JSON-RPC request gets correlation ID + trace span. Panic never crashes server."
    status: planned

  - advantage_id: gst_rfc7807_errors
    implemented_in: internal/mcp/errors.go
    mechanism: "Map Problem struct to JSON-RPC error response: status→code, title→message, detail→data, request_id→correlation"
    invariant: "All tool errors use structured Problem format. No bare string errors in MCP responses."
    status: planned

  - advantage_id: gst_repo_pattern
    implemented_in: internal/infra/postgres/
    mechanism: "Define SQL queries for matrix snapshots, computation logs, topology history → generate sqlc → implement Querier"
    invariant: "All persistence through Querier interface. SQL in .sql files, Go code generated, never hand-written."
    status: planned

  - advantage_id: gst_bootstrap_lifecycle
    implemented_in: cmd/mimic/internal/bootstrap/
    mechanism: "Copy bootstrap package → add C-core init stage → extend total budget → keep signal handling + graceful shutdown"
    invariant: "Startup completes in <30s or fails with actionable error. C-core initialized before MCP server binds."
    status: planned

  - advantage_id: gst_drain_mode
    implemented_in: internal/app/health/service.go
    mechanism: "SIGTERM → StartDrain() → health 503 → MCP server stops accepting new requests → existing tool calls complete → exit"
    invariant: "No in-flight tool call is interrupted by shutdown. Drain completes within shutdown timeout."
    status: planned

  - advantage_id: gst_network_policy
    implemented_in: cmd/mimic/internal/bootstrap/network_policy.go
    mechanism: "Load allowed egress hosts from config → enforce at startup → C-core outbound calls go through Go network layer"
    invariant: "C-core cannot contact any host not in allowed list. Policy enforced before MCP server binds."
    status: planned

  - advantage_id: gst_multi_agent_skills
    implemented_in: .agents/skills/
    mechanism: "Use go-coder, go-architect-spec, go-domain-invariant-spec, etc. for developing Mimic itself"
    invariant: "Skills used for development, not runtime. Not shipped in Mimic binary."
    status: reference

control:
  - advantage_id: gst_strict_interface
    verification: "Unit test: generate interface → implement stub → verify compiles and returns typed response"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_probe_health
    verification: "Integration test: C-core not loaded → /health/ready returns 503 with 'core' in failed probes"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_koanf_config
    verification: "Unit test: set env MIMIC_CORE_ENABLED=false → verify CoreConfig.Enabled is false"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_middleware_stack
    verification: "Integration test: send malformed JSON-RPC → verify correlation ID in response + recovery log"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_rfc7807_errors
    verification: "Unit test: return Problem{status:400} → verify JSON-RPC response with code=-32600 and detail field"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_repo_pattern
    verification: "Integration test: create matrix snapshot → list recent → verify record present with correct fields"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_bootstrap_lifecycle
    verification: "Integration test: start Mimic → verify /health/ready 200 within 30s"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_drain_mode
    verification: "Integration test: send SIGTERM during tool call → verify call completes, new calls rejected"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_network_policy
    verification: "Unit test: allow api.openai.com, deny 169.254.169.254 → verify enforcement"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never

  - advantage_id: gst_multi_agent_skills
    verification: "Manual review: skills exist and are syntactically valid TOML/MD"
    update_trigger: "Re-analyze when template releases new version"
    last_verified: never
```
