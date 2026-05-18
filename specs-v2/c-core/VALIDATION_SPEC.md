# C-Core Validation Specification

Exact validation logic, error codes, and pass/fail criteria for `ops_validate_chain()`.

Validation is the gate between PLAN and EXEC. No validated chain has ever caused an unrecoverable failure. Every failure mode is caught here.

---

## ValidationResult Structure

```c
typedef struct {
    bool is_valid;              // true = chain may execute
    uint32_t error_code;        // 0 = no error, else see below
    char error_msg[256];        // Human-readable error description
    
    uint32_t invalid_op_index;  // Index of failing packet (0-based)
    uint32_t conflict_op_pair[2]; // Indices of conflicting pair (if error_code=8)
    float total_energy;         // Sum of cost_tokens for all packets
    float estimated_latency_us; // Sum of cost_time_us for all packets
} ValidationResult;
```

---

## Validation Steps (in order)

Validation is a sequential pipeline. Step N only runs if steps 1..N-1 passed.

### Step 1: State Check

**What:** Core initialized? Count > 0? Packets pointer valid?

**Pass:** `g_initialized == true`, `count > 0`, `packets != NULL`.

**Fail:**
- `!g_initialized` → error_code = 1, error_msg = "Core not initialized"
- `count == 0` → error_code = 1, error_msg = "Empty chain"
- `packets == NULL` → error_code = 1, error_msg = "Null packet pointer"

**Why first:** No point checking opcodes if there's no core.

---

### Step 2: Opcode Validity

**What:** Every packet has a valid opcode.

**Pass:** For all i in [0, count): `packets[i].opcode > OP_NOP && packets[i].opcode < OP_MAX`.

**Fail:**
- `opcode == OP_NOP` in non-first position → error_code = 2, invalid_op_index = i, error_msg = "NOP at index N"
- `opcode >= OP_MAX` → error_code = 2, invalid_op_index = i, error_msg = "Invalid opcode N at index M"

**Why:** Prevents unregistered opcodes from reaching execution.

---

### Step 3: Registration Check

**What:** Every opcode has a registered OpCodeDef.

**Pass:** For all i: `g_op_registry[packets[i].opcode].execute != NULL`.

**Fail:**
- `g_op_registry[opcode].execute == NULL` → error_code = 3, invalid_op_index = i, error_msg = "Opcode N not registered"

**Why:** Prevents calls to undefined executors.

---

### Step 4: Argument Validity

**What:** Arguments are well-formed.

**Pass:**
- `arg_count ≤ MAX_ARGS` (16)
- For all j in [0, arg_count): `args[j].key[0] != '\0'` (non-empty key)
- No duplicate keys within a packet
- `type` in {0, 1, 2, 3, 4}
- If `type == 4` (blob): `buffer != NULL && buffer_size > 0`

**Fail:**
- `arg_count > MAX_ARGS` → error_code = 4, invalid_op_index = i
- Empty key → error_code = 4, invalid_op_index = i
- Duplicate key → error_code = 4, invalid_op_index = i
- Invalid type → error_code = 4, invalid_op_index = i
- Blob without buffer → error_code = 4, invalid_op_index = i

**Why:** Prevents malformed arguments from crashing executors.

---

### Step 5: FD Validity

**What:** File descriptor references are consistent.

**Pass:**
- If `fd_in != -1`: must be ≥ 0
- If `fd_out != -1`: must be ≥ 0
- `fd_in` and `fd_out` must not be the same value (would create a loop)
- Opening ops (OP_IO_OPEN, OP_NET_TCP_CONNECT) do not reference fd_in (they create new FDs)

**Fail:**
- Negative FD (other than -1) → error_code = 5, invalid_op_index = i
- fd_in == fd_out (and both != -1) → error_code = 5, invalid_op_index = i
- OP_IO_OPEN with fd_in != -1 → error_code = 5, invalid_op_index = i

