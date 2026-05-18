# Distillation Domain — Invariants

Rules that MUST hold for every process in the distillation domain.

---

## DINV-01: Source Repo Verified

**What it prevents:** Distillation from untrusted, corrupted, or non-existent repos.

**What it requires:** Before clone: verify repo URL is valid (https scheme, known host). After clone: verify git repository integrity (`git fsck`).

**Source of this rule:**
- Standard git hygiene.
- Security best practice.

**Consequence of violation:** Clone REJECTED with `invalid_repository`. Corrupted repo → discard, log.

---

## DINV-02: Blame Timestamp Valid

**What it prevents:** Invalid survival_index from bad timestamps.

**What it requires:** `blame_timestamp` must be parseable Unix timestamp. Author date used (not commit date). `survival_index = surviving_lines / total_lines` at current HEAD.

**Source of this rule:**
- embryo survival_index: git blame → surviving lines per commit.
- Statistical validity.

**Consequence of violation:** Invalid timestamp → survival_index set to 0.0, slot flagged for manual review.

---

## DINV-03: Extraction Hash Reproducible

**What it prevents:** Different extraction tool versions producing incompatible hashes.

**What it requires:** `extraction_hash = SHA256(extractor_binary + parameters + source_content)`. Same extractor + parameters + source = same hash. Hash recorded in artifact.

**Source of this rule:**
- Reproducibility principle.
- SLOT_SCHEMA.md.

**Consequence of violation:** Hash mismatch on re-extraction → slot flagged for review. If extractor changed → update hash and re-validate.

---

## DINV-04: Quality Gate Before Indexing

**What it prevents:** Low-quality patterns entering mesh.

**What it requires:** Before slot indexed: QAC assessment must pass (≤ 3 QAC failures). artifact_precision ≥ 0.8. survival_index > 0.0.

**Source of this rule:**
- Specs-v2 artifact schema: precision threshold.
- Quality gates: QAC-1..13.

**Consequence of violation:** Fails quality gate → slot ARCHIVED (not indexed). Log: "quality gate failed, archived".

---

## DINV-05: Counter Pattern Linked

**What it prevents:** Orphaned anti-patterns in mesh.

**What it requires:** If polarity == NEGATIVE, counter_pattern_id MUST exist in mesh. If counter not yet distilled, slot queued but NOT indexed until counter available.

**Source of this rule:**
- Artifact schema: polarity rules.
- RINV-02 (NEGATIVE with COUNTER).

**Consequence of violation:** NEGATIVE without counter → NOT indexed. Log: "missing counter_pattern".

---

## DINV-06: Multiple Sources Stacked

**What it prevents:** Overfitting to single source, lack of generality.

**What it requires:** High-confidence slots (survival_index > 0.8) should have ≥ 2 source repos. Single-source slots allowed but flagged `single_source: true`.

**Source of this rule:**
- Statistical robustness.
- embryo: cross-repo pattern validation.

**Consequence of violation:** Single-source slots indexed but marked. Retrieval ranking penalizes single-source vs multi-source.

---

## DINV-07: Distillation Does Not Block Runtime

**What it prevents:** Mesh queries blocked by long-running distillation.

**What it requires:** Distillation runs asynchronously. Mesh queries not blocked by distillation. New slots added via atomic append. Index updates batched and applied in background.

**Source of this rule:**
- Performance requirement.
- gastown: pressure gating.

**Consequence of violation:** Synchronous distillation → mesh queries timeout. Background processing required.

---

## DINV-08: Source Retention

**What it prevents:** Loss of provenance, inability to verify.

**What it requires:** Source repos retained for 90 days after last distillation. Commit SHAs recorded permanently. Source content not retained (only hashes), but can be re-cloned.

**Source of this rule:**
- Audit requirement.
- Reproducibility.

**Consequence of violation:** Source repo deleted → survival_index cannot be updated. Slot marked `source_unavailable`.

---

## DINV-09: Anti-Pattern Evidence Required

**What it prevents:** Vague anti-patterns without concrete basis.

**What it requires:** NEGATIVE slots MUST include `failure_evidence`: concrete description of observed failure (not abstract). Links to issue/PR where failure was observed.

**Source of this rule:**
- Evidence-based practice.
- Specs-v2 artifact schema.

**Consequence of violation:** NEGATIVE without evidence → NOT indexed. Log: "missing failure evidence".

---

## DINV-10: Polarity Correctness

**What it prevents:** Wrong polarity assignment, confusion between positive and negative.

**What it requires:** POSITIVE = what TO do. NEGATIVE = what NOT to do (must have counter). COUNTER = correct alternative to NEGATIVE. Polarity assigned by extraction tool, verified by QAC-7.

**Source of this rule:**
- Artifact schema polarity rules.
- Teaching clarity.

**Consequence of violation:** Wrong polarity → QAC-7 fail → slot archived.
