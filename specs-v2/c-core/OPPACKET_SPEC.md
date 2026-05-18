# C-Core OpPacket Specification

Exact memory layout, field semantics, and lifecycle of the fundamental execution unit.

Every operation in Mimic is an OpPacket. A chain of operations is an array of OpPackets. The entire system compiles down to manipulating these structures.

No implementation hints. Exact sizes and alignments.

---

## Structure Definition

```c
typedef struct {
    char key[OP_NAME_LEN];      // 32 bytes, null-terminated argument name
    char value[256];           // 256 bytes, null-terminated string value
    size_t value_len;          // sizeof(size_t), actual string length (excluding null)
    uint8_t type;              // 0=null, 1=int, 2=float, 3=string, 4=blob
} OpArg;
```

`sizeof(OpArg)` = 32 + 256 + sizeof(size_t) + 1, padded to alignment. On 64-bit: 32 + 256 + 8 + 1 = 297 → padded to 304 bytes (align to 8).

```c
typedef struct {
    // Identity
    uint32_t id;               // Unique packet ID within chain. Starts at 1.
    OpCode opcode;             // Operation code. uint32_t size for alignment.
    uint32_t flags;            // Operation-specific flags (override defaults)
    
    // Arguments
    OpArg args[MAX_ARGS];      // Up to 16 arguments
    uint8_t arg_count;         // Actual number of populated arguments
    uint8_t reserved[3];     // Padding to 4-byte boundary
    
    // Resources
    int32_t fd_in;             // Input file descriptor (-1 if unused)
    int32_t fd_out;            // Output file descriptor (-1 if unused)
    void* buffer;              // Pointer to blob data (owned by caller)
    size_t buffer_size;        // Size of blob data in bytes
    
    // Metadata
    uint64_t timestamp_ns;     // Submission timestamp (CLOCK_MONOTONIC)
    uint32_t timeout_ms;       // Timeout for this operation. 0 = use default.
    uint32_t retry_count;      // Remaining retries. 0 = no retry.
    
    // Result (populated by executor)
    int32_t result_code;       // ERR_OK (0) or negative error code
    size_t bytes_processed;    // Bytes read/written/moved by this op
    uint64_t latency_ns;       // Execution latency measured by executor
    
    // Chain linkage
    uint32_t prev_op_id;       // ID of previous packet in chain (0 if first)
    uint32_t next_op_id;       // ID of next packet in chain (0 if last)
    uint32_t chain_id;         // Identifier for the chain this packet belongs to
    
    // Padding to 64-byte cache line boundary
    uint8_t padding[4];        // Ensure total size is multiple of 64
} OpPacket;
```

### Size Calculation

On 64-bit Linux:
- OpArg: 304 bytes
- args[16]: 4864 bytes
- Rest of fields: 4 + 4 + 4 + 1 + 3 + 4 + 4 + 8 + 8 + 4 + 4 + 4 + 8 + 8 + 4 + 4 + 4 + 4 = 80 bytes
- Total: 4864 + 80 = 4944 bytes → padded to 4992 (nearest 64)

`sizeof(OpPacket)` = 4992 bytes.

An array of 1024 packets = ~4.9 MB. This fits in L3 cache on modern CPUs.

---

## Field Semantics

### id

- Unique within a chain. Monotonically increasing starting from 1.
- Used for rollback: `rollback_to_id = failed_id - 1`.
- Used for logging: every log line references `chain_id:packet_id`.

### opcode

- Must be < OP_MAX (0xFF).
- Must be registered via `ops_register()` before execution.
- Unregistered opcode → `ERR_NOT_REGISTERED`.

### flags

- Override the default flags for this opcode.
- If `OP_FLAG_DANGEROUS` is set in flags, it cannot be cleared by override.
- If `OP_FLAG_ATOMIC` is set, chain failure triggers rollback to packet before this one.
- If `OP_FLAG_READONLY` is set, the operation MUST NOT modify any state. Executor enforces this via pre/post state hash comparison.

