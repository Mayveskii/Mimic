# RAG Domain — Artifacts

How RAG processes are stored as mesh slots.

---

## Slot Structure for RAG Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_RAG` (10) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `mesh_query` / `hybrid_retrieval` / `semantic_search` |
| invariants | `["survival_signal_required", "negative_with_counter", "precision_threshold", "retrieval_path_recorded", "multi_tier_fallback", "int8_quantized", "query_bounded", "index_consistent"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | High |
| z_density | Medium |

---

## Pattern Codes

### mesh_query

```c
OpPacket chain[1] = {
    {OP_ORCH_PLAN,    .args = {{"query_type", "mesh"}, {"domain", ""}, {"tier", "linear"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-18 (grep workspace scan) | `rag_unindexed_scan` | `mesh_query` |
| AP-27 (flaky test without fix) | `rag_unverified_pattern` | `hybrid_retrieval` |

---

## Retrieval Path

RAG patterns retrieved via:
1. Linear: exact match.
2. Keyword: `survival_signal_required`, `multi_tier_fallback`.
3. Semantic: "how do I retrieve proven patterns?" → domain=rag.
