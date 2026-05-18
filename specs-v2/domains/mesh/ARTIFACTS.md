# Mesh Domain — Artifacts

How mesh processes are stored as mesh slots.

---

## Slot Structure for Mesh Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_MESH` (11) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `slot_storage` / `index_update` / `garbage_collection` |
| invariants | `["slot_immutable", "hash_integrity", "index_atomic", "backup_before_write", "slot_size_bounded", "domain_consistent", "polarity_rules", "survival_valid", "stats_updated", "gc_orphans"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high (mesh is fundamental) |
| z_density | Medium |

---

## Pattern Codes

### slot_storage

```c
OpPacket chain[1] = {
    {OP_MMAP_WRITE,    .args = {{"ptr", ""}, {"data", ""}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-11 (stale cached state) | `mesh_stale_index` | `index_update` |
| AP-15 (partial rollback) | `mesh_partial_cleanup` | `garbage_collection` |

---

## Retrieval Path

Mesh patterns retrieved via:
1. Linear: exact match.
2. Keyword: `slot_immutable`, `hash_integrity`.
3. Semantic: "how do I store patterns safely?" → domain=mesh.
