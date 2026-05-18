# Distillation Domain — From Source to Slot

How production repositories become proven knowledge that any model can use.

---

## What This Domain Does

Distillation transforms raw source code from production repositories into proven, indexed, searchable knowledge stored in mesh slots. A model does not need to read the entire codebase — it retrieves the distilled pattern, which is the essence: the proven approach, its preconditions, its invariants, its cost.

The domain covers: repository cloning, git blame analysis, survival index computation, pattern extraction, artifact encoding, compression, indexing, slot insertion.

---

## Processes

### distill_repository

**When to use:**  
When a new repository is added to repos-manifest.yaml and needs to be processed. Can be triggered manually or via CI.

**Goal:**  
Transform a production repository into a set of mesh slots containing proven patterns.

**Chain (semantically):**

1. **Clone** repository at pinned commit.
   - URL from repos-manifest.yaml.
   - Commit SHA pinned for reproducibility.
   - # UNCERTAIN: whether shallow clone is sufficient or full history needed for accurate blame

2. **Git blame** every source file.
   - For every file: `git blame --porcelain`.
   - Extract: author, commit, line count, timestamp for every line.

3. **Compute survival index** per commit.
   - For each commit: surviving_lines / total_lines_added.
   - surviving_lines = lines from this commit still present at HEAD.
   - total_lines_added = lines added by this commit.
   - survival_index ranges 0.0 to 1.0.

4. **Extract patterns** where survival_index ≥ 0.7.
   - Functions, types, critical decision patterns.
   - # UNCERTAIN: whether 0.7 is the optimal threshold; needs calibration across different repo types

5. **Handle revert commits**.
   - For every commit with message matching "Revert ..." pattern:
     - Create NEGATIVE artifact (polarity=NEGATIVE).
     - Record: original commit, revert commit, description of failure.
     - decision_survival = 0.0 (actively removed code).
   - For every NEGATIVE artifact:
     - Find or create COUNTER artifact (polarity=COUNTER) that replaces the failed approach.
     - Link: NEGATIVE.counter_pattern_id → COUNTER.id.

6. **Encode artifact** as protobuf.
   - Schema (from 09-DISTILLATION-ARTIFACTS.md):
     ```
     id: sha256(source_commit:path:line_range:modality)
     source_repo: "etcd-io/etcd"
     source_commit: "abc123..."
     domain: "distributed"
     layer: "code"
     modality: CODE
     pattern_name: "raft_apply_commit"
     pattern_code: <actual code/pattern>
     survival_index: 0.85
     z_density: 0.72
     decision_survival: 1.5  # for DECISION layer artifacts
     invariants: [
       {condition: "leader_elected == true", source: "BEHAVIOR.md #4", verification: "2-vote"}
     ]
     invariant_hash: sha256(serialized invariants)
     extracted_by: "distill.sh"
     extracted_at: "2025-05-17T..."
     extraction_hash: sha256("distill.sh v1.2 --threshold=0.7")
     token_count: <estimated>
     latency_us: <measured>
     memory_bytes: <measured>
     polarity: POSITIVE
     counter_pattern_id: ""
     anti_pattern_id: ""
     failure_evidence: ""
     qac_violated: ""
     ```

7. **Compress** artifact.
   - OP_COMPRESS_GZIP(artifact_bytes).
   - Compute compression_ratio = original_size / compressed_size.

8. **Verify integrity**.
   - sha256_hash(compressed_data) → stored_hash.
   - Verify: stored_hash matches artifact.id hash logic.

9. **Write to bmap**.
   - bmap_write_cell(slot_id, compressed_data).

10. **Index** slot.
    - si_insert(domain, layer, modality, pattern_name).
    - Key = "distributed:code:CODE:raft_apply_commit".
    - O(1) lookup by composite key.

