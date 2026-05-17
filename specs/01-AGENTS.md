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
├── ARCHITECTURE.md        ← Components, flows, boundaries
├── EXECUTION-SPACE.md     ← Agent task execution space, task types, dimensions
├── RESOURCES.md           ← Complete resource map, OpPacket translation
├── SCENARIOS.md           ← Execution patterns with chains, invariants, costs
├── BEHAVIOR.md            ← Formulas, invariants, phase transitions
├── SEMANTICS.md           ← Every function: name | input | output | invariant | source
├── MODULES.md             ← Per-module documentation, connections, state
├── mimicrya/
│   ├── behavior-sources.yaml  ← Mayveskii/* repos: which behaviors Mimic borrows
│   └── repos-manifest.yaml    ← Production repos: distillation status
├── docs/adr/              ← Architecture Decision Records
├── core/                  ← C-core (ops.c, ops.h, exec_*.c, bmap/)
├── internal/
│   ├── mcp/               ← MCP server (JSON-RPC, stdio/SSE/HTTP)
│   ├── tool/              ← MCP tool registry
│   ├── cgo/               ← CGO bridge → core/
│   ├── orchestrator/      ← Workflow state machine
│   ├── session/           ← Agent sessions
│   └── quality/           ← 2-vote verify, denial tracking
├── cmd/mimic/main.go      ← Entrypoint
├── data/
│   ├── extraction/        ← Distillation scripts
│   ├── seeds/             ← Initial slots
│   └── matrices/          ← Conflict/energy matrices
├── test/                  ← Integration tests
├── Makefile               ← build, test, lint, check, distill, release
├── .github/workflows/     ← ci.yml, release.yml, distill.yml
├── go.mod
└── Dockerfile
```

## Reading the Project (for any agent)

1. Read AGENTS.md (this file) — rules, structure, concepts
2. Read ARCHITECTURE.md — components, flows, boundaries, behavior source map
3. Read EXECUTION-SPACE.md — what agents can do, task types, execution dimensions
4. Read RESOURCES.md — complete resource map, OpPacket translation for every function
5. Read SCENARIOS.md — execution patterns with chains, invariants, costs
6. Read BEHAVIOR.md — formulas, invariants, phase transitions
7. Read SEMANTICS.md — every function: name | input | output | invariant | source
8. Read MODULES.md — per-module documentation, connections, current state
9. Read mimicrya/behavior-sources.yaml — where behaviors come from

## Two Sources of Knowledge

### Distillation (mimicrya/repos-manifest.yaml)
Production repos (etcd, k8s, go-ethereum and 90+ others) → git blame → survival index → best commits → mesh slots. These are proven patterns that survived in real systems.

### Mimicry (mimicrya/behavior-sources.yaml)
Mayveskii/* repos (bun, exa-mcp-server, gh-aw-mcpg, code-mode, opencode, agency-agents, etc.) → behavior selection: HOW to implement a function in Mimic. Bun showed how to orchestrate phases → Mimic implements a phase graph. gh-aw-mcpg showed how to route MCP → Mimic implements transport. This is not copying — it's selecting the best behavior to implement.

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
- `feature/X` → PR into dev → merge; dev → PR into main → tag → release

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
