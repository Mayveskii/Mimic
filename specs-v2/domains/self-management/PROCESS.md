# Self-Management Domain — Process Specification

How Mimic manages its own state: checkpoints, budget reallocation, strategy pivot, progress assessment, and autonomous recovery.

---

## Core Processes

### 1. Self-Checkpoint

```
Trigger: significant state change OR elapsed interval
    |
    v
[OP_SELF_CHECKPOINT_CREATE]
    - Snapshot current: session state, active hypotheses, budget remaining
    - Compute diff from last checkpoint
    - Store in mesh slot: domain="self-management", layer="checkpoint"
    - Verify: sha256 hash of snapshot
    |
    v
Result: checkpoint_id + size + hash
```

**Behavior**: Agent decides (or is configured) to checkpoint after significant work. Not just at session end. Configurable interval: `MIMIC_CHECKPOINT_INTERVAL_MINUTES`.

**Result**: Recoverable snapshot with hash verification.

**Why**: If a long task fails at 90%, agent resumes from last checkpoint, not from scratch.

---

### 2. Budget Reallocation

```
Subtask A consumes more than estimated:
    |
    v
[OP_SELF_BUDGET_REALLOCATE]
    - Compute: actual_cost_A vs estimated_cost_A
    - Compute: remaining_budget - (sum remaining_estimates)
    - If deficit: reduce estimates for B, C, D proportionally
    - If surplus: increase quality (more tests, deeper search)
    - Store new allocation as mesh slot
    |
    v
Result: updated budget plan + warning (if critical threshold crossed)
```

**Behavior**: Budget is not static. As tasks execute, actual costs update estimates. Agent rebalances.

**Result**: Task completes or agent warns early that budget is insufficient.

**Why**: Without reallocation, a single underestimated subtask aborts the entire chain.

---

### 3. Strategy Pivot

```
Same approach fails N times (default N=3, config: MIMIC_STRATEGY_PIVOT_THRESHOLD):
    |
    v
[OP_SELF_STRATEGY_PIVOT]
    - Record: approach_id + failure_count + failure_reasons
    - Query mesh: si_query_domain("self-management") for alternative approaches
    - Rank alternatives by: survival_index, z_density, prior success rate
    - Select best alternative
    - Log pivot decision with justification
    |
    v
Result: new_approach_id + rationale
```

**Behavior**: Agent recognizes when stuck. Queries mesh for proven alternative patterns. Switches.

**Result**: Task deblocks without human intervention.

**Why**: Humans pivot when stuck. Agents without pivoting retry indefinitely.

---

### 4. Progress Self-Assessment

```
Periodically (every checkpoint or on-demand):
    |
    v
[OP_SELF_PROGRESS_ASSESS]
    - Compare: completed_subtasks vs planned_subtasks
    - Compare: actual_quality vs target_quality (precision, test coverage)
    - Detect: divergence (agent pursuing wrong goal)
    - Detect: stalls (no progress in last K minutes)
    - If divergence detected: trigger OP_SELF_STRATEGY_PIVOT
    - If stall detected: trigger OP_SELF_STRATEGY_PIVOT
    |
    v
Result: progress_report {completion_pct, quality_score, divergence_flag, stall_flag}
```

**Behavior**: Agent monitors its own progress against the plan. Detects when going off-track.

**Result**: Early warning before wasted effort.

**Why**: Self-monitoring prevents agents from confidently working on the wrong problem for hours.

---

### 5. Context Summarize

```
Context growing large (> 50% of MIMIC_MAX_CONTEXT_SIZE_MB):
    |
    v
[OP_SELF_CONTEXT_SUMMARIZE]
    - Identify: key decisions, key findings, key invariants from current session
    - Compress: semantic summary (not just gzip — extract meaning)
    - Store: summary as mesh slot with high z-density
    - Mark: original detailed context as "archived", keep summary in active context
    |
    v
Result: freed_context_bytes + summary_slot_id
```

**Behavior**: Instead of dropping old context, extract key meaning and store it. Active context stays lean. Full history retrievable via RAG.

**Result**: Context never exceeds limit. Key information preserved.

**Why**: 1MB context limit forces truncation. Semantic summarization preserves meaning while reducing size.

---

## Domain-Specific OpCodes

| OpCode | Purpose | Safety |
|--------|---------|--------|
| OP_SELF_CHECKPOINT_CREATE | Snapshot current state | DANGEROUS (writes to mesh) |
| OP_SELF_CHECKPOINT_RESTORE | Load from snapshot | SAFE |
| OP_SELF_BUDGET_REALLOCATE | Rebalance budget across subtasks | SAFE |
| OP_SELF_STRATEGY_PIVOT | Switch approach after N failures | SAFE |
| OP_SELF_PROGRESS_ASSESS | Compare actual vs planned progress | READONLY |
| OP_SELF_CONTEXT_SUMMARIZE | Semantic compression of context | SAFE |

---

## Invariants

- SM-INV-1: A checkpoint MUST be taken before any destructive operation
- SM-INV-2: Budget reallocation MUST not exceed total session budget
- SM-INV-3: Strategy pivot MUST query at least 3 alternative approaches before selecting
- SM-INV-4: Progress assessment MUST run at every checkpoint + on-demand trigger
- SM-INV-5: Context summarization MUST preserve all invariants and decisions
- SM-INV-6: Every self-management decision MUST be logged with timestamp and rationale

---

## Sources

| Behavior | Source | Evidence |
|----------|--------|----------|
| Self-checkpoint | hashicorp/terraform | plan/apply state file pattern |
| Budget reallocation | Mayveskii/code-mode | iteration_budget + maxBudgetUsd |
| Strategy pivot | crewai/crewai | Task decomposition + retry patterns |
| Progress assessment | microsoft/autogen | Critic agent pattern |
| Context summarization | Mayveskii/openmythos | adaptive_computation_time |
| Self-healing | Mayveskii/gastown | rollback_on_failure + retry |

---

## Artifacts

Self-management artifacts stored in mesh:

```
slot_type: SELF_MANAGEMENT
layer: "checkpoint" | "budget" | "pivot" | "progress" | "context"
domain: "self-management"
payload: {<structured decision record>}
```
