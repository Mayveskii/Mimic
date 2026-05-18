# Meta-Invariant — No Side Effect Without Prior Validation

The single rule that unifies all quality gates and prevents all documented failures.

---

## What This Invariant Does

Every side effect (I/O, state mutation, resource allocation, network call, memory write) MUST be preceded by validation that the operation is safe, within budget, and not conflicting.

This is not a suggestion. It is the root cause analysis of every historical failure across binary-mesh, gonka, and all source repositories.

---

## What It Prevents

All 30 documented anti-patterns trace to a violation of this invariant:

| Anti-Pattern | What Went Wrong | Missing Validation |
|---|---|---|
| AP-01: break-on-failure cleanup | PruneEpoch failure stops all cleanup → unbounded disk | No validation PruneEpoch succeeded before advancing |
| AP-02: goroutine-per-item under mutex | Spawning goroutines under mutex → deadlock | No validation mutex free before spawning |
| AP-03: panic recovery instead of idempotent close | panic/recover for Close → non-deterministic | No validation Close is idempotent |
| AP-04: unvalidated input before I/O | Invalid data enters system → state corruption | No validation input before processing |
| AP-05: context injection without stability | buildMessages corruption → full revert needed | No validation structure stable before injection |
| AP-06: undifferentiated retry | 429 treated as 5xx → iteration budget consumed | No classification before retry decision |
| AP-07: hardcoded secrets | API keys committed to source → leaked | No validation secrets not in source |
| AP-08: index out of range | Backoff index exceeds bounds → panic | No bounds check before array access |
| AP-09: divide by zero | Division without zero-check → panic | No validation divisor ≠ 0 |
| AP-10: negative coin panic | NewInt64Coin panics on negative → runtime crash | No validation input before NewInt64Coin |
| AP-11: stale cached state | Decisions based on stale data → wrong actions | No validation state is current |
| AP-12: single verifier | One verifier for critical ops → single point of failure | No validation with 2-vote |
| AP-13: override on deny | Dangerous ops executed despite safety rules | No validation deny rules are absolute |
| AP-14: infinite wait | No timeout on streaming → hangs forever | No validation timeout set |
| AP-15: partial rollback | Cleanup only partial → orphaned resources | No validation rollback is complete |
| AP-16: flip-flop decision | Revert then re-revert within minutes | No validation decision has evidence |
| AP-17: blocking save | 1GB JSON write blocking startup 10+ min | No validation save is non-blocking |
| AP-18: grep workspace scan | grep -rl takes 3 minutes per call | No validation index exists |
| AP-19: toolloop eating iterations | 429 decrements iteration counter → no progress | No validation rate-limit handling |
| AP-20: regenerate seed without validation | Seed regeneration breaks existing state | No validation impact on existing state |
| AP-21: missing node ID | Batches without node_id → incorrect attribution | No validation node_id present |
| AP-22: unbounded queue hacky close | panic/recover for queue Close → data loss | No validation Close is thread-safe |
| AP-23: token overflow | Too many tokens in prompt → rate limit | No validation token count before call |
| AP-24: buildMessages corruption | Shared state mutation → message corruption | No validation shared state locked |
| AP-25: race in shared state | Concurrent access without synchronization | No validation lock held before access |
| AP-26: out-of-bound write | Writing beyond buffer → memory corruption | No validation bounds before write |
| AP-27: flaky test without fix | Tests skipped instead of fixed → false confidence | No validation test is deterministic |
| AP-28: missing overflow guard | Integer overflow → incorrect values | No validation before arithmetic |
| AP-29: swallowed error | Error silently swallowed → silent data loss | No validation error is surfaced |
| AP-30: missing cancellation boundary | Long operation blocks forever | No validation timeout exists |

---

## What It Requires

Before ANY side effect:

1. **Validate target exists and is accessible.**
   - File exists? Path is within workspace? Not sensitive?

2. **Validate operation is within budget.**
   - Estimated cost ≤ remaining budget? Both tokens and time?

3. **Validate operation does not conflict.**
   - conflict_matrix[op_i][op_j] = 0 for all pairs?
   - resource_bitmask does not overlap with in-flight operations?

4. **Validate operation maintains invariants.**
   - All scenario invariants satisfiable with current state?
   - No invariant would be violated by this operation?

5. **Validate permission.**
   - Never-rules do not block?
   - Permission pipeline returns allow?
   - If dangerous → explicit confirmation obtained?

6. **Validate state is current.**
   - Not relying on cached/stale state?
   - Observable reality matches assumptions?

---

## How It Is Checked

**Phase:** VALIDATE (before EXEC).

**Mechanism:**
- ops_validate_chain() called BEFORE ops_execute_chain().
- Pairwise conflict check: O(n²) for chain of n operations.
- Budget check: sum of costs ≤ remaining.
- Permission check: deny rules → classify → budget → allow.
- State freshness check: snapshot_diff(current, indexed).

**Failure handling:**
- Any validation failure → chain REJECTED.
- Model receives exact reason: which validation failed, which rule violated, what would have happened.
- State untouched. No partial execution.

---

## Why It Works

The invariant does not prevent failure. It prevents **unvalidated failure**.

When a failure occurs after validation:
- The failure is caught by VERIFY phase (2-vote).
- The failure is rolled back to pre-execution state.
- The failure is recorded as feedback artifact.
- The model receives honest report: "tried X, validation passed, execution failed, rolled back, here's why."

Without the invariant, failures happen silently or partially, corrupting state, consuming budget, and leaving the model unaware until much later.

---

## Cross-Domain Application

This invariant applies to ALL domains:
- **Git:** validate working tree before commit, validate fast-forward before merge.
- **Build:** validate compile before test, validate test pass before deploy.
- **IO:** validate file exists before read, validate path safe before write.
- **Network:** validate URL safe before fetch, validate timeout before call.
- **Process:** validate command safe before spawn, validate resource available before alloc.
- **Memory:** validate size safe before mmap, validate pointer valid before free.
- **System:** validate env safe before set, validate destination free before move.

No domain exempt. No operation exempt.
