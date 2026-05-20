# AGENTS.md — Mimic

## What This Is

Mimic is a standalone MCP server with a C-core. It is an **optional tool** that any AI-agent can call — the same way it calls bash, grep, or any other MCP tool. The agent is fully autonomous and works without Mimic. When the agent chooses to call Mimic, it gets help packaging an OpPacket chain — partially or fully — backed by distilled patterns from production code and behaviors borrowed from proven implementations.

## Rules

1. No action without user permission
2. Facts only — nothing invented
3. Semantics before code — no function without a row in SEMANTICS.md
4. No file edits without permission
5. Every non-trivial decision = ADR

## Project Structure

```
Mimic/
├── AGENTS.md              ← YOU ARE HERE. Entry point.
├── BRANCH-MAP.md          ← Branch topology, merge rules, per-branch scope
├── specs/                 ← All documentation (numbered, ordered)
│   ├── 00-SPEC-INDEX.md   ← Map of all docs + spec card schema
│   ├── ...through 08-MODULES.md
│   └── sources/*.md       ← 19 per-repo spec cards
├── mimicrya/
│   ├── behavior-sources.yaml  ← Mayveskii/* repos: which behaviors Mimic borrows
│   └── repos-manifest.yaml    ← Production repos: distillation status
├── docs/adr/              ← Architecture Decision Records (ADR-001..005)
├── docs/architecture/     ← Deployment, performance, autonomy roadmap
│   ├── DEPLOYMENT.md
│   ├── MESH_PERFORMANCE.md
│   ├── ROADMAP_AUTONOMY.md
│   └── ACTIONBYTES_SPEC.md
├── core/                  ← C-core (ops.c, ops.h, exec_*.c, bmap/)
│   ├── ops.h              ← 96 OpCodes
│   ├── ops.c
│   ├── mesh_op.c          ← OP_MESH_QUERY (0xC0)
│   └── pattern_op.c       ← OP_EXECUTE_PATTERN (0xD0)
├── internal/
│   ├── mcp/               ← MCP server (JSON-RPC, stdio/TCP)
│   │   ├── mcp.go         ← Server, fast path routing
│   │   ├── tool_schemas.go← 49 tool schemas with Group field
│   │   ├── mesh_handler.go← MESH_QUERY, EXECUTE_PATTERN, MESH_AUTO_APPLY
│   │   ├── projectmap_handler.go ← PROJECT_MAP_*, WORKSPACE_SYNTHESIZE
│   │   ├── plan_handler.go← PLAN_GENERATE (validates via C-core)
│   │   ├── exa_handler.go ← EXA_SEARCH web research
│   │   └── tcp.go         ← TCPTransport, ServeTCP
│   ├── tool/              ← MCP tool registry (legacy)
│   ├── cgo/               ← CGO bridge Go↔C
│   ├── embed/             ← TextEmbedding service client (int8 + float32)
│   ├── qdrant/            ← Qdrant REST client (vector search fallback)
│   ├── mesh/              ← Mesh registry, gob loader, ActionBytes decoder, text-native slots
│   │   ├── actionbytes.go ← !-delimited binary patch decoder
│   │   ├── text_slot.go   ← Markdown-native slots (ADR-005)
│   │   └── query.go       ← Hybrid search (qdrant primary → local fallback)
│   ├── projectmap/        ← SQLite+FTS5 indexer + workspace synthesizer
│   │   ├── projectmap.go  ← DB schema, WAL mode, symbol search
│   │   ├── scanner.go     ← Go regex parser (funcs, types, imports)
│   │   └── synthesize.go  ← Workspace → .mimic/workspace.graph.gob
│   ├── orchestrator/      ← Workflow state machine
│   │   └── plan.go        ← Plan generation, ValidatePlan, conflict matrix
│   ├── session/           ← Agent sessions
│   │   └── logger.go      ← JSONL session logging, pattern extraction
│   └── quality/           ← 2-vote verify, denial tracking
├── cmd/mimic/main.go              ← Entrypoint (serve, serve --tcp :1337)
├── cmd/migrate-gob-to-text/main.go ← ADR-005: gob → text-native migration tool
├── data/
│   ├── extraction/        ← Distillation scripts
│   ├── seeds/             ← Initial slots
│   ├── mesh/
│   │   ├── graphs/        ← 18 domain gob files
│   │   └── registry/      ← invariant_registry.json (curated text overlay)
│   └── matrices/          ← Conflict/energy matrices
├── test/
│   ├── battlefield/
│   │   ├── mesh_injection_benchmark.py   ← 4/4 PASS
│   │   ├── gonkagate_e2e_test.py         ← 6/6 PASS
│   │   ├── mimic_heavy_e2e_test.py       ← K8s, refactor, data flow, migration
│   │   ├── gonkagate_limits_test.py      ← Qwen 200K token stress test
│   │   └── three_tier_benchmark.py       ← (requires OpenRouter)
│   └── integration/
├── Makefile               ← build, test, lint, check, distill, release
├── .github/workflows/     ← ci.yml, release.yml, distill.yml
├── go.mod
├── Dockerfile
└── .workflows/
    └── mimic.service      ← systemd unit template
```

