# Anti-Patterns — Mimic

Negative artifacts distilled from proven failures across binary-mesh, gonka, and all 21 Mayveskii/* source repositories. Every anti-pattern links to a positive counter-pattern. Polarity: NEGATIVE.

---

## Schema

```yaml
anti_pattern:
  id: ap_<snake_case>
  what_goes_wrong: <one-line factual description>
  consequence: <what breaks>
  evidence: <source_commit or SuperInvariant>
  counter_pattern: <id of POSITIVE pattern that replaces this>
  counter_source: <which repo/spec defines the counter>
  qac_violated: <which QAC(s) this anti-pattern violates>
```

---

## AP-01: Break-on-Failure in Cleanup

```yaml
id: ap_break_on_failure_cleanup
what_goes_wrong: "Single PruneEpoch failure stops all subsequent epoch pruning → unbounded disk growth"
consequence: "Disk fills, service degrades, no recovery without manual intervention"
evidence: "gonka 86c686d92 — replace break-on-failure with continue+isolate pattern in cleanup()"
counter_pattern: gt_rollback_on_failure
counter_source: Mayveskii/gastown (cleanupOnError)
qac_violated: QAC-1, QAC-2
```

## AP-02: Goroutine-Per-Item Under Mutex

```yaml
id: ap_goroutine_per_item_under_mutex
what_goes_wrong: "Spawning goroutines for each epoch prune while holding mutex → blocks Store/Retrieve for entire cleanup duration"
consequence: "I/O blocked under lock, latency spike, potential deadlock"
evidence: "gonka d9c46cae0 — redesign cleanup() to prune sequentially outside mutex"
counter_pattern: gt_pressure_gating
counter_source: Mayveskii/gastown (checkPressure before dispatch)
qac_violated: QAC-4, QAC-11
```

## AP-03: Panic Recovery Instead of Idempotent Close

```yaml
id: ap_panic_recovery_close
what_goes_wrong: "Using panic/recover to make Close() idempotent → non-deterministic, not thread-safe, hides real panics"
consequence: "Real panics swallowed, double-close causes corruption, concurrent Close races"
evidence: "gonka 8d1cd00f3 — replace panic recovery with sync.Once for idempotent Close()"
counter_pattern: gst_drain_mode
counter_source: Mayveskii/go-service-template-rest (atomic.Bool draining flag)
qac_violated: QAC-2, QAC-7
```

## AP-04: Unvalidated Input Before I/O

```yaml
id: ap_unvalidated_input_before_io
what_goes_wrong: "Submitting new participant from dapi without validation → breaks validator downstream"
consequence: "Invalid data enters system, validator rejects, state inconsistency"
evidence: "gonka badcd8874 — don't submit new participant from dapi"
counter_pattern: gi_schema_validation
counter_source: Mayveskii/gitingest (No I/O operation started before schema validation passes)
qac_violated: QAC-2, QAC-9
```

## AP-05: Context Injection Without Structure Stability

```yaml
id: ap_context_injection_unstable
what_goes_wrong: "Injecting context into buildMessages without guaranteeing message structure stability → corruption"
consequence: "Corrupted messages, inference failures, requires full revert"
evidence: "binary-mesh c9e83c3 — revert to working state before buildMessages corruption; 11 revisions to stabilize"
counter_pattern: ha_prompt_caching
counter_source: Mayveskii/hermes-agent (stable system prompt prefix, never compressed)
qac_violated: QAC-6, QAC-10
```

## AP-06: Undifferentiated Retry on All Errors

```yaml
id: ap_undifferentiated_retry
what_goes_wrong: "Treating 429 rate-limit same as 5xx server error → wastes iterations on rate-limited endpoint"
consequence: "Iteration budget consumed on unretryable errors, valid retries starved"
evidence: "binary-mesh 0e36abd — separate 429 rate limit from 5xx server error backoff"
counter_pattern: ha_error_classifier_failover
counter_source: Mayveskii/hermes-agent (classify_api_error → retry/fallback/abort)
qac_violated: QAC-3, QAC-6
```

## AP-07: Hardcoded Secrets in Source

```yaml
id: ap_hardcoded_secrets
what_goes_wrong: "API keys (sk-, gp-) committed directly in source code → leaked on push"
consequence: "Key rotation required, potential unauthorized access, 3+ commits to fix"
evidence: "binary-mesh ff5ff82, 308856b — remove hardcoded INFERENCE_KEY, use env vars; b01d8e0 — fresh key after leak"
counter_pattern: ha_credential_pool
counter_source: Mayveskii/hermes-agent (credential pool with env fallback, rate-limited keys removed from pool)
qac_violated: QAC-8, QAC-12
```

## AP-08: Index Out of Range Without Guard

```yaml
id: ap_index_out_of_range
what_goes_wrong: "Backoff index calculation exceeds slice bounds → panic"
consequence: "Service crash, no graceful degradation"
evidence: "binary-mesh 4742339, c60c751 — panic in callInference backoff index out of range [5] with length 5"
counter_pattern: cm_validate_then_accept
counter_source: Mayveskii/caveman (no LLM output used without validation pass)
qac_violated: QAC-2, QAC-5
```

## AP-09: Divide by Zero Without Check

```yaml
id: ap_divide_by_zero
what_goes_wrong: "Division without zero-check in model assignment → panic"
consequence: "Service crash during critical path"
evidence: "gonka 683e15051 — fix: del by zero in inference-chain model_assignment.go"
counter_pattern: gt_atomic_allocation
counter_source: Mayveskii/gastown (AllocateAndAdd with flock + pending marker)
qac_violated: QAC-2, QAC-7
```

## AP-10: Negative Coin Value Without Guard

```yaml
id: ap_negative_coin_panic
what_goes_wrong: "sdk.NewInt64Coin panics on negative values — no pre-check"
consequence: "Runtime panic in economic operations, potential fund loss"
evidence: "gonka 75e31c233 — fix possible panics for negative coins, add Safe method for logging"
counter_pattern: bun_two_vote_verify
counter_source: Mayveskii/bun (2-vote adversarial verification for critical operations)
qac_violated: QAC-2, QAC-7
```

## AP-11: Stale Cached State

```yaml
id: ap_stale_cached_state
what_goes_wrong: "Trusting cached state files instead of observable reality → decisions based on stale data"
consequence: "Agent operates on outdated information, wrong decisions propagate"
evidence: "Mayveskii/gastown internal/witness/ — Zero False Confidence born from this failure"
counter_pattern: gt_zfc_state
counter_source: Mayveskii/gastown (tmux session is source of truth, not state files)
qac_violated: QAC-1, QAC-10
```

## AP-12: Single Verifier for Critical Operations

```yaml
id: ap_single_verifier_critical
what_goes_wrong: "One verifier for critical operations (git push, deploy, encrypt) → single point of failure"
consequence: "Bad deployment, data loss, security breach"
evidence: "Mayveskii/bun PR #30412 — two_vote_verify pattern explicitly designed to prevent this"
counter_pattern: bun_two_vote_verify
counter_source: Mayveskii/bun (2-vote adversarial verification + tiebreak)
qac_violated: QAC-2, QAC-7
```

## AP-13: Override on Deny Rules

```yaml
id: ap_override_deny_rules
what_goes_wrong: "Allowing override of deny rules via higher permission → dangerous ops executed anyway"
consequence: "Destructive operations (git reset, force push) executed despite safety rules"
evidence: "Mayveskii/bun PR #30412 — never-rules cannot be overridden by any permission mode; Mayveskii/code-mode — deny rules checked first, no override"
counter_pattern: bun_never_rules
counter_source: Mayveskii/bun (never git reset/checkout/restore/stash/rebase, violation = hard stop)
qac_violated: QAC-6, QAC-7
```

## AP-14: Infinite Wait on Dead Connection

```yaml
id: ap_infinite_wait_dead_connection
what_goes_wrong: "No timeout on streaming API calls → hangs forever on zombie provider"
consequence: "Agent frozen, no recovery, session wasted"
evidence: "Mayveskii/hermes-agent — 90s stale-stream detection + 60s read timeout"
counter_pattern: ha_streaming_health_check
counter_source: Mayveskii/hermes-agent (90s stale-stream detection + 60s read timeout)
qac_violated: QAC-3, QAC-5
```

## AP-15: Partial Rollback Leaving Orphaned Resources

```yaml
id: ap_partial_rollback_orphans
what_goes_wrong: "Cleanup only partially rolls back → orphaned resources leak"
consequence: "Resource leak under load, disk/memory exhaustion"
evidence: "Mayveskii/gastown — partial failure leaves no orphaned resources, rollback is complete or explicitly partial with log"
counter_pattern: gt_rollback_on_failure
counter_source: Mayveskii/gastown (cleanupOnError rolls back ALL created resources)
qac_violated: QAC-1, QAC-2
```

## AP-16: Flip-Flop Decision Without Evidence

```yaml
id: ap_flip_flop_decision
what_goes_wrong: "Reverting a decision and then re-reverting within minutes → no measurement, no evidence"
consequence: "Wasted iterations, instability, team confusion"
evidence: "binary-mesh 87c27d8→c03b244 — Revert then Revert of 'forbid shell exploration' within 2 minutes"
counter_pattern: vllm_measured_optimization_decision
counter_source: Mayveskii/vllm (every perf change must show measured before/after on real hardware)
qac_violated: QAC-6, QAC-13
```

## AP-17: Rag.Save Blocking Startup

```yaml
id: ap_blocking_save_on_startup
what_goes_wrong: "1GB JSON write blocking server startup for 10+ minutes"
consequence: "Service unavailable, health check failures, deployment rollback"
evidence: "binary-mesh 0e3165f — remove rag.Save() from startup; 1348d52 — skip JSON load when SQLite has data (15min→30s)"
counter_pattern: gst_bootstrap_lifecycle
counter_source: Mayveskii/go-service-template-rest (phased startup with budgets, total=30s)
qac_violated: QAC-3, QAC-5
```

## AP-18: Grep-Based Workspace Scan

```yaml
id: ap_grep_workspace_scan
what_goes_wrong: "grep -rl workspace scan for feature detection → 3 minutes per call"
consequence: "Every Solve() takes 3+ minutes, agent unusable"
evidence: "binary-mesh 05b31d8 — DumbIndex.DetectFeatures() replaces grep -rl (3min → O(1))"
counter_pattern: emb_projectmap_sqlite
counter_source: Mayveskii/embryo (SQLite FTS5, updated after every WRITE opcode)
qac_violated: QAC-3, QAC-5
```

## AP-19: Toolloop Eating Iterations on 429

```yaml
id: ap_toolloop_429_eating_iterations
what_goes_wrong: "429 rate-limit response decrements iteration counter → budget consumed without progress"
consequence: "Agent stops prematurely, task incomplete"
evidence: "binary-mesh 675601e — toolloop 429 iter-- was eating iterations, fixed with separate handling"
counter_pattern: ha_error_classifier_failover
counter_source: Mayveskii/hermes-agent (classify_api_error distinguishes retryable from non-retryable)
qac_violated: QAC-3, QAC-6
```

## AP-20: Regenerate Seed Without Validation

```yaml
id: ap_regenerate_seed_without_validation
what_goes_wrong: "Regenerating seed in distributed system without validating impact on existing state"
consequence: "State inconsistency across nodes, full revert required"
evidence: "gonka 80b2598b8 — Revert 'regenerate seed (#375)' (#378)"
counter_pattern: emb_orchestrator_pipeline
counter_source: Mayveskii/embryo (7-stage pipeline with validation before execution)
qac_violated: QAC-2, QAC-13
```

## AP-21: Missing Node ID in Batch Processing

```yaml
id: ap_missing_node_id_batch
what_goes_wrong: "Processing batches without node_id → legacy weight distribution without attribution"
consequence: "Incorrect reward/power distribution, security vulnerability"
evidence: "gonka 2002c9f37 — Multiple Fixes: add node_id to PoCBatch, remove legacy weight distribution"
counter_pattern: rustnet_immutable_after_set
counter_source: Mayveskii/rustnet (first-writer-wins for identity fields, prevents attribution conflicts)
qac_violated: QAC-2, QAC-7
```

## AP-22: Unbounded Queue Without Once

```yaml
id: ap_unbounded_queue_hacky_close
what_goes_wrong: "UnboundedQueue.Close() uses panic/recover instead of sync.Once → not thread-safe, not idempotent"
consequence: "Concurrent Close calls corrupt queue state, data loss"
evidence: "gonka 8d1cd00f3 — replace hacky panic recovery with sync.Once for idempotent Close"
counter_pattern: gst_drain_mode
counter_source: Mayveskii/go-service-template-rest (atomic.Bool draining flag, graceful shutdown)
qac_violated: QAC-2, QAC-7
```

## AP-23: Token Budget Overflow From System Prompt

```yaml
id: ap_system_prompt_token_overflow
what_goes_wrong: "Too many tokens in system prompt → rate limit from inference provider"
consequence: "Rate-limited, agent cannot proceed"
evidence: "binary-mesh 3620f8f — remove outline/RAG/topDirs from system prompt, too many tokens → rate limit"
counter_pattern: ha_context_compression
counter_source: Mayveskii/hermes-agent (preflight token detection + multi-pass compression)
qac_violated: QAC-3, QAC-5
```

## AP-24: BuildMessages Corruption

```yaml
id: ap_buildmessages_corruption
what_goes_wrong: "buildMessages function mutates shared state → message structure corruption"
consequence: "Inference calls fail, full restore required (c9e83c3)"
evidence: "binary-mesh c9e83c3 — revert to working state before buildMessages corruption; 11 revisions to fix"
counter_pattern: emb_binary_runtime
counter_source: Mayveskii/embryo (OpCode-based execution, no shared mutable message state)
qac_violated: QAC-4, QAC-6, QAC-10
```

## AP-25: Race in gettxoutsetinfo

```yaml
id: ap_race_condition_shared_state
what_goes_wrong: "Concurrent access to shared state without synchronization in gettxoutsetinfo"
consequence: "Incorrect blockchain state reported, consensus risk"
evidence: "bitcoin/bitcoin #34451 (from enricher) — fix race condition in gettxoutsetinfo"
counter_pattern: gt_atomic_allocation
counter_source: Mayveskii/gastown (AllocateAndAdd with flock + pending marker prevents TOCTOU)
qac_violated: QAC-4, QAC-11
```

## AP-26: Out-of-Bound Write Without Guard

```yaml
id: ap_oob_write
what_goes_wrong: "Writing beyond allocated buffer bounds → memory corruption"
consequence: "Security vulnerability, potential code execution"
evidence: "envoyproxy/envoy #44030 (from enricher) — Prevent out-of-bound writes"
counter_pattern: git_content_addressable_store
counter_source: Mayveskii/git (SHA-1 integrity verified on every read, no silent corruption)
qac_violated: QAC-2, QAC-8
```

## AP-27: Flaky Test Without Deterministic Fix

```yaml
id: ap_flaky_test_without_fix
what_goes_wrong: "Skipping flaky tests instead of fixing root cause → tests lose value"
consequence: "CI passes with real bugs, false confidence"
evidence: "Multiple: prometheus #18406, next.js #92162 #92199 #92198 — all skip flaky tests instead of fixing"
counter_pattern: vllm_measured_optimization_decision
counter_source: Mayveskii/vllm (every change must show measured before/after)
qac_violated: QAC-2, QAC-7
```

## AP-28: Missing Overflow Guard in Arithmetic

```yaml
id: ap_missing_overflow_guard
what_goes_wrong: "Arithmetic operations without overflow checking → integer overflow → incorrect values"
consequence: "Fund loss, incorrect rewards, consensus failure"
evidence: "SuperInvariant: 'arithmetic operation requires overflow guard' — 1140 members across 8 domains"
counter_pattern: bun_two_vote_verify
counter_source: Mayveskii/bun (critical operations always undergo 2-vote, including arithmetic results)
qac_violated: QAC-2, QAC-7
```

## AP-29: Swallowed Error Condition

```yaml
id: ap_swallowed_error
what_goes_wrong: "Error condition silently swallowed instead of surfaced → caller unaware of failure"
consequence: "Silent data loss, incorrect state, debugging nightmare"
evidence: "SuperInvariant: 'error condition must be surfaced not swallowed' — 1964 members; gonka #874 — ClaimRewards error swallowing"
counter_pattern: ha_error_classifier_failover
counter_source: Mayveskii/hermes-agent (classify_api_error → retry/fallback/abort, never silent)
qac_violated: QAC-2, QAC-7
```

## AP-30: Missing Cancellation Boundary

```yaml
id: ap_missing_cancellation_boundary
what_goes_wrong: "Long-running operation without timeout or cancellation → blocks forever"
consequence: "Agent frozen, resource exhaustion, no recovery"
evidence: "SuperInvariant: 'long-running operation requires cancellation boundary' — 590 members"
counter_pattern: gi_async_timeout_bounded
counter_source: Mayveskii/gitingest (async_timeout decorator, timeout → partial result + warning)
qac_violated: QAC-3, QAC-5
```

---

## Summary: Anti-Pattern to QAC Mapping

| QAC | Anti-Patterns | Count |
|-----|--------------|-------|
| QAC-1 | AP-01, AP-11, AP-15 | 3 |
| QAC-2 | AP-01, AP-03, AP-04, AP-08, AP-09, AP-10, AP-15, AP-20, AP-21, AP-22, AP-26, AP-27, AP-28, AP-29 | 14 |
| QAC-3 | AP-06, AP-07, AP-14, AP-17, AP-18, AP-19, AP-23, AP-30 | 8 |
| QAC-4 | AP-02, AP-24, AP-25 | 3 |
| QAC-5 | AP-08, AP-14, AP-17, AP-18, AP-23, AP-30 | 6 |
| QAC-6 | AP-05, AP-06, AP-13, AP-16, AP-19, AP-24 | 6 |
| QAC-7 | AP-03, AP-07, AP-09, AP-10, AP-12, AP-13, AP-21, AP-22, AP-26, AP-27, AP-28, AP-29 | 12 |
| QAC-8 | AP-07, AP-26 | 2 |
| QAC-9 | AP-04 | 1 |
| QAC-10 | AP-05, AP-11, AP-24 | 3 |
| QAC-11 | AP-02, AP-25 | 2 |
| QAC-12 | AP-07 | 1 |
| QAC-13 | AP-16, AP-20 | 2 |

Most violated: QAC-2 (Invariant Coverage, 14 anti-patterns), QAC-7 (Artifact Precision, 12), QAC-3 (Energy Cost, 8).
This confirms: the dominant failure mode is **missing validation before action** (QAC-2) and **unverified precision** (QAC-7).

---

## Domain Gap: 5 New Anti-Pattern Domains

| Domain | Anti-Patterns | Sources | Expected Yield |
|--------|-------------|---------|---------------|
| resource_cleanup_under_lock | AP-01, AP-02, AP-15 | gonka, binary-mesh | 20+ slots |
| context_structure_stability | AP-05, AP-24, AP-23 | binary-mesh (11 buildMessages revisions) | 15+ slots |
| input_validation_before_io | AP-04, AP-09, AP-10, AP-21 | gonka, bitcoin, envoy | 30+ slots |
| idempotent_close_cleanup | AP-03, AP-22 | gonka | 10+ slots |
| economic_invariant_enforcement | AP-09, AP-10, AP-21, AP-28 | gonka, bitcoin | 25+ slots |
