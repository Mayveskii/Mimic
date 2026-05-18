# Memory Domain — Sources

Where the memory domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- Mesh storage: slots stored in mmap-backed files.
- Index memory: B-tree and hash table structures in anonymous mmap.

**What Mimic does with them:**
Mesh slots stored in mmap regions. Anonymous mmap used for temporary buffers. MAP_PRIVATE ensures isolation.

**What Mimic does NOT copy:**
- Embryo's specific memory pool allocator.
- Embryo's GC integration (Mimic uses explicit free).

---

## gastown

**Principles taken:**
- Atomic allocation: memory allocated with collision check (flock-like for named regions).
- Rollback: free on failure.

**What Mimic does with them:**
Allocation tracked in context. Rollback frees all tracked regions.

---

## Standard POSIX Practice

**Principles taken:**
- Anonymous mmap for large buffers.
- msync before persistence.
- munmap on cleanup.
- PROT_READ for read-only data.

**What Mimic does with them:**
Standard POSIX mmap semantics. No custom allocator. Explicit lifecycle management.

---

## C Standard Library

**Principles taken:**
- Bounds checking: memcpy with size limit.
- NULL pointer checks.
- Size overflow checks before multiplication.

**What Mimic does with them:**
All memory operations include bounds and NULL checks. No raw memcpy without size validation.
