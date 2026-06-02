# Mimic — Configuration Variables Outcome

> Single authoritative document. Every customizable variable in Mimic.
> Schema: name | type | default | scope | invariant | source | description.
> If a variable exists in code but not here — CI `make check-config` fails.

---

## 1. Core Server

### `MIMIC_PORT`
- **type**: uint16
- **default**: 1337
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && != 22 && != 80`
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: Primary MCP server port for stdio/SSE/HTTP transport.
- **affects**: `internal/mcp/tcp.go`, `cmd/mimic/main.go`, `Dockerfile EXPOSE`

### `MIMIC_HTTP_PORT`
- **type**: uint16
- **default**: 1117
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && != MIMIC_PORT`
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: REST API and Prometheus metrics port.

### `MIMIC_ADMIN_PORT`
- **type**: uint16
- **default**: 1227
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && unique among all ports`
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: Management API and dynamic configuration port.

### `MIMIC_WS_PORT`
- **type**: uint16
- **default**: 1447
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && unique among all ports`
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: WebSocket real-time bidirectional transport port.

### `MIMIC_MESH_PORT`
- **type**: uint16
- **default**: 1557
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && unique among all ports`
- **source**: embryo (mesh_agent_cascade)
- **description**: Inter-node mesh communication (future distributed mode).

### `MIMIC_LOG_LEVEL`
- **type**: string
- **default**: info
- **scope**: env, runtime
- **invariant**: `in [debug, info, warn, error]`
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: Log verbosity. Debug enables packet dumps.

### `MIMIC_WORK_DIR`
- **type**: string
- **default**: `.` (process working directory)
- **scope**: env, runtime
- **invariant**: `dir_exists(path) || mkdir_p(path)`
- **source**: embryo (mcp_tools_registry)
- **description**: Root directory for all file operations and workspace indexing.

---

## 2. Budget & Limits

### `MIMIC_BUDGET_TOKENS`
- **type**: uint64
- **default**: 100000
- **scope**: env, runtime, session
- **invariant**: `> 0 && <= 10000000000`
- **source**: Mayveskii/code-mode (budget_enforcement)
- **description**: Hard limit on total LLM tokens consumed per agent session.

### `MIMIC_BUDGET_TIME_MS`
- **type**: uint64
- **default**: 3600000
- **scope**: env, runtime, session
- **invariant**: `> 0 && <= 86400000` (24h)
- **source**: Mayveskii/code-mode (budget_enforcement)
- **description**: Maximum session wall-clock time in milliseconds.

### `MIMIC_MAX_TASKS`
- **type**: uint16
- **default**: 50
- **scope**: env, runtime
- **invariant**: `> 0 && <= 10000`
- **source**: Mayveskii/code-mode (concurrency_control)
- **description**: Maximum concurrent tasks (semaphore limit).

### `MIMIC_MAX_CHAIN_LENGTH`
- **type**: uint16
- **default**: 1024
- **scope**: compile, runtime (C-core `#define`)
- **invariant**: `> 0 && <= 65535`
- **source**: 03-EXECUTION-SPACE.md
- **description**: Maximum OpPacket operations per single chain.

### `MIMIC_MAX_TOTAL_BUFFER`
- **type**: uint32
- **default**: 10485760 (10MB)
- **scope**: compile, runtime (C-core `#define`)
- **invariant**: `>= 4096 && <= 1073741824` (1GB)
- **source**: 03-EXECUTION-SPACE.md
- **description**: Maximum total result buffer across a chain.

---

## 3. RTK Compression

### `MIMIC_RTK_ENABLED`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false` (no side effects if false)
- **source**: Mayveskii/rtk (toml_filter_pipeline)
- **description**: Enable output compression before returning to agent.

### `MIMIC_RTK_MAX_LINES`
- **type**: uint32
- **default**: 100
- **scope**: env, runtime
- **invariant**: `>= 0 && <= 100000`
- **source**: Mayveskii/rtk (toml_filter_pipeline)
- **description**: Head+tail truncation: keep first N + last N lines. 0 = disable truncation.

