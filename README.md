# Mimic — Deterministic Tool Execution for AI Agents

**Mimic is not the agent.** It is an MCP tool that any AI model (Claude, GPT, Kimi, etc.) calls when it needs deterministic execution, rollback on failure, and validated chains — instead of guessing bash commands.

---

## What It Does

- **Validates before running** — conflict matrix, energy budget, permission pipeline prevent invalid tool sequences
- **Deterministic execution** — C-core with 96 OpCodes, measured latency, real syscalls (not stubs)
- **Rolls back on failure** — 3-phase inverse → cleanup → hash verify
- **Compresses output** — RTK reduces large outputs (git log, build output) by 95% to save context window
- **Searches the web** — Exa integration for real-time knowledge ingestion

## Architecture

```
AI Agent (autonomous, fully optional to use Mimic)
    ↓ JSON-RPC over stdio/TCP
MCP Server (Go) — tool routing, Exa handler, mesh query
    ↓ CGO Bridge
Orchestrator (Go) — classify → plan → validate → execute → verify
    ↓ OpPacket Chain
C-Core (C) — ops_execute_chain(), conflict matrix, energy costs
    ↓ real syscalls
OS — stat(), open(), git, make, curl
```

## Two Knowledge Sources

1. **Distillation** — 90+ production repos (etcd, k8s, go-ethereum...) → git blame → survival index → mesh slots
2. **Mimicry** — Mayveskii/* repos (bun, rtk, graphify, exa-mcp-server...) → behavior selection → implementation

Meshes are local `.gob` files (offline), indexable via `MESH_QUERY`.

## Roadmap to Autonomy

| Stage | Status | What changes |
|-------|--------|--------------|
| **0 — Passive Tool** | ✅ NOW | Agent asks → Mimic executes |
| **1 — Proactive** | ⏳ v0.4 | Mimic suggests next steps before being asked |
| **2 — Planning** | ⏳ v0.5 | Mimic generates multi-step plans with checkpoints |
| **3 — Generative** | ⏳ v0.6 | Mimic proposes new patterns from session logs |
| **4 — Autonomous** | ⏳ v0.7 | Self-directed learning loop, no human intervention |

## Collective Intelligence

When multiple agents use Mimic, their mesh graphs can merge. **mimic-server** (future) becomes a shared knowledge hub that aggregates survival indices across teams. Every agent execution improves the mesh for everyone else.

## Quick Start

```bash
# 1. Build
make build          # C-core + Go binary

# 2. Configure
cp .env.example .env
# Edit .env: EXA_API_KEY, MIMIC_MESH_DIR, etc.

# 3. Run
./bin/mimic serve   # stdio MCP
./bin/mimic serve --tcp :1337  # TCP mode

# 4. Test
make test           # Go tests
make check          # lint + test + semantics-check
```

## Specs (reading order)

1. `specs/01-AGENTS.md` — rules for agents working on Mimic
2. `specs/02-ARCHITECTURE.md` — components, flows, boundaries
3. `specs/03-EXECUTION-SPACE.md` — operations, dimensions, task types
4. `specs/12-EXA-RESEARCH.md` — how models use web search tools
5. `docs/adr/` — every non-trivial decision recorded

## Status

- **48 MCP tools** implemented (45 base + 3 Exa)
- **Mesh graphs**: 18 domain `.gob` files
- **Tests**: unit + e2e + battlefield benchmarks
- **Distribution**: binary releases, Docker Hub, npm (`@mayveskii/mimic`)
- **License**: MIT
