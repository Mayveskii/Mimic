# C-Core ExecContext Specification

Exact structure and semantics for execution context. The context tracks all resources opened during a chain, maintains state for rollback, and enforces session-level constraints.

---

## Structure Definition

```c
typedef struct {
    // Identity
    uint32_t context_id;       // Unique context ID (session-scoped)
    uint64_t start_time_ns;    // Chain start time (CLOCK_MONOTONIC)
    
    // File descriptor tracking
    int32_t* open_fds;         // Dynamic array of open FDs
    size_t fd_count;           // Number of tracked FDs
    size_t fd_capacity;        // Allocated capacity of open_fds array
    
    // Memory region tracking
    void** mmap_regions;       // Dynamic array of mmap pointers
    size_t* mmap_sizes;        // Parallel array of mmap sizes
    size_t mmap_count;         // Number of tracked mmaps
    size_t mmap_capacity;      // Allocated capacity
    
    // Execution statistics
    uint32_t current_op_id;    // Currently executing packet ID
    uint32_t total_ops;        // Total packets in chain
    uint32_t success_count;    // Successfully completed packets
    uint32_t error_count;      // Failed packets
    uint32_t retry_count;      // Total retries performed
    
    // State snapshots for rollback
    void* pre_state_blob;      // Serialized pre-chain state
    size_t pre_state_size;     // Size of pre_state_blob
    uint64_t pre_state_hash;   // xxHash64 of pre_state_blob
    
    // Resource conflict tracking
    uint64_t resource_bitmask; // Bitmask of locked resources
    uint8_t conflict_status;   // 0=none, 1=warning, 2=critical
    
    // Session linkage
    uint64_t session_budget_tokens;    // Remaining token budget
    uint64_t session_budget_time_ms;   // Remaining time budget
    uint8_t denial_count;             // Consecutive denials this session
    uint8_t circuit_broken;           // 0=normal, 1=circuit broken
    
    // Padding to 128-byte alignment
    uint8_t padding[36];
} ExecContext;
```

### Size

On 64-bit Linux:
- Base fields: 4 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 4 + 4 + 4 + 4 + 4 + 8 + 8 + 8 + 1 + 1 + 8 + 8 + 1 + 1 = 136 bytes
- Padding: 36 bytes → Total = 172 bytes → align to 192

`sizeof(ExecContext)` = 192 bytes (with 36-byte padding).

---

## Resource Bitmask Layout

The `resource_bitmask` is a 64-bit field where each bit represents a resource class. Two operations with overlapping bits MUST NOT execute concurrently or in conflicting order.

```
Bit 0  (0x0000000000000001): Git repository (primary working dir)
Bit 1  (0x0000000000000002): Git repository (secondary)
Bit 2  (0x0000000000000004): Git repository (tertiary)
Bit 3  (0x0000000000000008): Build output directory
Bit 4  (0x0000000000000010): Source file (single file lock)
Bit 5  (0x0000000000000020): Source directory
Bit 6  (0x0000000000000040): Network socket (outbound)
Bit 7  (0x0000000000000080): Network socket (inbound)
Bit 8  (0x0000000000000100): Process (PID space)
Bit 9  (0x0000000000000200): Environment variables
Bit 10 (0x0000000000000400): Session context (in-memory)
Bit 11 (0x0000000000000800): Mesh index (read)
Bit 12 (0x0000000000001000): Mesh index (write)
Bit 13 (0x0000000000002000): Mesh slots (read)
Bit 14 (0x0000000000004000): Mesh slots (write)
Bit 15 (0x0000000000008000): Backup/rollback storage
Bit 16 (0x0000000000010000): Credential pool
Bit 17 (0x0000000000020000): Temporary workspace
Bit 18 (0x0000000000040000): Cache directory
Bit 19 (0x0000000000080000): Log/output streams
Bit 20 (0x0000000000100000): mmap region (primary)
Bit 21 (0x0000000000200000): mmap region (secondary)
Bit 22-31: Reserved for domain-specific resources
Bit 32-63: Reserved for future use
```

### Bitmask Assignment by Opcode

| Opcode | Required Bits |
|---|---|
| OP_GIT_* | Bit 0 (or 1, 2 for multi-repo) |
| OP_BUILD_* | Bit 3 |
| OP_IO_READ/WRITE on file | Bit 4 |
| OP_IO_OPEN on directory | Bit 5 |
| OP_NET_* | Bit 6 |
| OP_PROC_* | Bit 8 |
| OP_SYS_ENV_* | Bit 9 |
| OP_SESS_* | Bit 10 |
| OP_MMAP_* | Bit 20 |
| OP_ORCH_PLAN (queries mesh) | Bit 11, 13 |
| OP_ORCH_EXEC | No bits (orchestrator is coordinator) |
| OP_HASH_* | No bits (pure compute) |

### Conflict Rules for Bitmask

