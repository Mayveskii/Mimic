# Session Domain — Artifacts

How session processes are stored as mesh slots.

---

## Slot Structure for Session Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_SESSION` (9) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `budget_tracking` / `context_management` / `snapshot_restore` |
| invariants | `["budget_accurate", "denial_count_accurate", "context_append_only", "snapshot_integrity", "compression_preserves", "session_isolation", "budget_warning", "denial_reason_logged"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high |
| z_density | Medium |

---

## Pattern Codes

### budget_tracking

```c
OpPacket chain[2] = {
    {OP_SESS_BUDGET_CHECK,  .args = {}},
    {OP_ORCH_RESPOND,       .args = {{"budget_remaining", ""}}}
};
```

### snapshot_restore

```c
OpPacket chain[2] = {
    {OP_SESS_SNAPSHOT,      .args = {{"path", ".mimic/snapshots/session.json.gz"}}},
    {OP_HASH_SHA256,        .args = {{"path", ".mimic/snapshots/session.json.gz"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-17 (blocking save) | `sess_blocking_save` | `snapshot_restore` |
| AP-23 (token overflow) | `sess_unbounded_context` | `context_management` |

---

## Retrieval Path

Session patterns retrieved via:
1. Linear: exact match.
2. Keyword: `budget_accurate`, `session_isolation`.
3. Semantic: "how do I manage session budget?" → domain=session.
