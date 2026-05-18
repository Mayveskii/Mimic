# Security Domain — Artifacts

How security processes are stored as mesh slots.

---

## Slot Structure for Security Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_SECURITY` (13) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `never_rules` / `difc_labels` / `credential_pool` / `path_sanitization` / `input_validation` / `audit_log` / `sandboxed_exec` / `side_channel_mitigation` |
| invariants | `["never_rules_absolute", "difc_enforced", "credential_isolated", "path_sanitized", "input_validated", "audit_immutable", "sandboxed_exec", "side_channel_mitigated"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high (security is critical) |
| z_density | Medium (complex rules) |

---

## Pattern Codes

### path_sanitization

```c
OpPacket chain[1] = {
    {OP_ORCH_VALIDATE,    .args = {{"path_safe", "true"}, {"workspace_boundary", "true"}}}
};
```

### credential_pool

```c
OpPacket chain[1] = {
    {OP_SESS_BUDGET_CHECK,  .args = {{"resource", "credential_pool"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-04 (unvalidated input) | `sec_unvalidated_input` | `input_validation` |
| AP-07 (hardcoded secrets) | `sec_hardcoded_key` | `credential_pool` |
| AP-13 (override on deny) | `sec_override_deny` | `never_rules` |
| AP-26 (out-of-bound write) | `sec_buffer_overflow` | `input_validation` |

---

## Retrieval Path

Security patterns retrieved via:
1. Linear: exact match.
2. Keyword: `never_rules_absolute`, `credential_isolated`.
3. Semantic: "how do I protect credentials?" → domain=security.
