# C-Core Opcode Specification

Exact OpCode enumeration, flag definitions, and semantic behavior for the Mimic physics-based core.

Every OpCode is a token. Every token has: a numeric value, a name, a safety level, required/forbidden flags, energy cost profile, and execution semantics.

No implementation hints. No "maybe". Every field is exact.

---

## Constants

```c
#define MAX_OPS             1024
#define MAX_PACKET_SIZE     256
#define MAX_ARGS            16
#define OP_NAME_LEN         32
#define OP_DESC_LEN         128

#define OP_FLAG_SAFE        0x01
#define OP_FLAG_READONLY    0x02
#define OP_FLAG_ATOMIC      0x04
#define OP_FLAG_REVERSIBLE  0x08
#define OP_FLAG_NETWORK     0x10
#define OP_FLAG_DISK        0x20
#define OP_FLAG_MEMORY      0x40
#define OP_FLAG_DANGEROUS   0x80
```

### Flag Semantics

- `OP_FLAG_SAFE (0x01)`: Operation cannot cause system-wide side effects. May be retried without harm.
- `OP_FLAG_READONLY (0x02)`: Operation reads only. No state mutation. Never conflicts with another readonly operation on the same resource.
- `OP_FLAG_ATOMIC (0x04)`: Operation must complete entirely or not at all. On chain failure, if any preceding op in the chain had this flag, the entire chain rolls back.
- `OP_FLAG_REVERSIBLE (0x08)`: Operation has an inverse. Rollback calls the inverse function with recorded state.
- `OP_FLAG_NETWORK (0x10)`: Operation involves network I/O. Subject to timeout. Retryable on transient errors.
- `OP_FLAG_DISK (0x20)`: Operation involves filesystem I/O. Subject to disk-full checks.
- `OP_FLAG_MEMORY (0x40)`: Operation involves memory allocation/mutation. Subject to OOM checks.
- `OP_FLAG_DANGEROUS (0x80)`: Operation can cause irreversible loss. ALWAYS requires explicit model confirmation. Bypasses auto-allow.

---

## OpCode Enumeration

