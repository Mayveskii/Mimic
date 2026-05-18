# C-Core Energy Cost Specification

Exact energy accounting formulas, cost assignment rules, and optimization principles.

Energy is not an abstract metric. It is the product of token cost and latency, summed over a chain. Minimizing energy means minimizing the time the model spends waiting for operations to complete, weighted by their token consumption.

---

## Formula

The Principle of Least Action applied to operation chains:

```
S(chain) = Σ (cost_tokens(op_i) × latency_ns(op_i))
```

Where:
- `cost_tokens(op_i)`: Estimated LLM token consumption to express this operation in natural language.
- `latency_ns(op_i)`: Wall-clock execution time of the operation.

This is NOT "maximize energy". It is "minimize the integral of cost over time". Lower S = more efficient chain.

### Token Cost Estimation

| Operation Type | cost_tokens | Rationale |
|---|---|---|
| No-op (NOP) | 0.0 | No model interaction |
| Read-only (status, diff, exists) | 1.0 | Single sentence: "check status" |
| Simple mutation (add, write, mkdir) | 2.0 | Sentence + parameters |
| Complex mutation (commit, merge) | 4.0 | Multiple sentences + validation |
| Network (fetch, push, HTTP) | 5.0 | High latency, model waits |
| Build (compile, test) | 6.0 | Long running, model blocked |
| Deploy | 8.0 | Critical, requires verification |
| Dangerous (kill, exec, rebase) | 10.0 | Maximum oversight required |
| Orchestrator meta-ops | 0.5-1.0 | Internal, no model waiting |

### Latency Estimation

| Operation Type | cost_time_us (estimated) | Typical Actual |
|---|---|---|
| NOP | 0.01 | < 1 us |
| Memory alloc/free | 0.1-1.0 | 1-10 us |
| File exists | 1.0 | 5-50 us |
| File read (small) | 2.0 | 10-100 us |
| File write (small) | 3.0 | 20-200 us |
| Git status | 5.0 | 5-50 ms |
| Git diff | 5.0 | 5-50 ms |
| Git commit | 10.0 | 10-100 ms |
| Git push | 50.0 | 500ms - 5s |
| HTTP GET | 50.0 | 100ms - 2s |
| TCP connect | 30.0 | 10ms - 1s |
| Compile (small) | 100.0 | 1-10s |
| Test suite | 500.0 | 5-60s |
| Deploy | 1000.0 | 10s - 5min |

Actual latencies are measured during execution and replace estimates. Estimates are used for validation (budget check before execution).

---

## Energy Cost Matrix

```c
// C core: static float g_energy_costs[OP_MAX][3];
// Index 0 = cost_tokens, Index 1 = cost_time_us, Index 2 = cost_memory_bytes

typedef struct {
    float cost_tokens;
    float cost_time_us;
    float cost_memory_bytes;
} EnergyCost;
```

Size: `256 × 3 × 4 bytes` = 3,072 bytes. Fits in L1 cache.

### Default Cost Assignment

