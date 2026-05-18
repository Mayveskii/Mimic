# C-Core Conflict Matrix Specification

Exact layout, population rules, and conflict levels for the operation conflict matrix.

The conflict matrix answers one question: can two operations run in the same chain, or concurrently, without corrupting state?

---

## Matrix Layout

```c
// C core: static uint8_t g_conflict_matrix[OP_MAX][OP_MAX];
// Go core: [][]ConflictLevel where ConflictLevel is uint8

typedef uint8_t ConflictLevel;

#define CONFLICT_NONE   0
#define CONFLICT_LOW    1   // Warning, logged, execution continues
#define CONFLICT_MEDIUM 2   // Requires attention, validation FAILS
#define CONFLICT_HIGH   3   // Critical, validation FAILS
#define CONFLICT_FATAL  4   // Unresolvable, validation FAILS, session logged
```

Matrix size: `OP_MAX × OP_MAX` = 256 × 256 = 65,536 bytes.
Symmetry: `matrix[i][j] == matrix[j][i]`. Only upper triangle stored in some implementations, but C core stores full for O(1) lookup.
Diagonal: `matrix[i][i]` = CONFLICT_LOW for most ops (same operation twice in chain is usually redundant but not dangerous). Exceptions:
- `OP_GIT_COMMIT × OP_GIT_COMMIT` = CONFLICT_HIGH (double commit is a bug).
- `OP_SYS_FILE_DELETE × OP_SYS_FILE_DELETE` = CONFLICT_HIGH (double delete = race or bug).
- `OP_BUILD_DEPLOY × OP_BUILD_DEPLOY` = CONFLICT_FATAL (double deploy to same target).

---

## Population Rules

### Rule 1: Domain Self-Conflicts

Operations within the same domain that touch the same resource conflict.

```
// Git: all git ops share resource bit 0 (primary repo)
// Any two git ops on same repo → at least LOW
OP_GIT_STATUS × OP_GIT_COMMIT = CONFLICT_MEDIUM (status during commit = dirty read)
OP_GIT_ADD × OP_GIT_COMMIT = CONFLICT_NONE (add then commit is the normal flow)
OP_GIT_COMMIT × OP_GIT_PUSH = CONFLICT_NONE (commit then push is normal)
OP_GIT_CHECKOUT × any_git_op = CONFLICT_HIGH (checkout changes working tree)
OP_GIT_MERGE × any_git_op = CONFLICT_HIGH (merge changes index and tree)
OP_GIT_REBASE × any_git_op = CONFLICT_FATAL (rebase is destructive and long)
OP_GIT_CLONE × OP_GIT_CLONE = CONFLICT_LOW (parallel clones of different repos = OK, same repo = race)
OP_GIT_FETCH × OP_GIT_PUSH = CONFLICT_MEDIUM (fetch then push on same remote = potential non-FF)

// Build: all build ops share resource bit 3 (build output dir)
OP_BUILD_COMPILE × OP_BUILD_COMPILE = CONFLICT_LOW (parallel compiles of different targets = OK)
OP_BUILD_COMPILE × OP_BUILD_CLEAN = CONFLICT_FATAL (clean during compile = lost objects)
OP_BUILD_LINK × OP_BUILD_COMPILE = CONFLICT_HIGH (link needs all objects)
OP_BUILD_TEST × OP_BUILD_COMPILE = CONFLICT_HIGH (test reads compiled binaries)
OP_BUILD_DEPLOY × any_build_op = CONFLICT_HIGH (deploy needs stable output)

// I/O: file ops share bit 4 (file) or bit 5 (directory)
OP_IO_READ × OP_IO_READ = CONFLICT_NONE (parallel reads are safe)
OP_IO_READ × OP_IO_WRITE = CONFLICT_MEDIUM (read during write = torn read)
OP_IO_WRITE × OP_IO_WRITE = CONFLICT_HIGH (parallel writes = corruption)
OP_IO_OPEN × OP_IO_CLOSE = CONFLICT_NONE (open then close is normal)
OP_IO_CLOSE × OP_IO_READ = CONFLICT_HIGH (read after close = EBADF)

// Network: all network ops share bit 6
OP_NET_HTTP_GET × OP_NET_HTTP_GET = CONFLICT_NONE (parallel GETs are safe)
OP_NET_HTTP_GET × OP_NET_HTTP_POST = CONFLICT_LOW (order matters but not structurally conflicting)
OP_NET_TCP_SEND × OP_NET_TCP_RECV = CONFLICT_NONE (send then recv is normal bidirectional)
OP_NET_TCP_CONNECT × OP_NET_TCP_CLOSE = CONFLICT_NONE (connect then close is normal)
OP_NET_WEBSOCKET × any_net_op = CONFLICT_LOW (websocket holds connection)

// Process: all process ops share bit 8 (PID space)
OP_PROC_SPAWN × OP_PROC_KILL = CONFLICT_NONE (spawn then kill is normal control)
OP_PROC_KILL × OP_PROC_KILL = CONFLICT_HIGH (double kill of same PID = race)
OP_PROC_WAIT × OP_PROC_KILL = CONFLICT_LOW (kill unblocks wait, order matters)
OP_PROC_SPAWN × OP_PROC_SPAWN = CONFLICT_NONE (parallel spawns are safe)

// Memory: all mmap ops share bit 20
OP_MMAP_ALLOC × OP_MMAP_FREE = CONFLICT_NONE (alloc then free is normal)
OP_MMAP_FREE × OP_MMAP_READ = CONFLICT_HIGH (use after free)
OP_MMAP_WRITE × OP_MMAP_SYNC = CONFLICT_NONE (write then sync is normal)
```

