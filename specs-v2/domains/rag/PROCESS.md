# RAG Domain — Retrieval-Augmented Generation

How Mimic finds and ranks proven knowledge for models.

---

## What This Domain Does

When a model needs a pattern, an approach, or a proven solution, it does not guess. It queries the mesh through Retrieval-Augmented Generation. RAG provides three access paths, ranked by intelligence required: linear lookup (fastest, no intelligence needed), keyword search (minimal intelligence), semantic search (full intelligence with survival-weighted ranking).

The domain covers: vector quantization, similarity search, domain filtering, survival-weighted ranking, top-k retrieval.

---

## Processes

### linear_retrieval

**When to use:**  
Model knows exactly what it needs: domain, layer, modality, pattern name. Fastest path.

**Goal:**  
O(1) exact lookup by composite key.

**Chain (semantically):**

1. Form composite key: `domain:layer:modality:pattern_name`.
2. Query: `si_query_domain_layer(domain, layer)`.
3. Filter by modality and pattern_name.
4. Return exact match or empty.

**Invariants:**
- Key format: `domain:layer:modality:pattern_name`.
- Lookup is O(1) — no intelligence needed.
- Every result has survival_index and z_density.

**Result:**
```
status: "success"
retrieval_path: "linear"
result: {
  pattern_name: "atomic_commit",
  survival_index: 0.85,
  z_density: 0.72,
  ...
}
```

---

### keyword_retrieval

**When to use:**  
Model knows keywords or invariant conditions but not exact pattern name.

**Goal:**  
Find all artifacts sharing the same invariants or keywords.

**Chain (semantically):**

1. Extract keywords from model query.
2. Compute invariant_hash from keywords.
3. Query: `si_query_state_hash(invariant_hash)`.
4. Return all artifacts with matching invariants.

**Invariants:**
- Results filtered by domain first.
- No result without survival_index.

**Result:**
```
status: "success"
retrieval_path: "keyword"
results: [artifact1, artifact2, ...]
```

---

### semantic_retrieval

**When to use:**  
Model expresses intent in natural language. No exact name known.

**Goal:**  
Find top-k most semantically similar patterns, re-ranked by proven quality.

**Chain (semantically):**

1. Quantize query: `int8_quantize(embed(query))` → int8 vector.
2. Filter by domain: `si_query_domain(domain)` → candidate slot set.
3. Compute similarity: `batch_cosine_int8(query_vec, candidate_vecs)` → similarity scores.
4. Rank candidates by score.
5. Re-rank top-k by survival_index × z_density boost.
6. Return ranked list.

**Invariants:**
- RAG without survival signal = unverified. Every result carries survival_index.
- Re-ranking boost: survival_index × z_density applied to top-k.
- No result without at least one invariant.

**Result:**
```
status: "success"
retrieval_path: "semantic"
results: [
  {pattern: "rollback_on_failure", similarity: 0.92, survival: 0.85, z_density: 0.81},
  {pattern: "cleanupOnError", similarity: 0.87, survival: 0.78, z_density: 0.65}
]
```

---

## Principles From Sources

### embryo (pkg/rag/)

**Principles taken:**
- 5-signal hybrid: vector_similarity + keyword_match + domain_filter + survival_index + z_density.
- Weighted combination: [0.3, 0.2, 0.2, 0.15, 0.15].
- qdrant client for vector search.

**What Mimic does with them:**
5 signals combined for ranking. Survival and Z-density as explicit signals, not afterthoughts.

### graphify

**Principles taken:**
- IDF-weighted search: exact 1000x > prefix 100x > substring 1x.
- Gap-ratio seed cutoff: stop expanding if gap < 20%.
- Hub-throttled traversal: skip high-degree nodes as transit.

**What Mimic does with them:**
Keyword matching uses three-tier precedence. Semantic traversal avoids hub overload.

### bin-mesh (RAG metrics)

**Principles taken:**
- AST symbols: 38829
- KW terms: 117230
- MD entries: 6367
- Comments: 4979
- Qdrant points: 180684
- Signal ratio: 92.4%

**What Mimic does with them:**
Signal ratio tracked per RAG query. Low ratio = more qdrant noise → improve indexing.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "rag" |
| layer | "process" |
| modality | "code" |
| pattern_name | "linear_retrieval" / "keyword_retrieval" / "semantic_retrieval" |
| invariants | ["survival_signal_required", "rerank_by_quality"] |

---

## Cross-Domain Conflicts

RAG domain is read-only. No conflicts with write domains.
