# Distillation Domain — Artifacts

How distillation processes are stored as mesh slots.

---

## Slot Structure for Distillation Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_DISTILLATION` (12) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `clone_blame` / `survival_extract` / `quality_gate` |
| invariants | `["source_verified", "blame_valid", "hash_reproducible", "quality_gate", "counter_linked", "multi_source", "async_runtime", "source_retained", "evidence_required", "polarity_correct"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | High |
| z_density | Medium |

---

## Pattern Codes

### clone_blame

```c
OpPacket chain[2] = {
    {OP_SYS_EXEC,        .args = {{"cmd", "git clone --depth 1 <repo> /tmp/repo"}}},
    {OP_SYS_EXEC,        .args = {{"cmd", "cd /tmp/repo && git blame -t <file>"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-27 (flaky test without fix) | `dist_skip_quality_gate` | `quality_gate` |
| AP-05 (context injection without stability) | `dist_unstable_extraction` | `survival_extract` |

---

## Retrieval Path

Distillation patterns retrieved via:
1. Linear: exact match.
2. Keyword: `quality_gate`, `blame_valid`.
3. Semantic: "how do I extract proven patterns?" → domain=distillation.
