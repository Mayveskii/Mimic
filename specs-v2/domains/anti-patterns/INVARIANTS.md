# Anti-Patterns Domain — Invariants

Rules that MUST hold for every process in the anti-patterns domain.

---

## AINV-01: Every Anti-Pattern Has Counter

**What it prevents:** Telling model what not to do without saying what to do.

**What it requires:** Every anti-pattern (NEGATIVE polarity) MUST link to a COUNTER pattern. No orphan anti-patterns. Counter pattern demonstrates the correct approach.

**Source of this rule:**
- Specs-v2 artifact schema.
- Teaching principle.

**Consequence of violation:** Anti-pattern without counter → NOT indexed. Log: "incomplete anti-pattern".

---

## AINV-02: Evidence Required

**What it prevents:** Vague, unproven anti-patterns.

**What it requires:** Every anti-pattern includes `failure_evidence`: concrete description of observed failure. Links to issue, PR, or commit where failure was documented.

**Source of this rule:**
- Evidence-based practice.
- binary-mesh history.

**Consequence of violation:** Missing evidence → anti-pattern NOT indexed.

---

## AINV-03: QAC Mapping Complete

**What it prevents:** Incomplete understanding of which quality gates are violated.

**What it requires:** Every anti-pattern maps to specific QAC codes it violates. At least one QAC mapping required. Up to all 13.

**Source of this rule:**
- Quality gate integration.
- Specs-v2 artifact schema.

**Consequence of violation:** Missing QAC mapping → anti-pattern indexed with warning.

---

## AINV-04: Root Cause Analysis

**What it prevents:** Symptom-level anti-patterns without understanding root cause.

**What it requires:** Every anti-pattern includes `missing_validation`: what specific validation was skipped (links to META_INVARIANT.md). Root cause = missing validation.

**Source of this rule:**
- META_INVARIANT.md: all 30 APs trace to missing validation.
- Root cause analysis principle.

**Consequence of violation:** Missing root cause → anti-pattern indexed but marked incomplete.

---

## AINV-05: Counter Pattern Tested

**What it prevents:** Counter pattern that doesn't actually work.

**What it requires:** Counter pattern linked from anti-pattern MUST have survival_index > 0.0 (proven in production). Counter pattern MUST have been successfully applied at least once (success_count > 0).

**Source of this rule:**
- Evidence-based practice.
- Survival signal requirement.

**Consequence of violation:** Counter without survival signal → anti-pattern NOT indexed until counter proven.

---

## AINV-06: No Duplicates

**What it prevents:** Redundant anti-patterns, mesh bloat.

**What it requires:** Before indexing, check for duplicate anti-patterns (same root cause, same violated QACs). Duplicate → merge or reject.

**Source of this rule:**
- Database normalization.
- Mesh efficiency.

**Consequence of violation:** Duplicate anti-pattern → REJECTED. Log: "duplicate of AP-NN".

---

## AINV-07: Naming Convention

**What it prevents:** Inconsistent anti-pattern identifiers.

**What it requires:** Anti-pattern IDs: `AP-NN` where NN = 01..99. Sequential. Counter pattern IDs: `CP-NN` matching anti-pattern. Counter slot name: `<domain>_<short_description>`.

**Source of this rule:**
- Naming standard.
- Traceability requirement.

**Consequence of violation:** Wrong ID format → REJECTED. Autocorrect attempted.

---

## AINV-08: Review Cycle

**What it prevents:** Stale anti-patterns, outdated counter patterns.

**What it requires:** Anti-patterns reviewed quarterly. If counter pattern survival_index drops below 0.5 → review counter. If anti-pattern no longer observed in 90 days → archive.

**Source of this rule:**
- Maintenance requirement.
- Standard review practice.

**Consequence of violation:** Stale anti-patterns remain in mesh. Quarterly review catches and archives.
