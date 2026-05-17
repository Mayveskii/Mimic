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
```

## Source Repository Analysis (per-repo spec cards)

Each Mayveskii/* repo has a structured spec card in `specs/sources/`:

```
specs/sources/bun.md
specs/sources/exa-mcp-server.md
specs/sources/gh-aw-mcpg.md
specs/sources/opencode-anomalyco.md
specs/sources/code-mode.md
specs/sources/embryo.md
specs/sources/gastown.md
specs/sources/rustnet.md
specs/sources/netbootxyz.md
specs/sources/git.md
specs/sources/gitingest.md
specs/sources/awesome-mcp-servers.md
specs/sources/agency-agents.md
specs/sources/openmythos.md
specs/sources/caveman.md
specs/sources/minbpe.md
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
