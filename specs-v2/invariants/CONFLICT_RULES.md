# Cross-Domain Conflict Rules

Rules that determine when operations from different domains cannot coexist in the same chain or execute concurrently.

These rules extend the per-domain conflict matrices. They are checked during VALIDATE phase after per-domain checks pass.

---

## Rule Set

### CR-01: Git Working Tree Lock

**What:** Any git operation that modifies the working tree or index holds a lock on the repository.

**Effect:** While git lock is held, no I/O write, build compile, or system file operation may target the same repository path.

**Applies to:** OP_GIT_COMMIT, OP_GIT_CHECKOUT, OP_GIT_MERGE, OP_GIT_REBASE, OP_GIT_ADD

**Conflicts with:** OP_IO_WRITE, OP_BUILD_COMPILE, OP_SYS_FILE_COPY, OP_SYS_FILE_MOVE, OP_SYS_FILE_DELETE

**Level:** CONFLICT_HIGH

---

### CR-02: Build Output Directory Lock

**What:** Build operations write to the build output directory.

**Effect:** While build is active, no I/O write to output directory, no git checkout that includes output dir, no system file operation on output dir.

**Applies to:** OP_BUILD_COMPILE, OP_BUILD_LINK, OP_BUILD_TEST (if it writes test output)

**Conflicts with:** OP_IO_WRITE, OP_GIT_CHECKOUT, OP_SYS_FILE_DELETE, OP_BUILD_CLEAN

**Level:** CONFLICT_HIGH

---

### CR-03: Environment Variable Global State

**What:** Environment variables are process-global.

**Effect:** After OP_SYS_ENV_SET, any operation that reads env variables may see the new value. This is not a conflict per se, but a data dependency that must be tracked.

**Applies to:** OP_SYS_ENV_SET

**Conflicts with:** OP_PROC_SPAWN (env inherited), OP_SYS_EXEC (env visible), OP_BUILD_COMPILE (makefile reads env)

**Level:** CONFLICT_LOW (warning, logged)

---

### CR-04: Network Socket Port Binding

**What:** TCP listen operations bind to a port.

**Effect:** While a port is bound, no other bind to the same port.

**Applies to:** OP_NET_TCP_CONNECT (if it calls bind), OP_NET_WEBSOCKET (if server)

**Conflicts with:** Same opcode on same port

**Level:** CONFLICT_HIGH

---

### CR-05: Process PID Space

**What:** Process operations refer to PIDs.

**Effect:** OP_PROC_KILL on a PID that OP_PROC_SPAWN created is normal. OP_PROC_KILL on a PID that OP_PROC_WAIT is waiting on is a race.

**Applies to:** OP_PROC_KILL, OP_PROC_WAIT

**Conflicts with:** Each other on same PID

**Level:** CONFLICT_MEDIUM (allowed but warned)

---

### CR-06: File Descriptor Lifecycle

**What:** FDs are integers that can be reused after close.

**Effect:** OP_IO_CLOSE followed by OP_IO_READ on the same fd number is a bug (stale FD). But different FD numbers are independent.

**Applies to:** OP_IO_CLOSE

**Conflicts with:** Any I/O op using the same fd_in/fd_out value

**Level:** CONFLICT_HIGH

---

### CR-07: Session Context Serialization

**What:** Session context is single-threaded append-only.

**Effect:** OP_SESS_CONTEXT_APPEND operations must be serialized. Parallel append would corrupt context order.

**Applies to:** OP_SESS_CONTEXT_APPEND

**Conflicts with:** OP_SESS_CONTEXT_APPEND

**Level:** CONFLICT_LOW (serialized automatically by session lock)

---

### CR-08: Mesh Index Write Lock

**What:** Mesh index updates require exclusive write lock.

**Effect:** While index is being updated (new slot indexed), no query may read the index.

**Applies to:** OP_ORCH_PLAN (if it triggers mesh storage of novel pattern)

**Conflicts with:** OP_ORCH_CLASSIFY (reads mesh for classification hints)

**Level:** CONFLICT_LOW (reads use MVCC snapshot)

---

### CR-09: Credential Pool Access

**What:** Credential pool is a singleton resource.

**Effect:** Any operation that reads credentials (OP_NET_HTTP_GET with auth, OP_BUILD_DEPLOY with SSH key) accesses the credential pool.

**Applies to:** OP_NET_HTTP_GET, OP_NET_HTTP_POST, OP_BUILD_DEPLOY, OP_GIT_PUSH (if authenticated)

**Conflicts with:** Credential pool refresh/update operations

**Level:** CONFLICT_LOW (read-sharing allowed)

---

### CR-10: Backup Storage Allocation

**What:** Backup creation allocates disk space.

**Effect:** If disk is near full, backup creation may fail. This is not a structural conflict but a resource exhaustion risk.

**Applies to:** OP_IO_WRITE (if file exists), OP_SYS_FILE_DELETE (creates backup), OP_BUILD_CLEAN

**Conflicts with:** Nothing structural

**Level:** CONFLICT_NONE (checked at runtime, not validation)

---

## Conflict Detection Algorithm

```c
CrossDomainConflict check_cross_domain_conflict(OpPacket* packets, uint32_t count) {
    for (uint32_t i = 0; i < count; i++) {
        for (uint32_t j = i + 1; j < count; j++) {
            OpCode op1 = packets[i].opcode;
            OpCode op2 = packets[j].opcode;
            
            // Check resource bitmask overlap
            uint64_t mask1 = get_resource_mask(op1, &packets[i]);
            uint64_t mask2 = get_resource_mask(op2, &packets[j]);
            
            if (mask1 & mask2) {
                // Overlap detected, check if readonly
                bool ro1 = is_readonly(op1) || (packets[i].flags & OP_FLAG_READONLY);
                bool ro2 = is_readonly(op2) || (packets[j].flags & OP_FLAG_READONLY);
                
                if (ro1 && ro2) {
                    // Both readonly: no conflict
                    continue;
                }
                
                // At least one is write: conflict level depends on rules
                ConflictLevel level = get_cross_domain_level(op1, op2);
                if (level > CONFLICT_NONE) {
                    return (CrossDomainConflict){.found = true, .pair = {i, j}, .level = level};
                }
            }
        }
    }
    return (CrossDomainConflict){.found = false};
}
```

---

## Resource Mask Assignment

See EXEC_CONTEXT_SPEC.md for full bitmask layout.

The `get_resource_mask()` function maps each opcode to its resource bits based on argument values:
- Git ops: bit 0 (or 1, 2 if `path` arg specifies different repo)
- Build ops: bit 3 (or derived from `target` arg)
- I/O ops: bit 4 or 5 (derived from `path` arg)
- Network ops: bit 6
- Process ops: bit 8
- System env ops: bit 9
- Session ops: bit 10
- Mesh ops: bit 11 or 12
- Memory ops: bit 20
- Utility ops: 0 (pure compute)

---

## Enforcement

Cross-domain conflicts are checked:
1. During VALIDATE phase (static: opcodes and argument paths known).
2. During parallel pipeline dispatch (dynamic: resource_bitmask of each context compared).
3. Never during single-chain sequential execution (chain is already serialized).

A cross-domain conflict at validation time → chain REJECTED with `ERR_CONFLICT` and specific rule reference.
A cross-domain conflict at parallel dispatch time → pipelines SERIALIZED, not rejected.
