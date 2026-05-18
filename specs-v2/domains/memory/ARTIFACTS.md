# Memory Domain — Artifacts

How memory processes are stored as mesh slots.

---

## Slot Structure for Memory Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_MEMORY` (5) |
| layer | `LAYER_PRIMITIVE` (0) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `mmap_alloc` / `mmap_free` / `mmap_sync` |
| invariants | `["size_validated", "pointer_tracked", "private_anonymous", "sync_before_visible", "cleanup_on_completion", "bounds_checked", "readonly_protected"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | High (mmap is fundamental, stable) |
| z_density | High (simple, highly compressible patterns) |

---

## Pattern Codes

### mmap_alloc

```c
OpPacket chain[2] = {
    {OP_SESS_BUDGET_CHECK,  .args = {{"type", "memory"}}},
    {OP_MMAP_ALLOC,         .args = {{"size", ""}, {"flags", "MAP_PRIVATE|MAP_ANONYMOUS"}}}
};
```

### mmap_free

```c
OpPacket chain[1] = {
    {OP_MMAP_FREE,          .args = {{"ptr", ""}, {"size", ""}}}
};
```

### mmap_sync

```c
OpPacket chain[1] = {
    {OP_MMAP_SYNC,          .args = {{"ptr", ""}, {"size", ""}, {"flags", "MS_SYNC"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-03 (panic recovery) | `mem_panic_recovery_free` | `mmap_free` |
| AP-26 (out-of-bound write) | `mem_buffer_overflow` | `mmap_alloc` |
| AP-28 (missing overflow guard) | `mem_size_overflow` | `mmap_alloc` |

---

## Retrieval Path

Memory patterns are rarely queried directly (they're primitives). They appear in:
1. Composite pattern decomposition (e.g., `build_and_test` includes memory allocation).
2. Keyword: `bounds_checked` for safety-critical code.
3. Semantic: "how do I safely allocate memory?" → domain=memory.
