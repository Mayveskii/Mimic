# MODULES.md — Mimic

Per-module documentation: what each module does, what resources it contains, how it connects to others.

> **NOTE**: Canonical specifications moved to `specs-v2/`. This file is historical.

---

## c-core/ — C-Core Engine

### What it does
Deterministic execution engine. Receives validated OpPacket chains, executes sequentially, measures latency, handles retries.

### Resources
- ops.c/ops.h: core engine (init, register, validate, execute)
- git_ops.c: OP_GIT_* executors
- git_scenarios.c: scenario chains (5 scenarios)
- mmap_ops.c: OP_MMAP_* executors
- libbmap.a: 39 symbols for mesh storage (sources exist in workspace)

### Current state
C-core sources in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`. Canonical specs in `specs-v2/c-core/`.

---

## internal/ — Go Modules (Stubs)

All internal packages are 1-line stubs:
- `internal/mcp/` — MCP server (not implemented)
- `internal/tool/` — Tool registry (not implemented)
- `internal/cgo/` — CGO bridge (not implemented)
- `internal/orchestrator/` — Workflow engine (not implemented)
- `internal/session/` — Agent sessions (not implemented)
- `internal/quality/` — Verification (not implemented)

**User writes code**. I review against specs-v2/.

---

## mimicrya/ — Knowledge Manifests

### Resources
- behavior-sources.yaml: 21+ repos with 146+ behaviors
- repos-manifest.yaml: 90+ production repos
- decision-patterns.yaml: Decision survival tracking

### Current state
behavior-sources.yaml has missing repo entry (fixed). repos-manifest.yaml has duplicates (to be cleaned).

---

## data/ — Extraction and Storage

### Resources
- extraction/: distillation scripts
- seeds/: initial mesh slots (empty)
- matrices/: conflict/energy matrices (empty)

### Current state
Not implemented. Distillation pipeline is stub in Makefile.

---

## specs-v2/ — Canonical Specifications

**This is the source of truth.**

- `c-core/`: 10 specs for exact C compilation
- `domains/`: 18 domains × 4 files = 72 files
- `patterns/`: 6 named scenarios
- `artifacts/`: 3 schemas
- `invariants/`: 2 cross-domain rules
- `blueprints/`: 3 templates

Total: 96+ spec files.
