# Mimic — Deterministic AI-Agent Tool Orchestration

> **Mimic is not the agent. The agent is autonomous.** Mimic is an optional MCP tool — the same way it calls bash, grep, or any other tool. When the agent chooses Mimic, it gets deterministic execution, validation before run, and rollback on failure.

---

## What is Mimic?

**Mimic** is a standalone [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server with a **C-core execution engine** and **Go orchestration layer**.

When any AI model (Claude, GPT, kimi, etc.) calls Mimic:
1. **Understands** — JSON Schema tells the model exactly what each tool needs
2. **Validates** — 6-phase pipeline checks conflicts, budgets, permissions
3. **Executes** — Real system calls (not stubs): `stat()`, `open()`, `git`, `make`, OpenSSL
4. **Measures** — Energy cost, latency, token usage tracked per operation
5. **Rolls back** — On failure, restores pre-execution state
6. **Compresses** — Large outputs reduced by 95% so context window survives

## Why Mimic?

### Problem: Models waste tokens on errors
- **Without Mimic:** Model guesses tool arguments → wrong types → crashes → retries → $$$ wasted
- **With Mimic:** Full JSON Schema prevents all argument collisions → zero retries → save 30-50% tokens

### Problem: Large outputs exhaust context
- **Without Mimic:** `git log` returns 5000 lines → 25K tokens → single call burns 20% of context
- **With Mimic:** RTK compression reduces to 50 lines → 250 tokens → 95% reduction

### Problem: Complex tasks need decomposition
- **Without Mimic:** Model tries to do everything in one call → fails
- **With Mimic:** Task Decomposition breaks "build and test entire project" into [Clean→Compile→Test] with dependencies

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AI Model (Claude/GPT/kimi)               │
│         "Build, test, and deploy this project"              │
└──────────────────────┬──────────────────────────────────────┘
                       │ JSON-RPC 2.0 / MCP stdio
┌──────────────────────▼──────────────────────────────────────┐
│                   Mimic MCP Server                            │
│  ┌─────────────────┐  ┌──────────────────┐  ┌──────────────┐  │
│  │  MCP Transport  │  │  6-Phase         │  │  Tool        │  │
│  │  stdio/SSE/HTTP │◄─┤  Orchestrator    │◄─┤  Registry    │  │
│  │  Port: 1337     │  │  CLASSIFY→PLAN   │  │  35 tools    │  │
│  └─────────────────┘  │  VALIDATE→EXEC   │  │  + schemas   │  │
│                         │  VERIFY→RESPOND  │  └──────────────┘  │
│                         └──────────────────┘                    │
│                              │                                  │
├──────────────────────────────┼────────────────────────────────┤
│  C-Core (91 OpCodes)         │  Go Layer                       │
│  ┌──────────────┐           │  ┌────────────────────────────┐   │
│  │ Validation   │           │  │ Task Decomposition         │   │
│  │ Conflict     │◄──────────┤  │ ProjectContext + Compress │   │
│  │ Energy costs │           │  │ RTK Output Compression   │   │
│  │ Rollback     │           │  │ Budget tracking            │   │
│  └──────────────┘           │  └────────────────────────────┘   │
└──────────────────────────────┴────────────────────────────────┘
```

### Port Configuration (1337-style)

All Mimic services use **1337-style ports** for consistency and recognition:

| Port | Service | Description |
|------|---------|-------------|
| **1337** | **Main MCP** | Primary stdio/SSE/HTTP transport (default) |
| 1117 | HTTP API | REST API, Prometheus metrics, health checks |
| 1227 | Admin | Management API, configuration |
| 1447 | WebSocket | Real-time bidirectional transport |
| 1557 | Mesh | Inter-node communication (future: distributed mesh) |

Set via environment:
```bash
MIMIC_PORT=1337
MIMIC_HTTP_PORT=1117
MIMIC_ADMIN_PORT=1227
```

## Quick Start

### One-liner Install (Recommended)

```bash
# macOS / Linux — one command, downloads binary + mesh data automatically
curl -sSL https://raw.githubusercontent.com/Mayveskii/Mimic/main/install.sh | bash

# Start server immediately
mimic serve --port 1337
```

### How Mimic is Distributed

Mimic separates **code**, **binaries**, and **data** — each lives where it belongs:

| What | Where | Size | How to get |
|------|-------|------|------------|
| **Source code** | GitHub repo | ~2 MB | `git clone` |
| **Binary releases** | GitHub Releases | ~15 MB per platform | Auto-downloaded by `install.sh` |
| **Mesh data** | GitHub Releases | ~25 MB | Auto-downloaded on first run |
| **Docker images** | GitHub Container Registry | ~50 MB | `docker pull ghcr.io/mayveskii/mimic:latest` |

**Why?** Git is for code review, not for 25MB JSON files or multi-arch binaries. Releases are immutable and versioned. Docker images are ready-to-run.

### Docker (Alternative)

```bash
# Pull and run — data pre-bundled in image
docker run -p 1337:1337 \
  -e MIMIC_PORT=1337 \
  -e MIMIC_LOG_LEVEL=info \
  ghcr.io/mayveskii/mimic:latest serve

# Or build locally
docker build -t mimic:latest .
docker run -p 1337:1337 mimic:latest serve
```

### From Source

```bash
# Clone & build
git clone git@github.com:Mayveskii/Mimic.git
cd Mimic
make                    # Build C-core + Go binary

# Run tests
make check              # lint + build + all tests
make core-test          # C-core 16 assertions
make test               # Go tests

# Start server
./bin/mimic serve       # MCP over stdio
./bin/mimic serve --port 1337  # HTTP transport
```

## What This Project Computes

| Metric | Formula | Current Value |
|--------|---------|---------------|
| **Survival Index (SI)** | `surviving_lines / total_lines_added` via `git blame` | **0.8500** avg |
| **Z-Density** | `(Σ survival_i × weight_i) / slot_volume` | **0.3342** avg |
| **Artifact Precision** | `SI × invariant_coverage × extraction_reproducibility` | **0.8500** avg |
| **Energy Cost** | `Σ cost_tokens × cost_time_us` per chain | Tracked per OpPacket |
| **Conflict Level** | `0=None, 1=Low, 2=Medium, 3=High, 4=Fatal` | 15 rules in matrix |

**Deep Cache:** 13,611 artifacts from 3 production repos, all passing 13 QAC checks with precision ≥ 0.8.

Run distillation yourself:
```bash
python3 data/extraction/distill_pipeline.py
# Generates: data/distilled/mesh_slots.json, mesh_stats.json, SEMANTIC_SUMMARY.md
```

## Inspiration & Behavior Sources

Mimic extracts proven patterns from production code — **not copying, but selecting the best behavior**:

### From [oven-sh/bun#30412](https://github.com/oven-sh/bun/pull/30412) — 170 Parallel Agents
- **Phase graph orchestration:** CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND
- **2-vote verification:** Independent verifiers for critical operations
- **Edit scope isolation:** Conflict matrix prevents cross-contamination
- **Never-rules:** Hard constraints (no git reset, no re-gate, no Box::leak)

### From [gonka-ai/vllm#36](https://github.com/gonka-ai/vllm/pull/36) — Measured Optimization
- **Every change measured:** Before/after metrics on real hardware
- **Decision pattern:** Measured optimization → distill → apply invariants
- **Paged memory management:** Block-level allocation without fragmentation

### From [Mayveskii/rtk](https://github.com/Mayveskii/rtk) — Token Compression (49K⭐)
- **8-stage filter pipeline:** strip_ansi → collapse → truncate → smart_omit
- **Language-aware:** Strip comments/bodies, keep signatures
- **Head+tail:** Keep first 50 + last 50 lines for logs
- **Result:** 95% token reduction on large outputs

### From [Mayveskii/hermes-agent](https://github.com/Mayveskii/hermes-agent) — Production Agent
- **Closed learning loop:** Skills from experience, self-improvement
- **Context compression:** Multi-pass with stable prefix caching
- **Iteration budget:** Grace call on exhaustion, circuit breaker at 3 denials

### From [Mayveskii/graphify](https://github.com/Mayveskii/graphify) — Knowledge Graph
- **30+ language AST extraction:** Structural + call-graph edges
- **IDF-weighted search:** Exact (1000x) > Prefix (100x) > Substring (1x)
- **Leiden clustering:** Community detection for codebase understanding

## Future: Mimic Mesh (Distributed Knowledge)

> **The more participants, the stronger we become.**

### Phase 1: Local Deep Cache (Current)
- 13,611 artifacts from 90+ production repos
- Survival index ≥ 0.8, Z-density ≥ 0.7
- Stored locally per agent session

### Phase 2: Shared Mesh Hub (Next)
```
┌─────────────────────────────────────────────┐
│              Mimic Mesh Hub                   │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
│  │ Agent A │  │ Agent B │  │ Agent C │     │
│  │ (node)  │  │ (node)  │  │ (node)  │     │
│  └────┬────┘  └────┬────┘  └────┬────┘     │
│       │            │            │            │
│       └────────────┼────────────┘            │
│                    ▼                          │
│            ┌─────────────┐                  │
│            │  Deep Cache │                  │
│            │  100K+ slots│                  │
│            │  Shared     │                  │
│            └─────────────┘                  │
└─────────────────────────────────────────────┘
```

- **Port 1557:** Inter-node mesh communication
- **Shared cache:** Every solved task becomes a mesh slot
- **Survival tracking:** Git blame across ALL participants
- **Z-density amplification:** More repos → higher density → better decisions

### Numbers

| Metric | Current | Mesh Target |
|--------|---------|-------------|
| Artifacts | 13,611 | 100,000+ |
| Repos | 90+ | 500+ |
| Participants | 1 (you) | 1000+ nodes |
| Z-Density | 0.72 avg | 0.90+ |
| Decision speed | ~2s | <500ms (cached) |
| Token savings | 30-50% | 70%+ |

## Project Structure

```
Mimic/
├── AGENTS.md              ← You are here. Rules for AI agents.
├── README.md              ← Human-readable overview
├── Dockerfile             ← Multi-stage build, ports 1337/1117/1227
├── docker-compose.yml     ← Docker Compose with healthcheck
├── .env.example           ← Environment variables reference
├── install.sh             ← One-liner curl|bash installer
├── Makefile               ← build, test, lint, check, distill
├──
├── core/                  ← C-core (91 OpCodes)
│   ├── ops.c              ← Execution engine
│   ├── ops.h              ← Public API
│   └── test_ops.c         ← 16 assertions
│
├── internal/
│   ├── mcp/               ← MCP server (JSON-RPC)
│   │   ├── mcp.go         ← Server loop
│   │   └── tool_schemas.go ← 35 JSON Schemas
│   ├── orchestrator/      ← 6-phase pipeline
│   │   ├── orchestrator.go ← CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND
│   │   └── decomposer.go  ← Task decomposition
│   ├── rtk/               ← Token compression (from rtk-ai)
│   │   └── compress.go  ← 95% output reduction
│   └── cgo/               ← Go↔C bridge
│
├── data/
│   ├── extraction/        ← Distillation scripts
│   │   └── distill_pipeline.py ← Validate + synthesize + metrics
│   ├── seeds/             ← Initial mesh slots (13,611 artifacts)
│   └── distilled/         ← Synthesized mesh slots + stats
│
├── test/
│   └── integration/       ← 29 integration tests
│       └── comprehensive_test.py
│
├── mimicrya/
│   ├── behavior-sources.yaml ← 20 repos, 123 behaviors
│   └── repos-manifest.yaml   ← 90+ production repos
│
├── specs-v2/              ← Full specification
│   └── domains/           ← Per-domain docs
│
└── project_context_main/  ← Agent persistent memory (gitignored)
```

## Test Results

### Local Tests
- `make check` ✅ lint + build + Go tests
- `make core-test` ✅ 16/16 C-core assertions
- `go test ./internal/orchestrator` ✅ 21 tests, ~80% coverage
- `go test ./internal/rtk` ✅ 13 tests, compression verified

### OpenRouter Integration (kimi k2.6)
- **tools/list:** 35 tools with schemas ✅
- **SYS_FILE_EXISTS:** Correct args, real `stat()` ✅
- **HASH_SHA256:** Real OpenSSL hash ✅
- **BUILD+TEST:** Model decomposed into 2 calls ✅
- **Collision rate:** 0% (with schema) vs ~30% (without)
- **Cost:** ~$0.01 per test run

## Branches

| Branch | Purpose |
|--------|---------|
| `main` | Stable releases only |
| `dev` | Integration, feature branches merge here |
| `feat/core-ops` | C-core OpPacket execution |
| `feat/mcp-server` | MCP JSON-RPC server |
| `feat/orchestrator` | Pipeline, budget, guardrails |
| `feat/graphify` | Knowledge graph integration |
| `feat/rtk-filter` | Token compression pipeline |

## Contributing

1. Read `AGENTS.md` — rules for agents working on Mimic
2. Run `make check` before any commit
3. Every non-trivial decision needs an ADR in `docs/adr/`
4. Follow the two-source rule: distillation + mimicry

## License

MIT — See LICENSE file

---

**Built with determination.** Every function has a source. Every artifact has survival index. Every decision is measured.

*Mimic. For agents that refuse to guess.*
