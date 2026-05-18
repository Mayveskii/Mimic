# specs-v2/ — Domain-Based Semantic Specification

What Mimic IS, what Mimic DOES, what results it predictably produces.
No implementation status. No made-up numbers. No timelines.
Pure semantics from sources, structured by domain.

---

## Directory Layout

```
specs-v2/
├── STRUCTURE.md          ← you are here. How to read these specs.
├── README.md             ← entry point: what Mimic is, for whom, why.
│
├── c-core/               ← exact specifications for C compilation
│   ├── OPCODE_SPEC.md    ← OpCode enum, flags, safety levels, error codes
│   ├── OPPACKET_SPEC.md  ← OpPacket struct layout, lifecycle, serialization
│   ├── EXEC_CONTEXT_SPEC.md ← ExecContext, resource bitmask, FD/mmap tracking
│   ├── VALIDATION_SPEC.md ← ops_validate_chain, error codes, 11 validation steps
│   ├── CONFLICT_MATRIX_SPEC.md ← conflict levels, population rules, matrix init
│   ├── ENERGY_COST_SPEC.md ← token costs, latency estimates, budget check formula
│   └── ROLLBACK_SPEC.md  ← rollback triggers, inverse mapping, state snapshots
│
├── domains/              ← every domain has its own directory
│   ├── git/              ← git workflows: commit, merge, branch, push
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── build/            ← compile, test, link, deploy
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── io/               ← file read, write, seek, open, close
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── network/          ← HTTP, TCP, requests
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── process/          ← spawn, wait, signal, kill
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── memory/           ← mmap, alloc, free, sync
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── system/           ← exec, env, dirs, files
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── utility/          ← hash, compress, encrypt
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── orchestrator/     ← classify, plan, validate, execute, verify, respond
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── session/          ← budget, denials, context flow
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── rag/              ← retrieval: linear, keyword, semantic
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── mesh/             ← slot storage, indexing, bmap
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── distillation/     ← clone, blame, survival, extract, slot
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── security/         ← DIFC, permissions, never-rules
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── quality/          ← 2-vote, conflict, energy, invariants
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── research/                    ← scientific research: hypotheses, experiments, literature
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── self-management/             ← checkpoint, budget reallocate, strategy pivot
│   │   ├── PROCESS.md
│   │   ├── INVARIANTS.md
│   │   ├── SOURCES.md
│   │   └── ARTIFACTS.md
│   ├── anti-patterns/    ← 30 documented failures and counters
│       ├── PROCESS.md
│       ├── INVARIANTS.md
│       ├── SOURCES.md
│       └── ARTIFACTS.md
│
├── blueprints/           ← reusable templates for domain specs
│   ├── DOMAIN_TEMPLATE.md
│   ├── INVARIANT_TEMPLATE.md
│   └── ARTIFACT_TEMPLATE.md
│
├── invariants/           ← cross-domain rules that apply everywhere
│   ├── META_INVARIANT.md ← no_side_effect_without_prior_validation
│   └── CONFLICT_RULES.md ← cross-domain conflict detection rules
│
├── patterns/             ← named scenario definitions (tokenized processes)
│   ├── atomic_commit.md
│   ├── safe_merge.md
│   ├── feature_branch.md
│   ├── hotfix.md
│   ├── ci_diff_check.md
│   └── build_and_test.md
│
└── artifacts/            ← how knowledge is stored in mesh
    ├── SLOT_SCHEMA.md    ← binary slot layout for compilation
    ├── ARTIFACT_SCHEMA.md ← JSON artifact structure for distillation
    └── FEEDBACK_SCHEMA.md ← feedback loop structure for learning
```

---

## How to Read a Domain Spec

Every domain directory contains:

```
domains/<name>/
├── PROCESS.md      ← what processes this domain provides, their behavior, results
├── INVARIANTS.md   ← rules that must hold for every process in this domain
├── SOURCES.md      ← which source repos inform this domain and what principles they bring
└── ARTIFACTS.md    ← how processes from this domain are stored as mesh slots
```

### PROCESS.md structure

Each process is described as:

```
## <process_name>

**When to use:**
**Goal:**
**Chain (semantically):**
**Hard constraints:**
**Invariants:**
**Result when successful:**
**Result when failed:**
**How a model uses this:**
```

No code. No "implemented X of Y". Only behavior and result.

### INVARIANTS.md structure

Each invariant:

```
## <invariant_name>

**What it prevents:**
**What it requires:**
**Source of this rule:**
**Consequence of violation:**
```

### SOURCES.md structure

```
## <source_repo>

**Principles taken:**
**What Mimic does with them:**
**What Mimic does NOT copy:**
```

### ARTIFACTS.md structure

```
## Slot Structure

| Field | Value |
|---|---|
| domain | <domain_enum> |
...

## Pattern Codes
### <pattern_name>
```c
OpPacket chain[N] = {...}
```

## Anti-Pattern Slots
| Anti-Pattern | Slot Name | counter_slot_id |

## Retrieval Path
```

---

## How to Read C-Core Specs

Every C-core spec is a compilation-ready document:

- **OPCODE_SPEC.md**: Exact enum values, flag constants, safety levels, string mappings, error codes.
- **OPPACKET_SPEC.md**: Exact struct layout with sizes, field semantics, lifecycle (create → validate → execute → rollback → destroy).
- **EXEC_CONTEXT_SPEC.md**: Exact struct layout, resource bitmask bit assignments, FD/mmap tracking, state snapshots, budget tracking.
- **VALIDATION_SPEC.md**: Exact 11 validation steps, error code assignments, performance requirements.
- **CONFLICT_MATRIX_SPEC.md**: Conflict levels, population rules, cross-domain rules, bitmask assignment.
- **ENERGY_COST_SPEC.md**: Token costs, latency estimates, budget formulas, optimization principles.
- **ROLLBACK_SPEC.md**: Rollback triggers, inverse operation mapping, best-effort cleanup, state snapshot format.

These specs describe the DESIRED binary interface. Implementation must match exactly for compilation.

---

## How Models Use This

A weak model wants to perform a task. It does not know HOW.

1. Model expresses intent to Mimic.
2. Mimic classifies intent → determines domain.
3. Mimic retrieves the tokenized process for that domain from mesh.
4. Model receives: the process (step by step), the invariants (what must hold), the constraints (what is forbidden).
5. Model does not improvise. It follows the tokenized process.
6. Result is deterministic because the process is proven, not guessed.

The intelligence is in the mesh, not in the model.

---

## Comment Convention

Any line starting with `#` is a comment on uncertainty:

```
# UNCERTAIN: precise threshold for this rule needs measurement on real data
# UNCERTAIN: whether this behavior applies to all git hosts or only specific ones
```

These comments mark places where the spec needs real-world validation.
They are not errors. They are flags for future measurement.

## No Implementation Status

No file contains "implemented", "pending", "planned", "2 of 9", or any status tracking.
These specs describe desired behavior. Implementation status lives outside this directory.