```c
typedef enum {
    OP_NOP = 0,
    
    // Memory Operations (0x10-0x14)
    OP_MMAP_ALLOC   = 0x10,
    OP_MMAP_FREE    = 0x11,
    OP_MMAP_READ    = 0x12,
    OP_MMAP_WRITE   = 0x13,
    OP_MMAP_SYNC    = 0x14,
    
    // I/O Operations (0x20-0x24)
    OP_IO_READ      = 0x20,
    OP_IO_WRITE     = 0x21,
    OP_IO_OPEN      = 0x22,
    OP_IO_CLOSE     = 0x23,
    OP_IO_SEEK      = 0x24,
    
    // Git Operations (0x30-0x3B)
    OP_GIT_INIT     = 0x30,
    OP_GIT_CLONE    = 0x31,
    OP_GIT_FETCH    = 0x32,
    OP_GIT_STATUS   = 0x33,
    OP_GIT_DIFF     = 0x34,
    OP_GIT_ADD      = 0x35,
    OP_GIT_COMMIT   = 0x36,
    OP_GIT_PUSH     = 0x37,
    OP_GIT_CHECKOUT = 0x38,
    OP_GIT_BRANCH   = 0x39,
    OP_GIT_MERGE    = 0x3A,
    OP_GIT_REBASE   = 0x3B,
    OP_GIT_TAG      = 0x3C,       // Create annotated tag
    OP_GIT_RESET    = 0x3D,       // Reset (NEVER-RULE: blocked without explicit emergency flag)
    
    // Build Operations (0x40-0x44)
    OP_BUILD_COMPILE = 0x40,
    OP_BUILD_LINK    = 0x41,
    OP_BUILD_TEST    = 0x42,
    OP_BUILD_DEPLOY  = 0x43,
    OP_BUILD_CLEAN   = 0x44,
    
    // Network Operations (0x50-0x56)
    OP_NET_HTTP_GET     = 0x50,
    OP_NET_HTTP_POST    = 0x51,
    OP_NET_TCP_CONNECT  = 0x52,
    OP_NET_TCP_SEND     = 0x53,
    OP_NET_TCP_RECV     = 0x54,
    OP_NET_TCP_CLOSE    = 0x55,
    OP_NET_WEBSOCKET    = 0x56,
    
    // Process Operations (0x60-0x64)
    OP_PROC_SPAWN   = 0x60,
    OP_PROC_WAIT    = 0x61,
    OP_PROC_KILL    = 0x62,
    OP_PROC_SIGNAL  = 0x63,
    
    // Utility Operations (0x70-0x75)
    OP_HASH_SHA256  = 0x70,
    OP_HASH_MD5     = 0x71,
    OP_COMPRESS_GZIP = 0x72,
    OP_DECOMPRESS_GZIP = 0x73,
    OP_ENCRYPT_AES  = 0x74,
    OP_DECRYPT_AES  = 0x75,
    
    // System Operations (0x80-0x89)
    OP_SYS_EXEC         = 0x80,
    OP_SYS_ENV_GET      = 0x81,
    OP_SYS_ENV_SET      = 0x82,
    OP_SYS_FILE_EXISTS  = 0x83,
    OP_SYS_DIR_CREATE   = 0x84,
    OP_SYS_DIR_REMOVE   = 0x85,
    OP_SYS_FILE_COPY    = 0x86,
    OP_SYS_FILE_MOVE    = 0x87,
    OP_SYS_FILE_DELETE  = 0x88,
    OP_SYS_CHMOD        = 0x89,
    
    // Session / Orchestrator Operations (0x90-0x9A)
    OP_SESS_BUDGET_CHECK    = 0x90,
    OP_SESS_CONTEXT_APPEND  = 0x91,
    OP_SESS_DENIAL_RECORD   = 0x92,
    OP_SESS_SNAPSHOT        = 0x93,
    OP_SESS_COMPRESS        = 0x94,
    OP_ORCH_CLASSIFY        = 0x95,
    OP_ORCH_PLAN            = 0x96,
    OP_ORCH_VALIDATE        = 0x97,
    OP_ORCH_EXEC            = 0x98,
    OP_ORCH_VERIFY          = 0x99,
    OP_ORCH_RESPOND         = 0x9A,
    
    // Research Operations (0xA0-0xAC)
    OP_RESEARCH_HYPOTHESIS_CREATE     = 0xA0,  // Create falsifiable hypothesis
    OP_RESEARCH_HYPOTHESIS_LOAD       = 0xA1,  // Load existing hypothesis
    OP_RESEARCH_HYPOTHESIS_INFERENCE  = 0xA2,  // Confirm/refute/refine
    OP_RESEARCH_EXPERIMENT_RUN        = 0xA3,  // Execute experiment
    OP_RESEARCH_RESULT_STORE          = 0xA4,  // Store result with hash
    OP_RESEARCH_STATISTICAL_TEST      = 0xA5,  // Compute significance/CI
    OP_RESEARCH_LITERATURE_FETCH      = 0xA6,  // Download/fetch paper
    OP_RESEARCH_LITERATURE_PARSE      = 0xA7,  // Extract structured info
    OP_RESEARCH_LITERATURE_INDEX      = 0xA8,  // Index as mesh slot
    OP_RESEARCH_CITATION_LINK         = 0xA9,  // Link cited papers
    OP_RESEARCH_LITERATURE_EMBED      = 0xAA,  // Generate embedding
    OP_RESEARCH_PROGRESS_STORE        = 0xAB,  // Store key findings
    OP_RESEARCH_CONTEXT_SUMMARIZE     = 0xAC,  // Semantic compression
    
    // Self-Management Operations (0xB0-0xB5)
    OP_SELF_CHECKPOINT_CREATE   = 0xB0,  // Snapshot current state
    OP_SELF_CHECKPOINT_RESTORE  = 0xB1,  // Load from snapshot
    OP_SELF_BUDGET_REALLOCATE   = 0xB2,  // Rebalance budget
    OP_SELF_STRATEGY_PIVOT      = 0xB3,  // Switch approach after N failures
    OP_SELF_PROGRESS_ASSESS     = 0xB4,  // Compare actual vs planned
    OP_SELF_CONTEXT_SUMMARIZE   = 0xB5,  // Semantic compression of context
    
    OP_MAX = 0xFF
} OpCode;
```

