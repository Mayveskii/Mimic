# Orchestrator Domain — Invariants

Rules that MUST hold for every process in the orchestrator domain.

---

## OINV-01: No EXEC Without VALIDATE

**What it prevents:** Execution of unvalidated chains, bypassing safety checks.

**What it requires:** OP_ORCH_EXEC MUST NOT execute a chain that has not passed OP_ORCH_VALIDATE in the same session. Validation result must be `is_valid == true`.

**Source of this rule:**
- META_INVARIANT.md: no side effect without prior validation.
- bun PR #30412: phase graph enforces gate transitions.

**Consequence of violation:** Execution REJECTED with `ERR_PERMISSION_DENY`.

---

## OINV-02: No Phase Skip

**What it prevents:** Partial process execution, missing checks.

**What it requires:** Full 6-phase sequence: CLASSIFY → PLAN → VALIDATE → EXEC → VERIFY → RESPOND. No phase may be skipped. VERIFY may be no-op for non-critical ops.

**Source of this rule:**
- bun PR #30412: phase graph with gate transitions.
- embryo orchestrator: 6-stage pipeline.

**Consequence of violation:** Chain REJECTED at validation if phases are out of order or missing.

---

## OINV-03: Rollback on Failure

**What it prevents:** Partial state corruption from failed chains.

**What it requires:** If any packet with `OP_FLAG_ATOMIC` fails during EXEC, rollback to pre-chain state. Rollback procedure documented in ROLLBACK_SPEC.md.

**Source of this rule:**
- gastown: rollback on failure.
- META_INVARIANT.md: partial failure leaves no orphans.

**Consequence of violation:** If rollback fails → `ERR_ROLLBACK_FAIL`. State may be partially modified. Honest report to model.

---

## OINV-04: Budget Never Exceeded

**What it prevents:** Resource exhaustion, runaway sessions.

**What it requires:** Energy budget check during VALIDATE. If total energy > remaining budget → chain REJECTED. During EXEC, if budget unexpectedly exhausted → immediate halt.

**Source of this rule:**
- code-mode: maxTurns + maxBudgetUsd.
- AP-23 (token overflow).

**Consequence of violation:** Chain REJECTED with `ERR_BUDGET_EXCEEDED`. Model receives consumption breakdown.

---

## OINV-05: Dangerous Ops Require Explicit Allow

**What it prevents:** Accidental execution of destructive operations.

**What it requires:** Any packet with `OP_FLAG_DANGEROUS` or safety level 0/1 requires explicit model confirmation. Auto-classifier CANNOT auto-allow. 2-vote for safety level 0.

**Source of this rule:**
- bun PR #30412: permission pipeline.
- AP-13 (override on deny).

**Consequence of violation:** Operation REJECTED with `ERR_PERMISSION_DENY`.

---

## OINV-06: Context Cumulative and Forward-Only

**What it prevents:** Stale context, circular references.

**What it requires:** Session context only APPENDS. Never modifies previous entries. Later phases see all earlier phase context. Context compression is append-only (new compressed blob replaces old, but old kept in history).

**Source of this rule:**
- embryo orchestrator: cumulative context flow.
- hermes-agent: context compression pipeline.

**Consequence of violation:** Context modification attempt → REJECTED. New context appended instead.

---

## OINV-07: Parallel Pipeline Isolation

**What it prevents:** Cross-pipeline contamination, race conditions.

**What it requires:** Parallel pipelines checked for resource_bitmask overlap before dispatch. Overlap → serialize. Max 10 concurrent pipelines. Each pipeline has independent ExecContext.

**Source of this rule:**
- bun PR #30412: ~170 agents with scoped isolation.
- CONFLICT_RULES.md.

**Consequence of violation:** Conflicting pipelines serialized automatically.

---

## OINV-08: Novel Patterns Recorded

**What it prevents:** Lost knowledge, repeated discoveries.

**What it requires:** If execution produces novel successful pattern (not in mesh), create feedback artifact. If precision ≥ 0.8 after distillation review, index as new slot.

**Source of this rule:**
- hermes-agent: closed learning loop.
- embryo hunt system.

**Consequence of violation:** Novelty not recorded → knowledge lost. No penalty, but missed optimization opportunity.

---

## OINV-09: Honest Response Always

**What it prevents:** Hidden failures, false confidence.

**What it requires:** RESPOND phase includes complete result: status, artifacts, metrics, validation report, partial results, failure reasons, rollback status. No omission of negative information.

**Source of this rule:**
- META_INVARIANT.md: complete honest response.
- gastown: zero false confidence.

**Consequence of violation:** Incomplete response detected by response schema validation. Forced to include missing fields.

---

## OINV-10: Circuit Break on Denials

**What it prevents:** Repeated failed attempts, wasted budget.

**What it requires:** 3 consecutive permission denials → circuit broken. Manual reset required. All ops return `ERR_CIRCUIT_BREAK` until reset.

**Source of this rule:**
- Standard circuit breaker pattern.
- AP-19 (toolloop eating iterations).

**Consequence of violation:** 4th denial triggers circuit break. Session locked.
