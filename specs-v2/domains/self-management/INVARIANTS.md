# Self-Management Domain — Invariants

Every invariant in the self-management domain.

---

## SM-INV-1: Pre-Destructive Checkpoint

A checkpoint MUST be taken before any destructive operation.

```
validation: destructive_op.preceded_by(OP_SELF_CHECKPOINT_CREATE) within last 60s
violation: no checkpoint before destructive mutation → state loss on failure
```

---

## SM-INV-2: Budget Ceiling

Budget reallocation MUST not exceed total session budget.

```
validation: sum(new_subtask_budgets) <= MIMIC_BUDGET_TOKENS (total)
violation: over-allocation → mid-chain abort on final subtask
```

---

## SM-INV-3: Pivot Diversity

Strategy pivot MUST query at least 3 alternative approaches before selecting.

```
validation: alternatives_queried >= 3 AND alternatives_queried >= mesh_results_count / 2
violation: selecting first alternative → may pick bad pattern
```

---

## SM-INV-4: Assessment Frequency

Progress assessment MUST run at every checkpoint + on-demand trigger.

```
validation: last_assessment_timestamp <= last_checkpoint_timestamp + 1s
violation: missed assessment → divergence/stall goes undetected
```

---

## SM-INV-5: Invariant Preservation

Context summarization MUST preserve all invariants and decisions.

```
validation: summary contains all invariant_ids from original context AND all decision_records
violation: lost invariant → agent violates rule it no longer remembers
```

---

## SM-INV-6: Decision Logging

Every self-management decision MUST be logged with timestamp and rationale.

```
validation: decision_log_entry exists for every OP_SELF_* opcode with fields: timestamp, opcode, trigger, rationale, result
violation: opaque decisions → non-debuggable failures
```

---

## SM-INV-7: No Infinite Pivot

Strategy pivot MUST not occur more than MIMIC_STRATEGY_PIVOT_THRESHOLD consecutive times on same subtask.

```
validation: pivot_count(subtask_id) <= MIMIC_STRATEGY_PIVOT_THRESHOLD OR (pivot_count > threshold AND human_approval == true)
violation: infinite loop of approach switching → no progress
```

---

## SM-INV-8: Checkpoint Verification

Restored checkpoint MUST hash-match stored hash.

```
validation: sha256(checkpoint_data) == checkpoint.stored_hash
violation: corrupted checkpoint → agent resumes into bad state
```
