# C-Core Environment Configuration

All tunable parameters available via environment variables. Every parameter has a default value. All are optional.

The C-core reads env at `ops_init()` time. Values are cached. No re-reading during session (determinism).

---

## Core Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_WORKSPACE` | string | `.` | any valid path | Workspace root. All paths resolved against this. |
| `MIMIC_LOG_LEVEL` | enum | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` | Log verbosity. |
| `MIMIC_LOG_PATH` | string | `stderr` | path or `stderr` | Log destination. |
| `MIMIC_MAX_OPS_PER_CHAIN` | uint32 | 1024 | 1..65536 | Max packets in single chain. |
| `MIMIC_MAX_CONCURRENT_CHAINS` | uint32 | 10 | 1..256 | Max parallel pipelines. |
| `MIMIC_PACKET_POOL_SIZE` | uint32 | 1024 | 1..65536 | Pre-allocated packet pool count. |

---

## Budget Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_BUDGET_TOKENS` | uint64 | 100000 | 0..UINT64_MAX | Session token budget. 0 = unlimited. |
| `MIMIC_BUDGET_TIME_MS` | uint64 | 3600000 | 0..UINT64_MAX | Session time budget (ms). 0 = unlimited. Override for research: 86400000 (24h). |

---

## Research Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_RESEARCH_MODE` | bool | false | true, false | Enable research session semantics (larger context, checkpoint interval). |
| `MIMIC_CHECKPOINT_INTERVAL_MINUTES` | uint32 | 30 | 1..1440 | Auto-checkpoint interval during long tasks. |
| `MIMIC_CHECKPOINT_MAX_RETAINED` | uint32 | 100 | 1..10000 | Max checkpoints retained per session. |
| `MIMIC_SESSION_HISTORY_ENABLED` | bool | true | true, false | Enable querying past session snapshots. |
| `MIMIC_SESSION_HISTORY_RETENTION_DAYS` | uint32 | 365 | 1..3650 | How long to keep session history. |
| `MIMIC_RAG_QUERY_MAX_CHARS` | uint32 | 32768 | 1024..131072 | Max RAG query length for research mode. |
| `MIMIC_STRATEGY_PIVOT_THRESHOLD` | uint8 | 3 | 1..255 | Failures before strategy pivot triggers. |
| `MIMIC_SUBTASK_AUTO_DECOMPOSE` | bool | true | true, false | Auto-break large tasks into subtasks. |
| `MIMIC_SELF_CHECKPOINT_ENABLED` | bool | true | true, false | Allow agent to trigger checkpoints. |

---

## Timeout Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_TIMEOUT_DEFAULT_MS` | uint32 | 30000 | 1..UINT32_MAX | Default timeout for unspecified ops. |
| `MIMIC_TIMEOUT_NETWORK_MS` | uint32 | 30000 | 1..UINT32_MAX | Network operation timeout. |
| `MIMIC_TIMEOUT_GIT_MS` | uint32 | 300000 | 1..UINT32_MAX | Git operation timeout. |
| `MIMIC_TIMEOUT_BUILD_MS` | uint32 | 3600000 | 1..UINT32_MAX | Build/test timeout. |
| `MIMIC_TIMEOUT_PROC_MS` | uint32 | 300000 | 1..UINT32_MAX | Process spawn/wait timeout. |
| `MIMIC_TIMEOUT_TCP_CONNECT_MS` | uint32 | 10000 | 1..UINT32_MAX | TCP connect timeout. |
| `MIMIC_TIMEOUT_HEALTH_CHECK_MS` | uint32 | 30000 | 1..UINT32_MAX | Deploy health check timeout. |

---