### args / arg_count

- args[0] through args[arg_count-1] are valid.
- args[arg_count] through args[15] MUST be zeroed (memset to 0).
- arg_count > MAX_ARGS → validation failure `ERR_INVALID_ARG_COUNT`.
- Each arg has a key (parameter name) and a string value. Even numeric values are stored as strings in `value` and parsed by the executor.
- type field indicates how to interpret value: 0=null (arg present but no value), 1=int (parse with atoi), 2=float (parse with atof), 3=string (use as-is), 4=blob (value contains base64 or hex, decode to buffer).

### fd_in / fd_out

- File descriptors managed by ExecContext.
- fd_in = -1 means no input FD.
- fd_out = -1 means no output FD.
- Any FD used MUST be tracked in ExecContext.open_fds. Executor validates FD is known before use.
- Closing an FD removes it from tracking. Using closed FD = `ERR_INVALID_FD`.

### buffer / buffer_size

- buffer points to caller-allocated memory.
- buffer_size = 0 means no buffer.
- If buffer != NULL and buffer_size > 0, the executor may read from or write to this memory.
- buffer is NOT owned by OpPacket. Caller must free after chain completion.
- For large data (files, HTTP bodies), use FD-based I/O instead of buffer.

### timestamp_ns

- Set by `ops_packet_init()` using `ops_get_time_ns()` (CLOCK_MONOTONIC).
- Used for latency calculation: `latency_ns = end_time - timestamp_ns`.
- Also used for timeout: `deadline = timestamp_ns + timeout_ms * 1000000`.

### timeout_ms

- 0 means use system default (30 seconds).
- Network ops should have explicit timeout (5-30s).
- Git ops should have explicit timeout (60-300s).
- Build ops should have explicit timeout (300-3600s).
- If `ops_get_time_ns() > deadline` during execution → `ERR_TIMEOUT`.

### retry_count

- Number of times to retry on transient failure.
- Transient failures: ERR_NETWORK, ERR_TIMEOUT, ERR_DISK_FULL (may resolve).
- Non-transient failures: ERR_INVALID_OPCODE, ERR_PERMISSION_DENY, ERR_CONFLICT, ERR_CIRCUIT_BREAK (never retry).
- On retry: packet is re-executed with `retry_count--`. Chain position unchanged.
- Max total retries per chain: 10. Exceeding → `ERR_CHAIN_REJECTED`.

### result_code

- Set by executor after execution.
- 0 (ERR_OK) = success.
- Negative = error. See OPCODE_SPEC.md for full error code list.
- If result_code != 0 and retry_count > 0 → retry logic applies.
- If result_code != 0 and retry_count == 0 and OP_FLAG_ATOMIC → rollback.
- If result_code != 0 and retry_count == 0 and no atomic flag → chain stops, partial results returned.

### bytes_processed

- Number of bytes transferred by this operation.
- I/O ops: bytes read or written.
- Network ops: bytes sent or received.
- Memory ops: bytes allocated, freed, or copied.
- Git ops: 0 (semantic operations have no byte count).
- Build ops: bytes of output (object code, test output).
- Used for energy accounting: `total_bytes += bytes_processed`.

### latency_ns

- Executor-measured wall time for this operation only.
- Does NOT include validation time, planning time, or inter-op overhead.
- Used for: metrics reporting, energy calculation, timeout enforcement.
- Sum of all packet latencies = chain execution time.

### prev_op_id / next_op_id / chain_id

- Set by `ops_execute_chain()` before execution.
- prev_op_id = 0 for first packet in chain.
- next_op_id = 0 for last packet in chain.
- chain_id is a session-unique identifier. Used for logging and resource grouping.
- During rollback: traverse backwards using prev_op_id, execute inverse for each reversible op.

---

## Lifecycle

### Creation

```c
OpPacket pkt;
ops_packet_init(&pkt, OP_GIT_STATUS);
ops_packet_set_string(&pkt, "path", "/workspace/repo");
```

