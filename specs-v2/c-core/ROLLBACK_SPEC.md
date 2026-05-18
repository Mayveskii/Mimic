# C-Core Rollback Specification

Exact rollback semantics, inverse operation mapping, and state restoration procedures.

Rollback is the guarantee that a failed atomic chain leaves the system in the same state as before the chain started. Without rollback, atomicity is a lie.

---

## When Rollback Triggers

1. **Atomic chain failure**: Any packet with `OP_FLAG_ATOMIC` fails, and any preceding packet was executed.
2. **Explicit rollback request**: Orchestrator requests rollback after VERIFY failure.
3. **Validation failure after partial execution**: This should never happen (validation runs before exec), but defensive code handles it.

---

## Rollback State Machine

```
CHAIN_EXECUTING
    ↓ packet[i] fails
    ↓ (packet[i].flags & OP_FLAG_ATOMIC) || (any previous packet flags & OP_FLAG_ATOMIC)
ROLLBACK_INIT
    ↓
ROLLBACK_INVERSE  ← for j = i-1 down to 0
    ↓ reversible? → execute inverse
    ↓ not reversible? → best-effort cleanup
ROLLBACK_VERIFY
    ↓ state hash matches pre_state_hash?
    ↓ yes → ROLLBACK_SUCCESS
    ↓ no → ROLLBACK_FAIL (irreversible change detected)
```

---

## Inverse Operation Mapping

Every reversible operation MUST have a registered inverse. The inverse is stored in `OpCodeDef`:

```c
typedef struct {
    OpCode opcode;
    char name[OP_NAME_LEN];
    char description[OP_DESC_LEN];
    int (*execute)(OpPacket* packet);
    
    // NEW: Inverse operation
    OpCode inverse_opcode;     // OP_NOP if not reversible
    int (*inverse_execute)(OpPacket* original, OpPacket* inverse_result);
    
    uint32_t required_flags;
    uint32_t forbidden_flags;
    float cost_tokens;
    float cost_time_us;
    float cost_memory_bytes;
    uint8_t safety_level;
    bool is_reversible;
    bool is_atomic;
} OpCodeDef;
```

### Inverse Table

| Opcode | Inverse | Inverse Logic |
|---|---|---|
| OP_MMAP_ALLOC | OP_MMAP_FREE | Free with same pointer and size |
| OP_MMAP_FREE | OP_MMAP_ALLOC | Re-alloc with saved pointer and size (from pre-state snapshot) |
| OP_IO_OPEN | OP_IO_CLOSE | Close the FD |
| OP_IO_CLOSE | OP_IO_OPEN | Re-open with same path and mode (from pre-state snapshot) |
| OP_SYS_DIR_CREATE | OP_SYS_DIR_REMOVE | Remove directory (if empty) |
| OP_SYS_FILE_COPY | OP_SYS_FILE_DELETE | Delete the destination file |
| OP_SYS_FILE_MOVE | OP_SYS_FILE_MOVE | Move back (swap src/dst) |
| OP_SYS_ENV_SET | OP_SYS_ENV_SET | Set back to previous value (from pre-state snapshot) |
| OP_GIT_ADD | OP_GIT_ADD | `git reset HEAD <paths>` (unstage) |
| OP_GIT_COMMIT | — | NOT REVERSIBLE (push to remote = irreversible) |
| OP_GIT_BRANCH | — | NOT REVERSIBLE (if branch pushed to remote) |
| OP_GIT_CHECKOUT | OP_GIT_CHECKOUT | Checkout back to previous branch (from pre-state snapshot) |
| OP_BUILD_COMPILE | — | NOT REVERSIBLE (output files overwritten) |
| OP_BUILD_CLEAN | — | NOT REVERSIBLE (deleted files) |
| OP_PROC_SPAWN | OP_PROC_KILL | Kill the PID |
| OP_NET_TCP_CONNECT | OP_NET_TCP_CLOSE | Close the socket |
| OP_SESS_CONTEXT_APPEND | OP_SESS_CONTEXT_APPEND | Truncate context to pre-append length |