**Note:** FD existence in tracking cannot be checked at validation time (FDs are opened during execution). Validation only checks structural validity.

---

### Step 6: Pairwise Conflict Check

**What:** No pair of operations in the chain conflicts.

**Algorithm:**
```
for i = 0 to count-1:
  for j = i+1 to count-1:
    if conflict_matrix[packets[i].opcode][packets[j].opcode] > ConflictNone:
      FAIL
```

**Pass:** All pairs have conflict level = ConflictNone.

**Fail:**
- Any pair has conflict → error_code = 8, conflict_op_pair = {i, j}
- error_msg = "Conflict between op N (NAME) and op M (NAME): LEVEL"

**Conflict Levels:**
- ConflictNone (0): No conflict. Continue.
- ConflictLow (1): Warning. Not a validation failure but logged.
- ConflictMedium (2): Requires attention. Validation FAIL.
- ConflictHigh (3): Critical. Validation FAIL.
- ConflictFatal (4): Unresolvable. Validation FAIL.

**Symmetry:** `conflict_matrix[i][j] == conflict_matrix[j][i]`. Only upper triangle checked.

**Why:** Prevents operations that would corrupt state if run in the same chain.

---

### Step 7: Energy Budget Check

**What:** Total estimated energy does not exceed remaining budget.

**Calculation:**
```
total_energy = Σ g_energy_costs[packets[i].opcode][0]  // cost_tokens
total_latency_us = Σ g_energy_costs[packets[i].opcode][1]  // cost_time_us
```

**Pass:** `total_energy ≤ ctx->session_budget_tokens` AND `total_latency_us ≤ ctx->session_budget_time_ms * 1000`.

**Fail:**
- `total_energy > budget_tokens` → error_code = 6, error_msg = "Energy N > budget M"
- `total_latency_us > budget_time` → error_code = 6, error_msg = "Latency N us > budget M us"

**Why:** Prevents chains that would exhaust session resources.

---

### Step 8: Permission Check

**What:** No dangerous operation without explicit permission.

**Algorithm:**
```
for i = 0 to count-1:
  def = &g_op_registry[packets[i].opcode]
  
  // Check DANGEROUS flag
  if (def->flags & OP_FLAG_DANGEROUS) || (packets[i].flags & OP_FLAG_DANGEROUS):
    if (!session_has_explicit_allow(packets[i].opcode, packets[i].chain_id)):
      FAIL
  
  // Check safety level
  if (def->safety_level == 0):  // CRITICAL
    if (!session_has_2vote_verify(packets[i].chain_id)):
      FAIL
  
  // Check circuit break
  if (ctx->circuit_broken):
    FAIL
```

**Pass:** All dangerous ops have explicit allow. All critical ops have 2-vote. Circuit not broken.

**Fail:**
- Dangerous without allow → error_code = 7, invalid_op_index = i, error_msg = "Permission denied: dangerous op N"
- Critical without 2-vote → error_code = 7, invalid_op_index = i, error_msg = "Permission denied: critical op N requires 2-vote"
- Circuit broken → error_code = 15, error_msg = "Circuit broken: manual reset required"

**Why:** Enforces safety policy before any execution.

---

### Step 9: Readonly Integrity Check

**What:** Operations marked READONLY do not have write-side effects.

**Check:** For all packets with `OP_FLAG_READONLY`:
- Opcode must be from whitelist of known-readonly ops: OP_GIT_STATUS, OP_GIT_DIFF, OP_IO_READ, OP_IO_SEEK, OP_HASH_*, OP_COMPRESS_*, OP_DECOMPRESS_*, OP_SYS_FILE_EXISTS, OP_SYS_ENV_GET, OP_SESS_BUDGET_CHECK, OP_ORCH_CLASSIFY, OP_ORCH_VALIDATE.
- If opcode is not in whitelist but has READONLY flag → FAIL.

