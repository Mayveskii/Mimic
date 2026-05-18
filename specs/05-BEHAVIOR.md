# BEHAVIOR.md — Mimic

## Formulas and Invariants

Every formula drives code. Every coefficient comes from measurement.

> **NOTE**: Canonical specifications moved to `specs-v2/`. This file is historical. Read `specs-v2/invariants/META_INVARIANT.md` and `specs-v2/domains/*/PROCESS.md` for current specs.

---

### 1. Principle of Least Action

The engine selects the chain with minimum action S:

```
S = Σᵢ (cost_tokensᵢ × cost_time_usᵢ)
```

**Invariant**: if S(chain_A) < S(chain_B) and both are valid, chain_A is preferred.
**Implementation**: `ops_calculate_action()` in ops.c
**Source**: Principle of Least Action (physics)

---

### 2. Conflict Matrix

```
conflict[op1][op2] ∈ {0, 1}
0 = compatible, 1 = conflict
```

**Population rules** (from behavior sources):
- OP_SYS_EXEC × OP_SYS_EXEC = 1 (race condition, from ops.c)
- DELETE × WRITE = 1 (write_after_delete, from validator.go)
- WRITE × READ without SYNC = 1 (read_after_write_no_sync, from validator.go)
- Same-domain operations with overlapping resources = 1 (edit scope isolation, from Mayveskii/bun)

**Invariant**: if conflict[op1][op2] = 1, these ops CANNOT be in the same chain.
**Implementation**: `g_conflict_matrix` in ops.c, `conflictRules` in validator.go
**Canonical spec**: `specs-v2/c-core/CONFLICT_MATRIX_SPEC.md`

---

### 3. Energy Cost Matrix

```
energy[op] = [cost_tokens, cost_time_us, cost_memory_bytes]
```

**Measured values** (from ops.c ops_register_builtins):

| OpCode | cost_tokens | cost_time_us | cost_memory_bytes |
|--------|------------|-------------|-------------------|
| OP_NOP | 0.0 | 0.01 | 0.0 |
| OP_SYS_FILE_EXISTS | 1.0 | 10.0 | 0.0 |
| OP_SYS_DIR_CREATE | 2.0 | 50.0 | 4096.0 |

**Invariant**: Σ cost_tokensᵢ ≤ budget_tokens
**Implementation**: `g_energy_costs` in ops.c
**Canonical spec**: `specs-v2/c-core/ENERGY_COST_SPEC.md`

---

### 4. Survival Index

```
survival(commit) = surviving_lines / total_lines_added
```

**Thresholds** (require calibration on real data):
- survival ≥ 0.7 → slot candidate
- survival < 0.1 → discard
- 0.1 ≤ survival < 0.7 → partial pattern, manual review

**Invariant**: survival = 1.0 is impossible for active projects (code evolves)
**Source**: Distilled from git blame analysis (see `specs-v2/domains/distillation/`)

---

### 5. Z-density

```
Z(slot) = (Σᵢ survivalᵢ × weightᵢ) / slot_volume
```

**Invariant**: Z(slot_A) > Z(slot_B) → slot_A contains more proven knowledge per unit volume
**Implementation**: `z_density_compute` in libbmap.a
**Note**: libbmap.a sources exist in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`. No "no source files" claims.

---

### 6. Phase Transitions (Workflow State Machine)

```
CLASSIFY ──[intent identified]──→ PLAN
PLAN ──[chain built]──→ VALIDATE
VALIDATE ──[no conflicts, budget ok]──→ EXEC
EXEC ──[success]──→ VERIFY
VERIFY ──[2-vote pass]──→ RESPOND
```

Rollback: VALIDATE fail → PLAN, EXEC fail → PLAN, VERIFY fail → CLASSIFY.

**Invariant**: no EXEC without passed VALIDATE.
**Source**: Mayveskii/bun (PR #30412 phase graph)
**Canonical spec**: `specs-v2/domains/orchestrator/PROCESS.md`

---

### 7. Permission Pipeline

```
deny_rules → classify(auto) → budget_check → allow_rules
```

**Denial tracking**: 3 consecutive denies → circuit break.
**Invariant**: OP_FLAG_DANGEROUS (0x80) always requires explicit allow.
**Source**: Mayveskii/code-mode
**Canonical spec**: `specs-v2/domains/security/PROCESS.md`

---

### 8. 2-Vote Adversarial Verify

```
verify_result = vote(executor_A, executor_B)
if vote_A == vote_B → return vote_A
if vote_A ≠ vote_B → tiebreak(verified_result)
```

**Invariant**: critical operations always undergo 2-vote.
**Source**: Mayveskii/bun (PR #30412)
**Canonical spec**: `specs-v2/domains/quality/PROCESS.md`

---

### 9. DIFC Security (6-phase pipeline)

```
1. label_agent    → agent clearance
2. label_resource → resource classification
3. coarse_check   → clearance ≥ classification?
4. execute        → if pass
5. label_response → response classification
6. fine_filter    → remove what agent must not see
```

**Invariant**: information flows only from ≥ clearance to ≤ clearance, never the reverse.
**Source**: Mayveskii/gh-aw-mcpg
**Canonical spec**: `specs-v2/domains/security/PROCESS.md`

---

### 10. Mimicry Control

Reproducing behavior from a source in a new context:

```
mimic(source_behavior, context) = {
    precondition = source_behavior.preconditions
    context_check = precondition.evaluate(context)
    if context_check → implement(source_behavior.pattern)
    if !context_check → degrade(source_behavior.next_lower_tier)
}
```

**Invariant**: mimicry without preconditions = guessing. Every borrowed behavior must have preconditions.
**Source**: mimicrya/behavior-sources.yaml

---

### 11. Workspace Indexing

Mimic indexes the agent's workspace into bmap slots for fast context retrieval:

```
workspace_index = {
    file_tree:    si_insert(path, domain="workspace", layer="tree")
    symbols:      si_insert(symbol, domain="workspace", layer="symbols")
    dependencies: si_insert(dep, domain="workspace", layer="deps")
    git_state:    si_insert(branch/diff/stash, domain="git", layer="state")
}
```

Query: `si_query_domain_layer("workspace", "symbols")` → all indexed symbols.

**Invariant**: index is stale after any WRITE operation without re-index.
**Canonical spec**: `specs-v2/domains/rag/PROCESS.md`

---

### 12. Binary RAG (Retrieval-Augmented Generation)

Binary vector search over mesh slots for pattern retrieval per `specs-v2/domains/rag/PROCESS.md`:

```
binary_rag(query, domain) = {
    q_vec = int8_quantize(embed(query))          → int8 vector
    candidates = si_query_domain(domain)           → slot set
    scores = batch_cosine_int8(q_vec, candidates)  → similarity scores
    ranked = sort(candidates, by=scores, desc)     → ranked slots
    return ranked[0..k]                            → top-k results
}
```

5-signal hybrid (from Mayveskii/embryo pkg/rag/):
1. Vector similarity (batch_cosine_int8)
2. Keyword match (si_query_state_hash)
3. Domain filter (si_query_domain)
4. Survival score (survival index of slot's source commit)
5. Z-density (z_density_compute)

**Invariant**: RAG without survival signal = unverified retrieval. Every result must carry survival index.

---

### 13. Context Flow

Context passes through the execution pipeline per `specs-v2/domains/session/PROCESS.md`:

```
agent_intent → CLASSIFY → classified_intent → PLAN → planned_chain → VALIDATE → validated_chain → EXEC → execution_context → VERIFY → verify_result → RESPOND → response
```

Each phase enriches context. Context is cumulative — later phases see everything from earlier phases.

**Invariant**: no phase receives context from a phase that hasn't executed.
**Canonical spec**: `specs-v2/domains/session/PROCESS.md`

---

### 14. Multi-task Pipeline Execution

Multiple independent pipelines execute concurrently with isolation per `specs-v2/domains/orchestrator/PROCESS.md`:

```
pipelines = [pipeline_A, pipeline_B, pipeline_C]

