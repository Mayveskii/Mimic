# Quality Domain — Invariants

Rules that MUST hold for every process in the quality domain.

---

## QINV-01: 2-Vote on Critical Operations

**What it prevents:** Single point of failure, confirmation bias.

**What it requires:** Safety level 0 operations (CRITICAL) MUST have 2-vote verification: Verifier A checks correctness, Verifier B checks invariants. Both pass → VERIFIED. Both fail → REJECTED. Disagreement → tiebreak or manual escalation.

**Source of this rule:**
- bun PR #30412: 2-vote adversarial verify.
- AP-12 (single verifier).

**Consequence of violation:** Single-vote critical operation → REJECTED at validation.

---

## QINV-02: Conflict Check Pairwise

**What it prevents:** Conflicting operations in same chain, race conditions.

**What it requires:** Every chain validated with O(n²) pairwise conflict check. Any conflict > CONFLICT_NONE → REJECTED.

**Source of this rule:**
- CONFLICT_MATRIX_SPEC.md.
- bun PR #30412: resource bitmask overlap check.

**Consequence of violation:** Conflicting chain REJECTED with `ERR_CONFLICT`.

---

## QINV-03: Budget Check Before Execution

**What it prevents:** Runaway chains, budget exhaustion.

**What it requires:** Total estimated energy ≤ remaining budget before EXEC. Tokens and time checked separately.

**Source of this rule:**
- ENERGY_COST_SPEC.md.
- code-mode: budget enforcement.

**Consequence of violation:** Chain REJECTED with `ERR_BUDGET_EXCEEDED`.

---

## QINV-04: Permission Pipeline Enforced

**What it prevents:** Unauthorized dangerous operations.

**What it requires:** Permission pipeline: never-rules → classify → budget → allow. 3 consecutive denials → circuit break.

**Source of this rule:**
- bun PR #30412: permission pipeline.
- AP-13 (override on deny).

**Consequence of violation:** Unauthorized operation REJECTED with `ERR_PERMISSION_DENY`.

---

## QINV-05: Timeout on All Blocking Operations

**What it prevents:** Infinite hangs, dead sessions.

**What it requires:** Every blocking operation has timeout. Network: 30s. Git: 300s. Build: 3600s. Process: 300s.

**Source of this rule:**
- AP-14 (infinite wait).
- AP-30 (missing cancellation boundary).

**Consequence of violation:** Operation returns `ERR_TIMEOUT`.

---

## QINV-06: Rollback on Atomic Failure

**What it prevents:** Partial state corruption.

**What it requires:** Any chain with atomic op that fails → rollback to pre-chain state. Rollback procedure in ROLLBACK_SPEC.md.

**Source of this rule:**
- gastown: rollback on failure.
- META_INVARIANT.md.

**Consequence of violation:** If rollback fails → `ERR_ROLLBACK_FAIL`.

---

## QINV-07: Error Classification Before Retry

**What it prevents:** Blind retries, wasted budget.

**What it requires:** Errors classified: retryable (timeout, 5xx, DNS), permanent (4xx, auth, SSL), rate-limit (429). Only retryable errors retried.

**Source of this rule:**
- hermes-agent: error classifier.
- AP-06 (undifferentiated retry).

**Consequence of violation:** Wrong retry → `retry_count` wasted. After 3 wrong retries → circuit break.

---

## QINV-08: Honest Reporting

**What it prevents:** Hidden failures, false confidence.

**What it requires:** All results include: status, artifacts, metrics, validation report, partial results (if any), failure reasons, rollback status.

**Source of this rule:**
- gastown: zero false confidence.
- META_INVARIANT.md.

**Consequence of violation:** Incomplete response → schema validation forces inclusion.

---

## QINV-09: Quality Gate Thresholds

**What it prevents:** Low-quality patterns in mesh.

**What it requires:** artifact_precision ≥ 0.8 for indexing. ≤ 3 QAC failures. survival_index > 0.0.

**Source of this rule:**
- Specs-v2 artifact schema.
- embryo checkpoints.

**Consequence of violation:** Fails quality gate → archived, not indexed.

---

## QINV-10: Measured Thresholds from Real Data

**What it prevents:** Made-up thresholds that don't match reality.

**What it requires:** Quality thresholds derived from measured data on real projects. binary-mesh server metrics: L0 alert < 7000, L4 < 1500, L6 < 1500, L8 < 7500, L9 < 6000.

**Source of this rule:**
- binary-mesh measurements.
- embryo checkpoints.

**Consequence of violation:** Thresholds not calibrated → false positives/negatives. Recalibration required.
