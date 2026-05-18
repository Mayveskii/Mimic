# Process Domain — Artifacts

How process processes are stored as mesh slots.

---

## Slot Structure for Process Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_PROCESS` (4) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `proc_spawn` / `proc_wait` / `proc_signal_kill` |
| invariants | `["command_validated", "sandboxed", "timeout_enforced", "resource_limited", "pid_tracked", "sigterm_first", "own_session_only", "exit_code_captured", "output_bounded", "env_sanitized"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | From safe process management patterns across repos |
| z_density | Medium |

---

## Pattern Codes

### proc_spawn

```c
OpPacket chain[3] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", ""}}},  // validate command
    {OP_ORCH_VALIDATE,    .args = {{"sandbox", "true"}, {"limits", "true"}}},
    {OP_PROC_SPAWN,       .args = {{"cmd", ""}, {"args", ""}, {"timeout_ms", "300000"}}}
};
```

### proc_wait

```c
OpPacket chain[1] = {
    {OP_PROC_WAIT,        .args = {{"pid", ""}, {"timeout_ms", "300000"}}}
};
```

### proc_signal_kill

```c
OpPacket chain[3] = {
    {OP_ORCH_VALIDATE,    .args = {{"pid_owned", "true"}}},
    {OP_PROC_SIGNAL,      .args = {{"pid", ""}, {"signal", "15"}}},  // SIGTERM
    {OP_PROC_WAIT,        .args = {{"pid", ""}, {"timeout_ms", "5000"}}},
    // If still alive:
    {OP_PROC_KILL,        .args = {{"pid", ""}}}  // SIGKILL
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-04 (unvalidated input) | `proc_unvalidated_command` | `proc_spawn` |
| AP-14 (infinite wait) | `proc_no_timeout` | `proc_wait` |
| AP-15 (partial rollback) | `proc_orphaned_children` | `proc_spawn` |

---

## Retrieval Path

Process patterns retrieved via:
1. Linear: exact name match.
2. Keyword: `sandboxed`, `timeout_enforced`.
3. Semantic: "how do I safely run a command?" → domain=process.