for each pipeline in pipelines:
    check: conflict_matrix[pipeline_A.chain] × [pipeline_B.chain] = all 0
    if conflict → serialize (A first, then B)
    if no conflict → parallel
```

Isolation rules (from Mayveskii/bun edit scope isolation):
- Each pipeline edits only its own scope (file set, domain)
- Shared resources (git index, build cache) require serialization
- Conflict matrix extended to cross-pipeline: resource_bitmask overlap → conflict

**Invariant**: no two pipelines write to the same resource simultaneously.

---

### 15. Constant Data Compression

Data in mesh slots is compressed at rest and decompressed on access per `specs-v2/c-core/OPPACKET_SPEC.md`:

```
slot_write(data) = {
    compressed = OP_COMPRESS_GZIP(data)
    hash = sha256_hash(compressed)
    bmap_write_cell(slot_id, compressed)
    store_metadata(slot_id, hash, original_size, compressed_size)
}

slot_read(slot_id) = {
    compressed = bmap_read_cell(slot_id)
    stored_hash = get_metadata(slot_id).hash
    verify: sha256_hash(compressed) == stored_hash
    data = OP_DECOMPRESS_GZIP(compressed)
    return data
}
```

**Invariant**: every slot write is compressed. Every slot read verifies hash before decompress.

---

## Formula Control

1. Every formula has an implementation in code (noted where)
2. Coefficients are measured, not invented
3. If a formula has no implementation — it is a TODO
4. CI checks: SEMANTICS.md is in sync with code
5. Formula change = ADR (why changed, what was measured)

---

### 16. Meta-Invariant: No Side Effect Without Prior Validation

```
no_side_effect_without_prior_validation
```

Every side effect (I/O, state mutation, resource allocation, network call) MUST be preceded by validation that the operation is safe, within budget, and not conflicting.

**Source**: Historical analysis of 453 binary-mesh commits (222 fixes), 1074 gonka commits (386 fixes), 148121 enricher slots across 8 domains. Root cause of all 30 documented anti-patterns (`specs-v2/domains/anti-patterns/`).

**Canonical spec**: `specs-v2/invariants/META_INVARIANT.md`

**Violations and consequences**:

| Violation | Source | QAC |
|-----------|--------|-----|
| No validation PruneEpoch succeeded before advancing | gonka 86c686d92 | QAC-1, QAC-2 |
| No validation mutex free before spawning goroutine | gonka d9c46cae0 | QAC-4, QAC-11 |
| No validation buildMessages structure stable | binary-mesh c9e83c3 | QAC-6, QAC-10 |
| No validation input before NewInt64Coin | gonka 75e31c233 | QAC-2, QAC-7 |
| No validation secrets not in source | binary-mesh ff5ff82 | QAC-8, QAC-12 |
| No classification before retry decision | binary-mesh 0e36abd | QAC-3, QAC-6 |
| No idempotency validation on Close | gonka 8d1cd00f3 | QAC-2, QAC-7 |
| No bounds check before array access | binary-mesh 4742339 | QAC-2, QAC-5 |

**Implementation**: ops_validate_chain before ops_execute_chain. Quality of validation tracked via QAC-4.

**Cross-reference**: 10-QUALITY-GATES.md Section 2 (Meta-Invariant), `specs-v2/domains/anti-patterns/PROCESS.md` (30 anti-patterns all traceable to this invariant).