### Opcode Value Ranges by Domain

| Range | Domain | Count |
|---|---|---|
| 0x00 | nop | 1 |
| 0x10-0x1F | memory | 5 |
| 0x20-0x2F | io | 5 |
| 0x30-0x3D | git | 14 |
| 0x40-0x4F | build | 5 |
| 0x50-0x5F | network | 7 |
| 0x60-0x6F | process | 4 |
| 0x70-0x7F | utility | 6 |
| 0x80-0x8F | system | 10 |
| 0x90-0x9A | session/orchestrator | 11 |
| 0xA0-0xAC | research | 13 |
| 0xB0-0xB5 | self-management | 6 |
| 0xB6-0xFE | reserved | 74 |

Reserved opcodes MUST return `ERR_INVALID_OPCODE` if executed unregistered.

---

## Default Flag Assignment

Every opcode has a default flag set used by `ops_register_builtins()`. Custom registrations may override with `forbidden_flags` enforcement.

| Opcode | Default Flags | Safety Level | Atomic | Reversible |
|---|---|---|---|---|
| OP_NOP | SAFE \| READONLY | 3 | yes | yes |
| OP_MMAP_ALLOC | MEMORY | 2 | yes | no |
| OP_MMAP_FREE | MEMORY | 2 | yes | no |
| OP_MMAP_READ | SAFE \| READONLY | 3 | yes | yes |
| OP_MMAP_WRITE | MEMORY \| DISK | 2 | yes | no |
| OP_MMAP_SYNC | DISK | 2 | yes | no |
| OP_IO_READ | SAFE \| READONLY | 3 | yes | yes |
| OP_IO_WRITE | DISK | 2 | yes | no |
| OP_IO_OPEN | DISK | 2 | yes | no |
| OP_IO_CLOSE | DISK | 2 | yes | no |
| OP_IO_SEEK | SAFE \| READONLY | 3 | yes | yes |
| OP_GIT_INIT | DISK | 2 | no | no |
| OP_GIT_CLONE | NETWORK \| DISK | 1 | no | no |
| OP_GIT_FETCH | NETWORK | 2 | no | no |
| OP_GIT_STATUS | SAFE \| READONLY | 3 | yes | yes |
| OP_GIT_DIFF | SAFE \| READONLY | 3 | yes | yes |
| OP_GIT_ADD | DISK | 2 | yes | no |
| OP_GIT_COMMIT | DISK \| DANGEROUS | 0 | no | no |
| OP_GIT_PUSH | NETWORK \| DANGEROUS | 0 | no | no |
| OP_GIT_CHECKOUT | DISK \| DANGEROUS | 1 | no | no |
| OP_GIT_BRANCH | DISK | 2 | no | no |
| OP_GIT_MERGE | DISK \| DANGEROUS | 0 | no | no |
| OP_GIT_REBASE | DISK \| DANGEROUS | 0 | no | no |
| OP_GIT_TAG | DISK | 1 | no | no |
| OP_GIT_RESET | DISK \| DANGEROUS | 0 | no | no |
| OP_BUILD_COMPILE | DISK | 2 | no | no |
| OP_BUILD_LINK | DISK | 2 | no | no |
| OP_BUILD_TEST | SAFE | 3 | no | yes |
| OP_BUILD_DEPLOY | NETWORK \| DANGEROUS | 0 | no | no |
| OP_BUILD_CLEAN | DISK \| DANGEROUS | 1 | no | no |
| OP_NET_HTTP_GET | NETWORK | 2 | yes | yes |
| OP_NET_HTTP_POST | NETWORK | 2 | yes | no |
| OP_NET_TCP_CONNECT | NETWORK | 2 | no | no |
| OP_NET_TCP_SEND | NETWORK | 2 | yes | no |
| OP_NET_TCP_RECV | NETWORK | 2 | yes | yes |
| OP_NET_TCP_CLOSE | NETWORK | 2 | yes | no |
| OP_NET_WEBSOCKET | NETWORK | 1 | no | no |
| OP_PROC_SPAWN | DANGEROUS | 0 | no | no |
| OP_PROC_WAIT | SAFE | 3 | yes | yes |
| OP_PROC_KILL | DANGEROUS | 0 | no | no |
| OP_PROC_SIGNAL | DANGEROUS | 1 | no | no |
| OP_HASH_SHA256 | SAFE \| READONLY | 3 | yes | yes |
| OP_HASH_MD5 | SAFE \| READONLY | 3 | yes | yes |
| OP_COMPRESS_GZIP | SAFE | 3 | yes | yes |
| OP_DECOMPRESS_GZIP | SAFE | 3 | yes | yes |
| OP_ENCRYPT_AES | SAFE | 2 | yes | yes |
| OP_DECRYPT_AES | SAFE | 2 | yes | yes |
| OP_SYS_EXEC | DANGEROUS | 0 | no | no |
| OP_SYS_ENV_GET | SAFE \| READONLY | 3 | yes | yes |
| OP_SYS_ENV_SET | DANGEROUS | 1 | no | no |
| OP_SYS_FILE_EXISTS | SAFE \| READONLY | 3 | yes | yes |
| OP_SYS_DIR_CREATE | DISK | 2 | yes | no |
| OP_SYS_DIR_REMOVE | DISK \| DANGEROUS | 1 | no | no |
| OP_SYS_FILE_COPY | DISK | 2 | yes | no |
| OP_SYS_FILE_MOVE | DISK | 2 | yes | no |
| OP_SYS_FILE_DELETE | DISK \| DANGEROUS | 1 | no | no |
| OP_SYS_CHMOD | DISK \| DANGEROUS | 1 | no | no |
| OP_SESS_BUDGET_CHECK | SAFE \| READONLY | 3 | yes | yes |
| OP_SESS_CONTEXT_APPEND | MEMORY | 2 | yes | no |
| OP_SESS_DENIAL_RECORD | SAFE | 3 | yes | no |
| OP_SESS_SNAPSHOT | DISK | 2 | yes | no |
| OP_SESS_COMPRESS | SAFE | 3 | yes | yes |
| OP_ORCH_CLASSIFY | SAFE \| READONLY | 3 | yes | yes |
| OP_ORCH_PLAN | SAFE | 3 | yes | no |
| OP_ORCH_VALIDATE | SAFE \| READONLY | 3 | yes | yes |
| OP_ORCH_EXEC | DANGEROUS | 0 | no | no |
| OP_ORCH_VERIFY | SAFE \| READONLY | 3 | yes | yes |
| OP_ORCH_RESPOND | SAFE | 3 | yes | no |
| OP_RESEARCH_HYPOTHESIS_CREATE | SAFE | 3 | yes | no |
| OP_RESEARCH_HYPOTHESIS_LOAD | SAFE \| READONLY | 3 | yes | yes |
| OP_RESEARCH_HYPOTHESIS_INFERENCE | SAFE | 3 | yes | no |
| OP_RESEARCH_EXPERIMENT_RUN | DANGEROUS | 1 | no | no |
| OP_RESEARCH_RESULT_STORE | DISK \| DANGEROUS | 2 | yes | no |
| OP_RESEARCH_STATISTICAL_TEST | SAFE \| READONLY | 3 | yes | yes |
| OP_RESEARCH_LITERATURE_FETCH | NETWORK | 2 | yes | yes |
| OP_RESEARCH_LITERATURE_PARSE | SAFE | 3 | yes | yes |
| OP_RESEARCH_LITERATURE_INDEX | DISK \| DANGEROUS | 2 | yes | no |
| OP_RESEARCH_CITATION_LINK | SAFE | 3 | yes | no |
| OP_RESEARCH_LITERATURE_EMBED | MEMORY | 2 | yes | yes |
| OP_RESEARCH_PROGRESS_STORE | DISK | 2 | yes | no |
| OP_RESEARCH_CONTEXT_SUMMARIZE | SAFE | 3 | yes | yes |
| OP_SELF_CHECKPOINT_CREATE | DISK | 2 | yes | no |
| OP_SELF_CHECKPOINT_RESTORE | SAFE | 3 | yes | no |
| OP_SELF_BUDGET_REALLOCATE | SAFE \| READONLY | 3 | yes | yes |
| OP_SELF_STRATEGY_PIVOT | SAFE | 3 | yes | no |
| OP_SELF_PROGRESS_ASSESS | SAFE \| READONLY | 3 | yes | yes |
| OP_SELF_CONTEXT_SUMMARIZE | SAFE | 3 | yes | yes |