**Hard constraints:**
- Same commit + same tool version + same parameters → same bmap. Non-reproducible = broken distillation.
- Every artifact must have ≥1 invariant. No slot without invariant.
- Every NEGATIVE artifact must link to COUNTER.
- Only artifacts with artifact_precision ≥ 0.8 enter shared deep cache.
- Revert commits must become NEGATIVE artifacts. 100%. No exceptions.

**Invariants:**
- survival_index is computed from git blame, not guessed.
- extraction_hash matches current extractor version.
- compressed data verified by sha256 before indexing.
- slot key format: domain:layer:modality:pattern_name.

**Result when successful:**
```
status: "success"
repo: "etcd-io/etcd"
commit: "abc123..."
slots_created: 47
slots_by_domain: {
  "distributed": 23,
  "consensus": 12,
  "storage": 12
}
negative_artifacts: 3
counter_artifacts: 3
compression_ratio_avg: 2.3
```

**Result when failed:**
```
status: "failure"
repo: "etcd-io/etcd"
reason: "git_blame_parse_error" | "survival_index_computation_failed" | "bmap_write_failed"
partial_slots: 12  # how many succeeded before failure
rollback: "delete partial slots, notify operator"
```

**How a model uses this:**  
Model does not call distillation directly. Distillation is a background/administrative process. Models benefit from distilled slots through mesh_query. When a model asks "how does etcd handle consensus?" → orchestrator queries mesh → retrieves raft_apply_commit slot with survival_index 0.85, from etcd abc123, with invariant "leader_elected == true". Model learns the proven pattern, applies it.

---

### query_mesh_precision

**When to use:**  
When evaluating whether a retrieved pattern is trustworthy enough to apply.

**Goal:**  
Assess the precision and reliability of a mesh slot before using it in a process.

**Chain (semantically):**

1. Retrieve slot by key.
2. Check artifact_precision:
   - artifact_precision = survival_index × invariant_coverage × extraction_reproducibility.
   - precision = 1.0 → fully proven, fully verified, exactly reproducible.
   - precision < 0.8 → local use only, not shared.
3. Check temporal consistency:
   - source_commit has new_child_count > 0 → re-compute survival_index within 24h.
   - drift > 0.1 from recorded → flag for re-validation.
4. Check polarity:
   - POSITIVE → proceed.
   - NEGATIVE → must have counter_pattern_id; if not → reject slot.
   - COUNTER → replace linked NEGATIVE.
5. Return precision assessment.

**Hard constraints:**
- NEGATIVE without COUNTER → NEVER used. Rejected immediately.
- precision < 0.8 → NEVER shared to deep cache. Local only.
- extraction_hash mismatch → re-extract, do not use cached.

**Invariants:**
- Every result carries survival_index.
- Every result carries z_density.
- RAG without survival signal = unverified retrieval.

**Result when successful:**
```
status: "trusted"
slot_id: "sha256..."
precision: 0.92
survival_index: 0.85
invariant_coverage: 0.95
extraction_reproducibility: 1.0
recommendation: "safe_to_apply"
```

**Result when precision is low:**
```
status: "caution"
slot_id: "sha256..."
precision: 0.65
recommendation: "local_use_only"
reason: "invariant_coverage_low: only 1 invariant found"
```

**How a model uses this:**  
Orchestrator automatically checks precision before including slot in planned chain. Model never sees low-precision slots unless explicitly requesting "show me everything including untrusted." Default = only precision ≥ 0.8.

---

## Principles From Sources

### embryo (pkg/survival/)

**Principles taken:**
- git blame → survival index:
  1. surviving_lines / total_lines_added per commit.
  2. survival ≥ 0.7 → slot candidate.
  3. survival < 0.1 → discard.
  4. 0.1 ≤ survival < 0.7 → partial pattern, manual review.
- Survival index is deterministic for same commit at same point in history.

**What Mimic does with them:**
Exact survival computation. Threshold 0.7 for slot extraction. Below 0.1 = discard. Between = flag for review.

### embryo (pkg/mesh/)