### `MIMIC_RTK_STRIP_ANSI`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: Mayveskii/rtk (toml_filter_pipeline)
- **description**: Strip ANSI escape sequences from tool output.

### `MIMIC_RTK_COLLAPSE_BLANKS`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: Mayveskii/rtk (toml_filter_pipeline)
- **description**: Collapse consecutive blank lines to single blank.

### `MIMIC_RTK_MODE`
- **type**: string
- **default**: standard
- **scope**: env, runtime
- **invariant**: `in [minimal, standard, aggressive, verbatim]`
- **source**: Mayveskii/caveman (intensity_level_config)
- **description**: Compression intensity. Minimal = strip only ANSI. Aggressive = strip comments/bodies. Verbatim = no compression.

---

## 4. Deep Cache & Mesh

### `MIMIC_CACHE_ENABLED`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: embryo (mesh_graph_slots)
- **description**: Enable local deep cache for pattern lookup.

### `MIMIC_CACHE_DIR`
- **type**: string
- **default**: /var/lib/mimic/cache (Linux), ~/.mimic/cache (macOS)
- **scope**: env, runtime
- **invariant**: `dir_exists || mkdir_p`
- **source**: embryo (mesh_graph_slots)
- **description**: Path to cache directory for mesh slots and matrices.

### `MIMIC_CACHE_MIN_SI`
- **type**: float64
- **default**: 0.8
- **scope**: env, runtime
- **invariant**: `>= 0.0 && <= 1.0`
- **source**: embryo (survival_index)
- **description**: Minimum survival index for cached patterns to be considered proven.

### `MIMIC_CACHE_MAX_SIZE_MB`
- **type**: uint64
- **default**: 1024
- **scope**: env, runtime
- **invariant**: `>= 100 && <= 1048576` (1TB)
- **source**: embryo (mesh_graph_slots)
- **description**: Maximum cache size in megabytes. LRU eviction when exceeded.

### `MIMIC_MESH_DIR`
- **type**: string
- **default**: data/distilled/
- **scope**: env, runtime
- **invariant**: `dir_exists || error`
- **source**: embryo (mesh_graph_slots)
- **description**: Path to mesh graph files (.gob, .json).

### `MIMIC_SEED_DIR`
- **type**: string
- **default**: data/seeds/
- **scope**: env, runtime
- **invariant**: `dir_exists || mkdir_p`
- **source**: ADR-0009 (auto-ingestion)
- **description**: Where raw research seeds are written before distillation.

### `MIMIC_AUTO_RESEARCH`
- **type**: bool
- **default**: false
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: ADR-0009 (auto-ingestion)
- **description**: Enable orchestrator auto-research on mesh query gaps.

### `MIMIC_AUTO_RESEARCH_MIN_SIM`
- **type**: float64
- **default**: 0.4
- **scope**: env, runtime
- **invariant**: `>= 0.0 && <= 1.0`
- **source**: ADR-0009 (auto-ingestion)
- **description**: Similarity threshold below which a mesh gap is declared.

### `MIMIC_AUTO_RESEARCH_MAX_RESULTS`
- **type**: uint8
- **default**: 5
- **scope**: env, runtime
- **invariant**: `>= 1 && <= 20`
- **source**: ADR-0009 (auto-ingestion)
- **description**: How many Exa results to fetch per auto-research gap.

---

## 5. Security

### `MIMIC_CONFIRM_DANGEROUS`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: Mayveskii/code-mode (permission_pipeline)
- **description**: Require explicit confirmation for safety_level=0 operations.

### `MIMIC_CIRCUIT_BREAKER_THRESHOLD`
- **type**: uint8
- **default**: 3
- **scope**: env, runtime
- **invariant**: `>= 1 && <= 20`
- **source**: Mayveskii/code-mode (denial_tracking)
- **description**: Max consecutive denials before circuit breaker trips.

### `MIMIC_CIRCUIT_BREAKER_COOLDOWN`
- **type**: uint16
- **default**: 60
- **scope**: env, runtime
- **invariant**: `>= 1 && <= 3600`
- **source**: Mayveskii/code-mode (denial_tracking)
- **description**: Cooldown period in seconds after circuit breaker trips.

