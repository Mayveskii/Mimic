# Memory Domain — MMap, Alloc, Free, Sync

How Mimic manages memory-mapped regions for large data operations.

---

## What This Domain Does

Memory operations allocate, read, write, and synchronize memory-mapped regions. These are used for large buffers (RAG vectors, mesh slots, session snapshots) where heap allocation would be inefficient or impossible. The model never directly manipulates memory — it receives tokens that operate on allocated regions.

---

## Processes

### mmap_alloc

**When to use:**  
When a process needs a large contiguous memory region (RAG vectors, compressed slots, session state).

**Goal:**  
Allocate memory-mapped region with specific properties.

**Chain (semantically):**

1. Request size validated (≤ max allowed, typically 1GB).
2. Allocate: `OP_MMAP_ALLOC(size)` with MAP_PRIVATE | MAP_ANONYMOUS.
3. Return pointer + size to requesting process.

**Hard constraints:**
- Size must be > 0.
- Size must be ≤ session memory budget.
- Pointer tracked in ExecContext.

**Invariants:**
- Allocated region is private (not shared between processes).
- Anonymous (not backed by file).
- Tracked in ExecContext for cleanup on chain completion.

---

### mmap_free

**When to use:**  
When a memory region is no longer needed.

**Goal:**  
Release memory region.

**Chain (semantically):**

1. Validate pointer is tracked in ExecContext.
2. Validate pointer ≠ NULL, size ≠ 0.
3. Free: `OP_MMAP_FREE(ptr, size)`.
4. Remove from ExecContext tracking.

**Hard constraints:**
- Never free NULL pointer.
- Never free untracked pointer.
- Size must match allocation size.

**Invariants:**
- Freed region no longer accessible.
- ExecContext updated to reflect deallocation.

---

### mmap_sync

**When to use:**  
When data in memory region must be persisted to disk or made visible to other processes.

**Goal:**  
Synchronize memory region with backing storage.

**Chain (semantically):**

1. Validate pointer is tracked.
2. Sync: `OP_MMAP_SYNC(ptr, size)` with MS_SYNC.
3. Verify sync completed.

**Hard constraints:**
- Sync only tracked regions.
- MS_SYNC blocks until complete.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "memory" |
| layer | "process" |
| modality | "code" |
| invariants | ["private_anonymous", "tracked_in_context", "size_matches"] |

---

## Cross-Domain Conflicts

Memory domain conflicts with:
- **build domain**: cannot free memory during compilation (race with linker).
- **system domain**: cannot alloc while system exec is modifying process state.