### Non-Reversible Operations

The following operations are NEVER reversible. If they appear before an atomic op in a chain, the chain cannot be fully rolled back:

- `OP_GIT_COMMIT` (once pushed, forever in history)
- `OP_GIT_PUSH` (irreversible network action)
- `OP_GIT_MERGE` (merge commit is in history)
- `OP_GIT_REBASE` (rewrites history, irreversible)
- `OP_BUILD_DEPLOY` (deployed code is live)
- `OP_SYS_EXEC` (arbitrary command, no inverse)
- `OP_PROC_KILL` (killed process cannot be resurrected)
- `OP_SYS_FILE_DELETE` (deleted file, unless backup exists)
- `OP_SYS_DIR_REMOVE` (removed directory)
- `OP_IO_WRITE` (overwritten file, unless backup exists)

These ops can still be used in chains, but:
1. They MUST NOT appear before an atomic op unless a snapshot exists before them.
2. They trigger `ERR_ATOMIC_BREAK` at validation if so positioned.
3. Best-effort cleanup attempts to restore from backups but logs irreversible changes.

---

## Rollback Procedure

```c
int ops_rollback_chain(OpPacket* packets, uint32_t failed_index, ExecContext* ctx) {
    if (!ctx->pre_state_blob) {
        // No snapshot taken = cannot rollback
        return ERR_ROLLBACK_FAIL;
    }
    
    // Phase 1: Execute inverses for reversible ops
    for (int j = (int)failed_index - 1; j >= 0; j--) {
        OpCodeDef* def = &g_op_registry[packets[j].opcode];
        
        if (def->is_reversible && def->inverse_execute) {
            OpPacket inverse;
            ops_packet_init(&inverse, def->inverse_opcode);
            
            // Populate inverse args from original packet + pre-state snapshot
            int inv_result = def->inverse_execute(&packets[j], &inverse);
            
            if (inv_result != 0) {
                // Inverse failed = partial rollback
                // Log which operation could not be reversed
                // Continue with best-effort cleanup
            }
        } else {
            // Not reversible: attempt best-effort cleanup
            ops_best_effort_cleanup(&packets[j], ctx);
        }
    }
    
    // Phase 2: Resource cleanup
    // Close all tracked FDs
    for (size_t k = 0; k < ctx->fd_count; k++) {
        close(ctx->open_fds[k]);
    }
    ctx->fd_count = 0;
    
    // Sync and free all tracked mmaps
    for (size_t k = 0; k < ctx->mmap_count; k++) {
        ops_mmap_sync(ctx->mmap_regions[k], ctx->mmap_sizes[k]);
        ops_mmap_free(ctx->mmap_regions[k], ctx->mmap_sizes[k]);
    }
    ctx->mmap_count = 0;
    
    // Phase 3: State verification
    uint64_t current_hash = ops_compute_state_hash(ctx);
    if (current_hash == ctx->pre_state_hash) {
        return ERR_OK;  // Rollback successful
    } else {
        // State mismatch = irreversible change leaked through
        return ERR_ROLLBACK_FAIL;
    }
}
```

---

## Best-Effort Cleanup

For non-reversible operations, attempt the following:

| Operation | Best-Effort Cleanup |
|---|---|
| OP_IO_WRITE | If backup exists: restore from `.mimic/backups/<file>.<timestamp>.sha256.gz` |
| OP_SYS_FILE_DELETE | If backup exists: restore from backup. If not: log permanent loss. |
| OP_SYS_DIR_REMOVE | If backup exists: restore directory tree. If not: log permanent loss. |
| OP_SYS_FILE_COPY | Delete the copied file. |
| OP_SYS_FILE_MOVE | Move back (may fail if src was overwritten). |
| OP_BUILD_COMPILE | Delete output artifacts from this compilation. |
| OP_BUILD_TEST | No cleanup needed (test does not mutate source). |
| OP_GIT_COMMIT | `git reset --soft HEAD~1` (uncommit, keep changes staged). |
| OP_GIT_BRANCH | `git branch -D <name>` (delete branch). |
| OP_GIT_CHECKOUT | `git checkout -` (checkout previous branch). |
| OP_PROC_SPAWN | `kill -9 <pid>` if still running. |
| OP_NET_HTTP_POST | No cleanup (POST already sent). Log the irreversible action. |

