# BRANCH-MAP — Mimic Development Topology

Dynamic map of branches, their purpose, merge rules, and current status.

## Branch Hierarchy

```
main                        ← production-ready releases only
 └── dev                    ← integration branch: all feature work lands here first
      ├── feat/core-ops     ← C-core: OpPacket chain execution, OpCode dispatch
      ├── feat/core-bmap    ← C-core: bmap rewrite (.o → .c), slot/invariant/SI
      ├── feat/mcp-server   ← Go: MCP JSON-RPC server, tool registry, transport
      ├── feat/mcp-bridge   ← Go→C: CGo bridge, type marshaling, lifecycle
      ├── feat/orchestrator ← Go: pipeline stages, budget, guardrails, delegation
      ├── feat/graphify    ← Go: knowledge graph integration (graphify bridge)
      ├── feat/rtk-filter  ← Go: token compression integration (rtk bridge)
      ├── feat/config      ← Go: koanf config system, layered overlays
      ├── feat/observability← Go: OTel tracing, Prometheus metrics, health probes
      └── fix/*             ← hotfix branches, merge straight to dev
```

## Branch Rules

| Branch | Base | Merge to | CI Required | Protection |
|--------|------|----------|-------------|------------|
| `main` | — | — | `make check` + tag | Force-push denied, 1 review |
| `dev` | `main` | `main` (squash on release) | `make check` | Force-push denied |
| `feat/*` | `dev` | `dev` (squash) | `make check` | None |
| `fix/*` | `dev` | `dev` (squash) | `make check` | None |

### Merge Flow

```
feat/core-ops ──→ dev ──→ main (on release tag)
feat/mcp-server ──→ dev
fix/opcode-collision ──→ dev
```

- **Squash merge** into `dev` — clean history, one commit per feature increment
- **Squash merge** into `main` — only on release, tagged `v0.X.Y`
- **No direct push** to `main` or `dev` — all changes via PR
- **Feature branches** are disposable — delete after merge

## Per-Branch Scope

### `feat/core-ops`
- **Scope**: C-core opcode execution, OpPacket chain, OpCodeDef registry, dispatch
- **Files touched**: `core/ops.c`, `core/ops.h`, `core/dispatch.c`, `core/test/ops_test.c`
- **Blocked by**: None (foundational)
- **Blocks**: `feat/mcp-bridge` (needs stable C API)
- **Key invariants**: `ops_execute_chain` returns same result for same opcode input as embryo Go `BinaryRuntime`
- **Target**: 27+ libops.a symbols implemented in .c

### `feat/core-bmap`
- **Scope**: Rewrite libbmap.a from .o to .c — slot_index, inv_create/add/find_similar, si_create/insert/query
- **Files touched**: `core/bmap/*.c`, `core/bmap/*.h`, `core/test/bmap_test.c`
- **Blocked by**: None (can start in parallel with core-ops)
- **Blocks**: `feat/orchestrator` (needs slot/invariant API)
- **Key invariants**: 39 bmap symbols match embryo Go `mapstore`/`mesh` behavior

### `feat/mcp-server`
- **Scope**: Go MCP server — JSON-RPC handler, tool registry, stdio/HTTP transport, tool dispatch
- **Files touched**: `internal/mcp/*.go`, `cmd/mimic/main.go`
- **Blocked by**: None (can develop tool stubs without C-core)
- **Blocks**: Nothing (other branches integrate into it)
- **Key invariants**: MCP spec compliance — correct JSON-RPC 2.0 responses for tools/list, tools/call

### `feat/mcp-bridge`
- **Scope**: CGo bridge — type marshaling (Go↔C OpPacket), lifecycle (init/destroy), error mapping
- **Files touched**: `internal/cgo/*.go`, `core/bridge.h`
- **Blocked by**: `feat/core-ops` (needs stable C API signatures)
- **Blocks**: `feat/orchestrator` (needs Go callable C functions)
- **Key invariants**: Zero-copy where possible, <1μs per OpPacket crossing boundary

### `feat/orchestrator`
- **Scope**: 6-stage pipeline, iteration budget, tool guardrails, delegation, memory nudge, trajectory
- **Files touched**: `internal/orchestrator/*.go`
- **Blocked by**: `feat/mcp-bridge` (needs Go→C call path)
- **Blocks**: None (downstream integration branch)
- **Key invariants**: No EXEC without passed VALIDATE. Budget exceeded → stop.