```c
void ops_init_energy_matrix(void) {
    // Default: zero cost for unregistered ops
    memset(g_energy_costs, 0, sizeof(g_energy_costs));
    
    // Memory ops
    g_energy_costs[OP_MMAP_ALLOC][0] = 1.0f;   g_energy_costs[OP_MMAP_ALLOC][1] = 1.0f;   g_energy_costs[OP_MMAP_ALLOC][2] = 4096.0f;
    g_energy_costs[OP_MMAP_FREE][0] = 0.5f;     g_energy_costs[OP_MMAP_FREE][1] = 0.5f;     g_energy_costs[OP_MMAP_FREE][2] = 0.0f;
    g_energy_costs[OP_MMAP_READ][0] = 1.0f;     g_energy_costs[OP_MMAP_READ][1] = 1.0f;     g_energy_costs[OP_MMAP_READ][2] = 0.0f;
    g_energy_costs[OP_MMAP_WRITE][0] = 1.5f;    g_energy_costs[OP_MMAP_WRITE][1] = 1.5f;    g_energy_costs[OP_MMAP_WRITE][2] = 0.0f;
    g_energy_costs[OP_MMAP_SYNC][0] = 1.0f;     g_energy_costs[OP_MMAP_SYNC][1] = 5.0f;     g_energy_costs[OP_MMAP_SYNC][2] = 0.0f;
    
    // I/O ops
    g_energy_costs[OP_IO_OPEN][0] = 1.0f;       g_energy_costs[OP_IO_OPEN][1] = 2.0f;       g_energy_costs[OP_IO_OPEN][2] = 0.0f;
    g_energy_costs[OP_IO_CLOSE][0] = 0.5f;      g_energy_costs[OP_IO_CLOSE][1] = 1.0f;      g_energy_costs[OP_IO_CLOSE][2] = 0.0f;
    g_energy_costs[OP_IO_READ][0] = 1.0f;        g_energy_costs[OP_IO_READ][1] = 2.0f;        g_energy_costs[OP_IO_READ][2] = 0.0f;
    g_energy_costs[OP_IO_WRITE][0] = 2.0f;       g_energy_costs[OP_IO_WRITE][1] = 3.0f;       g_energy_costs[OP_IO_WRITE][2] = 0.0f;
    g_energy_costs[OP_IO_SEEK][0] = 0.5f;        g_energy_costs[OP_IO_SEEK][1] = 1.0f;        g_energy_costs[OP_IO_SEEK][2] = 0.0f;
    
    // Git ops
    g_energy_costs[OP_GIT_STATUS][0] = 1.0f;     g_energy_costs[OP_GIT_STATUS][1] = 5000.0f;  // 5ms
    g_energy_costs[OP_GIT_DIFF][0] = 1.0f;       g_energy_costs[OP_GIT_DIFF][1] = 5000.0f;
    g_energy_costs[OP_GIT_ADD][0] = 2.0f;          g_energy_costs[OP_GIT_ADD][1] = 10000.0f;     // 10ms
    g_energy_costs[OP_GIT_COMMIT][0] = 4.0f;      g_energy_costs[OP_GIT_COMMIT][1] = 50000.0f;  // 50ms
    g_energy_costs[OP_GIT_PUSH][0] = 5.0f;        g_energy_costs[OP_GIT_PUSH][1] = 1000000.0f;   // 1s
    g_energy_costs[OP_GIT_FETCH][0] = 5.0f;       g_energy_costs[OP_GIT_FETCH][1] = 500000.0f;  // 500ms
    g_energy_costs[OP_GIT_CLONE][0] = 5.0f;        g_energy_costs[OP_GIT_CLONE][1] = 30000000.0f; // 30s
    g_energy_costs[OP_GIT_CHECKOUT][0] = 4.0f;     g_energy_costs[OP_GIT_CHECKOUT][1] = 20000.0f; // 20ms
    g_energy_costs[OP_GIT_MERGE][0] = 4.0f;        g_energy_costs[OP_GIT_MERGE][1] = 50000.0f;
    g_energy_costs[OP_GIT_REBASE][0] = 10.0f;      g_energy_costs[OP_GIT_REBASE][1] = 100000.0f;  // 100ms
    
    // Build ops
    g_energy_costs[OP_BUILD_COMPILE][0] = 6.0f;    g_energy_costs[OP_BUILD_COMPILE][1] = 1000000.0f;  // 1s
    g_energy_costs[OP_BUILD_LINK][0] = 3.0f;        g_energy_costs[OP_BUILD_LINK][1] = 100000.0f;      // 100ms
    g_energy_costs[OP_BUILD_TEST][0] = 6.0f;        g_energy_costs[OP_BUILD_TEST][1] = 5000000.0f;     // 5s
    g_energy_costs[OP_BUILD_DEPLOY][0] = 8.0f;       g_energy_costs[OP_BUILD_DEPLOY][1] = 30000000.0f;  // 30s
    g_energy_costs[OP_BUILD_CLEAN][0] = 2.0f;         g_energy_costs[OP_BUILD_CLEAN][1] = 50000.0f;
    
    // Network ops
    g_energy_costs[OP_NET_HTTP_GET][0] = 5.0f;      g_energy_costs[OP_NET_HTTP_GET][1] = 500000.0f;    // 500ms
    g_energy_costs[OP_NET_HTTP_POST][0] = 5.0f;      g_energy_costs[OP_NET_HTTP_POST][1] = 500000.0f;
    g_energy_costs[OP_NET_TCP_CONNECT][0] = 3.0f;    g_energy_costs[OP_NET_TCP_CONNECT][1] = 300000.0f;  // 300ms
    g_energy_costs[OP_NET_TCP_SEND][0] = 2.0f;       g_energy_costs[OP_NET_TCP_SEND][1] = 100000.0f;    // 100ms
    g_energy_costs[OP_NET_TCP_RECV][0] = 2.0f;       g_energy_costs[OP_NET_TCP_RECV][1] = 100000.0f;
    g_energy_costs[OP_NET_TCP_CLOSE][0] = 0.5f;      g_energy_costs[OP_NET_TCP_CLOSE][1] = 1000.0f;
    g_energy_costs[OP_NET_WEBSOCKET][0] = 5.0f;      g_energy_costs[OP_NET_WEBSOCKET][1] = 200000.0f;
    
    // Process ops
    g_energy_costs[OP_PROC_SPAWN][0] = 10.0f;        g_energy_costs[OP_PROC_SPAWN][1] = 100000.0f;
    g_energy_costs[OP_PROC_WAIT][0] = 1.0f;           g_energy_costs[OP_PROC_WAIT][1] = 1000000.0f;
    g_energy_costs[OP_PROC_KILL][0] = 10.0f;          g_energy_costs[OP_PROC_KILL][1] = 10000.0f;
    g_energy_costs[OP_PROC_SIGNAL][0] = 5.0f;          g_energy_costs[OP_PROC_SIGNAL][1] = 10000.0f;
    
    // System ops
    g_energy_costs[OP_SYS_EXEC][0] = 10.0f;          g_energy_costs[OP_SYS_EXEC][1] = 1000000.0f;
    g_energy_costs[OP_SYS_FILE_EXISTS][0] = 1.0f;      g_energy_costs[OP_SYS_FILE_EXISTS][1] = 1.0f;
    g_energy_costs[OP_SYS_DIR_CREATE][0] = 2.0f;       g_energy_costs[OP_SYS_DIR_CREATE][1] = 50.0f;
    g_energy_costs[OP_SYS_DIR_REMOVE][0] = 2.0f;       g_energy_costs[OP_SYS_DIR_REMOVE][1] = 100.0f;
    g_energy_costs[OP_SYS_FILE_COPY][0] = 2.0f;         g_energy_costs[OP_SYS_FILE_COPY][1] = 1000.0f;
    g_energy_costs[OP_SYS_FILE_MOVE][0] = 2.0f;         g_energy_costs[OP_SYS_FILE_MOVE][1] = 500.0f;
    g_energy_costs[OP_SYS_FILE_DELETE][0] = 2.0f;        g_energy_costs[OP_SYS_FILE_DELETE][1] = 100.0f;
    g_energy_costs[OP_SYS_ENV_GET][0] = 0.5f;            g_energy_costs[OP_SYS_ENV_GET][1] = 1.0f;
    g_energy_costs[OP_SYS_ENV_SET][0] = 1.0f;            g_energy_costs[OP_SYS_ENV_SET][1] = 1.0f;
    g_energy_costs[OP_SYS_CHMOD][0] = 1.0f;               g_energy_costs[OP_SYS_CHMOD][1] = 10.0f;
    
    // Utility ops
    g_energy_costs[OP_HASH_SHA256][0] = 1.0f;            g_energy_costs[OP_HASH_SHA256][1] = 10.0f;
    g_energy_costs[OP_HASH_MD5][0] = 1.0f;                g_energy_costs[OP_HASH_MD5][1] = 5.0f;
    g_energy_costs[OP_COMPRESS_GZIP][0] = 1.0f;           g_energy_costs[OP_COMPRESS_GZIP][1] = 100.0f;
    g_energy_costs[OP_DECOMPRESS_GZIP][0] = 1.0f;         g_energy_costs[OP_DECOMPRESS_GZIP][1] = 100.0f;
    g_energy_costs[OP_ENCRYPT_AES][0] = 1.0f;             g_energy_costs[OP_ENCRYPT_AES][1] = 50.0f;
    g_energy_costs[OP_DECRYPT_AES][0] = 1.0f;             g_energy_costs[OP_DECRYPT_AES][1] = 50.0f;
    
    // Session/Orchestrator ops
    g_energy_costs[OP_SESS_BUDGET_CHECK][0] = 0.5f;      g_energy_costs[OP_SESS_BUDGET_CHECK][1] = 1.0f;
    g_energy_costs[OP_SESS_CONTEXT_APPEND][0] = 0.5f;     g_energy_costs[OP_SESS_CONTEXT_APPEND][1] = 10.0f;
    g_energy_costs[OP_SESS_DENIAL_RECORD][0] = 0.5f;      g_energy_costs[OP_SESS_DENIAL_RECORD][1] = 1.0f;
    g_energy_costs[OP_SESS_SNAPSHOT][0] = 1.0f;            g_energy_costs[OP_SESS_SNAPSHOT][1] = 1000.0f;
    g_energy_costs[OP_SESS_COMPRESS][0] = 0.5f;             g_energy_costs[OP_SESS_COMPRESS][1] = 100.0f;
    g_energy_costs[OP_ORCH_CLASSIFY][0] = 0.5f;             g_energy_costs[OP_ORCH_CLASSIFY][1] = 1000.0f;
    g_energy_costs[OP_ORCH_PLAN][0] = 1.0f;                  g_energy_costs[OP_ORCH_PLAN][1] = 5000.0f;
    g_energy_costs[OP_ORCH_VALIDATE][0] = 0.5f;              g_energy_costs[OP_ORCH_VALIDATE][1] = 1000.0f;
    g_energy_costs[OP_ORCH_EXEC][0] = 0.5f;                  g_energy_costs[OP_ORCH_EXEC][1] = 10.0f;
    g_energy_costs[OP_ORCH_VERIFY][0] = 1.0f;                 g_energy_costs[OP_ORCH_VERIFY][1] = 2000.0f;
    g_energy_costs[OP_ORCH_RESPOND][0] = 0.5f;                g_energy_costs[OP_ORCH_RESPOND][1] = 100.0f;
}
```