`ops_packet_init()` MUST:
1. memset entire struct to 0
2. Set opcode
3. Set id = next_id++ (global atomic counter)
4. Set timestamp_ns = ops_get_time_ns()
5. Set timeout_ms = 0, retry_count = 0
6. Set fd_in = fd_out = -1
7. Set chain_id = 0 (set by chain builder)

### Validation

```c
ValidationResult vr = ops_validate_chain(packets, count);
```

Validation checks (before any execution):
1. Every opcode is registered.
2. Every opcode is < OP_MAX.
3. arg_count ≤ MAX_ARGS.
4. No duplicate keys within a packet's args.
5. No conflict between any pair of opcodes (conflict_matrix).
6. Sum of energy costs ≤ session budget.
7. No dangerous op without explicit permission.
8. No readonly op that modifies state (pre-check via state snapshot).
9. FD references are valid (if fd_in/fd_out != -1, must be in tracking).
10. buffer != NULL implies buffer_size > 0.

### Execution

```c
int result = ops_execute_chain(packets, count, &ctx);
```

Execution order:
1. Set prev/next links.
2. Set chain_id in context.
3. Record initial state snapshot (for rollback).
4. For each packet i = 0 to count-1:
   a. Check deadline (timestamp_ns + timeout_ms).
   b. Lookup executor for opcode.
   c. Call executor(&packets[i]).
   d. Record latency_ns.
   e. If result != 0 and retry_count > 0: retry (goto 4a with retry_count--).
   f. If result != 0 and OP_FLAG_ATOMIC: trigger rollback.
   g. If result != 0 and no atomic: return partial (packets[0..i] executed, i+1..count-1 not).
   h. Track resources: if op opens FD, add to ctx.open_fds. If op mmap's, add to ctx.mmap_regions.
5. On chain completion: close all tracked FDs, sync all tracked mmaps, free temporary resources.

### Rollback

Rollback triggered when:
- Any packet in chain has OP_FLAG_ATOMIC and any packet fails.
- Explicit rollback request from orchestrator.

Rollback procedure:
1. Identify last successfully executed packet index: `last_success = failure_index - 1`.
2. For i = last_success down to 0:
   a. If packets[i].flags & OP_FLAG_REVERSIBLE:
      - Look up inverse opcode (from OpCodeDef).
      - Execute inverse with recorded state snapshot.
   b. If not reversible:
      - Log irreversible state change.
      - Attempt best-effort cleanup (close FDs, free mmaps).
3. Restore resource tracking to pre-chain state.
4. Return `ERR_ROLLBACK_FAIL` if any inverse execution failed, else `ERR_OK`.

### Destruction

OpPackets are stack-allocated or caller-allocated. No free() needed for struct itself.
Blobs referenced by `buffer` MUST be freed by caller after chain completion.
Args with blob references (type=4) may contain decoded data that also needs freeing.

---

## Thread Safety

OpPacket is NOT thread-safe for mutation. Read-only access from multiple threads is safe.
During chain execution, only the executor thread may modify result fields.
Retry logic re-executes the SAME packet struct (modifies result_code, latency_ns, retry_count).

---

## Serialization

OpPacket serializes to JSON for logging and mesh storage:

```json
{
  "id": 1,
  "opcode": 51,
  "opcode_name": "GIT_STATUS",
  "flags": 3,
  "flags_names": ["SAFE", "READONLY"],
  "args": [
    {"key": "path", "value": "/workspace/repo", "type": 3}
  ],
  "arg_count": 1,
  "fd_in": -1,
  "fd_out": -1,
  "buffer_size": 0,
  "timestamp_ns": 1715923456789012345,
  "timeout_ms": 30000,
  "retry_count": 0,
  "result_code": 0,
  "bytes_processed": 0,
  "latency_ns": 1450000,
  "prev_op_id": 0,
  "next_op_id": 2,
  "chain_id": 42
}
```

Deserialization must validate all constraints (opcode range, arg_count, etc.) before creating in-memory struct.
