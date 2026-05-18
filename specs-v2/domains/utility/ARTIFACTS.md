# Utility Domain — Artifacts

How utility processes are stored as mesh slots.

---

## Slot Structure for Utility Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_UTILITY` (7) |
| layer | `LAYER_PRIMITIVE` (0) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `hash_sha256` / `compress_gzip` / `encrypt_aes` |
| invariants | `["hash_deterministic", "compression_reversible", "encryption_authenticated", "key_isolated", "input_bounded", "side_effect_free"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | Very high (utility patterns are universal) |
| z_density | Very high (simple, highly compressible) |

---

## Pattern Codes

### hash_sha256

```c
OpPacket chain[1] = {
    {OP_HASH_SHA256,    .args = {{"data", ""}}}
};
```

### compress_gzip

```c
OpPacket chain[1] = {
    {OP_COMPRESS_GZIP,  .args = {{"data", ""}, {"level", "6"}}}
};
```

### encrypt_aes

```c
OpPacket chain[1] = {
    {OP_ENCRYPT_AES,    .args = {{"data", ""}, {"key_id", ""}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-07 (hardcoded secrets) | `util_inline_key` | `encrypt_aes` |
| AP-28 (missing overflow guard) | `util_hash_overflow` | `hash_sha256` |

---

## Retrieval Path

Utility patterns are rarely queried directly (primitives). They appear in:
1. Composite pattern decomposition.
2. Semantic: "how do I hash a file?" → domain=utility.
