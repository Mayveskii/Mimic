# Mesh Domain — Invariants

Rules that MUST hold for every process in the mesh domain.

---

## MSHINV-01: Slot Immutable After Creation

**What it prevents:** Corruption of proven patterns, loss of historical data.

**What it requires:** Once a slot is written to mesh storage, it is NEVER modified. Updates create NEW slots with new slot_id. Old slots marked `SLOT_FLAG_ARCHIVED`.

**Source of this rule:**
- Immutable data principle.
- Audit trail requirement.

**Consequence of violation:** Attempt to modify slot → REJECTED. New slot created instead.

---

## MSHINV-02: Hash Chain Integrity

**What it prevents:** Corrupted or tampered slot content.

**What it requires:** Every slot header and text content has SHA-256 hash. On read, hash verified. Mismatch → slot quarantined, error logged, previous version used if available.

**Source of this rule:**
- Data integrity standard.
- SLOT_SCHEMA.md.

**Consequence of violation:** Corrupted slot skipped during load. Log: "slot hash mismatch, quarantined". Mesh may lose one pattern but remains operational.

---

## MSHINV-03: Index Atomic Updates

**What it prevents:** Inconsistent index state, missing or ghost entries.

**What it requires:** Index updates are atomic. New slots appended first, then index updated. If crash during index update → next open rebuilds index from slots. No partial index state visible.

**Source of this rule:**
- Database ACID principles.
- Standard indexing best practice.

**Consequence of violation:** Inconsistent index detected on open. Auto-rebuild triggered. May cause temporary performance degradation.

---

## MSHINV-04: Backup Before Major Write

**What it prevents:** Catastrophic data loss from corruption or bugs.

**What it requires:** Before any write operation that modifies > 1% of mesh file, create backup `.bmap.YYYYMMDD.HHMMSS`. Keep last 7 backups. Auto-cleanup after 30 days.

**Source of this rule:**
- Standard backup practice.
- gastown: rollback capability.

**Consequence of violation:** No backup for major write → warning logged. If corruption occurs without backup → data loss possible. Manual recovery required.

---

## MSHINV-05: Slot Size Bounded

**What it prevents:** Individual slots consuming excessive storage, OOM on load.

**What it requires:** Max slot size: header (5200 bytes) + text (65536 bytes) = 70736 bytes. Text > 64KB → truncate or split into multiple linked slots.

**Source of this rule:**
- Resource management.
- Standard document size limits.

**Consequence of violation:** Oversized slot → REJECTED at storage. Text truncated or split.

---

## MSHINV-06: Domain Consistency

**What it prevents:** Slots in wrong domain, mis-categorized patterns.

**What it requires:** `domain` field MUST be valid DomainEnum (0-15). `layer` field valid LayerEnum (0-2). `modality` field valid ModalityEnum (0-3). Invalid values → REJECTED.

**Source of this rule:**
- Schema validation.
- Type safety.

**Consequence of violation:** Invalid domain/layer/modality → slot REJECTED. Model receives "invalid domain classification".

---

## MSHINV-07: Polarity Rules Enforced

**What it prevents:** Incomplete anti-patterns, orphaned negative slots.

**What it requires:** `polarity == NEGATIVE` → `counter_slot_id` MUST be non-zero and valid. `polarity == COUNTER` → linked from at least one NEGATIVE. `polarity == POSITIVE` → no counter required.

**Source of this rule:**
- Artifact schema polarity rules.
- RINV-02 (NEGATIVE with COUNTER).

**Consequence of violation:** NEGATIVE without counter → slot NOT indexed. Log: "incomplete anti-pattern". COUNTER without backlink → allowed but flagged for review.

---

## MSHINV-08: Survival Index Non-Negative

**What it prevents:** Invalid survival signal, division by zero in ranking.

**What it requires:** `survival_index` ≥ 0.0 and ≤ 1.0. `z_density` ≥ 0.0 and ≤ 1.0. `artifact_precision` ≥ 0.0 and ≤ 1.0. Out of range → REJECTED.

**Source of this rule:**
- Statistical validity.
- Quality signal integrity.

**Consequence of violation:** Invalid quality signals → slot NOT indexed. Log: "invalid quality signals".

---

## MSHINV-09: Retrieval Statistics Updated

**What it prevents:** Stale usage data, incorrect ranking.

**What it requires:** `retrieval_count`, `success_count`, `failure_count` updated atomically on every access. Race-safe counter increments (atomic operations).

**Source of this rule:**
- Analytics accuracy.
- Standard counter best practice.

**Consequence of violation:** Stale counters → ranking may be inaccurate. Not critical but affects optimization decisions.

---

## MSHINV-10: Garbage Collection of Orphans

**What it prevents:** Storage bloat from unused slots.

**What it requires:** Monthly scan for orphaned slots: `retrieval_count == 0` AND `age > 90 days` AND `survival_index < 0.1`. Orphaned slots moved to cold storage (compressed archive). NOT deleted.

**Source of this rule:**
- Storage management.
- Data retention policy.

**Consequence of violation:** Storage grows indefinitely. Performance degrades over time. GC keeps storage bounded.