## Resource Limits

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_MAX_FD_PER_CONTEXT` | uint32 | 1024 | 1..65536 | Max tracked file descriptors per chain. |
| `MIMIC_MAX_MMAP_PER_CONTEXT` | uint32 | 1024 | 1..65536 | Max tracked mmap regions per chain. |
| `MIMIC_MAX_BUFFER_SIZE` | size_t | 1073741824 | 1..SIZE_MAX | Max buffer size (1GB). |
| `MIMIC_MAX_HTTP_BODY_SIZE` | size_t | 52428800 | 1..SIZE_MAX | Max HTTP response body (50MB). |
| `MIMIC_MAX_TEXT_BODY_SIZE` | size_t | 10485760 | 1..SIZE_MAX | Max text response body (10MB). |
| `MIMIC_MAX_CONTEXT_SIZE_MB` | uint64 | 64 | 1..1024 | Max session context size (MB). Default raised for research mode. |
| `MIMIC_MAX_RESEARCH_CONTEXT_MB` | uint64 | 256 | 1..4096 | Max research session context (MB). Higher limit for long-running research. |
| `MIMIC_MAX_SLOT_TEXT_SIZE` | uint32 | 65536 | 1..1048576 | Max slot text content (64KB). Literature > 64KB uses linked slots. |

---

## Security Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_SANDBOX_ENABLED` | bool | true | true, false | Enable Landlock/Seatbelt sandbox. |
| `MIMIC_WORKSPACE_BOUNDARY_ENFORCED` | bool | true | true, false | Enforce path within workspace. |
| `MIMIC_CIRCUIT_BREAK_THRESHOLD` | uint8 | 3 | 1..255 | Consecutive denials before circuit break. |
| `MIMIC_CIRCUIT_BREAK_AUTO_RESET` | bool | false | true, false | Auto-reset circuit after timeout. |
| `MIMIC_CIRCUIT_BREAK_RESET_TIMEOUT_MS` | uint32 | 300000 | 1..UINT32_MAX | Auto-reset timeout if enabled (5min). |
| `MIMIC_NEVER_RULES_OVERRIDE` | bool | false | true, false | **DEBUG ONLY**: disable never-rules. |
| `MIMIC_ALLOW_PRIVATE_IP` | bool | false | true, false | Allow private IP access (SSRF bypass). |
| `MIMIC_ALLOW_HTTP_SCHEME` | bool | false | true, false | Allow http:// (default https only). |

---

## Quality Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_QUALITY_PRECISION_THRESHOLD` | float | 0.80 | 0.0..1.0 | Min artifact_precision for indexing. |
| `MIMIC_QUALITY_MAX_QAC_FAILURES` | uint8 | 3 | 0..13 | Max QAC failures for indexing. |
| `MIMIC_2VOTE_ENABLED` | bool | true | true, false | Enable 2-vote verification. |
| `MIMIC_2VOTE_VERIFIER_COUNT` | uint8 | 2 | 1..255 | Number of verifiers (default 2). |

---

## Distillation Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_DISTILL_REPO_RETENTION_DAYS` | uint32 | 90 | 1..365 | Source repo retention after last distillation. |
| `MIMIC_DISTILL_BATCH_SIZE` | uint32 | 100 | 1..10000 | New slots per batch before index update. |
| `MIMIC_DISTILL_ASYNC` | bool | true | true, false | Run distillation in background. |
| `MIMIC_DISTILL_MAX_SOURCES_PER_SLOT` | uint8 | 4 | 1..16 | Max source repos per slot. |

---

## Mesh Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_MESH_PATH` | string | `.mimic/mesh/mesh.bmap` | any path | Mesh storage file path. |
| `MIMIC_MESH_BACKUP_COUNT` | uint8 | 7 | 1..255 | Max mesh backups to retain. |
| `MIMIC_MESH_INDEX_REBUILD_INTERVAL_HOURS` | uint32 | 24 | 1..168 | Index integrity check interval. |
| `MIMIC_MESH_GC_INTERVAL_DAYS` | uint32 | 30 | 1..365 | Garbage collection interval. |
| `MIMIC_MESH_PAGE_SIZE` | uint32 | 4096 | 512..65536 | Slot alignment page size. |