---

## Safety Levels

| Level | Name | Meaning | Auto-Allow | 2-Vote Required |
|---|---|---|---|---|
| 0 | CRITICAL | Destructive, irreversible, production-affecting | NEVER | ALWAYS |
| 1 | DANGEROUS | Can lose data or break state | NEVER | ON FLAG_DANGEROUS |
| 2 | MUTATING | Side effects but recoverable | IF not denied | NO |
| 3 | SAFE | Readonly or no side effects | ALWAYS | NO |

Circuit break rule: 3 consecutive denials at any level → manual mode only for remainder of session.

---

## Semantics Per Domain

### Memory (0x10-0x14)

Every memory operation operates on anonymous mmap regions. No file-backed mmap in core. File-backed I/O uses IO domain.

- `OP_MMAP_ALLOC`: Allocates anonymous mmap. Args: `size` (int). Returns: `ptr` (blob handle).
- `OP_MMAP_FREE`: Frees mmap region. Args: `ptr` (blob handle), `size` (int). Must match alloc size.
- `OP_MMAP_READ`: Read from mmap. Args: `ptr`, `offset`, `length`. Returns: data blob.
- `OP_MMAP_WRITE`: Write to mmap. Args: `ptr`, `offset`, `data`. Overwrites in place.
- `OP_MMAP_SYNC`: msync(MS_SYNC). Args: `ptr`, `size`.

