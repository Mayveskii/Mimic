# Artifact Blueprint — TEMPLATE

Copy this to describe how artifacts are stored: domains/<name>/ARTIFACTS.md

---

## How This Domain's Processes Become Slots

### Slot Structure

```
domain:     "<domain_name>"
layer:      "process" | "invariant" | "decision" | "pattern"
modality:   "code" | "text" | "diagram" | "table" | "metric"
pattern_name: "<identifier>"
pattern_code: "<semantic description of process>"
invariants: [
  {condition: "...", source: "...", verification: "..."}
]
survival_index: <float>
z_density: <float>
polarity: POSITIVE | NEGATIVE | COUNTER
counter_pattern_id: "<id>" (if NEGATIVE)
extraction_hash: "<sha256>"
```

### Storage Pipeline

1. Process identified and validated
2. Protobuf encoded
3. OP_COMPRESS_GZIP
4. OP_HASH_SHA256
5. bmap_write_cell
6. si_insert(domain, layer, modality, pattern_name)

### Retrieval Paths

1. Linear: domain + layer + modality + pattern_name → exact slot
2. Keyword: si_query_state_hash(invariant_hash) → all matching
3. Semantic: int8_quantize → batch_cosine_int8 → survival+Z boost

---

## Domain-Specific Fields

Any fields unique to this domain's artifacts.
