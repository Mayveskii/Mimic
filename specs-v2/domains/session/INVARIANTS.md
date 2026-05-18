# Session Domain — Invariants

Rules that MUST hold for every process in the session domain.

---

## SINV-01: Budget Tracked Accurately

**What it prevents:** Budget overruns, undefined behavior when budget exhausted.

**What it requires:** `session_budget_tokens` and `session_budget_time_ms` decremented on every operation. Validation checks total energy ≤ remaining before EXEC. No operation starts if budget insufficient.

**Source of this rule:**
- code-mode: maxTurns + maxBudgetUsd.
- OINV-04 (budget never exceeded).

**Consequence of violation:** Chain REJECTED with `ERR_BUDGET_EXCEEDED`.

---

## SINV-02: Denial Count Accurate

**What it prevents:** Circuit break miscalculation, false resets.

**What it requires:** `denial_count` incremented on every `ERR_PERMISSION_DENY`. Reset to 0 on any successful operation. Never negative.

**Source of this rule:**
- Standard circuit breaker.
- OINV-10 (circuit break on denials).

**Consequence of violation:** Incorrect denial count → premature or delayed circuit break.

---

## SINV-03: Context Append-Only

**What it prevents:** Context corruption, circular references, tampering with history.

**What it requires:** Session context ONLY appends new entries. Previous entries immutable. Compression creates new blob, old kept in history. Total context size bounded (default 1MB, configurable).

**Source of this rule:**
- hermes-agent: context compression pipeline.
- OINV-06 (context cumulative and forward-only).

**Consequence of violation:** Context modification attempt → REJECTED. New entry appended instead.

---

## SINV-04: Snapshot Integrity Verified

**What it prevents:** Corrupted snapshots, failed recovery.

**What it requires:** OP_SESS_SNAPSHOT writes compressed state + SHA-256 hash. On read, hash verified. Mismatch → reject snapshot, log corruption.

**Source of this rule:**
- Standard backup integrity.
- UINV-01 (hash deterministic).

**Consequence of violation:** Snapshot REJECTED. Previous snapshot used (if any). If no valid snapshot → session starts fresh.

---

## SINV-05: Compression Preserves Semantics

**What it prevents:** Context loss from over-compression.

**What it requires:** Context compression preserves: all operations performed, all results, all denials, all budget changes. Compressed context must be decompressible and re-parsable.

**Source of this rule:**
- hermes-agent: multi-pass compression with stable prefix.
- Data integrity principle.

**Consequence of violation:** Decompression failure → use previous snapshot. Log compression error.

---

## SINV-06: Session Isolation

**What it prevents:** Cross-session data leakage, resource conflicts.

**What it requires:** Each session has independent context, budget, resource tracking. Sessions do not share mutable state. Read-only shared data (mesh slots) is safe.

**Source of this rule:**
- Standard multi-tenant security.
- bun PR #30412: edit scope isolation.

**Consequence of violation:** Cross-session access attempt → REJECTED with `ERR_PERMISSION_DENY`.

---

## SINV-07: Budget Warnings Before Exhaustion

**What it prevents:** Sudden halt mid-operation.

**What it requires:** When budget < 20% remaining, warning appended to context. When < 10%, CRITICAL warning. Model receives warning in RESPOND.

**Source of this rule:**
- UX best practice.
- code-mode: budget visibility.

**Consequence of violation:** No warning. Budget exhausted unexpectedly. Model surprised by halt.

---

## SINV-08: Denial Reason Logged

**What it prevents:** Repeated identical denials, debugging difficulty.

**What it requires:** Every denial logged with: reason, operation attempted, denied_by (rule), suggested_alternative (if known).

**Source of this rule:**
- Debugging best practice.
- META_INVARIANT.md: complete honest response.

**Consequence of violation:** Missing denial log → debugging harder. No functional impact.