### I/O (0x20-0x24)

File descriptor based I/O. FDs tracked in ExecContext.

- `OP_IO_OPEN`: Opens file. Args: `path`, `mode` (r/rw/a). Returns: `fd`.
- `OP_IO_CLOSE`: Closes FD. Args: `fd`. Removes from ExecContext tracking.
- `OP_IO_READ`: Reads from FD. Args: `fd`, `offset`, `length`. Returns: data.
- `OP_IO_WRITE`: Writes to FD. Args: `fd`, `offset`, `data`.
- `OP_IO_SEEK`: Seeks FD. Args: `fd`, `whence` (0=SET, 1=CUR, 2=END), `offset`.

### Git (0x30-0x3B)

Git operations run in a specified working directory. All git ops share a resource_bitmask bit per repo path.

- `OP_GIT_STATUS`: Returns working tree state as structured data.
- `OP_GIT_DIFF`: Returns diff output. Args: `pathspec` (optional), `cached` (bool for staged).
- `OP_GIT_ADD`: Stages files. Args: `paths` (array of strings).
- `OP_GIT_COMMIT`: Creates commit. Args: `message`, `author`. Returns: `hash`.
- `OP_GIT_CHECKOUT`: Switches branch or ref. Args: `target`. DANGEROUS if uncommitted changes exist.
- `OP_GIT_BRANCH`: Creates branch. Args: `name`, `start_point`.
- `OP_GIT_MERGE`: Merges. Args: `source`. DANGEROUS always.
- `OP_GIT_REBASE`: Rebases. Args: `upstream`. DANGEROUS always.
- `OP_GIT_INIT`: Init repo. Args: `path`.
- `OP_GIT_CLONE`: Clone repo. Args: `url`, `path`, `depth`.
- `OP_GIT_FETCH`: Fetch refs. Args: `remote`.
- `OP_GIT_PUSH`: Push refs. Args: `remote`, `refspec`. DANGEROUS always.

### Build (0x40-0x44)

- `OP_BUILD_COMPILE`: Compile. Args: `target`, `flags`. Captures stdout/stderr.
- `OP_BUILD_LINK`: Link. Args: `inputs`, `output`.
- `OP_BUILD_TEST`: Run tests. Args: `filter`, `timeout_ms`. Returns: pass/fail counts.
- `OP_BUILD_DEPLOY`: Deploy. Args: `target`, `version`. DANGEROUS always.
- `OP_BUILD_CLEAN`: Clean build artifacts. Args: `target`. DANGEROUS (deletes files).

### Network (0x50-0x56)