### Rule 2: Cross-Domain Conflicts

Operations from different domains that share resource bits conflict.

```
// Git (bit 0) + I/O (bit 4/5) on same repo path
OP_GIT_STATUS × OP_IO_READ (on .git/index) = CONFLICT_MEDIUM (git locks index during status)
OP_GIT_COMMIT × OP_IO_WRITE (any file in repo) = CONFLICT_HIGH (commit updates index and tree)
OP_GIT_CHECKOUT × OP_IO_READ (any file in repo) = CONFLICT_HIGH (checkout may overwrite file)

// Build (bit 3) + I/O (bit 4/5) on build output
OP_BUILD_COMPILE × OP_IO_WRITE (to output dir) = CONFLICT_HIGH (compiler writes, external write corrupts)
OP_BUILD_TEST × OP_IO_READ (test binary) = CONFLICT_NONE (test reads its own binary)

// Network (bit 6) + Git (bit 0) via remote
OP_NET_HTTP_GET × OP_GIT_FETCH = CONFLICT_NONE (different protocols)
OP_NET_TCP_CONNECT × OP_GIT_PUSH = CONFLICT_LOW (push opens its own TCP connection)

// System (bit 9 env) + any op that reads env
OP_SYS_ENV_SET × OP_SYS_ENV_GET = CONFLICT_LOW (set then get sees new value, not a bug)
OP_SYS_ENV_SET × OP_PROC_SPAWN = CONFLICT_MEDIUM (spawn inherits env, set after spawn has no effect)
```

### Rule 3: Orchestrator Meta-Ops

Orchestrator ops (0x90-0x9A) do not conflict with domain ops because they are meta-operations. They operate on in-memory session state, not shared resources.

```
OP_ORCH_CLASSIFY × any_op = CONFLICT_NONE
OP_ORCH_PLAN × any_op = CONFLICT_NONE
OP_ORCH_VALIDATE × any_op = CONFLICT_NONE
OP_ORCH_EXEC × any_op = CONFLICT_NONE (but OP_ORCH_EXEC holds chain lock)
OP_ORCH_VERIFY × any_op = CONFLICT_NONE
OP_ORCH_RESPOND × any_op = CONFLICT_NONE

// Session ops (bit 10)
OP_SESS_BUDGET_CHECK × any_op = CONFLICT_NONE
OP_SESS_CONTEXT_APPEND × OP_SESS_CONTEXT_APPEND = CONFLICT_LOW (append order matters)
OP_SESS_DENIAL_RECORD × any_op = CONFLICT_NONE
```

### Rule 4: Utility Ops

Utility ops (hash, compress, encrypt) are pure functions. They conflict with nothing.

```
OP_HASH_SHA256 × any_op = CONFLICT_NONE
OP_HASH_MD5 × any_op = CONFLICT_NONE
OP_COMPRESS_GZIP × any_op = CONFLICT_NONE
OP_DECOMPRESS_GZIP × any_op = CONFLICT_NONE
OP_ENCRYPT_AES × any_op = CONFLICT_NONE
OP_DECRYPT_AES × any_op = CONFLICT_NONE
```

Exception: `OP_ENCRYPT_AES` × `OP_DECRYPT_AES` with same key = CONFLICT_LOW (order-dependent, not state-corrupting).

### Rule 5: Dangerous Ops Always Conflict with Themselves

Any operation with `OP_FLAG_DANGEROUS` has `CONFLICT_HIGH` with itself in the same chain.

```
OP_GIT_COMMIT × OP_GIT_COMMIT = CONFLICT_HIGH
OP_GIT_PUSH × OP_GIT_PUSH = CONFLICT_HIGH
OP_GIT_CHECKOUT × OP_GIT_CHECKOUT = CONFLICT_HIGH
OP_BUILD_DEPLOY × OP_BUILD_DEPLOY = CONFLICT_FATAL
OP_PROC_KILL × OP_PROC_KILL = CONFLICT_HIGH
OP_SYS_EXEC × OP_SYS_EXEC = CONFLICT_HIGH
```

---

## Conflict Matrix Initialization

