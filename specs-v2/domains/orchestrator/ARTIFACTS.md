# Orchestrator Domain — Artifacts

How orchestrator processes are stored as mesh slots.

---

## Slot Structure for Orchestrator Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_ORCHESTRATOR` (8) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `intent_to_validated_chain` / `mesh_query` / `parallel_pipeline` |
| invariants | `["no_exec_without_validate", "no_phase_skip", "rollback_on_failure", "budget_never_exceeded", "dangerous_requires_allow", "context_cumulative", "parallel_isolation", "novelty_recorded", "honest_response", "circuit_break"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high (orchestrator is core) |
| z_density | Medium (complex interactions) |

---

## Pattern Codes

### intent_to_validated_chain

```c
OpPacket chain[6] = {
    {OP_ORCH_CLASSIFY,    .args = {{"intent", ""}}},
    {OP_ORCH_PLAN,        .args = {{"domain", ""}, {"scenario", ""}}},
    {OP_ORCH_VALIDATE,    .args = {{"opchain", ""}}},
    {OP_ORCH_EXEC,        .args = {{"opchain", ""}, {"context_id", ""}}},
    {OP_ORCH_VERIFY,      .args = {{"results", ""}, {"verifier_count", "2"}}},
    {OP_ORCH_RESPOND,     .args = {{"results", ""}, {"metrics", ""}}}
};
```

### mesh_query

```c
OpPacket chain[1] = {
    {OP_ORCH_PLAN,        .args = {{"query_type", "mesh"}, {"domain", ""}}}
};
// Internal: linear → keyword → semantic retrieval
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-01 (break-on-failure cleanup) | `orch_no_cleanup` | `intent_to_validated_chain` |
| AP-06 (undifferentiated retry) | `orch_blind_retry` | `intent_to_validated_chain` |
| AP-12 (single verifier) | `orch_single_verify` | `intent_to_validated_chain` |
| AP-23 (token overflow) | `orch_unbounded_context` | `intent_to_validated_chain` |

---

## Retrieval Path

Orchestrator patterns are core and retrieved frequently:
1. Linear: exact match by scenario name.
2. Keyword: invariant hashes for `no_exec_without_validate`.
3. Semantic: "how do I safely execute an operation?" → domain=orchestrator.
