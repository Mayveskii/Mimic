Happy how , my dear friends, from cloud ! This is quite sad that persons that were against this repo's being borned now are watchiong this, but the same time it's a big pleasure to be dropped from your ship.
So it's borned , awaitingg for cost$/token reduce so soon , guys )) Have a good time with it!


# Mimic — MCP Server for Deterministic AI-Agent Tool Orchestration

Mimic is a standalone [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server with a C-core execution engine. When an AI agent chooses to call Mimic, it gets deterministic OpPacket chains — validated before execution, measured during execution, and rolled back on failure.

> **Mimic does not replace the agent.** The agent is fully autonomous. Mimic is an optional tool — the same way it calls bash, grep, or any other MCP tool.

---

## Quick Start

```bash
# Clone & build
git clone git@github.com:Mayveskii/Mimic.git
cd Mimic
make

# Run tests
make test           # Go tests
make core-test      # C-core tests (16 assertions)
make check          # lint + test + build + semantics check

# Start MCP server
./bin/mimic serve   # JSON-RPC over stdio
```

---

## What This Project Computes

| Metric | Formula | Source |
|--------|---------|--------|
| **Survival Index (SI)** | `surviving_lines / total_lines_added` via `git blame` | `compute_survival.py` |
| **Z-Density** | `(Σ survival_i × weight_i) / slot_volume` | `compute_zdensity.py` |
| **Artifact Precision** | `SI × invariant_coverage × extraction_reproducibility` | `quality_gate.py` |
| **Energy Cost** | `Σ cost_tokens × cost_time_us` per chain | `ops.c` |
| **Conflict Level** | `0=None, 1=Low, 2=Medium, 3=High, 4=Fatal` | `g_conflict_matrix` |

**Deep Cache threshold**: `artifact_precision ≥ 0.8` (all 13,611 current artifacts qualify).

---

## Architecture Overview

```
┌──────────────────────────────────────────────┐
│              AI Agent (client)               │
│  Chooses to call Mimic via MCP stdio/SSE     │
└──────────────────┬───────────────────────────┘
                   │ JSON-RPC 2.0
┌──────────────────▼───────────────────────────┐
│   MCP Server (internal/mcp/mcp.go)           │
│   Tools/list, tools/call, initialize, ping   │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│   Orchestrator (internal/orchestrator/)      │
│   CLASSIFY → PLAN → VALIDATE → EXEC →        │
│   VERIFY → RESPOND                           │
│   - Budget tracking (tokens + time)          │
│   - Circuit breaker (3 denials → manual)     │
│   - 2-vote verification for critical ops     │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│   CGO Bridge (internal/cgo/cgo.go)           │
│   Go Packet[] → C OpPacketEx[], exec →       │
│   ValidationResult, ChainResult              │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│   C-Core Execution Engine (core/ops.c)       │
│   - 91 OpCodes registered                    │
│   - 9-step validation pipeline               │
│   - Conflict matrix (15 rules)               │
│   - Energy cost matrix (91 entries)          │
│   - Rollback engine (3-phase)                │
│   - I/O, System, Build executors (real)      │
│   - Research, Self-mgmt stubs (registered)   │
└──────────────────────────────────────────────┘
```

---

## Domain Map: How to Strengthen Each Area

Mimic is organized into **12 execution domains** plus 4 support domains. Each domain has `PROCESS.md` (workflow), `INVARIANTS.md` (rules), `ARTIFACTS.md` (examples), and `SOURCES.md` (behavior provenance).

| Domain | OpCodes | Status | How to Strengthen |
|--------|---------|--------|-------------------|
| **Memory** (0x10-0x1F) | 5 | ✅ Complete | `mmap_alloc/free/sync` tested; next: add OOM guards |
| **I/O** (0x20-0x2F) | 5 | ✅ Complete | `open/read/write/close/seek` tested; next: FD leak detector |
| **Git** (0x30-0x3D) | 14 | ⚠️ Partial | `status/diff/add/commit/checkout/branch` work; next: scenarios layer |
| **Build** (0x40-0x4F) | 5 | ✅ Complete | `compile/link/test/deploy/clean` work; next: artifact caching |
| **Network** (0x50-0x5F) | 7 | ⚠️ Partial | `http_get/post/tcp_close` work; next: TCP connect/send/recv |
| **Process** (0x60-0x6F) | 4 | ⚠️ Partial | `spawn/wait/kill/signal` work; next: fork+exec, PID tracking |
| **Utility** (0x70-0x7F) | 6 | ⏳ Stubs | `hash/compress/encrypt` registered; next: OpenSSL integration |
| **System** (0x80-0x8F) | 10 | ✅ Complete | `file_exists/dir_create/copy/move/delete/chmod/env_get/env_set/exec` tested |
| **Session** (0x90-0x9F) | 5 | ⏳ Stubs | Budget/context/denial/snapshot/compress registered; next: Go session store |
| **Orchestrator** (0x95-0x9A) | 6 | ✅ Complete | 6-phase pipeline implemented with budget + circuit breaker |
| **Research** (0xA0-0xAC) | 13 | ⏳ Stubs | Hypothesis/experiment/literature registered; next: arXiv API integration |
| **Self-Management** (0xB0-0xB5) | 6 | ⏳ Stubs | Checkpoint/strategy/assess registered; next: state machine |

**Support domains** (no OpCodes, pure Go):
| Domain | Role | Strengthen By |
|--------|------|---------------|
| **Distillation** | Extract patterns from production repos | Add more repos, improve invariant inference |
| **Quality** | 13 QAC gates | Add measured baselines from binary-mesh traces |
| **RAG** | Retrieval-augmented generation | Implement mesh query over slot index |
| **Anti-Patterns** | Negative pattern detection | Link revert commits to counter_patterns |

---

## Quality Assurance: 13 QAC Gates

Every artifact, operation, and decision must satisfy:

| # | Gate | Threshold | Status |
|---|------|-----------|--------|
| QAC-1 | Survival Index from git blame | ≥ 0.7 | ✅ All seeds |
| QAC-2 | ≥ 1 invariant per artifact | invariant_count ≥ 3 | ✅ All seeds |
| QAC-3 | Energy cost measured (or N/A for distillation) | na at extraction | ✅ All seeds |
| QAC-4 | Conflict matrix domain valid | known domain | ✅ All seeds |
| QAC-5 | Z-density > 0 | computed from slot | ✅ All seeds |
| QAC-6 | Decision consistency | zero contradictions | ✅ All seeds |
| QAC-7 | Artifact precision | ≥ 0.8 for deep cache | ✅ 13,611/13,611 |
| QAC-8 | Multimodal integrity | implicit for text | ✅ All seeds |
| QAC-9 | Anti-pattern polarity links | counter_pattern_id set | ✅ All seeds |
| QAC-10 | Temporal consistency | blame_timestamp present | ✅ All seeds |
| QAC-11 | Cross-domain conflicts | na for knowledge artifacts | ✅ All seeds |
| QAC-12 | Provenance chain | extraction.hash present | ✅ All seeds |
| QAC-13 | Revert detection | N/A for positive artifacts | ✅ All seeds |

**Verdict**: `DEEP_CACHE` — all artifacts ready for shared mesh use.

---

## Branch Strategy

Per [BRANCH-MAP.md](BRANCH-MAP.md):

```
main  ──→ production releases (tagged v0.X.Y)
  └── dev  ──→ integration branch (this branch)
       ├── feat/core-ops       ← C-core: 91 OpCodes + validation + rollback
       ├── feat/core-bmap      ← bmap rewrite: 39 libbmap functions → .c
       ├── feat/mcp-server     ← MCP JSON-RPC server, transport
       ├── feat/mcp-bridge     ← CGO bridge Go↔C
       ├── feat/orchestrator   ← 6-phase pipeline, budget, circuit breaker
       ├── feat/graphify       ← knowledge graph integration
       ├── feat/rtk-filter     ← token compression pipeline
       ├── feat/config         ← koanf layered config
       ├── feat/observability  ← OTel, Prometheus, health probes
       └── fix/*               ← hotfixes → dev
```

**Rules**:
- **Squash merge** into `dev` — one commit per feature increment
- **`make check` must pass** before any merge
- **No direct push** to `main` or `dev` — all changes via PR

---

## Data Pipeline

```
Production Repo (etcd, k8s, go-ethereum, ...)
    ↓ git clone --depth 1
    ↓ git blame -t <file>  → compute_survival.py
    ↓ survival_index = surviving_lines / total_lines_added
    ↓ extract_patterns.py  → function-level chunks
    ↓ encode_artifacts_v2.py
        - UUID artifact_id
        - 3+ domain-specific invariants
        - inline QAC assessment
        - artifact_precision = SI × inv_cov × reproducibility
    ↓ quality_gate.py (13 checks)
        - verdict: REJECT / REVIEW_PENDING / LOCAL_ONLY / DEEP_CACHE
    ↓ data/seeds/<repo>-artifacts.json
    ↓ mimicrya/repos-manifest.yaml (updated with slots, z_density)
```

**Current status**:

| Repo | Slots | Z-Density | Avg Precision | Verdict |
|------|-------|-----------|---------------|---------|
| Mayveskii/etcd | 11,551 | 0.3282 | 0.8500 | DEEP_CACHE |
| Mayveskii/rtk | 1,952 | 0.3707 | 0.8500 | DEEP_CACHE |
| Mayveskii/vllm | 108 | 0.3185 | 0.8500 | DEEP_CACHE |

---

## Known Limitations & Next Priority

1. **libbmap.a rewrite** — ADR-0005: 39 storage functions need .c implementation (slot index, invariants, snapshots, cosine similarity)
2. **Session layer** — Go-side budget tracking, denial logging, 2-vote verification hooks (currently stubs in c-core)
3. **Decision extraction** — PR comment distillation from Mayveskii/etcd/rtk/vllm → decision-patterns.yaml → conflict matrix expansion
4. **Observability** — OTel traces, Prometheus metrics, health probes
5. **Research domain** — arXiv API, hypothesis management, statistical testing

---

## License

MIT — see repository for full text.

---

## Resources

- **Specs**: `specs-v2/README.md` — canonical specification index
- **Architecture**: `specs-v2/STRUCTURE.md` — component map, data flows, file layout
- **C-Core**: `specs-v2/c-core/OPCODE_SPEC.md` — 91 OpCodes, flags, costs, safety levels
- **Quality Gates**: `specs/10-QUALITY-GATES.md` — 13 QAC definitions and thresholds
- **Behavior Sources**: `mimicrya/behavior-sources.yaml` — 19 repos, 117 behaviors
- **Distillation Sources**: `mimicrya/repos-manifest.yaml` — 90+ production repos
- **ADRs**: `docs/adr/0001-0006.md` — architecture decisions with measured impact