---

## 6. External APIs

### `EXA_API_KEY`
- **type**: string
- **default**: "" (empty)
- **scope**: env, secret
- **invariant**: `len == 0 || len >= 20` (no default, no example keys)
- **source**: Mayveskii/exa-mcp-server (exa_rate_limit_management)
- **description**: API key for Exa.ai search and content fetch. Empty = Exa tools disabled.

### `EXA_BASE_URL`
- **type**: string
- **default**: https://api.exa.ai
- **scope**: env, runtime
- **invariant**: `valid_url_scheme_https`
- **source**: Mayveskii/exa-mcp-server (exa_web_search)
- **description**: Exa API endpoint. Override for enterprise proxy or local mock.

### `EXA_MAX_RESULTS`
- **type**: uint8
- **default**: 10
- **scope**: env, runtime
- **invariant**: `>= 1 && <= 100`
- **source**: Mayveskii/exa-mcp-server (exa_web_search)
- **description**: Default numResults for EXA_SEARCH.

### `EXA_TIMEOUT_MS`
- **type**: uint32
- **default**: 30000
- **scope**: env, runtime
- **invariant**: `>= 1000 && <= 300000` (5 min)
- **source**: Mayveskii/exa-mcp-server (exa_rate_limit_management)
- **description**: Per-request timeout for Exa API calls.

### `EXA_RETRY_MAX`
- **type**: uint8
- **default**: 3
- **scope**: env, runtime
- **invariant**: `>= 0 && <= 10`
- **source**: Mayveskii/exa-mcp-server (exa_rate_limit_management)
- **description**: Max retries on 429 / 5xx responses.

### `EXA_RETRY_BACKOFF_BASE_MS`
- **type**: uint32
- **default**: 1000
- **scope**: env, runtime
- **invariant**: `>= 100 && <= 60000`
- **source**: Mayveskii/exa-mcp-server (exa_rate_limit_management)
- **description**: Base delay in ms for exponential backoff: `base * 2^attempt`.

### `MIMIC_QDRANT_URL`
- **type**: string
- **default**: http://localhost:6333
- **scope**: env, runtime
- **invariant**: `valid_url`
- **source**: embryo (hybrid_rag)
- **description**: Qdrant vector database endpoint.

### `MIMIC_QDRANT_COLLECTION`
- **type**: string
- **default**: binary_mesh_chunks
- **scope**: env, runtime
- **invariant**: `len > 0`
- **source**: embryo (hybrid_rag)
- **description**: Qdrant collection name for vector search.

### `MIMIC_EMBED_ENDPOINT`
- **type**: string
- **default**: http://localhost:1137
- **scope**: env, runtime
- **invariant**: `valid_url`
- **source**: embryo (hybrid_rag)
- **description**: TextEmbedding service endpoint (all-MiniLM-L6-v2).

---

## 7. C-Core Compile-Time

### `MAX_CHAIN_LENGTH`
- **type**: uint16
- **default**: 1024
- **scope**: compile (C-core `#define`)
- **invariant**: `> 0 && <= 65535`
- **source**: 03-EXECUTION-SPACE.md
- **description**: Maximum operations per chain. Fixed at compile for stack allocation.

### `MAX_TOTAL_BUFFER_SIZE`
- **type**: uint32
- **default**: 10485760 (10MB)
- **scope**: compile (C-core `#define`)
- **invariant**: `>= 4096`
- **source**: 03-EXECUTION-SPACE.md
- **description**: Total buffer cap across chain execution.

### `MIMIC_CC`
- **type**: string
- **default**: gcc
- **scope**: compile (Makefile)
- **invariant**: `executable_in_path`
- **source**: C build convention
- **description**: C compiler for libcore.a.

### `MIMIC_CFLAGS`
- **type**: string
- **default**: `-O2 -Wall -fPIC`
- **scope**: compile (Makefile)
- **invariant**: `no -D_FORTIFY_SOURCE conflicts`
- **source**: C build convention
- **description**: Additional C compiler flags.

