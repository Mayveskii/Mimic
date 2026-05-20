# SPEC-INDEX — Mimic Project Documentation

Reading order. Every document has a number. Read in order.

```
00-SPEC-INDEX.md          ← YOU ARE HERE. Map of all documents.
01-AGENTS.md              ← Rules for agents working on Mimic.
02-ARCHITECTURE.md        ← System: components, flows, boundaries.
03-EXECUTION-SPACE.md     ← What agents can do: operations, dimensions, task types.
04-SCENARIOS.md           ← Execution patterns: chains, invariants, costs.
05-BEHAVIOR.md            ← Formulas, invariants, phase transitions (15 formulas).
06-SEMANTICS.md           ← Every function: name | input | output | invariant | source.
07-RESOURCES.md           ← Complete resource map, OpPacket translation table.
08-MODULES.md             ← Per-module: what, resources, connections, state.
09-DISTILLATION-ARTIFACTS.md ← Domain coverage, atomic artifacts, feedback, multimodality, precision formula.
10-QUALITY-GATES.md      ← Quality axis conditions (QAC-1..13), anti-pattern polarity, measured thresholds, efficiency metrics.
11-CONFIGURATION.md      ← Every customizable variable: name | type | default | scope | invariant | source | description.
12-EXA-RESEARCH.md       ← How models use Exa tools: discovery, extraction, synthesis workflow.
```

## Source Repository Analysis (per-repo spec cards)

Each Mayveskii/* repo has a structured spec card in `specs/sources/`:

```
specs/sources/rtk.md
specs/sources/graphify.md
specs/sources/go-service-template-rest.md
specs/sources/hermes-agent.md
specs/sources/gastown.md
specs/sources/bun.md
specs/sources/rustnet.md
specs/sources/exa-mcp-server.md
specs/sources/gh-aw-mcpg.md
specs/sources/opencode-anomalyco.md
specs/sources/code-mode.md
specs/sources/embryo.md
specs/sources/git.md
specs/sources/gitingest.md
specs/sources/netbootxyz.md
specs/sources/agency-agents.md
specs/sources/openmythos.md
specs/sources/caveman.md
specs/sources/minbpe.md
specs/sources/awesome-mcp-servers.md
```

## Spec Card Schema (every source repo follows this)

```yaml
repo: Mayveskii/<name>
url: https://github.com/Mayveskii/<name>
language: <primary language>
status: pending | partial | analyzed
last_sync: <date or "never">

description: |
  <1-2 sentence factual description of what this repo IS>

advantages:                           # WHAT this repo has that Mimic needs
  - id: <unique_id>
    what: <what capability/pattern/knowledge this repo provides>
    evidence: <where in the repo this is proven (file:line or commit)>

applications:                         # HOW Mimic applies each advantage
  - advantage_id: <links to advantages.id>
    implemented_in: <Mimic module/path>
    mechanism: <specific mechanism of application>
    invariant: <how correctness is verified>
    status: pending | planned | implemented

control:                              # HOW application correctness is tracked
  - advantage_id: <links to advantages.id>
    verification: <test, invariant check, 2-vote, ADR>
    update_trigger: <what causes re-analysis of this source>
    last_verified: <date or "never">
```

## Cross-references

- Source spec cards → BEHAVIOR.md (formulas #6-15 reference source repos)
- Source spec cards → ARCHITECTURE.md (behaviors table)
- Source spec cards → MODULES.md (behaviors implemented in modules)
- Source spec cards → SEMANTICS.md (functions marked with source repo)
- behavior-sources.yaml → source spec cards (machine-readable summary)
- 09-DISTILLATION-ARTIFACTS.md → repos-manifest.yaml (coverage targets)
- 09-DISTILLATION-ARTIFACTS.md → decision-patterns.yaml (decision survival)
- 09-DISTILLATION-ARTIFACTS.md → artifact.proto (deep cache exchange format)

## Architecture Documentation (docs/architecture/)

Operational and strategic docs for running and evolving Mimic in production:

| Document | Purpose | Audience |
|----------|---------|----------|
| DEPLOYMENT.md | Docker Compose, systemd units, bootstrap scripts, resource requirements | DevOps, SRE |
| MESH_PERFORMANCE.md | Query latency analysis, HNSW/IVF migration plan, tiered storage | Back-end engineer |
| ROADMAP_AUTONOMY.md | 4-stage autonomy roadmap (passive → proactive → generative → autonomous) | Product, architect |
| ACTIONBYTES_SPEC.md | Embryo binary patch format (`!`-delimited decoder spec) | Core contributor |

## Architecture & Strategy (docs/architecture/)

Cross-cutting documents for production ops and long-term evolution:

```
docs/architecture/DEPLOYMENT.md         ← Docker Compose, systemd, bootstrap, resource requirements
docs/architecture/MESH_PERFORMANCE.md   ← From brute-force to HNSW/IVF; tiered storage
docs/architecture/ROADMAP_AUTONOMY.md   ← Stage 0→4: passive tool → autonomous agent
docs/adr/005-text-native-mesh.md        ← ADR-005: markdown slots (GiT text-token analogy)
```

## GiT ↔ Mimic Mapping

| GiT Principle | Mimic Implementation |
|---------------|---------------------|
| Universal text tokens (vision→text) | TextSlot markdown (gob→text, ADR-005) |
| Multi-task joint training | Cross-domain edges (SlotLink, qdrant-primary) |
| Auto-regressive generation | ValidatePlan + generative OpPacket chains |
| Zero-shot | Mesh.AutoApply with adaptive threshold |
| Emergence (tasks improve each other) | Self-improving mesh via session logging |

