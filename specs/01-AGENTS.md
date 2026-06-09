# AGENTS.md — Mimic

## Rules

1. No action without user permission
2. Facts only — nothing invented
3. Semantics before code — no function without a row in SEMANTICS.md
4. No file edits without permission
5. Every non-trivial decision = ADR

## Project Structure

```
Mimic/
├── specs-v2/               ← CANONICAL SPECIFICATIONS (read this first)
│   ├── README.md           ← What Mimic IS and DOES
│   ├── STRUCTURE.md        ← How to navigate specs-v2
│   ├── c-core/             ← Exact specs for C compilation
│   │   ├── OPCODE_SPEC.md
│   │   ├── OPPACKET_SPEC.md
│   │   ├── EXEC_CONTEXT_SPEC.md
│   │   ├── VALIDATION_SPEC.md
│   │   ├── ROLLBACK_SPEC.md
│   │   ├── ENERGY_COST_SPEC.md
│   │   ├── CONFLICT_MATRIX_SPEC.md
│   │   ├── ENV_CONFIG.md          ← All env vars with defaults
│   │   ├── RPC_FORMAT.md          ← MCP JSON-RPC + binary mesh wire
│   │   ├── BUILD_CONFIG.md        ← Compile flags, feature toggles
│   │   ├── MESH_EXCHANGE.md       ← Cross-node slot serialization
│   │   └── MEMORY_LAYOUT.md       ← Arena-based linear access
│   ├── domains/            ← 16+ domains × 4 files each
│   │   ├── git/build/io/network/process/memory/system/utility
│   │   ├── orchestrator/session/rag/mesh/distillation/security/quality/anti-patterns
│   │   └── research/                  ← NEW: scientific research workflows
│   │   └── self-management/           ← NEW: checkpoint, pivot, budget reallocate
│   ├── patterns/           ← Named scenario chains
│   ├── artifacts/          ← Slot schema, artifact schema, feedback schema
│   ├── invariants/         ← Cross-domain rules
│   └── blueprints/         ← Templates for new specs
│
├── specs/                  ← OLD monolithic specs (being replaced by specs-v2)
│   ├── 00-SPEC-INDEX.md
│   ├── 01-AGENTS.md       ← YOU ARE HERE (this file)
│   ├── 02-ARCHITECTURE.md
│   ├── ... (see SPEC-INDEX for full list)
│   └── sources/            ← Per-repo spec cards (17 repos)
│
├── mimicrya/
│   ├── behavior-sources.yaml   ← 21 repos, 146 behaviors (Mayveskii/*)
│   ├── repos-manifest.yaml     ← 90+ production repos (distillation targets)
│   └── decision-patterns.yaml  ← Decision survival tracking
│
├── docs/adr/               ← Architecture Decision Records (0001-0005)
├── c-core/                 ← C-core sources (dev branch)
│   ├── ops.c/ops.h
│   ├── git_*.c
│   ├── matrix/
│   └── libbmap.a (39 functions)
├── data/
│   ├── extraction/         ← Distillation scripts
│   ├── seeds/              ← Initial slots (empty)
│   └── matrices/           ← Conflict/energy matrices (empty)
├── Makefile                ← build, lint, check, distill, release
├── Dockerfile              ← Container build
├── go.mod                  ← Go module (orchestrator + MCP bridge)
└── cmd/mimic/main.go       ← Entrypoint (stub — user writes code)
```

## Reading the Project

### For agents contributing to Mimic:

1. Read `specs-v2/README.md` — what Mimic IS
2. Read `specs-v2/STRUCTURE.md` — how to navigate
3. Choose domain → read `specs-v2/domains/<name>/PROCESS.md`
4. Read `specs-v2/invariants/META_INVARIANT.md` — root rule
5. Read old specs (00-10) for historical context ONLY
6. Read `mimicrya/behavior-sources.yaml` — where behaviors come from

## Two Sources of Knowledge