**Fail:**
- Non-whitelist op with READONLY → error_code = 4, invalid_op_index = i, error_msg = "Invalid READONLY flag on op N"

**Why:** Prevents false confidence. An op with READONLY must genuinely be readonly.

---

### Step 10: Atomic Chain Consistency

**What:** If chain contains atomic ops, the entire chain must be rollbackable.

**Check:**
- Find all packets with `OP_FLAG_ATOMIC`.
- For each such packet at index i: verify packets[0..i-1] are all reversible OR there exists a snapshot point before i.
- A snapshot point is any packet with `OP_FLAG_REVERSIBLE` that records state.

**Simplified rule:** If ANY packet has OP_FLAG_ATOMIC, then ALL preceding packets must be reversible, OR a snapshot must exist before the first atomic op.

**Pass:** Atomic flag consistency holds.

**Fail:**
- Atomic op at i with non-reversible op before i and no snapshot → error_code = 14, invalid_op_index = i, error_msg = "Atomic break: cannot rollback to before index N"

**Why:** Ensures that atomic rollback is actually possible.

---

### Step 11: Buffer Consistency

**What:** Buffer pointers and sizes are consistent.

**Pass:**
- If `buffer == NULL`: `buffer_size` must be 0.
- If `buffer != NULL`: `buffer_size` must be > 0.
- `buffer_size` must be ≤ 1GB (hard limit to prevent OOM).

**Fail:**
- NULL buffer with size > 0 → error_code = 4
- Non-NULL buffer with size == 0 → error_code = 4
- buffer_size > 1GB → error_code = 12 (OOM)

---

## Error Code Summary

| Code | Name | Step | Description |
|---|---|---|---|
| 0 | ERR_OK | — | Validation passed |
| 1 | ERR_INVALID_STATE | 1 | Core not initialized or empty chain |
| 2 | ERR_INVALID_OPCODE | 2 | Opcode out of range |
| 3 | ERR_NOT_REGISTERED | 3 | Opcode not registered |
| 4 | ERR_INVALID_ARG | 4, 9, 11 | Argument malformed or invalid flag combination |
| 5 | ERR_INVALID_FD | 5 | File descriptor reference invalid |
| 6 | ERR_BUDGET_EXCEEDED | 7 | Energy or latency budget exceeded |
| 7 | ERR_PERMISSION_DENY | 8 | Permission denied or circuit broken |
| 8 | ERR_CONFLICT | 6 | Pairwise conflict detected |
| 12 | ERR_OOM | 11 | Buffer too large |
| 14 | ERR_ATOMIC_BREAK | 10 | Atomic op without rollback path |
| 15 | ERR_CIRCUIT_BREAK | 8 | Circuit breaker active |

---

## ValidationResult Semantics

- `is_valid == true`: Chain MAY execute. All checks passed.
- `is_valid == false`: Chain MUST NOT execute. Exact reason in error_code and error_msg.
- `invalid_op_index`: 0-based index of first failing packet. Only valid if is_valid == false.
- `conflict_op_pair`: Two indices {i, j} where i < j and conflict exists. Only valid if error_code == 8.
- `total_energy`: Sum of token costs. Always computed, even on failure (shows how much would have been needed).
- `estimated_latency_us`: Sum of time costs. Always computed.

---

## Performance

Validation is O(count²) due to pairwise conflict check.
For count = 1024: ~1M pairs. With uint8 matrix lookup: ~1M operations = sub-millisecond.
All other steps are O(count).
Validation must complete in < 1ms for chains up to 100 ops, < 10ms for chains up to 1000 ops.

---

## Thread Safety

`ops_validate_chain()` reads global state (`g_op_registry`, `g_conflict_matrix`, `g_energy_costs`).
These globals are immutable after initialization. Safe for concurrent read-only access.
Session state (`ctx->budget`, `circuit_broken`) is read but not modified. Safe for concurrent validation of different contexts.