---

## Budget Check

```c
bool ops_check_budget(OpPacket* packets, uint32_t count, ExecContext* ctx) {
    float total_tokens = 0.0f;
    float total_time_us = 0.0f;
    
    for (uint32_t i = 0; i < count; i++) {
        OpCode op = packets[i].opcode;
        total_tokens += g_energy_costs[op][0];
        total_time_us += g_energy_costs[op][1];
    }
    
    // Convert to common units: tokens and milliseconds
    float total_time_ms = total_time_us / 1000.0f;
    
    bool tokens_ok = total_tokens <= ctx->session_budget_tokens;
    bool time_ok = total_time_ms <= ctx->session_budget_time_ms;
    
    if (!tokens_ok || !time_ok) {
        // Store breakdown in ValidationResult
        // Model receives: "Budget exceeded: tokens N > M, time P > Q ms"
    }
    
    return tokens_ok && time_ok;
}
```

---

## Measured vs Estimated

After execution, actual latencies replace estimates in session metrics. This feedback improves future estimates.

```c
void ops_update_energy_estimate(OpCode opcode, float measured_latency_us) {
    // Exponential moving average: new = 0.7 * old + 0.3 * measured
    float old = g_energy_costs[opcode][1];
    g_energy_costs[opcode][1] = 0.7f * old + 0.3f * measured_latency_us;
}
```

Alpha = 0.3 provides fast adaptation to environment changes without overreacting to outliers.

---

## Optimization: Chain Reordering

Given a set of independent operations (no conflicts), reorder to minimize S:

```
// Sort by (cost_tokens / latency) descending
// Operations that provide the most "tokens per unit time" first
// = operations that are cheap in tokens but fast in execution
// This minimizes total "waiting time weighted by token cost"
```

Example:
- Op A: tokens=5, latency=1000us → ratio=0.005
- Op B: tokens=1, latency=100us → ratio=0.01
- Op C: tokens=2, latency=200us → ratio=0.01

Optimal order: B, C, A (or C, B, A - same ratio).
This is only applicable when operations are independent (no conflicts, no dependencies).

In practice, chain order is determined by semantics, not energy. Energy optimization is secondary to correctness. The energy matrix is primarily for budget enforcement and reporting.