1. Two operations with ANY shared bit MUST be serialized.
2. Read-only operations (OP_FLAG_READONLY) on same bit: ALLOW parallel if both are readonly.
3. One readonly + one write on same bit: MUST serialize (write after read).
4. Two writes on same bit: MUST serialize.
5. OP_GIT_COMMIT (write to bit 0) + OP_GIT_STATUS (readonly on bit 0): ALLOW parallel ONLY if status reads from index (not working tree). If status reads working tree, serialize after commit.
6. OP_BUILD_COMPILE (bit 3) + OP_IO_READ on source file in build dir (bit 4): ALLOW parallel (readonly on source does not conflict with write to output).
7. OP_SYS_ENV_SET (bit 9) + any op that reads env: MUST serialize (env is global).
8. OP_ORCH_EXEC does not set bits but checks bits of the chain it executes.

---

## FD Tracking

### Adding an FD

When an operation opens a file descriptor:
1. FD is added to `open_fds[fd_count]`.
2. `fd_count++`.
3. If `fd_count == fd_capacity`, realloc with doubling strategy.
4. Max tracked FDs: 1024 per context.

### Removing an FD

When an operation closes a file descriptor:
1. Search `open_fds` for the FD value.
2. If found: swap with last element, `fd_count--`.
3. If not found: log warning (double-close or untracked FD).

### Validation

Before any operation uses an FD:
1. If fd_in != -1: must exist in `open_fds`.
2. If fd_out != -1: must exist in `open_fds`.
3. Operation OP_IO_CLOSE: FD must be in tracking.
4. Operation that opens new FD: must not exceed 1024 tracked.

### Cleanup

On chain completion (success or failure):
1. Close all remaining FDs in `open_fds`.
2. Free `open_fds` array.
3. Set `fd_count = fd_capacity = 0`.

---

## Mmap Tracking

### Adding a Region

When OP_MMAP_ALLOC succeeds:
1. Pointer added to `mmap_regions[mmap_count]`.
2. Size added to `mmap_sizes[mmap_count]`.
3. `mmap_count++`.
4. Bit 20 set in `resource_bitmask`.

### Removing a Region

When OP_MMAP_FREE succeeds:
1. Search `mmap_regions` for pointer.
2. If found: call munmap, swap with last, `mmap_count--`.
3. If `mmap_count == 0`: clear bit 20.

### Cleanup

On chain completion:
1. For all remaining regions: `ops_mmap_sync()` then `ops_mmap_free()`.
2. Free `mmap_regions` and `mmap_sizes` arrays.
3. Set `mmap_count = mmap_capacity = 0`.

---

## State Snapshots

### Recording Pre-State

Before chain execution:
1. Serialize all tracked resources into `pre_state_blob`:
   - Current working directory.
   - Environment variables (filtered, no secrets).
   - Git HEAD hash (if in git repo).
   - List of tracked FDs and their paths (via /proc/self/fd).
   - List of tracked mmap regions.
2. Compute `pre_state_hash = xxHash64(pre_state_blob)`.

### Rollback

On rollback request:
1. Re-read current state.
2. Compute diff against `pre_state_blob`.
3. For each changed resource:
   - Git repo: `git checkout` to pre-state HEAD.
   - Files: restore from backup copies (created during write ops).
   - Env: restore from pre-state values.
   - FDs: close all, reopen pre-state FDs (best effort).
4. Verify: re-compute state hash, compare to `pre_state_hash`.
5. If mismatch: `ERR_ROLLBACK_FAIL`. Log detailed diff.

---

## Budget Tracking

### Token Budget

`session_budget_tokens` is decremented by:
- `OpCodeDef.cost_tokens` for each executed packet.
- `ops_calculate_action()` total for the chain (validation check).
- Context append cost (OP_SESS_CONTEXT_APPEND bytes / 4 ≈ tokens).

If budget would go negative during VALIDATE: chain REJECTED with `ERR_BUDGET_EXCEEDED`.
If budget goes negative during EXEC (unexpected, e.g., cost underestimated): immediate halt, no new ops started.

### Time Budget

`session_budget_time_ms` is decremented by actual chain latency.
If timeout would exceed remaining time: `ERR_TIMEOUT` with "budget exhausted" reason.

### Denial Tracking

Every `ERR_PERMISSION_DENY` increments `denial_count`.
Every successful operation resets `denial_count` to 0.
If `denial_count >= 3`: set `circuit_broken = 1`.
When `circuit_broken == 1`: ALL ops return `ERR_CIRCUIT_BREAK`. Only manual reset clears this.

---

## Thread Safety

ExecContext is NOT thread-safe. All operations on a context must be from a single thread or externally synchronized.
Parallel chains use separate ExecContexts. Resource bitmask conflicts between contexts are checked by the orchestrator before parallel dispatch.

---

## Context Lifecycle

```
1. ALLOC:   ctx = calloc(1, sizeof(ExecContext)); ctx->context_id = next_id++;
2. INIT:    ops_execute_chain() sets ctx->start_time_ns, total_ops, budgets from session.
3. EXEC:    Executor modifies success_count, error_count, current_op_id, tracks resources.
4. RESULT:  Chain completes. Resources cleaned. Metrics extracted.
5. FREE:    free(ctx->open_fds); free(ctx->mmap_regions); free(ctx->mmap_sizes);
            free(ctx->pre_state_blob); free(ctx);
```

`pre_state_blob` is freed in step 5 regardless of rollback success. If rollback needed state and failed, the failure is logged but blob is still freed (state is lost, which is why rollback must succeed).
