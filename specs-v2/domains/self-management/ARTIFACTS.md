# Self-Management Domain — Artifacts

Mesh slots produced by the self-management domain.

---

## Slot Types

| Type | layer | Purpose |
|------|-------|---------|
| CHECKPOINT | "checkpoint" | Full state snapshot at a point in time |
| BUDGET_PLAN | "budget" | Allocated budget per subtask |
| PIVOT_LOG | "pivot" | Record of strategy change |
| PROGRESS_REPORT | "progress" | Assessment output |
| CONTEXT_SUMMARY | "context" | Semantic compression of session |

---

## Checkpoint Slot Structure

```json
{
  "slot_type": "SELF_MANAGEMENT",
  "layer": "checkpoint",
  "domain": "self-management",
  "payload": {
    "checkpoint_id": "chk-uuid",
    "session_id": "sess-uuid",
    "timestamp": "ISO-8601",
    "sequence_number": 7,
    "state_snapshot": {
      "session_state": "<compressed>",
      "active_hypotheses": ["hyp-uuid-1"],
      "budget_remaining": {"tokens": 50000, "time_ms": 1800000},
      "open_resources": [...]
    },
    "hash": "sha256-of-snapshot",
    "size_bytes": 204800
  }
}
```

---

## Budget Plan Slot Structure

```json
{
  "slot_type": "SELF_MANAGEMENT",
  "layer": "budget",
  "domain": "self-management",
  "payload": {
    "plan_id": "bud-uuid",
    "session_id": "sess-uuid",
    "total_budget": {"tokens": 100000, "time_ms": 3600000},
    "subtasks": [
      {"id": "sub-1", "estimated_tokens": 20000, "actual_tokens": 25000, "status": "done"},
      {"id": "sub-2", "estimated_tokens": 30000, "actual_tokens": 0, "status": "pending"}
    ],
    "reallocation_event": true,
    "trigger": "sub-1 overran by 25%"
  }
}
```

---

## Pivot Log Slot Structure

```json
{
  "slot_type": "SELF_MANAGEMENT",
  "layer": "pivot",
  "domain": "self-management",
  "payload": {
    "pivot_id": "pvt-uuid",
    "session_id": "sess-uuid",
    "timestamp": "ISO-8601",
    "subtask_id": "sub-2",
    "old_approach_id": "app-old",
    "new_approach_id": "app-new",
    "failure_count": 3,
    "failure_reasons": ["timeout", "bad_result", "timeout"],
    "alternatives_queried": 5,
    "rationale": "Approach app-old timed out 3 times. app-new has survival_index=0.82.",
    "mesh_query": "si_query_domain('orchestrator', min_survival=0.7)"
  }
}
```

---

## Progress Report Slot Structure

```json
{
  "slot_type": "SELF_MANAGEMENT",
  "layer": "progress",
  "domain": "self-management",
  "payload": {
    "report_id": "prg-uuid",
    "session_id": "sess-uuid",
    "timestamp": "ISO-8601",
    "planned_subtasks": 10,
    "completed_subtasks": 6,
    "completion_pct": 0.60,
    "quality_score": 0.85,
    "divergence_detected": false,
    "stall_detected": false,
    "last_checkpoint_id": "chk-uuid-6"
  }
}
```

---

## OpPacket Chains for Self-Management

### Full checkpoint cycle

```
[0] OP_SELF_CHECKPOINT_CREATE         {session_id: "...", trigger: "interval"}
[1] OP_SELF_PROGRESS_ASSESS           {session_id: "..."}
[2] OP_SELF_CONTEXT_SUMMARIZE         {session_id: "...", target_size_mb: 0.5}
```

### Budget overrun handling

```
[0] OP_SELF_PROGRESS_ASSESS           {session_id: "..."}
[1] OP_SELF_BUDGET_REALLOCATE         {session_id: "...", trigger_subtask: "sub-1"}
[2] OP_SELF_STRATEGY_PIVOT            {subtask_id: "sub-2", threshold: 3}  // if needed
```