- `OP_NET_HTTP_GET`: HTTP GET. Args: `url`, `headers`, `timeout_ms`. Returns: status, body.
- `OP_NET_HTTP_POST`: HTTP POST. Args: `url`, `headers`, `body`, `timeout_ms`.
- `OP_NET_TCP_CONNECT`: TCP connect. Args: `host`, `port`, `timeout_ms`. Returns: `fd`.
- `OP_NET_TCP_SEND`: Send on TCP FD. Args: `fd`, `data`.
- `OP_NET_TCP_RECV`: Recv on TCP FD. Args: `fd`, `length`, `timeout_ms`. Returns: data.
- `OP_NET_TCP_CLOSE`: Close TCP FD. Args: `fd`.
- `OP_NET_WEBSOCKET`: WebSocket handshake. Args: `url`, `protocols`.

### Process (0x60-0x64)

- `OP_PROC_SPAWN`: Spawn process. Args: `cmd`, `args`, `env`, `cwd`. Returns: `pid`. DANGEROUS always.
- `OP_PROC_WAIT`: Wait for PID. Args: `pid`, `timeout_ms`. Returns: exit_code.
- `OP_PROC_KILL`: Send SIGKILL. Args: `pid`. DANGEROUS always.
- `OP_PROC_SIGNAL`: Send signal. Args: `pid`, `signal` (int). DANGEROUS if signal != SIGTERM.

### Utility (0x70-0x75)

- `OP_HASH_SHA256`: SHA-256 hash. Args: `data`. Returns: `hash` (hex string).
- `OP_HASH_MD5`: MD-5 hash. Args: `data`. Returns: `hash` (hex string).
- `OP_COMPRESS_GZIP`: Gzip compress. Args: `data`, `level` (1-9).
- `OP_DECOMPRESS_GZIP`: Gzip decompress. Args: `data`.
- `OP_ENCRYPT_AES`: AES-256-GCM encrypt. Args: `data`, `key_id`.
- `OP_DECRYPT_AES`: AES-256-GCM decrypt. Args: `data`, `key_id`.

### System (0x80-0x89)

- `OP_SYS_EXEC`: Execute shell command. Args: `cmd`. DANGEROUS always.
- `OP_SYS_ENV_GET`: Get env var. Args: `name`. Returns: `value`.
- `OP_SYS_ENV_SET`: Set env var. Args: `name`, `value`. DANGEROUS always.
- `OP_SYS_FILE_EXISTS`: Check existence. Args: `path`. Returns: `exists` (bool).
- `OP_SYS_DIR_CREATE`: Create directory. Args: `path`, `mode`.
- `OP_SYS_DIR_REMOVE`: Remove directory. Args: `path`, `recursive` (bool). DANGEROUS if recursive.
- `OP_SYS_FILE_COPY`: Copy file. Args: `src`, `dst`.
- `OP_SYS_FILE_MOVE`: Move file. Args: `src`, `dst`.
- `OP_SYS_FILE_DELETE`: Delete file. Args: `path`. DANGEROUS always.
- `OP_SYS_CHMOD`: Change mode. Args: `path`, `mode`. DANGEROUS if mode makes file world-writable.

### Session / Orchestrator (0x90-0x9A)

These are meta-operations that do not leave the process boundary. They operate on in-memory session state.

- `OP_SESS_BUDGET_CHECK`: Query remaining budget. No args. Returns: `tokens_remaining`, `time_remaining_ms`.
- `OP_SESS_CONTEXT_APPEND`: Append to session context. Args: `data`, `compression` (bool).
- `OP_SESS_DENIAL_RECORD`: Record a denial. Args: `reason`, `level`. Triggers circuit break check.
- `OP_SESS_SNAPSHOT`: Snapshot session state. Args: `path`. Writes compressed snapshot to path.
- `OP_SESS_COMPRESS`: Compress data. Args: `data`. Returns: compressed.
- `OP_ORCH_CLASSIFY`: Classify intent. Args: `intent_text`. Returns: `domain`, `safety_level`, `scenario_match`.
- `OP_ORCH_PLAN`: Build plan. Args: `domain`, `scenario`, `args`. Returns: `opchain` (serialized OpPacket array).
- `OP_ORCH_VALIDATE`: Validate chain. Args: `opchain`. Returns: `valid` (bool), `conflicts`, `budget`, `permissions`.
- `OP_ORCH_EXEC`: Execute validated chain. Args: `opchain`, `context_id`. Returns: `results`.
- `OP_ORCH_VERIFY`: Verify critical result. Args: `result`, `verifier_count` (default 2). Returns: `votes`.
- `OP_ORCH_RESPOND`: Assemble response. Args: `results`, `metrics`. Returns: structured response.

