# Quality Domain — Sources

Where the quality domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- pkg/checkpoints/: C6-C10 gates with quality checkpoints.
- pkg/hunt/: 17-module hunt pipeline includes quality assessment.

**What Mimic does with them:**
Quality gates at C6-C10. Hunt pipeline assesses pattern quality. Checkpoints gate progression.

**What Mimic does NOT copy:**
- Embryo's specific checkpoint implementation.
- Embryo's Go-based quality scoring.

---

## bun (PR #30412)

**Principles taken:**
- 2-vote adversarial verify.
- Permission pipeline.
- Phase graph with gates.

**What Mimic does with them:**
2-vote on critical ops. Permission pipeline gates all operations. Phase transitions validated.

---

## gastown

**Principles taken:**
- ZFC state: zero false confidence.
- Rollback on failure.
- Pressure gating.

**What Mimic does with them:**
Quality checks from observable reality. Rollback on failure. System load gates dispatch.

---

## hermes-agent

**Principles taken:**
- Error classification: retryable vs permanent vs auth vs rate-limit.
- Streaming health: stale-stream detection.

**What Mimic does with them:**
Errors classified before retry. Health monitoring for quality.

---

## graphify

**Principles taken:**
- Confidence labels on extracted patterns.
- IDF-weighted relevance for quality ranking.

**What Mimic does with them:**
Extraction confidence feeds artifact_precision. Relevance ranking for quality.

---

## caveman

**Principles taken:**
- Sensitive path protection: quality of data includes no credential leakage.
- File type detection: correct handling preserves quality.

**What Mimic does with them:**
Quality checks include security scan.

---

## binary-mesh Measurements

**Principles taken:**
- L0 (Completeness): 8000 measured.
- L4 (Usefulness): 2100 measured.
- L6 (Speed): 1931 measured.
- L8 (Correctness): 8005 measured.
- L9 (Completion): 7072 measured.

**What Mimic does with them:**
Thresholds derived from real data. L4 and L9 under-sampled (23 and 50 samples) → marked UNCERTAIN.

---

## Standard Practice

**Principles taken:**
- Pair programming / code review (2-vote analogy).
- Budget management.
- Timeout enforcement.
- Rollback procedures.
- Error classification.

**What Mimic does with them:**
Standard software quality practices applied to AI operations.