### `feat/graphify`
- **Scope**: Knowledge graph — AST extraction → graph → IDF search → subgraph rendering
- **Files touched**: `internal/graphify/*.go`, `internal/mcp/tools_graph.go`
- **Blocked by**: `feat/mcp-server` (needs tool registration)
- **Blocks**: None
- **Key invariants**: IDF-weighted search returns same ranking as graphify Python `serve.py`

### `feat/rtk-filter`
- **Scope**: Token compression — TOML filter pipeline, language-aware stripping, smart truncation
- **Files touched**: `internal/rtk/*.go`, `internal/mcp/tools_rtk.go`
- **Blocked by**: `feat/mcp-server` (needs tool registration)
- **Blocks**: None
- **Key invariants**: 60-90% token reduction on supported commands, zero information loss on safety-critical output

### `feat/config`
- **Scope**: Koanf layered config — CoreConfig, MCPConfig, FeatureFlags, validation, strict unknown-key
- **Files touched**: `internal/config/*.go`
- **Blocked by**: None
- **Blocks**: All other feat branches (needs config types)
- **Key invariants**: Config loads in <100ms, strict validation rejects unknown keys

### `feat/observability`
- **Scope**: OTel tracing, Prometheus metrics, health probes, startup budgets, drain mode
- **Files touched**: `internal/observability/*.go`, `internal/app/health/service.go`
- **Blocked by**: `feat/config` (needs observability config types)
- **Blocks**: None
- **Key invariants**: Every C-core call spans a trace. Every tool call increments a counter.

## Development Cadence

```
Phase 1 (Foundation):   feat/config + feat/core-ops + feat/core-bmap
Phase 2 (Server):       feat/mcp-server + feat/mcp-bridge
Phase 3 (Intelligence): feat/orchestrator + feat/graphify + feat/rtk-filter
Phase 4 (Production):   feat/observability → dev → main
```

Phases can overlap. `feat/config` starts first (unblocks everything).
`feat/core-ops` and `feat/core-bmap` run in parallel.
`feat/mcp-server` starts as soon as tool stubs are defined.

## Status Tracking

| Branch | Created | Last Activity | Open PRs | Merged PRs | Status |
|--------|---------|----------------|----------|------------|--------|
| `main` | initial | — | 0 | 0 | protected |
| `dev` | initial | 2025-05-17 | 0 | 1 | active |
| `feat/core-ops` | — | — | 0 | 0 | pending |
| `feat/core-bmap` | — | — | 0 | 0 | pending |
| `feat/mcp-server` | — | — | 0 | 0 | pending |
| `feat/mcp-bridge` | — | — | 0 | 0 | pending |
| `feat/orchestrator` | — | — | 0 | 0 | pending |
| `feat/graphify` | — | — | 0 | 0 | pending |
| `feat/rtk-filter` | — | — | 0 | 0 | pending |
| `feat/config` | — | — | 0 | 0 | pending |
| `feat/observability` | — | — | 0 | 0 | pending |

## New Source Repository Integration

### rtk (Rust Token Killer) → `feat/rtk-filter`
- **Integration path**: Port the 8-stage TOML filter pipeline (`toml_filter.rs`) to Go as `internal/rtk/pipeline.go`
- **C-core port**: Language-aware code stripping (MinimalFilter/AggressiveFilter) to `core/filter.c`
- **MCP tools**: `rtk_run`, `rtk_read`, `rtk_gain`, `rtk_discover`
- **Key behavior**: Trust-on-first-use for user-defined filters (security model)
- **Spec card**: `specs/sources/rtk.md`

### graphify → `feat/graphify`
- **Integration path**: Port IDF-weighted search + BFS/DFS traversal to `internal/graphify/search.go`
- **C-core port**: Node/Edge structs, MinHash dedup, Union-Find to `core/graph.c`
- **MCP tools**: `query_graph`, `get_node`, `get_neighbors`, `god_nodes`, `graph_stats`, `shortest_path`
- **Key behavior**: Three-tier precedence (exact > prefix > substring), hub throttling, hot-reload
- **Spec card**: `specs/sources/graphify.md`

### go-service-template-rest → `feat/config` + `feat/observability` + `feat/mcp-server`
- **Integration path**: Clone template → rebrand → extend with C-core config + MCP transport
- **Direct reuse**: Bootstrap lifecycle, config system (koanf), health probes, middleware stack, RFC 7807 errors, OTel setup, Dockerfile, Makefile targets
- **Replace**: OpenAPI REST endpoints → MCP JSON-RPC tool definitions
- **Key behavior**: Orchestrator-first workflow, phased startup with budgets, drain mode
- **Spec card**: `specs/sources/go-service-template-rest.md`