## Reading the Project (for any agent)

1. Read specs/00-SPEC-INDEX.md — map of all documents, reading order
2. Read specs/01-AGENTS.md — rules for agents working on Mimic
3. Read specs/02-ARCHITECTURE.md — components, flows, boundaries
4. Read specs/03-EXECUTION-SPACE.md — what agents can do, task types, dimensions
5. Read specs/04-SCENARIOS.md — execution patterns with chains, invariants, costs
6. Read specs/05-BEHAVIOR.md — formulas, invariants, phase transitions
7. Read specs/06-SEMANTICS.md — every function: name | input | output | invariant | source
8. Read specs/07-RESOURCES.md — complete resource map, OpPacket translation
9. Read specs/08-MODULES.md — per-module documentation, connections, state
10. Read specs/sources/*.md — per-repo spec cards: advantages → applications → control
11. Read docs/architecture/DEPLOYMENT.md — production deployment guide (Docker Compose, systemd, bootstrap)
12. Read docs/architecture/MESH_PERFORMANCE.md — scaling plan, bottlenecks, HNSW migration
13. Read docs/architecture/ROADMAP_AUTONOMY.md — from passive tool to autonomous agent (4 stages)

## Two Sources of Knowledge

### Distillation (mimicrya/repos-manifest.yaml)
Production repos (etcd, k8s, go-ethereum and 90+ others) → git blame → survival index → best commits → mesh slots. These are proven patterns that survived in real systems.

### Mimicry (mimicrya/behavior-sources.yaml)
Mayveskii/* repos (hermes-agent, gastown, bun, rtk, graphify, go-service-template-rest, exa-mcp-server, gh-aw-mcpg, code-mode, opencode, etc.) → behavior selection: HOW to implement a function in Mimic. Hermes-agent showed closed learning loop → Mimic implements curator + skill nudge. Gastown showed multi-agent watchdog → Mimic implements 3-tier health. Rtk showed TOML filter pipeline → Mimic implements token compression. Graphify showed IDF-weighted graph search → Mimic implements knowledge graph. Go-service-template-rest showed bootstrap lifecycle → Mimic implements phased startup. This is not copying — it's selecting the best behavior to implement.

## Commands

```bash
make              # Build libcore.a + mimic binary
make test         # Run all tests
make lint         # Check code
make check        # lint + test + semantics-check
make distill      # Run distillation
make release      # Build binaries + checksums
make docker       # Build docker image
```

## Branches

- `main` — stable, only via PR + CI green + review
- `dev` — integration, feature branches from here
- `feat/core-ops` — C-core OpPacket execution
- `feat/core-bmap` — bmap rewrite (.o → .c)
- `feat/mcp-server` — MCP JSON-RPC server
- `feat/mcp-bridge` — CGo bridge Go↔C
- `feat/orchestrator` — pipeline, budget, guardrails
- `feat/graphify` — knowledge graph integration
- `feat/rtk-filter` — token compression
- `feat/config` — koanf layered config
- `feat/observability` — OTel, Prometheus, health
- `fix/*` — hotfixes → dev

Full topology: `BRANCH-MAP.md`

## Releases

- GitHub Release: binaries + SHA256 checksums
- Docker Hub: same binary in container, immutable tags
- Update: new tag → new container → rolling recreate
- Security: image verification on update

## Key Concepts

- **OpPacket chain** — ordered sequence of deterministic operations, validated BEFORE execution
- **Conflict matrix** — [N×N] matrix defining which operations cannot run together
- **Energy cost matrix** — [N×3] matrix: cost_tokens, cost_time_us, cost_memory_bytes
- **Z-density** — density of proven knowledge in a mesh slot
- **Survival index** — fraction of commit lines still present at HEAD (git blame → metric)
- **Mimicry** — selecting behavior from best implementations in Mayveskii/* repos
- **Distillation** — git blame → survival → extract → bmap slot → Z-density
- **Behavior source** — a repo from which Mimic borrows behavior (approach, not code)
- **Embryo** — Mayveskii/embryo: the repo where Mimic was born. Go implementations (pkg/do/, pkg/mesh/, pkg/hunt/, cmd/mcp/) informed the C-core and internal/ design.
- **mimic-server** — FUTURE: shared knowledge hub for multiple clients. Not part of current scope.
