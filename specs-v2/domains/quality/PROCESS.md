# Quality Domain — 2-Vote, Conflict, Energy, Metrics

How Mimic verifies correctness and measures itself.

---

## What This Domain Does

Quality operations ensure that every execution is correct, every result is verified, and every cost is measured. The 2-vote adversarial verify catches errors that single verification misses. Conflict detection prevents invalid chains before execution. Energy tracking ensures budget compliance. Metrics provide observability into the system.

---

## Processes

### two_vote_verify

**When to use:**  
Before returning result for critical operations (deploy, merge, encrypt, economic operations).

**Goal:**  
Two independent verifiers check the result. Consensus → accept. Disagreement → tiebreak.

**Chain (semantically):**

1. **Verifier A** checks output correctness.
   - Did the operation produce expected result?
   - Are outputs consistent with inputs?

2. **Verifier B** checks invariant preservation.
   - Did the operation maintain all declared invariants?
   - Were any invariants violated?

3. **Compare votes.**
   - A=pass, B=pass → VERIFIED.
   - A=fail, B=fail → REJECTED, rollback.
   - A≠B → tiebreak via additional check or manual escalation.

**Hard constraints:**
- Critical operations ALWAYS undergo 2-vote. No exceptions.
- Single verifier pass = insufficient for critical ops.
- Disagreement must be resolved, never ignored.

**Invariants:**
- Every critical operation has 2 independent verifications.
- Tiebreak always produces definitive result.
- Verification results logged with timestamps.

**Result:**
```
status: "verified"
vote_a: "pass"
vote_b: "pass"
confidence: 1.0
```

---

### conflict_check

**When to use:**  
During VALIDATE phase for every chain.

**Goal:**  
Detect incompatible operations before execution.

**Chain (semantically):**

1. For every pair (op_i, op_j) where i < j:
   - Check conflict_matrix[op_i][op_j].
   - If 1 → REJECT chain.
2. For cross-pipeline chains:
   - Check resource_bitmask overlap.
   - If overlap → serialize.

**Invariants:**
- No two conflicting operations in same chain.
- No two pipelines write to same resource simultaneously.

---

### energy_track

**When to use:**  
During PLAN and EXEC phases.

**Goal:**  
Track estimated vs actual energy cost.

**Chain (semantically):**

1. PLAN: estimate cost = Σ(cost_tokens × cost_time_us).
2. VALIDATE: check estimated ≤ budget.
3. EXEC: measure actual latency per operation with CLOCK_MONOTONIC.
4. If actual > 1.2× estimated → abort, rollback, record overestimate.
5. Update energy matrix with measured values.

**Invariants:**
- Budget never exceeded.
- Measured values replace estimates over time.
- Overestimates trigger ADR investigation.

---

## Principles From Sources

### bun

**Principles taken:**
- 2-vote adversarial verify: two independent verifiers + tiebreak.
- Edit scope isolation: resource_bitmask per operation.

### code-mode

**Principles taken:**
- Budget enforcement: maxTurns + maxBudgetUsd.
- Denial tracking: 3 consecutive → circuit break.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "quality" |
| layer | "process" |
| modality | "metric" |
| invariants | ["2vote_for_critical", "conflict_check_before_exec", "budget_never_exceeded"] |
