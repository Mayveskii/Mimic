# Anti-Patterns Domain — Artifacts

How anti-patterns are stored as mesh slots.

---

## Slot Structure for Anti-Pattern Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_ANTI_PATTERNS` (15) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_TEXT` (1) |
| pattern_name | `AP-NN_<short_desc>` / `CP-NN_<short_desc>` |
| invariants | `["has_counter", "evidence_required", "qac_mapped", "root_cause_analysis", "counter_tested", "no_duplicates", "naming_convention", "review_cycle"]` |
| polarity | `POLARITY_NEGATIVE` (1) for AP, `POLARITY_COUNTER` (2) for CP |
| survival_index | From evidence strength (not git blame) |
| z_density | High (text-heavy, structured) |

---

## Pattern Codes

Anti-patterns are documentation patterns, not executable OpPacket chains:

```json
{
  "anti_pattern_id": "AP-05",
  "name": "context_injection_without_structure_stability",
  "description": "Mutating shared state (like buildMessages) without guaranteeing structure stability leads to corruption.",
  "evidence": "binary-mesh buildMessages corruption caused full revert (April 2025)",
  "missing_validation": "structure_stable_before_injection",
  "qac_violated": ["QAC-2", "QAC-7"],
  "counter_pattern_id": "CP-05",
  "counter_pattern_name": "atomic_commit"
}
```

---

## Counter Pattern Linkage

| AP | Counter | Domain |
|---|---|---|
| AP-01 | `intent_to_validated_chain` | orchestrator |
| AP-02 | `parallel_build_shards` | build |
| AP-03 | `mmap_free` | memory |
| AP-04 | `write_file` | io |
| AP-05 | `atomic_commit` | git |
| AP-06 | `error_classify` | quality |
| AP-07 | `credential_pool` | security |
| AP-08 | `input_validation` | security |
| AP-09 | `input_validation` | security |
| AP-10 | `input_validation` | security |
| AP-11 | `safe_merge` | git |
| AP-12 | `two_vote_verify` | quality |
| AP-13 | `never_rules` | security |
| AP-14 | `tcp_managed` | network |
| AP-15 | `rollback_atomic` | quality |
| AP-16 | `feature_branch` | git |
| AP-17 | `snapshot_restore` | session |
| AP-18 | `mesh_query` | rag |
| AP-19 | `http_get_post` | network |
| AP-20 | `input_validation` | security |
| AP-21 | `input_validation` | security |
| AP-22 | `mmap_free` | memory |
| AP-23 | `context_management` | session |
| AP-24 | `atomic_commit` | git |
| AP-25 | `parallel_build_shards` | build |
| AP-26 | `write_file` | io |
| AP-27 | `quality_gate` | quality |
| AP-28 | `mmap_alloc` | memory |
| AP-29 | `honest_reporting` | quality |
| AP-30 | `tcp_managed` | network |

---

## Retrieval Path

Anti-patterns retrieved via:
1. Linear: exact AP-NN ID.
2. Keyword: `qac_violated` contains specific QAC code.
3. Semantic: "what can go wrong with X?" → domain=anti-patterns + related domain.

Anti-patterns are teaching material. Retrieved when model asks about risks or when validation detects potential violation.
