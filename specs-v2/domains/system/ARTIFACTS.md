# System Domain — Artifacts

How system processes are stored as mesh slots.

---

## Slot Structure for System Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_SYSTEM` (6) |
| layer | `LAYER_PRIMITIVE` (0) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `exec_validated` / `env_isolated` / `path_workspace` / `destructive_confirmed` / `copy_integrity` / `mkdir_idempotent` / `chmod_restricted` |
| invariants | `["exec_validated", "env_isolated", "path_workspace", "destructive_confirmed", "copy_integrity", "mkdir_idempotent", "chmod_restricted", "exists_check_fresh"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | High |
| z_density | High |

---

## Pattern Codes

### exec_validated

```c
OpPacket chain[2] = {
    {OP_ORCH_VALIDATE,  .args = {{"cmd_safe", "true"}}},
    {OP_SYS_EXEC,       .args = {{"cmd", ""}, {"args", ""}, {"timeout_ms", "300000"}}}
};
```

### mkdir_idempotent

```c
OpPacket chain[1] = {
    {OP_SYS_DIR_CREATE, .args = {{"path", ""}, {"mode", "0755"}, {"recursive", "true"}}}
};
```

### copy_integrity

```c
OpPacket chain[3] = {
    {OP_SYS_FILE_COPY,  .args = {{"src", ""}, {"dst", ""}}},
    {OP_HASH_SHA256,    .args = {{"path", "src"}}},
    {OP_HASH_SHA256,    .args = {{"path", "dst"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-04 (unvalidated input) | `sys_unvalidated_exec` | `exec_validated` |
| AP-07 (hardcoded secrets) | `sys_env_leak` | `env_isolated` |
| AP-15 (partial rollback) | `sys_partial_delete` | `destructive_confirmed` |

---

## Retrieval Path

System patterns retrieved via:
1. Linear: exact name match.
2. Keyword: `path_workspace`, `destructive_confirmed`.
3. Semantic: "how do I safely delete a file?" → domain=system.