```c
void ops_init_conflict_matrix(void) {
    // Default: all operations are compatible
    memset(g_conflict_matrix, CONFLICT_NONE, sizeof(g_conflict_matrix));
    
    // Set diagonal for dangerous ops
    for (int i = 0; i < OP_MAX; i++) {
        if (g_op_registry[i].flags & OP_FLAG_DANGEROUS) {
            g_conflict_matrix[i][i] = CONFLICT_HIGH;
        }
    }
    
    // Git domain conflicts
    g_conflict_matrix[OP_GIT_STATUS][OP_GIT_COMMIT] = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_GIT_COMMIT][OP_GIT_STATUS] = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_GIT_CHECKOUT][OP_GIT_ADD] = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_ADD][OP_GIT_CHECKOUT] = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_MERGE][OP_GIT_COMMIT] = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_COMMIT][OP_GIT_MERGE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_REBASE][OP_GIT_STATUS] = CONFLICT_FATAL;
    g_conflict_matrix[OP_GIT_STATUS][OP_GIT_REBASE] = CONFLICT_FATAL;
    
    // Build domain conflicts
    g_conflict_matrix[OP_BUILD_COMPILE][OP_BUILD_CLEAN] = CONFLICT_FATAL;
    g_conflict_matrix[OP_BUILD_CLEAN][OP_BUILD_COMPILE] = CONFLICT_FATAL;
    g_conflict_matrix[OP_BUILD_LINK][OP_BUILD_COMPILE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_COMPILE][OP_BUILD_LINK] = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_TEST][OP_BUILD_COMPILE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_COMPILE][OP_BUILD_TEST] = CONFLICT_HIGH;
    
    // I/O domain conflicts
    g_conflict_matrix[OP_IO_READ][OP_IO_WRITE] = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_IO_WRITE][OP_IO_READ] = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_IO_WRITE][OP_IO_WRITE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_CLOSE][OP_IO_READ] = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_READ][OP_IO_CLOSE] = CONFLICT_HIGH;
    
    // Memory domain conflicts
    g_conflict_matrix[OP_MMAP_FREE][OP_MMAP_READ] = CONFLICT_HIGH;
    g_conflict_matrix[OP_MMAP_READ][OP_MMAP_FREE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_MMAP_FREE][OP_MMAP_WRITE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_MMAP_WRITE][OP_MMAP_FREE] = CONFLICT_HIGH;
    
    // Cross-domain: Git + I/O
    g_conflict_matrix[OP_GIT_COMMIT][OP_IO_WRITE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_WRITE][OP_GIT_COMMIT] = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_CHECKOUT][OP_IO_READ] = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_READ][OP_GIT_CHECKOUT] = CONFLICT_HIGH;
    
    // Cross-domain: Build + I/O on output dir
    g_conflict_matrix[OP_BUILD_COMPILE][OP_IO_WRITE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_WRITE][OP_BUILD_COMPILE] = CONFLICT_HIGH;
    
    // Process domain
    g_conflict_matrix[OP_PROC_KILL][OP_PROC_KILL] = CONFLICT_HIGH;
    g_conflict_matrix[OP_PROC_WAIT][OP_PROC_KILL] = CONFLICT_LOW;
    g_conflict_matrix[OP_PROC_KILL][OP_PROC_WAIT] = CONFLICT_LOW;
    
    // Session
    g_conflict_matrix[OP_SESS_CONTEXT_APPEND][OP_SESS_CONTEXT_APPEND] = CONFLICT_LOW;
}
```

---

## Conflict Level Semantics

| Level | Name | Validation Result | Log Level | Action |
|---|---|---|---|---|
| 0 | NONE | Pass | — | Execute |
| 1 | LOW | Pass | WARN | Execute, log warning |
| 2 | MEDIUM | FAIL | ERROR | Reject chain |
| 3 | HIGH | FAIL | ERROR | Reject chain |
| 4 | FATAL | FAIL | FATAL | Reject chain, mark session |

A FATAL conflict increments the session denial count and may trigger circuit break if repeated.

---

## Resource Bitmask vs Conflict Matrix

Two orthogonal systems:

1. **Conflict Matrix**: Static, pre-defined. Checks if two opcodes CAN coexist in a chain. Determined by opcode semantics, not runtime state.
2. **Resource Bitmask**: Dynamic, runtime. Checks if two in-flight chains hold the same resource. Determined by what files/sockets/processes are actually open.

Usage:
- Validation (PLAN phase): conflict matrix + resource_bitmask of current session state.
- Execution (EXEC phase): resource_bitmask updated after each op.
- Parallel chains: resource_bitmask compared between contexts.

---

## Adding New Conflicts

When a new opcode is registered:
1. Define its default resource bits (which bits it touches).
2. Add explicit conflict rules with all opcodes that share ANY bit AND are not readonly.
3. If the new op is dangerous, set diagonal to CONFLICT_HIGH.
4. Document the conflict in CONFLICT_RULES.md (this file).
5. Update validation tests to verify the conflict is caught.

Never add runtime-generated conflicts. All conflicts must be statically defined and reviewed.