### Distillation (mimicrya/repos-manifest.yaml)
Production repos (etcd, k8s, go-ethereum, terraform, autogen, crewai, and 90+ others) → git blame → survival index → best commits → mesh slots. These are proven patterns that survived in real systems. All repos tracked; analysis is ongoing.

### Mimicry (mimicrya/behavior-sources.yaml)
Mayveskii/* repos (bun, exa-mcp-server, gh-aw-mcpg, code-mode, opencode, hermes-agent, vllm, gastown, openmythos, and others) → behavior selection: HOW to implement a function in Mimic. Bun showed how to orchestrate phases → Mimic implements a phase graph. gh-aw-mcpg showed how to route MCP → Mimic implements transport. This is not copying — it's selecting the best behavior to implement.

## Commands

```bash
make              # Build libcore.a + mimic binary
make lint         # Check code (go vet + gofmt)
make check        # lint + test + semantics-check
make distill      # Run distillation pipeline
make release      # Build binaries + checksums
docker build .    # Build container image
```

## Branches

- `main` — stable releases only (tags)
- `dev` — C-core development, specs, and canonical specifications
- `embryo` — Go implementations (pkg/, internal/, cmd/), documentation, tools, tests
- Feature branches: `feat/*` → PR → squash merge to dev → squash merge to main on release tag

## Key Concepts

- **OpPacket chain** — ordered sequence of deterministic operations, validated BEFORE execution
- **Conflict matrix** — [OP_MAX × OP_MAX] matrix defining which operations cannot run together
- **Energy cost matrix** — [OP_MAX × 3] matrix: cost_tokens, cost_time_us, cost_memory_bytes
- **Z-density** — density of proven knowledge in a mesh slot
- **Survival index** — fraction of commit lines still present at HEAD (git blame → metric)
- **Mimicry** — selecting behavior from best implementations in Mayveskii/* repos
- **Distillation** — git blame → survival → extract → bmap slot → Z-density
- **6-phase pipeline** — CLASSIFY → PLAN → VALIDATE → EXEC → VERIFY → RESPOND
- **Meta-invariant** — no_side_effect_without_prior_validation (root cause of all 30 APs)
- **Research mode** — long-running sessions with checkpoint/resume, large context, and self-management
- **mimic-server** — FUTURE: shared knowledge hub for multiple clients. Not part of current scope.

## Capabilities

| Capability | Status | Source |
|------------|--------|--------|
| Phase graph (6-phase pipeline) | Specified | specs-v2/domains/orchestrator/ |
| 2-vote adversarial verify | Specified | specs-v2/domains/quality/ |
| Conflict matrix + energy cost | Specified | specs-v2/c-core/ |
| Budget tracking (tokens/time/memory) | Specified | specs-v2/domains/session/ |
| RAG retrieval (3-tier hybrid) | Specified | specs-v2/domains/rag/ |
| Session snapshot + resume | Specified | specs-v2/domains/session/ |
| Rollback on failure | Specified | specs-v2/c-core/ROLLBACK_SPEC.md |
| Checkpoint for long tasks | NEW | specs-v2/domains/self-management/ |
| Research hypothesis tracking | NEW | specs-v2/domains/research/ |
| Tool chaining / tool loop | Specified | specs-v2/domains/orchestrator/ |
| Parallel pipelines (max 10) | Specified | specs-v2/domains/orchestrator/ |
| Mesh exchange (cross-node) | Specified | specs-v2/c-core/MESH_EXCHANGE.md |
| MCP JSON-RPC server | Specified | specs-v2/c-core/RPC_FORMAT.md |
| DIFC security (6-phase) | Specified | specs-v2/domains/security/ |

## Notes

- All specs in specs-v2 are in English. Zero Russian in specs.
- Every line in specs-v2 follows: behavior → result → why
- No "implemented/pending/done" in spec text. Only desired behavior.
- libbmap.a sources exist in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`
- Semantics check script does not exist yet: `make semantics-check` is stubbed
- Old specs (00-11) describe historical design; canonical specs moved to specs-v2/
