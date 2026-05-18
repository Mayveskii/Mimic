# IO Domain — Artifacts

How IO processes are stored as mesh slots.

---

## Slot Structure for IO Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_IO` (2) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `read_file` / `write_file` / `delete_file` |
| invariants | `["file_exists_before_read", "backup_before_overwrite", "sensitive_path_blocked", "workspace_boundary", "read_back_verification", "index_update_after_write"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | From safe file operation patterns across repos |
| z_density | Medium (file operations are common, high frequency) |

---

## Pattern Codes

### read_file

```c
OpPacket chain[4] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", ""}}},
    {OP_IO_OPEN,          .args = {{"path", ""}, {"mode", "r"}}},
    {OP_IO_READ,          .args = {{"fd", ""}, {"length", ""}}},
    {OP_HASH_SHA256,      .args = {{"data", ""}}}  // hash of content
};
```

### write_file

```c
OpPacket chain[6] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", ""}}},  // check parent
    {OP_SYS_DIR_CREATE,   .args = {{"path", ""}, {"recursive", "true"}}},  // create parent if needed
    {OP_IO_OPEN,          .args = {{"path", ""}, {"mode", "w"}}},
    {OP_IO_WRITE,         .args = {{"fd", ""}, {"data", ""}}},
    {OP_IO_CLOSE,         .args = {{"fd", ""}}},
    {OP_HASH_SHA256,      .args = {{"path", ""}}}  // verify written
};
```

### delete_file

```c
OpPacket chain[3] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", ""}}},
    {OP_GIT_STATUS,       .args = {{"path", ""}}},  // check if tracked
    {OP_SYS_FILE_DELETE,  .args = {{"path", ""}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-04 (unvalidated input before I/O) | `io_unvalidated_write` | `write_file` |
| AP-15 (partial rollback) | `io_partial_cleanup` | `write_file` |
| AP-26 (out-of-bound write) | `io_buffer_overflow` | `write_file` |

---

## Retrieval Path

IO patterns are retrieved via:
1. Linear: exact name match.
2. Keyword: invariant hash for "backup_before_overwrite", "sensitive_path_blocked".
3. Semantic: "how do I safely write a file?" → domain=io.

Survival index: file operation safety from repos with robust file handling (golang stdlib, rust std).
