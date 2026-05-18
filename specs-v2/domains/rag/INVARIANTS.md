# RAG Domain — Invariants

Rules that MUST hold for every process in the RAG domain.

---

## RINV-01: Survival Signal Required

**What it prevents:** Retrieval of unverified, low-quality patterns.

**What it requires:** Every retrieved slot MUST have `survival_index > 0.0`. If survival_index is 0.0 or missing, result tagged `unverified, use with caution`.

**Source of this rule:**
- embryo hunt system: survival_index is primary quality signal.
- AP-27 (flaky test without fix): unverified data → false confidence.

**Consequence of violation:** Unverified results flagged. Model warned: "result lacks survival signal, apply with caution".

---

## RINV-02: NEGATIVE With COUNTER

**What it prevents:** Telling model what NOT to do without saying what TO do.

**What it requires:** If retrieved slot has `polarity == NEGATIVE`, it MUST include valid `counter_slot_id`. Model receives both: the anti-pattern AND the correct alternative.

**Source of this rule:**
- Specs-v2 artifact schema: Polarity rules.
- Teaching principle: negative examples need positive counterpart.

**Consequence of violation:** NEGATIVE slot without counter → result REJECTED from retrieval. Log: "incomplete anti-pattern, missing counter".

---

## RINV-03: Precision Threshold Enforced

**What it prevents:** Low-precision artifacts polluting mesh.

**What it requires:** If `artifact_precision < 0.8`, result tagged `low_precision, use with caution`. If < 0.5, result NOT returned (filtered at retrieval).

**Source of this rule:**
- Specs-v2 artifact schema: QAC assessment.
- Quality gating: QAC-7 (Artifact Precision).

**Consequence of violation:** Low-precision results filtered or flagged.

---

## RINV-04: Retrieval Path Recorded

**What it prevents:** Black-box retrieval, debugging difficulty.

**What it requires:** Every query result includes `retrieval_path`: "linear" | "keyword" | "semantic". For semantic: include top-k count and reranking method.

**Source of this rule:**
- Transparency requirement.
- Debugging best practice.

**Consequence of violation:** Missing path → result marked incomplete. Log warning.

---

## RINV-05: Multi-Tier Fallback

**What it prevents:** Empty results, failed queries.

**What it requires:** Retrieval follows cascade: linear (exact) → keyword (invariant match) → semantic (vector similarity). If tier N fails, auto-fallback to tier N+1. If all fail → return `no_results` with suggestions.

**Source of this rule:**
- embryo hunt system: 3-tier cascade (local→medium→top).
- graphify: exact > prefix > substring.

**Consequence of violation:** Single-tier failure → empty results. Cascade ensures coverage.

---

## RINV-06: Semantic Search Int8 Quantized

**What it prevents:** Slow vector similarity, high memory usage.

**What it requires:** Semantic search uses int8 quantization for embeddings. Cosine similarity computed with INT8 arithmetic. Top-k reranked by survival_index × z_density.

**Source of this rule:**
- Performance optimization.
- graphify: IDF-weighted search.

**Consequence of violation:** Float32 search allowed but slower. Warning: "using float32, consider int8 for performance".

---

## RINV-07: Query Context Bounded

**What it prevents:** Infinite query loops, context explosion.

**What it requires:** Query input bounded: max 4096 characters. Max 10 results per query. Pagination supported for larger result sets.

**Source of this rule:**
- Resource management.
- AP-23 (token overflow).

**Consequence of violation:** Query REJECTED with `ERR_INVALID_ARG` if input too large.

---

## RINV-08: Index Consistency

**What it prevents:** Stale or corrupted index results.

**What it requires:** Index rebuild triggered after batch slot insertion (>100 new slots). Background integrity check runs daily. Corrupted index entries detected and quarantined.

**Source of this rule:**
- Standard database maintenance.
- embryo projectmap: auto-index on WRITE.

**Consequence of violation:** Stale index → results may reference deleted slots. Integrity check detects and removes.
