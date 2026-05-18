# Memory Domain — Invariants

Rules that MUST hold for every process in the memory domain.

---

## MINV-01: Size Validated Before Alloc

**What it prevents:** OOM, allocation of impossibly large regions.

**What it requires:** OP_MMAP_ALLOC size > 0 and ≤ session memory budget AND ≤ 1GB (global max). Zero or negative size = REJECTED.

**Source of this rule:**
- Standard memory management.
- AP-28 (missing overflow guard): integer overflow → incorrect values.

**Consequence of violation:** Allocation REJECTED with `ERR_INVALID_ARG` or `ERR_OOM`.

---

## MINV-02: Pointer Tracked Before Free

**What it prevents:** Double-free, use-after-free, freeing untracked memory.

**What it requires:** OP_MMAP_FREE validates pointer exists in ExecContext.mmap_regions. Size must match allocation size. Pointer must not be NULL.

**Source of this rule:**
- EXEC_CONTEXT_SPEC.md: mmap tracking.
- AP-03 (panic recovery instead of idempotent close).

**Consequence of violation:** Free REJECTED with `ERR_INVALID_ARG`. Untracked pointer = warning logged, NOT freed (prevents corruption).

---

## MINV-03: Private Anonymous Mapping

**What it prevents:** Shared memory corruption, inter-process data leakage.

**What it requires:** OP_MMAP_ALLOC uses MAP_PRIVATE | MAP_ANONYMOUS only. No MAP_SHARED. No file-backed mmap (use IO domain for file-backed).

**Source of this rule:**
- Security isolation principle.
- Standard practice for temporary buffers.

**Consequence of violation:** Request with MAP_SHARED or file-backed → REJECTED with `ERR_PERMISSION_DENY`.

---

## MINV-04: Sync Before Visibility

**What it prevents:** Data loss on crash, torn writes.

**What it requires:** Before data in mmap region is considered "persisted" (passed to another operation, written to disk), OP_MMAP_SYNC MUST be called. MS_SYNC (blocking) for critical data. MS_ASYNC for non-critical.

**Source of this rule:**
- Standard POSIX mmap semantics.
- gastown: observable reality requires sync for durability.

**Consequence of violation:** Data may be lost on crash. Warning: "data not synced before handoff".

---

## MINV-05: Cleanup on Chain Completion

**What it prevents:** Memory leaks, resource exhaustion across sessions.

**What it requires:** ALL tracked mmap regions freed on chain completion (success or failure). Auto-cleanup at chain end. Manual free preferred but not required.

**Source of this rule:**
- EXEC_CONTEXT_SPEC.md: cleanup procedure.
- AP-15 (partial rollback): orphaned resources.

**Consequence of violation:** Memory leak detected at chain cleanup. Warning logged. If leak persists across chains → `ERR_OOM` on future allocations.

---

## MINV-06: No Write Beyond Bounds

**What it prevents:** Buffer overflows, memory corruption.

**What it requires:** OP_MMAP_WRITE offset + length ≤ allocation size. Bounds checked before write.

**Source of this rule:**
- AP-26 (out-of-bound write).
- Standard bounds checking.

**Consequence of violation:** Write REJECTED with `ERR_INVALID_ARG`. Offset and size reported in error.

---

## MINV-07: Read-Only Regions Protected

**What it prevents:** Accidental modification of read-only data (slot storage, index).

**What it requires:** Regions marked read-only (PROT_READ only) cannot be written. OP_MMAP_WRITE on read-only region = REJECTED.

**Source of this rule:**
- Memory protection principle.
- Slot storage immutability.

**Consequence of violation:** Write REJECTED with `ERR_PERMISSION_DENY`. Attempt logged.
