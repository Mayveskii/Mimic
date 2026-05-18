# Anti-Patterns Domain — What Not To Do

30 documented failures and their positive counter-patterns.

---

## What This Domain Does

Anti-patterns are NEGATIVE artifacts in the mesh. Each anti-pattern documents a failure mode: what went wrong, why, the evidence, and the counter-pattern that should be used instead. Every NEGATIVE artifact links to a COUNTER artifact (QAC-9). Models learn "what NOT to do" AND "what TO do instead."

---

## Catalog

### AP-01: Break-on-Failure in Cleanup

**What goes wrong:** Single PruneEpoch failure stops all subsequent epoch pruning → unbounded disk growth.

**Consequence:** Disk fills, service degrades, no recovery without manual intervention.

**Evidence:** gonka 86c686d92 — replace break-on-failure with continue+isolate pattern in cleanup().

**Counter-pattern:** gt_rollback_on_failure (gastown) — cleanupOnError rolls back all created resources.

**QAC violated:** QAC-1 (no validation PruneEpoch succeeded), QAC-2 (missing invariant).

---

### AP-02: Goroutine-Per-Item Under Mutex

**What goes wrong:** Spawning goroutines for each epoch prune while holding mutex → blocks Store/Retrieve.

**Consequence:** I/O blocked under lock, latency spike, potential deadlock.

**Evidence:** gonka d9c46cae0 — redesign cleanup() to prune sequentially outside mutex.

**Counter-pattern:** gt_pressure_gating (gastown) — checkPressure before dispatch.

**QAC violated:** QAC-4 (conflict), QAC-11 (cross-domain).

---

### AP-03: Panic Recovery Instead of Idempotent Close

**What goes wrong:** Using panic/recover to make Close() idempotent → non-deterministic.

**Consequence:** Real panics swallowed, double-close causes corruption.

**Evidence:** gonka 8d1cd00f3 — replace with sync.Once for idempotent Close().

**Counter-pattern:** gst_drain_mode (go-service-template-rest) — atomic.Bool draining flag.

**QAC violated:** QAC-2, QAC-7.

---

### AP-04: Unvalidated Input Before I/O

**What goes wrong:** Submitting new participant without validation → breaks validator downstream.

**Consequence:** Invalid data enters system, state inconsistency.

**Evidence:** gonka badcd8874.

**Counter-pattern:** gi_schema_validation (gitingest) — No I/O before schema validation passes.

**QAC violated:** QAC-2, QAC-9.

---

### AP-05: Context Injection Without Structure Stability

**What goes wrong:** Injecting context into buildMessages without guaranteeing structure stability → corruption.

**Consequence:** Corrupted messages, inference failures, full revert.

**Evidence:** binary-mesh c9e83c3 — 11 revisions to stabilize.

**Counter-pattern:** ha_prompt_caching (hermes-agent) — stable system prompt prefix.

**QAC violated:** QAC-6, QAC-10.

---

### AP-06: Undifferentiated Retry on All Errors

**What goes wrong:** Treating 429 same as 5xx → wastes iterations.

**Consequence:** Iteration budget consumed on unretryable errors.

**Evidence:** binary-mesh 0e36abd.

**Counter-pattern:** ha_error_classifier_failover — classify_api_error → retry/fallback/abort.

**QAC violated:** QAC-3, QAC-6.

---

### AP-07: Hardcoded Secrets in Source

**What goes wrong:** API keys committed in source code → leaked on push.

**Consequence:** Key rotation required, potential unauthorized access.

**Evidence:** binary-mesh ff5ff82, 308856b, b01d8e0.

**Counter-pattern:** ha_credential_pool (hermes-agent) — multi-key rotation with env fallback.

**QAC violated:** QAC-8, QAC-12.

---

### AP-08: Index Out of Range Without Guard

**What goes wrong:** Backoff index exceeds slice bounds → panic.

**Consequence:** Service crash.

**Evidence:** binary-mesh 4742339, c60c751.

**Counter-pattern:** cm_validate_then_accept (caveman) — no LLM output used without validation.

**QAC violated:** QAC-2, QAC-5.

---

### AP-09: Divide by Zero Without Check

**What goes wrong:** Division without zero-check → panic.

**Consequence:** Service crash during critical path.

**Evidence:** gonka 683e15051.

**Counter-pattern:** gt_atomic_allocation (gastown) — AllocateAndAdd with flock.

**QAC violated:** QAC-2, QAC-7.

---

### AP-10: Negative Coin Value Without Guard

**What goes wrong:** NewInt64Coin panics on negative values — no pre-check.

**Consequence:** Runtime panic in economic operations.

**Evidence:** gonka 75e31c233.

**Counter-pattern:** bun_two_vote_verify — 2-vote for critical operations.

**QAC violated:** QAC-2, QAC-7.

---

## Meta-Invariant

All 30 anti-patterns trace to violation of:

**no_side_effect_without_prior_validation**

See invariants/META_INVARIANT.md.

---

## Artifact Storage

Anti-patterns stored as NEGATIVE artifacts with linked COUNTER artifacts.

| Field | Value |
|-------|-------|
| domain | "anti-patterns" |
| layer | "decision" |
| modality | "text" |
| polarity | NEGATIVE |
| counter_pattern_id | ID of POSITIVE replacement |
| failure_evidence | source_commit + description |
| qac_violated | which QACs this failure violates |