---

## Backup Creation

Before any non-reversible or destructive operation, create backup:

```c
int ops_create_backup(const char* path) {
    // 1. Compute hash of current file content
    // 2. Compress with gzip -6 (speed/size balance)
    // 3. Store at .mimic/backups/<filename>.<timestamp>.<hash>.gz
    // 4. Return backup path
}
```

Backup policy:
- Created before OP_IO_WRITE on existing file.
- Created before OP_SYS_FILE_MOVE on existing destination.
- Created before OP_SYS_FILE_COPY if destination exists.
- Created before OP_SYS_FILE_DELETE.
- Created before OP_BUILD_CLEAN (backup entire output dir).
- Created before OP_GIT_CHECKOUT if uncommitted changes exist (git stash).
- NOT created before OP_GIT_COMMIT (git history IS the backup).
- NOT created before OP_PROC_KILL (process state cannot be backed up).

Backups are cleaned up after:
- 24 hours (temporary backups).
- Successful rollback (immediately).
- Orphaned backup detection: if backup file older than 7 days, delete.

---

## State Snapshot Format

The `pre_state_blob` is a serialized representation of system state before chain execution:

```
[4 bytes]  magic = 0x52424C4B ('RBLK')
[4 bytes]  version = 1
[8 bytes]  timestamp_ns
[8 bytes]  cwd_hash (xxHash64 of CWD string)
[4 bytes]  env_count
[env_count × variable]:
  [2 bytes] name_len
  [name_len bytes] name
  [2 bytes] value_len
  [value_len bytes] value
[4 bytes]  git_repo_count
[repo_count × repo]:
  [2 bytes] path_len
  [path_len bytes] path
  [20 bytes] HEAD hash (binary)
  [4 bytes] branch_name_len
  [branch_name_len bytes] branch_name
[4 bytes]  fd_count
[fd_count × fd]:
  [4 bytes] fd number
  [2 bytes] path_len
  [path_len bytes] path (from /proc/self/fd/)
[4 bytes]  mmap_count
[mmap_count × mmap]:
  [8 bytes] pointer (as uint64)
  [8 bytes] size
[8 bytes]  state_hash (xxHash64 of entire blob)
```

The blob is compressed with LZ4 for storage efficiency.
The hash at the end is verified before rollback to detect blob corruption.

---

## Rollback Guarantees

| Scenario | Guarantee |
|---|---|
| All ops reversible | Full rollback, state hash matches |
| Some ops non-reversible, backups exist | Full rollback if backups restore successfully |
| Some ops non-reversible, no backups | Best-effort, state hash mismatch likely, logged |
| No pre_state_blob | ERR_ROLLBACK_FAIL, no attempt made |
| Inverse execution fails | Partial rollback, remaining ops cleaned best-effort, logged |
| State hash mismatch after rollback | ERR_ROLLBACK_FAIL, detailed diff logged |

The golden rule: rollback failure is always logged with full details. The model receives honest report: "operation failed, rollback attempted, state may be partially modified, here is the diff."

---

## Performance

- Snapshot creation: O(state_size). Typically < 10ms for normal sessions.
- Rollback execution: O(chain_length × inverse_cost). Typically < 100ms for 10-op chains.
- State hash verification: O(state_size).
- Total rollback time budget: 5 seconds. Exceeding → force cleanup + ERR_ROLLBACK_FAIL.