---

## Backup Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_BACKUP_PATH` | string | `.mimic/backups/` | any path | Backup directory for file overwrites. |
| `MIMIC_BACKUP_RETENTION_DAYS` | uint32 | 7 | 1..365 | Backup retention. |
| `MIMIC_SNAPSHOT_PATH` | string | `.mimic/snapshots/` | any path | Session snapshot directory. |
| `MIMIC_SNAPSHOT_RETENTION_DAYS` | uint32 | 90 | 1..365 | Snapshot retention. |

---

## Network Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_NETWORK_RATE_LIMIT_PER_MIN` | uint32 | 60 | 1..10000 | Max requests per minute per domain. |
| `MIMIC_NETWORK_MAX_CONCURRENT` | uint32 | 100 | 1..10000 | Max concurrent network connections. |
| `MIMIC_NETWORK_DNS_CACHE_TTL_S` | uint32 | 300 | 0..3600 | DNS cache TTL (0 = disabled). |
| `MIMIC_NETWORK_RETRY_MAX` | uint32 | 3 | 0..10 | Max retries per operation. |
| `MIMIC_NETWORK_RETRY_BACKOFF_MS` | uint32 | 1000 | 1..60000 | Exponential backoff base (ms). |

---

## MCP Server Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_MCP_ENABLED` | bool | true | true, false | Enable MCP JSON-RPC server. |
| `MIMIC_MCP_TRANSPORT` | enum | `stdio` | `stdio`, `tcp`, `unix` | MCP transport type. |
| `MIMIC_MCP_PORT` | uint16 | 0 | 1..65535 | TCP port (0 = auto-assign). |
| `MIMIC_MCP_UNIX_SOCKET` | string | `.mimic/mcp.sock` | any path | Unix socket path. |
| `MIMIC_MCP_LOG_LEVEL` | enum | `WARN` | `DEBUG`, `INFO`, `WARN`, `ERROR` | MCP-specific log level. |
| `MIMIC_MCP_IDLE_TIMEOUT_MS` | uint32 | 300000 | 1..UINT32_MAX | Idle connection timeout. |
| `MIMIC_MCP_MAX_REQUEST_SIZE` | size_t | 1048576 | 1..SIZE_MAX | Max JSON-RPC request size (1MB). |
| `MIMIC_MCP_MAX_RESPONSE_SIZE` | size_t | 10485760 | 1..SIZE_MAX | Max JSON-RPC response size (10MB). |

---

## Build Parameters

| Variable | Type | Default | Range | Description |
|---|---|---|---|---|
| `MIMIC_BUILD_PARALLEL_SHARDS` | uint32 | 10 | 1..256 | Max concurrent build shards. |
| `MIMIC_BUILD_SYSTEM_DETECT` | bool | true | true, false | Auto-detect build system. |
| `MIMIC_BUILD_TIMEOUT_MS` | uint32 | 3600000 | 1..UINT32_MAX | Default build timeout override. |
| `MIMIC_BUILD_COVERAGE_ENABLED` | bool | true | true, false | Collect test coverage. |

---

## Reading Convention

All environment variables are read once at `ops_init()`. Values cached in global config struct. Changes after init are ignored for determinism.