**Principles taken:**
- Mesh slots indexed by domain + layer + state_hash.
- Slot index enables O(1) lookup by composite key.
- Domain groups related capabilities.

**What Mimic does with them:**
Slots stored in bmap with domain:layer:modality:pattern_name key. Query by any combination.

### bin-mesh (enricher)

**Principles taken:**
- API calls made: 72603, slots created: 32949, efficiency: 45.4%.
- PR yield: 97.6% (nearly every PR produces ≥1 slot).
- Throughput: 18.67 slots/cycle.
- Under-sampled axes: L4 (Usefulness, 23 samples), L9 (Completion, 50 samples).
- Signal ratio: 92.4% (RAG quality).

**What Mimic does with them:**
Distillation targets efficiency ≥ 60% (Phase 2), ≥ 75% (Phase 4). L4 and L9 require more data points. Signal ratio tracked for RAG quality.

### bin-mesh (SuperInvariants)

**Principles taken:**
Top SuperInvariants from 148121 slots:
- "incorrect logic must be replaced" — 57724 members → universal invariant.
- "nullable reference must be guarded" — 9409 members → nil-check rule.
- "concurrent shared state requires lock" — 1936 members → conflict rule.
- "error condition must be surfaced" — 1964 members → error handling.

**What Mimic does with them:**
SuperInvariants feed QAC-2 (Invariant Coverage) and QAC-4 (Conflict Matrix). Every extracted artifact must map to at least one SuperInvariant.

---

## Artifact Storage

How distillation artifacts are stored:

### Slot Key
```
key = "{domain}:{layer}:{modality}:{pattern_name}"
```

### Slot Data (protobuf, compressed)
```
message Artifact {
  string id = 1;                    // sha256(source_commit:path:line_range:modality)
  string source_repo = 2;           // "etcd-io/etcd"
  string source_commit = 3;         // "abc123..."
  string domain = 4;                // "distributed"
  string layer = 5;                 // "code" / "decision" / "diagram"
  Modality modality = 6;            // CODE, TEXT, IMAGE, DIAGRAM, TABLE, METRIC
  string pattern_name = 7;          // "raft_apply_commit"
  string pattern_code = 8;          // actual code/pattern text
  float survival_index = 9;         // 0.0-1.0
  float z_density = 10;             // Z(slot) = (Σ survival_i × weight_i) / slot_volume
  float decision_survival = 11;     // for DECISION layer
  repeated Invariant invariants = 12;
  string invariant_hash = 13;       // sha256 of serialized invariants
  string extracted_by = 14;         // "distill.sh"
  string extracted_at = 15;         // ISO 8601
  string extraction_hash = 16;      // sha256 of tool + params
  int32 token_count = 17;
  int32 latency_us = 18;
  int32 memory_bytes = 19;
  Polarity polarity = 20;           // POSITIVE / NEGATIVE / COUNTER
  string counter_pattern_id = 21;   // for NEGATIVE
  string anti_pattern_id = 22;      // for COUNTER
  string failure_evidence = 23;     // for NEGATIVE
  string qac_violated = 24;         // for NEGATIVE
}
```

### Storage Pipeline
```
raw_repo → git_blame → survival_index → extract_patterns → protobuf_encode
    → OP_COMPRESS_GZIP → OP_HASH_SHA256 → bmap_write_cell → si_insert
```

### Retrieval Paths
1. **Linear** (no intelligence): `si_query_domain_layer("distributed", "code")` → exact match.
2. **Keyword** (minimal): `si_query_state_hash(invariant_hash)` → matching artifacts.
3. **Semantic** (full): `int8_quantize(query)` → `batch_cosine_int8` → top-k → re-rank by survival × z_density.

---

## Cross-Domain Conflicts

Distillation domain does not conflict with execution domains. It is a background process. However:
- Distillation READs from git domain (clone, blame).
- Distillation WRITES to mesh domain (slots, index).
- Concurrent distillation of same repo → serialize to prevent bmap corruption.
