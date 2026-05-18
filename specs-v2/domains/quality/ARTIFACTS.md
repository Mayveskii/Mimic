# Quality Domain — Artifacts

How quality processes are stored as mesh slots.

---

## Slot Structure for Quality Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_QUALITY` (14) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `two_vote_verify` / `conflict_check` / `budget_check` / `permission_pipeline` / `timeout_enforced` / `rollback_atomic` / `error_classify` / `honest_report` / `quality_gate` |
| invariants | `["two_vote_critical", "conflict_pairwise", "budget_before_exec", "permission_pipeline", "timeout_all_blocking", "rollback_atomic", "error_classify_before_retry", "honest_reporting", "quality_thresholds", "measured_thresholds"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high (quality is core) |
| z_density | Medium |

---

## Pattern Codes

### two_vote_verify

```c
OpPacket chain[2] = {
    {OP_ORCH_VERIFY,    .args = {{"verifier_a", "correctness"}, {"verifier_b", "invariants"}}},
    {OP_ORCH_RESPOND,   .args = {{"verification", "true"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-01 (break-on-failure cleanup) | `quality_no_cleanup` | `rollback_atomic` |
| AP-06 (undifferentiated retry) | `quality_blind_retry` | `error_classify` |
| AP-12 (single verifier) | `quality_single_verify` | `two_vote_verify` |
| AP-27 (flaky test without fix) | `quality_skip_test` | `quality_gate` |

---

## Retrieval Path

Quality patterns retrieved via:
1. Linear: exact match.
2. Keyword: `two_vote_critical`, `conflict_pairwise`.
3. Semantic: "how do I ensure quality?" → domain=quality.