---

## Error Codes for Opcode Operations

```c
#define ERR_OK              0
#define ERR_INVALID_OPCODE  -1
#define ERR_NOT_REGISTERED  -2
#define ERR_NO_EXECUTOR     -3
#define ERR_VALIDATION_FAIL -4
#define ERR_CHAIN_REJECTED  -5
#define ERR_BUDGET_EXCEEDED -6
#define ERR_PERMISSION_DENY -7
#define ERR_CONFLICT        -8
#define ERR_TIMEOUT         -9
#define ERR_NETWORK         -10
#define ERR_DISK_FULL       -11
#define ERR_OOM             -12
#define ERR_ROLLBACK_FAIL   -13
#define ERR_ATOMIC_BREAK    -14
#define ERR_CIRCUIT_BREAK   -15
```

---

## String Names

Every opcode maps to a canonical string name used in logging, mesh slots, and JSON serialization.

```
OP_NOP              -> "NOP"
OP_MMAP_ALLOC       -> "MMAP_ALLOC"
OP_MMAP_FREE        -> "MMAP_FREE"
OP_MMAP_READ        -> "MMAP_READ"
OP_MMAP_WRITE       -> "MMAP_WRITE"
OP_MMAP_SYNC        -> "MMAP_SYNC"
OP_IO_READ          -> "IO_READ"
OP_IO_WRITE         -> "IO_WRITE"
OP_IO_OPEN          -> "IO_OPEN"
OP_IO_CLOSE         -> "IO_CLOSE"
OP_IO_SEEK          -> "IO_SEEK"
OP_GIT_INIT         -> "GIT_INIT"
OP_GIT_CLONE        -> "GIT_CLONE"
OP_GIT_FETCH        -> "GIT_FETCH"
OP_GIT_STATUS       -> "GIT_STATUS"
OP_GIT_DIFF         -> "GIT_DIFF"
OP_GIT_ADD          -> "GIT_ADD"
OP_GIT_COMMIT       -> "GIT_COMMIT"
OP_GIT_PUSH         -> "GIT_PUSH"
OP_GIT_CHECKOUT     -> "GIT_CHECKOUT"
OP_GIT_BRANCH       -> "GIT_BRANCH"
OP_GIT_MERGE        -> "GIT_MERGE"
OP_GIT_REBASE       -> "GIT_REBASE"
OP_GIT_TAG          -> "GIT_TAG"
OP_GIT_RESET        -> "GIT_RESET"
OP_BUILD_COMPILE    -> "BUILD_COMPILE"
OP_BUILD_LINK       -> "BUILD_LINK"
OP_BUILD_TEST       -> "BUILD_TEST"
OP_BUILD_DEPLOY     -> "BUILD_DEPLOY"
OP_BUILD_CLEAN      -> "BUILD_CLEAN"
OP_NET_HTTP_GET     -> "NET_HTTP_GET"
OP_NET_HTTP_POST    -> "NET_HTTP_POST"
OP_NET_TCP_CONNECT  -> "NET_TCP_CONNECT"
OP_NET_TCP_SEND     -> "NET_TCP_SEND"
OP_NET_TCP_RECV     -> "NET_TCP_RECV"
OP_NET_TCP_CLOSE    -> "NET_TCP_CLOSE"
OP_NET_WEBSOCKET    -> "NET_WEBSOCKET"
OP_PROC_SPAWN       -> "PROC_SPAWN"
OP_PROC_WAIT        -> "PROC_WAIT"
OP_PROC_KILL        -> "PROC_KILL"
OP_PROC_SIGNAL      -> "PROC_SIGNAL"
OP_HASH_SHA256      -> "HASH_SHA256"
OP_HASH_MD5         -> "HASH_MD5"
OP_COMPRESS_GZIP    -> "COMPRESS_GZIP"
OP_DECOMPRESS_GZIP  -> "DECOMPRESS_GZIP"
OP_ENCRYPT_AES      -> "ENCRYPT_AES"
OP_DECRYPT_AES      -> "DECRYPT_AES"
OP_SYS_EXEC         -> "SYS_EXEC"
OP_SYS_ENV_GET      -> "SYS_ENV_GET"
OP_SYS_ENV_SET      -> "SYS_ENV_SET"
OP_SYS_FILE_EXISTS  -> "SYS_FILE_EXISTS"
OP_SYS_DIR_CREATE   -> "SYS_DIR_CREATE"
OP_SYS_DIR_REMOVE   -> "SYS_DIR_REMOVE"
OP_SYS_FILE_COPY    -> "SYS_FILE_COPY"
OP_SYS_FILE_MOVE    -> "SYS_FILE_MOVE"
OP_SYS_FILE_DELETE  -> "SYS_FILE_DELETE"
OP_SYS_CHMOD        -> "SYS_CHMOD"
OP_SESS_BUDGET_CHECK    -> "SESS_BUDGET_CHECK"
OP_SESS_CONTEXT_APPEND  -> "SESS_CONTEXT_APPEND"
OP_SESS_DENIAL_RECORD   -> "SESS_DENIAL_RECORD"
OP_SESS_SNAPSHOT        -> "SESS_SNAPSHOT"
OP_SESS_COMPRESS        -> "SESS_COMPRESS"
OP_ORCH_CLASSIFY        -> "ORCH_CLASSIFY"
OP_ORCH_PLAN            -> "ORCH_PLAN"
OP_ORCH_VALIDATE        -> "ORCH_VALIDATE"
OP_ORCH_EXEC            -> "ORCH_EXEC"
OP_ORCH_VERIFY          -> "ORCH_VERIFY"
OP_ORCH_RESPOND         -> "ORCH_RESPOND"
OP_RESEARCH_HYPOTHESIS_CREATE      -> "RESEARCH_HYPOTHESIS_CREATE"
OP_RESEARCH_HYPOTHESIS_LOAD        -> "RESEARCH_HYPOTHESIS_LOAD"
OP_RESEARCH_HYPOTHESIS_INFERENCE   -> "RESEARCH_HYPOTHESIS_INFERENCE"
OP_RESEARCH_EXPERIMENT_RUN         -> "RESEARCH_EXPERIMENT_RUN"
OP_RESEARCH_RESULT_STORE           -> "RESEARCH_RESULT_STORE"
OP_RESEARCH_STATISTICAL_TEST       -> "RESEARCH_STATISTICAL_TEST"
OP_RESEARCH_LITERATURE_FETCH       -> "RESEARCH_LITERATURE_FETCH"
OP_RESEARCH_LITERATURE_PARSE       -> "RESEARCH_LITERATURE_PARSE"
OP_RESEARCH_LITERATURE_INDEX       -> "RESEARCH_LITERATURE_INDEX"
OP_RESEARCH_CITATION_LINK          -> "RESEARCH_CITATION_LINK"
OP_RESEARCH_LITERATURE_EMBED       -> "RESEARCH_LITERATURE_EMBED"
OP_RESEARCH_PROGRESS_STORE         -> "RESEARCH_PROGRESS_STORE"
OP_RESEARCH_CONTEXT_SUMMARIZE      -> "RESEARCH_CONTEXT_SUMMARIZE"
OP_SELF_CHECKPOINT_CREATE          -> "SELF_CHECKPOINT_CREATE"
OP_SELF_CHECKPOINT_RESTORE         -> "SELF_CHECKPOINT_RESTORE"
OP_SELF_BUDGET_REALLOCATE          -> "SELF_BUDGET_REALLOCATE"
OP_SELF_STRATEGY_PIVOT             -> "SELF_STRATEGY_PIVOT"
OP_SELF_PROGRESS_ASSESS            -> "SELF_PROGRESS_ASSESS"
OP_SELF_CONTEXT_SUMMARIZE          -> "SELF_CONTEXT_SUMMARIZE"
```

Reverse mapping (string → opcode) uses exact match, case-sensitive. Unknown string → `OP_NOP` with error.