### `CGO_ENABLED`
- **type**: string (boolish)
- **default**: 1
- **scope**: compile (env)
- **invariant**: `in [0, 1]`
- **source**: Go build convention
- **description**: Enable CGO bridge to C-core. 0 = Go-only (stubs).

---

## 8. Docker / Compose

### `MIMIC_DOCKER_HEALTHCHECK_INTERVAL`
- **type**: string (duration)
- **default**: 30s
- **scope**: compose, runtime
- **invariant**: `parse_duration >= 5s && <= 5m`
- **source**: Mayveskii/go-service-template-rest (probe_health_interface)
- **description**: Docker healthcheck interval.

### `MIMIC_DOCKER_HEALTHCHECK_TIMEOUT`
- **type**: string (duration)
- **default**: 10s
- **scope**: compose, runtime
- **invariant**: `parse_duration >= 1s && <= 60s`
- **source**: Mayveskii/go-service-template-rest (probe_health_interface)
- **description**: Docker healthcheck timeout.

### `MIMIC_DOCKER_RESTART_POLICY`
- **type**: string
- **default**: unless-stopped
- **scope**: compose, runtime
- **invariant**: `in [no, always, unless-stopped, on-failure]`
- **source**: Docker convention
- **description**: Container restart policy.

---

## 9. Client / npm

### `MIMIC_VERSION`
- **type**: string (semver)
- **default**: "" (use package.json version)
- **scope**: client, env
- **invariant**: `len == 0 || matches_semver`
- **source**: ADR-0007 (npm distribution)
- **description**: Override binary version for npm install script. Empty = auto-detect from package.json.

### `MIMIC_BINARY_CACHE_DIR`
- **type**: string
- **default**: ~/.mimic/bin
- **scope**: client, env
- **invariant**: `dir_exists || mkdir_p`
- **source**: ADR-0007 (npm distribution)
- **description**: Local cache directory for downloaded native binaries.

### `MIMIC_SKIP_POSTINSTALL`
- **type**: bool
- **default**: false
- **scope**: client, env
- **invariant**: `true || false`
- **source**: ADR-0007 (npm distribution)
- **description**: Skip binary download during npm install (useful for offline/air-gapped).

---

## 10. Data Pipeline (Sync / Distill / Reach)

### `MIMIC_SYNC_INTERVAL_HOURS`
- **type**: uint16
- **default**: 168 (7 days)
- **scope**: env, cron
- **invariant**: `>= 1 && <= 720` (30 days)
- **source**: repos-manifest.yaml workflow
- **description**: Interval between `sync repos` runs in Data CI.

### `MIMIC_DISTILL_BATCH_SIZE`
- **type**: uint16
- **default**: 10
- **scope**: env, runtime (distill script)
- **invariant**: `>= 1 && <= 100`
- **source**: data/extraction/distill_pipeline.py
- **description**: Number of repos to process per distill batch.

### `MIMIC_DISTILL_MIN_SURVIVAL`
- **type**: float64
- **default**: 0.8
- **scope**: env, runtime (distill script)
- **invariant**: `>= 0.0 && <= 1.0`
- **source**: embryo (survival_index)
- **description**: Minimum survival index for a commit to be distilled into a slot.

### `MIMIC_REACH_ENABLED`
- **type**: bool
- **default**: true
- **scope**: env, runtime
- **invariant**: `true || false`
- **source**: ADR-0009 (auto-ingestion)
- **description**: Enable web reach (Exa) for knowledge gap closure.

---

## CI Drift Detection

Every variable above is checked by `scripts/check_config_consistency.py`:
1. Parse `11-CONFIGURATION.md` → extract all `### \`NAME\`` blocks.
2. Grep codebase for each `NAME` in `.go`, `.c`, `.h`, `.yaml`, `.sh`, `Makefile`.
3. If variable declared in code but missing in doc → **FAIL**.
4. If variable in doc but no reference in code → **WARN** (may be future/planned).
5. Run on every PR via `make check-config`.

---

*Version: 2026-05-20 | Status: Draft | Next: populate all 50+ variables, add CI check script.*
