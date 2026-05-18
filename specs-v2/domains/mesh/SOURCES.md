# Mesh Domain — Sources

Where the mesh domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- Mesh graph slots: domain/layer/state_hash indexed.
- Mapstore: slot storage with invariant registry.
- Project map: SQLite-based file/symbol navigation.

**What Mimic does with them:**
Slots stored in mmap-backed binary format. Indexed by domain, layer, modality, name. Invariant registry for keyword lookup.

**What Mimic does NOT copy:**
- Embryo's specific SQLite schema.
- Embryo's Go mapstore implementation.

---

## graphify

**Principles taken:**
- AST extraction: two-pass structural + call-graph.
- IDF-weighted search.
- Hub-throttled traversal.

**What Mimic does with them:**
Index structure supports graph-like traversal. Hub patterns (very common) not over-represented.

---

## bun (PR #30412)

**Principles taken:**
- Session enrichment: every operation logged and indexed.
- Edit scope isolation: resource tracking.

**What Mimic does with them:**
Slot operations logged. Resource bitmask tracks mesh access.

---

## caveman

**Principles taken:**
- File type detection for slot content.
- Sensitive path protection for slot metadata.

**What Mimic does with them:**
Slot content typed appropriately. No credential leakage in slot metadata.

---

## Standard Database Practice

**Principles taken:**
- Append-only storage for audit trail.
- Atomic index updates.
- Backup before major writes.
- Bounded document sizes.
- Garbage collection.

**What Mimic does with them:**
Standard database/storage practices applied to mesh.
