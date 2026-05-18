# specs-v2 — Mimic Semantic Specification

What Mimic IS. What Mimic DOES. What results it predictably produces.

No implementation status. No made-up numbers. No timelines.
Pure semantics from sources, structured by domain.

---

## What is Mimic?

Mimic is a model amplifier. A weak model + Mimic = a strong model.

The model expresses intent. Mimic translates intent into validated, tokenized, measured processes that the model follows step by step. The intelligence is not in the model — it is in the mesh (proven patterns from production code), the conflict matrix (enforced prohibitions), and the energy matrix (measured costs).

## How It Works

```
Model: "commit these files safely"
    ↓
Mimic: CLASSIFY → git domain, atomic_commit scenario
    ↓
Mimic: PLAN → instantiate tokenized process [status → diff → add → commit]
    ↓
Mimic: VALIDATE → conflict check, budget check, permission check
    ↓
Mimic: EXEC → run process, measure latency per step
    ↓
Mimic: VERIFY → (if critical) 2-vote verification
    ↓
Mimic: RESPOND → commit hash + metrics + traceability
    ↓
Model: receives result, follows process, never improvises
```

## Directory Structure

```
specs-v2/
├── STRUCTURE.md          ← how to read these specs
├── README.md             ← this file
│
├── domains/              ← every domain has its own spec
│   ├── git/              ← git workflows
│   ├── build/            ← compile, test, deploy
│   ├── io/               ← file operations
│   ├── network/          ← HTTP, TCP
│   ├── process/          ← spawn, wait, signal
│   ├── memory/           ← mmap, alloc, free
│   ├── system/           ← exec, env, dirs
│   ├── utility/          ← hash, compress, encrypt
│   ├── orchestrator/     ← 6-phase pipeline
│   ├── session/          ← budget, denials, context
│   ├── rag/              ← retrieval
│   ├── mesh/             ← slot storage
│   ├── distillation/     ← from repo to slot
│   ├── security/         ← permissions, never-rules
│   ├── quality/          ← 2-vote, conflict, energy
│   ├── research/         ← scientific research: hypotheses, experiments, literature
│   ├── self-management/  ← checkpoint, budget reallocate, strategy pivot
│   └── anti-patterns/    ← 30 documented failures
│
├── blueprints/           ← templates for domain specs
│
├── invariants/           ← cross-domain rules
│   └── META_INVARIANT.md ← no_side_effect_without_prior_validation
│
├── patterns/             ← named tokenized processes
│   └── atomic_commit.md  ← example pattern
│
└── artifacts/            ← how knowledge is stored
```

## How to Use This

### For a developer implementing Mimic

1. Choose domain → read `domains/<name>/PROCESS.md`
2. Understand processes → what behavior, what result, what constraints
3. Read sources → `SOURCES.md` in each domain (created per-domain)
4. Read invariants → `invariants/META_INVARIANT.md` applies to all
5. Read patterns → `patterns/` for named scenario templates

### For a model using Mimic

Models do not read these specs directly. Models call Mimic MCP tools. These specs describe what Mimic does internally to serve the model.

## Key Principles

1. **Behavior → Result → Why** — every line describes what happens and why, not where code lives.
2. **No implementation status** — "implemented", "pending", "2 of 9" are banned. Only desired behavior.
3. **No made-up numbers** — measured values only. Uncertainty marked with `# UNCERTAIN:` comment.
4. **Sources inform, not dictate** — we learn principles from Mayveskii/* repos, not copy their code.
5. **libbmap.a sources exist** in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`. No "no source files" lies.
6. **Tokenization is the core** — every operation is an OpPacket token. Processes are token chains.
7. **Mesh stores proven knowledge** — slots are compressed, hashed, indexed, retrievable.
8. **Meta-invariant rules all** — no side effect without prior validation.

## Comment Convention

Any line starting with `# UNCERTAIN:` marks a place needing real-world measurement or calibration.

These are not errors. They are flags for future validation.

---

## Reading Order

1. `STRUCTURE.md` — how to read these specs
2. `domains/orchestrator/PROCESS.md` — the 6-phase pipeline
3. `domains/git/PROCESS.md` — example domain with full semantics
4. `invariants/META_INVARIANT.md` — the rule that prevents all failures
5. Any specific domain you are implementing
6. `patterns/` — named scenario templates

---

## What This Is NOT

- Not a code specification. No function signatures, no type definitions.
- Not an implementation status tracker. No "done", "pending", "planned".
- Not a project plan. No timelines, no phases, no roadmaps.
- Not a data dump. Every line serves a behavioral purpose.
