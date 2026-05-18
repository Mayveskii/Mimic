# Session Domain — State, Budget, Context

How Mimic maintains state across model interactions.

---

## What This Domain Does

Every model interaction with Mimic happens within a session. The session tracks budget (tokens, time), denials (permissions refused), context flow (what happened in previous phases), and state across the 6-phase pipeline. Sessions compress at rest and verify on restore.

The domain covers: session creation, budget tracking, denial counting, context accumulation, session expiry.

---

## Processes

### session_create

**When to use:**  
First time a model calls Mimic in a conversation.

**Goal:**  
Create a new session with budget, tracking, and context storage.

**Chain (semantically):**

1. Generate session ID.
2. Set budget:
   - budget_tokens: configurable (default: model-dependent).
   - budget_time_ns: configurable (default: 1 hour).
3. Initialize counters:
   - operations_executed: 0.
   - consecutive_denials: 0.
   - total_denials: 0.
4. Initialize context flow:
   - classified_intent: null.
   - planned_chain: null.
   - validated_chain: null.
   - execution_result: null.
   - verify_result: null.
   - response: null.

**Invariants:**
- Every session has positive budget on creation.
- Every session has unique ID.

---

### budget_check

**When to use:**  
Before any operation execution.

**Goal:**  
Verify that operation cost fits within remaining budget.

**Chain (semantically):**

1. Compute estimated cost: Σ(cost_tokens × cost_time_us) for planned chain.
2. Compare: estimated ≤ remaining.
3. If exceeds → REJECT chain, notify model with breakdown.
4. If fits → proceed, reserve budget.

**Hard constraints:**
- Operation NEVER executes if estimated cost > remaining budget.
- Actual cost tracked during execution.
- If actual cost > 1.2× estimated → abort, rollback, record overestimate.

**Invariants:**
- Budget never negative.
- Model notified of budget state before every turn.
- Both token and time limits enforced independently.

---

### denial_track

**When to use:**  
After any permission check that results in denial.

**Goal:**  
Track denials, trigger circuit break if thresholds exceeded.

**Chain (semantically):**

1. Increment consecutive_denials.
2. Increment total_denials.
3. Check thresholds:
   - consecutive_denials ≥ 3 → circuit break to manual mode.
   - total_denials ≥ 20 → permanent manual mode for session.
4. Notify model of denial reason and current state.

**Hard constraints:**
- Circuit break → manual mode only. No auto-approval until model explicitly resets.
- Permanent manual mode → cannot be overridden within session.

---

### context_flow

**When to use:**  
Between every phase of the 6-phase pipeline.

**Goal:**  
Pass enriched context from one phase to the next, forward only.

**Chain (semantically):**

1. After CLASSIFY → classified_intent passed to PLAN.
2. After PLAN → planned_chain + estimated_cost passed to VALIDATE.
3. After VALIDATE → validated_chain + ValidationResult passed to EXEC.
4. After EXEC → execution_result + ExecContext passed to VERIFY.
5. After VERIFY → verify_result passed to RESPOND.
6. After RESPOND → full context compressed and stored.

**Invariants:**
- Context flows forward only.
- Rollback signals carry failure context backward (VALIDATE fail → PLAN gets reason).
- Every phase sees everything from all previous phases.
- No phase receives context from a phase that hasn't executed.

---

### session_persist

**When to use:**  
On session end or timeout.

**Goal:**  
Compress session state, verify hash, write to disk.

**Chain (semantically):**

1. Compress: `OP_COMPRESS_GZIP(session_state)`.
2. Hash: `sha256_hash(compressed)`.
3. Write to disk.
4. Verify on restore: hash matches.

**Invariants:**
- Compressed state verified by hash.
- No unverified restore.

---

## Principles From Sources

### code-mode

**Principles taken:**
- Budget: maxTurns + maxBudgetUsd per session.
- Denial tracking: 3 consecutive → circuit break; 20 total → permanent manual.

**What Mimic does with them:**
Exact thresholds. Exact behavior. No override.

### hermes-agent

**Principles taken:**
- IterationBudget.consume() gates loop.
- Budget exhausted → one grace call → stop.

**What Mimic does with them:**
Budget consumed per operation. Grace call = one extra operation after budget exhausted, to allow proper shutdown.

### go-service-template-rest

**Principles taken:**
- Phased startup with budgets (config=10s, probe=15s, telemetry=2s, total=30s).
- Graceful shutdown with drain.

**What Mimic does with them:**
Session lifecycle has phases with time budgets. Shutdown drains pending operations.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "session" |
| layer | "state" |
| modality | "metric" |
| pattern_name | "session_budget" / "denial_tracking" / "context_flow" |

---

## Cross-Domain Conflicts

Session domain coordinates all other domains. Budget limits apply across all domains.