```c
typedef struct {
    // Core
    char workspace[PATH_MAX];
    uint32_t log_level;
    char log_path[PATH_MAX];
    uint32_t max_ops_per_chain;
    uint32_t max_concurrent_chains;
    uint32_t packet_pool_size;
    
    // Budget
    uint64_t budget_tokens;
    uint64_t budget_time_ms;
    uint64_t budget_memory_mb;
    float budget_warn_threshold;
    float budget_critical_threshold;
    
    // Timeouts
    uint32_t timeout_default_ms;
    uint32_t timeout_network_ms;
    uint32_t timeout_git_ms;
    uint32_t timeout_build_ms;
    uint32_t timeout_proc_ms;
    uint32_t timeout_tcp_connect_ms;
    uint32_t timeout_health_check_ms;
    
    // Resources
    uint32_t max_fd_per_context;
    uint32_t max_mmap_per_context;
    size_t max_buffer_size;
    size_t max_http_body_size;
    size_t max_text_body_size;
    uint64_t max_context_size_mb;
    uint64_t max_research_context_mb;  // UNCERTAIN: optimal size needs measurement
    uint32_t max_slot_text_size;
    
    // Security
    bool sandbox_enabled;
    bool workspace_boundary_enforced;
    uint8_t circuit_break_threshold;
    bool circuit_break_auto_reset;
    uint32_t circuit_break_reset_timeout_ms;
    bool never_rules_override;      // DEBUG ONLY
    bool allow_private_ip;          // SSRF bypass
    bool allow_http_scheme;          // HTTP allowed
    
    // Quality
    float quality_precision_threshold;
    uint8_t quality_max_qac_failures;
    bool two_vote_enabled;
    uint8_t two_vote_verifier_count;
    
    // Distillation
    uint32_t distill_repo_retention_days;
    uint32_t distill_batch_size;
    bool distill_async;
    uint8_t distill_max_sources_per_slot;
    
    // Mesh
    char mesh_path[PATH_MAX];
    uint8_t mesh_backup_count;
    uint32_t mesh_index_rebuild_interval_hours;
    uint32_t mesh_gc_interval_days;
    uint32_t mesh_page_size;
    
    // Backup
    char backup_path[PATH_MAX];
    uint32_t backup_retention_days;
    char snapshot_path[PATH_MAX];
    uint32_t snapshot_retention_days;
    
    // Network
    uint32_t network_rate_limit_per_min;
    uint32_t network_max_concurrent;
    uint32_t network_dns_cache_ttl_s;
    uint32_t network_retry_max;
    uint32_t network_retry_backoff_ms;
    
    // MCP
    bool mcp_enabled;
    uint32_t mcp_transport;     // 0=stdio, 1=tcp, 2=unix
    uint16_t mcp_port;
    char mcp_unix_socket[PATH_MAX];
    uint32_t mcp_log_level;
    uint32_t mcp_idle_timeout_ms;
    size_t mcp_max_request_size;
    size_t mcp_max_response_size;
    
    // Research / Self-Management
    bool research_mode;
    uint32_t checkpoint_interval_minutes;
    uint32_t checkpoint_max_retained;
    bool session_history_enabled;
    uint32_t session_history_retention_days;
    uint32_t rag_query_max_chars;
    uint8_t strategy_pivot_threshold;
    bool subtask_auto_decompose;
    bool self_checkpoint_enabled;
    
} MimicConfig;

Size: ~2.2KB. Fits in L1 cache.

---

## Optional Compilation

Some parameters affect only optional features:

- `MIMIC_MCP_*`: Only compiled when `MIMIC_ENABLE_MCP` is defined at build time.
- `MIMIC_SANDBOX_ENABLED`: Only compiled when `MIMIC_ENABLE_SANDBOX` is defined.
- `MIMIC_DISTILL_*`: Only compiled when `MIMIC_ENABLE_DISTILLATION` is defined.
- `MIMIC_MESH_*`: Only compiled when `MIMIC_ENABLE_MESH` is defined (core always enabled).
- `MIMIC_BUILD_*`: Only relevant when `MIMIC_ENABLE_BUILD_DOMAIN` is defined.
- `MIMIC_RESEARCH_*`: Only compiled when `MIMIC_ENABLE_RESEARCH_DOMAIN` is defined.
- `MIMIC_CHECKPOINT_*`: Only compiled when `MIMIC_ENABLE_SELF_MANAGEMENT` is defined.

See `BUILD_CONFIG.md` for compile-time flags.
