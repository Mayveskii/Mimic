# Mesh Domain — Slot Storage and Index

How proven knowledge is stored, indexed, and retrieved.

---

## What This Domain Does

The mesh is the knowledge store. Every proven pattern, every validated process, every learned invariant becomes a slot in the mesh. Slots are compressed, hashed, indexed, and retrievable in O(1) by composite key or via semantic search.

The domain covers: bmap storage, slot indexing, invariant registry, snapshot management, drift detection.

---

## Processes

### slot_write

**When to use:**  
After distillation produces a new artifact, or after a model execution creates new proven knowledge.

**Goal:**  
Store artifact as compressed, verified slot in bmap, indexed for retrieval.

**Chain (semantically):**

1. Encode artifact as protobuf.
2. Compress: `OP_COMPRESS_GZIP(artifact_bytes)`.
3. Hash: `sha256_hash(compressed)` → stored_hash.
4. Write: `bmap_write_cell(slot_id, compressed)`.
5. Verify: read back, hash matches.
6. Index: `si_insert(domain, layer, modality, pattern_name)`.

**Hard constraints:**
- Every write is compressed. No uncompressed data in bmap.
- Every read verifies hash before decompress.
- Compression ratio tracked. Alert if ratio < 1.5 (possible corruption).

**Invariants:**
- slot_id = sha256 of content.
- hash(compressed) == stored_hash.
- index key = domain:layer:modality:pattern_name.

**Result:**
```
status: "success"
slot_id: "sha256..."
compressed_size: 2048
original_size: 4096
compression_ratio: 2.0
indexed: true
```

---

### slot_read

**When to use:**  
When retrieving a pattern or process from mesh.

**Goal:**  
Read slot, verify integrity, decompress, return artifact.

**Chain (semantically):**

1. Lookup index: `si_query_domain_layer(domain, layer)`.
2. Read cell: `bmap_read_cell(slot_id)`.
3. Verify hash: `sha256_hash(compressed) == stored_hash`.
4. Decompress: `OP_DECOMPRESS_GZIP(compressed)`.
5. Parse protobuf → artifact.
6. Return artifact.

**Hard constraints:**
- Hash mismatch → reject slot, mark as corrupted, log.
- Decompress failure → reject slot, log.

**Invariants:**
- No slot read without hash verification.
- No decompressed data without hash match.

**Result:**
```
status: "success"
artifact: <protobuf object>
verification: "sha256_match"
```

---

### snapshot_create

**When to use:**  
When capturing a point-in-time state of the entire mesh for drift detection or backup.

**Goal:**  
Create signed snapshot of current mesh state.

**Chain (semantically):**

1. Build snapshot: `snapshot_build(bmap)`.
2. Sign: `snapshot_sign(snapshot, key)`.
3. Save: `snapshot_write(snapshot, path)`.

**Invariants:**
- Snapshot is point-in-time consistent.
- Signature prevents tampering.

---

### drift_detect

**When to use:**  
When comparing current mesh state against previous snapshot.

**Goal:**  
Detect what changed in the mesh since last snapshot.

**Chain (semantically):**

1. Load previous snapshot.
2. Build current snapshot.
3. Diff: `snapshot_diff(current, previous)`.
4. Report: added slots, removed slots, modified slots.

**Invariants:**
- Diff is symmetric: shows what A has that B doesn't and vice versa.

---

## Principles From Sources

### embryo (pkg/mesh/, pkg/mapstore/)

**Principles taken:**
- Slots indexed by domain + layer + state_hash.
- Invariant registry: every slot has ≥1 invariant.
- Slot deduplication: `inv_dedup_check` prevents duplicate invariants.

**What Mimic does with them:**
Composite key indexing. Invariant registry per slot. Dedup checks before insert.

### embryo (pkg/survival/)

**Principles taken:**
- Survival index per slot: surviving_lines / total_lines_added.
- Z-density: (Σ survival_i × weight_i) / slot_volume.

**What Mimic does with them:**
Slots carry survival_index and z_density. Low survival = low trust. High z-density = high knowledge density.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "mesh" |
| layer | "infrastructure" |
| modality | "code" |
| invariants | ["compressed_and_verified", "index_consistent", "no_duplicate_slots"] |

---

## Cross-Domain Conflicts

Mesh domain is storage infrastructure. No direct conflicts, but:
- Concurrent slot writes to same domain → serialize to prevent index corruption.
- Snapshot during write → snapshot sees consistent state (read-lock or copy-on-write).
