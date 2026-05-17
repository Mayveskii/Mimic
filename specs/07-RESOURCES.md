# RESOURCES.md — Mimic

Complete map of every resource in Mimic: what it is, what it does for agent operations, how it translates to the OpPacket language.

---

## Resource Translation Principle

Every agent operation (bash command, git operation, file read, build step) is translated into an OpPacket — a structured, validated, deterministic unit. The agent never sees OpCodes; it sees MCP tools with human-readable names. Mimic translates intent → OpPacket chain → validated execution → result.

```
Agent intent: "commit these files safely"
    ↓
Mimic translation:
    OpPacket[0]: OP_GIT_STATUS  (args: {path: "."})
    OpPacket[1]: OP_GIT_DIFF    (args: {staged: true})
    OpPacket[2]: OP_GIT_ADD     (args: {files: [...]})
    OpPacket[3]: OP_GIT_COMMIT  (args: {message: "..."})
    ↓
Validation: conflict_matrix[0..3] = all 0, energy_budget = 8.0 tokens, latency ≈ 15000μs
    ↓
Execution: ops_execute_chain(packets, 4, ctx)
    ↓
Result: {success: true, commit: "abc123", latency_ns: 12400000}
```

---

## C-Core Resources (core/)

### ops.c / ops.h — Core Engine

| Resource | Purpose for Agent Operations | OpPacket Translation |
|----------|------------------------------|---------------------|
| ops_init | Initialize the engine before any operation | Called once at startup, no OpPacket |
| ops_register | Register an OpCode definition with executor, cost, safety | Internal, no OpPacket |
| ops_register_builtins | Register NOP, FILE_EXISTS, DIR_CREATE | Internal, no OpPacket |
| ops_execute | Execute a single OpPacket | Agent tool call → single OpPacket |
| ops_execute_chain | Execute a validated sequence of OpPackets | Agent tool call → scenario → multiple OpPackets |
| ops_validate_chain | Validate chain without executing (dry-run) | Agent `validate` tool → ValidationResult |
| ops_check_conflict | Check if two OpCodes conflict | Internal, used by validate_chain |
| ops_calculate_action | Compute action S = Σ(cost_tokens × cost_time_us) | Internal, used by orchestrator to choose cheapest chain |
| ops_packet_init | Create an OpPacket with auto-incremented ID | Internal, called by chain builder |
| ops_packet_set_string | Set a string argument (key→value) on an OpPacket | Internal, called by chain builder |
| ops_packet_set_int | Set an integer argument on an OpPacket | Internal, called by chain builder |
| ops_mmap_alloc | Allocate memory-mapped region | Agent operation needing large buffers |
| ops_mmap_free | Free mmap region | Cleanup after buffer operation |
| ops_mmap_sync | Sync mmap region to disk | Agent operation needing durability |

### git_ops.c — Git Operations

| OpCode | Agent Operation | OpPacket Args | Result |
|--------|----------------|---------------|--------|
| OP_GIT_INIT | Initialize a git repository | {path: "."} | repo created |
| OP_GIT_CLONE | Clone a repository | {url: "...", path: "..."} | repo cloned |
| OP_GIT_FETCH | Fetch remote changes | {remote: "origin"} | refs updated |
| OP_GIT_COMMIT | Commit staged changes | {message: "..."} | commit hash |
| OP_GIT_PUSH | Push to remote | {remote: "origin", branch: "main"} | refs pushed |
| OP_GIT_DIFF | Show differences | {staged: true} or {base: "...", head: "..."} | diff output |
| OP_GIT_STATUS | Show working tree status | {path: "."} | status output |
| OP_GIT_CHECKOUT | Switch branch/commit | {ref: "..."} | HEAD moved |
| OP_GIT_BRANCH | Create/list branches | {name: "..."} or {list: true} | branch info |
| OP_GIT_MERGE | Merge branches | {source: "...", target: "..."} | merge result |
| OP_GIT_REBASE | Rebase onto branch | {onto: "..."} | rebase result |

### git_scenarios.c — Pre-built Scenario Chains

| Scenario | Agent Intent | OpPacket Chain | Invariants |
|----------|-------------|----------------|------------|
| atomic_commit | "commit safely without breaking anything" | OP_GIT_STATUS → OP_GIT_DIFF → OP_GIT_ADD → OP_GIT_COMMIT | Status must be clean after commit |
| safe_merge | "merge without force-push" | OP_GIT_FETCH → OP_GIT_DIFF → OP_GIT_MERGE (ff-only) | Fast-forward only, no merge commits |
| feature_branch | "create a feature branch" | OP_GIT_BRANCH → OP_GIT_CHECKOUT | Branch must not exist |
| hotfix | "hotfix and merge into target" | OP_GIT_BRANCH → OP_GIT_COMMIT → OP_GIT_CHECKOUT(target) → OP_GIT_MERGE | Must merge back into target |
| ci_diff_check | "check for whitespace errors" | OP_GIT_DIFF(base, head) → check output | No trailing whitespace, no merge conflict markers |

### mmap_ops.c — Memory Operations

| OpCode | Agent Operation | OpPacket Args | Result |
|--------|----------------|---------------|--------|
| OP_MMAP_ALLOC | Allocate memory region | {size: N} | pointer + size |
| OP_MMAP_FREE | Free memory region | {ptr, size} | freed |
| OP_MMAP_READ | Read from mmap region | {ptr, offset, len} | bytes read |
| OP_MMAP_WRITE | Write to mmap region | {ptr, offset, data} | bytes written |
| OP_MMAP_SYNC | Sync to disk | {ptr, size} | synced |

### git_search_ops.c — ⚠️ NEEDS FIX

Opcode collision 0x50-0x5F with NET range. Must be remapped to 0x60+ before use.

### libbmap.a — Binary Map Storage (39 symbols, no .c sources)

These resources form the **knowledge layer** — mesh slots, invariants, scoring, snapshots. They store and retrieve distilled patterns.

| Resource Group | Functions | Purpose for Agent Operations | OpPacket Connection |
|---------------|-----------|------------------------------|---------------------|
| bmap | open/close/read_cell/write_cell/write/free_cell/cell_serialized_size | Store and retrieve mesh slots (distilled patterns) | Slots referenced by query_slots MCP tool |
| slot_index | si_create/destroy/insert/query_domain/query_domain_layer/query_state_hash/build_from_bmap/result_free | Index slots by domain, layer, state hash for fast lookup | Agent: "find patterns for git domain" → si_query_domain("git") |
| invariant | inv_create/destroy/add/find_similar/load/save/dedup_check | Store and check preconditions for mimicry | Every borrowed behavior must have an invariant (BEHAVIOR.md #10) |
| gnk_scorer | gnk_compute/score_domains/result_free | Score domains by relevance and quality | Internal: rank which domains have best patterns |
| snapshot | snapshot_build/load/write/sign/diff/diff_free/free | Point-in-time state of the mesh | Agent: compare current mesh vs previous snapshot → drift_detect |
| matrix_ops | layer_walk/drift_detect | Walk mesh layers, detect drift between states | Agent: "has this domain's patterns changed?" → drift_detect |
| math | cosine_f32/cosine_int8/batch_cosine_int8/int8_quantize | Vector similarity for pattern matching | Agent: "find similar patterns" → inv_find_similar → cosine |
| hash | sha256_hash | Integrity verification | Agent: verify slot content unchanged |
| z_density | z_density_compute | Quality metric for mesh slots | Agent: "how good is this domain's knowledge?" → z_density_compute |

---

## CGO Bridge Resources (internal/cgo/)

### cgo_wrapper.go — Go ↔ C Translation

| Resource | Purpose | How Agent Operations Flow Through |
|----------|---------|----------------------------------|
| Init() | Start C-core | Called at Mimic startup |
| Shutdown() | Stop C-core | Called at Mimic shutdown |
| Execute(pkt) | Execute one OpPacket | Single MCP tool call → toCPacket → ops_execute → fromCPacket |
| ExecuteChain(pkts, ctx) | Execute validated chain | Scenario MCP tool call → array toCPacket → ops_execute_chain → array fromCPacket |
| ValidateChain(pkts) | Dry-run validation | Agent `validate` tool → toCPacket → ops_validate_chain → ValidationResult |
| CheckConflict(op1, op2) | Check two operations | Internal orchestrator use |
| CalculateAction(pkts) | Compute chain cost | Internal orchestrator: choose cheapest chain |
| MMapAlloc/Free/Sync | Memory management | Internal: large buffer operations |

### validator.go — Pre-flight Checks

| Resource | Purpose | How It Serves Agent Operations |
|----------|---------|-------------------------------|
| NewValidator() | Create validator with default rules | Called at startup |
| ValidateChain() | Go-side chain validation | Before CGO: max 1024 ops, max 10MB buffers, per-op rules, conflict rules |
| ValidateOne() | Single packet validation | Internal: check before adding to chain |
| GetOpDefinition() | ⚠️ Stub — only 3 explicit definitions | Needs full implementation for all 46 OpCodes |

**Go-side validation rules:**
- `valid_opcode`: reject OP_NOP and opcode ≥ OP_MAX
- `max_args`: reject > 16 args per op
- `buffer_size`: reject single buffer > 1MB
- `timeout`: reject timeout > 300000ms (5 min)
- `retry_count`: reject retry > 10
- `write_after_delete`: DELETE then WRITE = conflict
- `read_after_write_no_sync`: WRITE then READ without SYNC (currently no-op — needs implementation)

### helpers.go — Builder Pattern

| Resource | Purpose | Agent Operation Connection |
|----------|---------|--------------------------|
| NewOpPacket(opcode, args) | Create packet with SAFE flag | Chain builder entry point |
| WithFlag(flag) | Add flags (READONLY, ATOMIC, etc.) | Mark operation properties |
| WithTimeout(ms) | Set execution timeout | Agent specifies "this must complete in 5s" |
| WithRetry(count) | Set retry count on failure | Agent specifies "retry up to 3 times" |
| WithStringArg(key, value) | Add string argument | Agent provides path, message, URL, etc. |
| WithIntArg(key, value) | Add integer argument | Agent provides size, count, mode, etc. |

---

## Embryo Resources (Mayveskii/embryo)

The repo where Mimic was born. Go implementations that informed the C-core and internal/ design.

| Embryo Resource | What It Does | Mimic Equivalent | Status |
|----------------|-------------|-----------------|--------|
| pkg/do/BinaryRuntime | OpCodes 0x10-0x41, tool loop, session, inference pool, hooks | core/ops.c + internal/orchestrator/ | c-core derived from this |
| pkg/orchestrator/ | Pipeline: state→mesh→DIRECT→classify→exec→flywheel→respond | internal/orchestrator/ | planned |
| pkg/agent/ | MeshAgent + Cascade (3-tier: local→medium→top) | internal/orchestrator/ | planned |
| pkg/hunt/ (17 files) | Hunt system: assess, compress, github, meshscan | internal/tool/ | planned |
| pkg/mesh/ | Mesh graphs, slots, domains | core/ (bmap) | libbmap.a derived from this |
| pkg/rag/ | 5-signal hybrid RAG + qdrant | internal/orchestrator/ | planned |
| pkg/projectmap/ | SQLite navigation | internal/orchestrator/ | planned |
| pkg/survival/ | git blame → Survival Index | core/ (distillation) | planned |
| pkg/mapstore/ | Slot, store, invariant_registry | core/ (bmap) | planned |
| cmd/mcp/ | 45+ MCP tools, JSON-RPC, HTTP pooling, circuit breaker | internal/mcp/ + internal/tool/ | planned |

---

## Behavior Source Resources (mimicrya/behavior-sources.yaml)

| Source | Behaviors Borrowed | How They Translate to Agent Operations |
|--------|-------------------|---------------------------------------|
| Mayveskii/bun | phase_graph, two_vote_verify, edit_scope_isolation, pre_post_hooks, never_rules, lifetime_classify | Orchestrator phases, quality checks, permission rules, middleware |
| Mayveskii/exa-mcp-server | web_search_tool, web_fetch_tool, mcp_tool_pattern | Agent can search web, fetch URLs, tools follow define→validate→execute pattern |
| Mayveskii/gh-aw-mcpg | mcp_gateway_routed/unified, difc_security, circuit_breaker, oauth_pkce, wasm_guards | MCP transport modes, security pipeline, resilience, auth |
| Mayveskii/opencode-anomalyco- | tool_registry, tool_loop, mcp_transport | Tool registration, execution loop, stdio/SSE/HTTP transport |
| Mayveskii/code-mode | permission_pipeline, denial_tracking, auto_classifier, budget_enforcement, concurrency_control | Permission, circuit breaking, AI classification, budget, parallelism |
| Mayveskii/embryo | binary_runtime, orchestrator_pipeline, mesh_agent_cascade, hunt_system, mesh_graph_slots, hybrid_rag, projectmap_sqlite, survival_index, mapstore_slots_invariants, mcp_tools_registry | Core engine, orchestration, mesh, hunt, RAG, navigation, survival, storage, MCP tools |

### Not Yet Analyzed

- Mayveskii/gastown — unknown behaviors
- Mayveskii/rustnet — unknown behaviors
- Mayveskii/netboot.xyz — unknown behaviors
- Mayveskii/git — unknown behaviors
- Mayveskii/gitingest — unknown behaviors
- Mayveskii/agency-agents — unknown behaviors
- Mayveskii/openmythos — unknown behaviors
- Mayveskii/caveman — unknown behaviors
- Mayveskii/minbpe — unknown behaviors

---

## Distillation Resources (mimicrya/repos-manifest.yaml)

90+ production repositories for pattern extraction. Current status: all pending.

| Category | Count | Top Repos | Pattern Types |
|----------|-------|-----------|--------------|
| current | 16 | etcd, k8s, go-ethereum, envoy, istio | Distributed consensus, controllers, p2p, proxy, mesh |
| new | 4 | autogen, langchain, lyfe | Multi-agent orchestration |
| recommended | 15 | redis, kafka, vault, openssl, postgres | Persistence, streaming, secrets, crypto, SQL |
| agentic | 10 | crewai, openai/swarm, agentops | Agent frameworks, observability |
| git | 10 | git/git, libgit2, go-git, terraform | Git internals, IaC |
| network | 7 | nginx, node, turbo, react | Web server, event loop, build, UI |
| python | 6 | cpython, pip, jupyter, pandas | Runtime, packaging, notebooks, data |
| llm | 5 | transformers, vllm, ollama | Model loading, inference, serving |
| swe | 4 | copilot-cli, sourcegraph | AI coding, code search |
| os | 4 | linux, llvm, gdb | Kernel, compiler, debugging |
| security | 3 | sanitizers, defender | Security scanning, memory safety |
| hardware | 3 | qemu, docker, riscv | Emulation, containers, ISA |

---

## Engine Subsystem Resources

### Workspace Indexing

| Resource | Function | Purpose for Agent Operations |
|----------|----------|------------------------------|
| si_insert | Add entry to slot index | Index workspace files, symbols, deps on startup and after writes |
| si_query_domain | Query by domain | Agent: "show me workspace structure" → si_query_domain("workspace") |
| si_query_domain_layer | Query by domain + layer | Agent: "find function X" → si_query_domain_layer("workspace", "symbols") |
| si_query_state_hash | Query by content hash | Agent: "has this file changed?" → si_query_state_hash(hash) |
| si_build_from_bmap | Build index from bmap | Startup: load workspace index from disk |
| snapshot_diff | Compare current vs indexed | Detect staleness: if diff ≠ empty → re-index |
| drift_detect | Detect drift over time | Monitor: "how much has workspace changed since last session?" |

### Binary RAG

| Resource | Function | Purpose for Agent Operations |
|----------|----------|------------------------------|
| int8_quantize | Convert float vectors to int8 | Compress query vector for fast comparison |
| batch_cosine_int8 | Batch similarity search | Find top-k similar slots in one pass |
| cosine_f32 | Float similarity | Precise similarity for top-k re-ranking |
| si_query_domain | Filter by domain | RAG signal 3: only search in relevant domain |
| inv_find_similar | Find similar invariants | RAG signal 2: keyword/structure match |
| z_density_compute | Compute slot quality | RAG signal 5: prefer high Z-density slots |
| survival(commit) | Commit survival score | RAG signal 4: prefer high-survival patterns |

RAG flow: query → int8_quantize → si_query_domain (filter) → batch_cosine_int8 (rank) → top-k → cosine_f32 (re-rank) → survival + Z-density (boost) → result

### Context Flow

| Phase | Context In | Context Out | Structure |
|-------|-----------|-------------|-----------|
| CLASSIFY | session (budget, denials), workspace index | classified_intent (domain, safety, scenario_hint) | Go struct |
| PLAN | classified_intent, RAG results | planned_chain (OpPacket[]), estimated_cost | Go struct + C OpPacket[] |
| VALIDATE | planned_chain, estimated_cost | validated_chain + ValidationResult | C ValidationResult |
| EXEC | validated_chain, ExecContext | execution_result (per-op latency, result_code) | C ExecContext |
| VERIFY | execution_result, invariant_checks | verify_result (2-vote outcome) | Go struct |
| RESPOND | all previous context | result + metrics + slot_suggestions | MCP response |

Context is cumulative and forward-flowing. Rollback sends failure context backward (VALIDATE fail → PLAN gets failure reason).

### Constant Compression

| Resource | Compression | Decompression | Verification |
|----------|------------|---------------|--------------|
| bmap slot write | OP_COMPRESS_GZIP → bmap_write_cell | bmap_read_cell → OP_DECOMPRESS_GZIP | sha256_hash before decompress |
| session persist | OP_COMPRESS_GZIP → file write | file read → OP_DECOMPRESS_GZIP | sha256_hash on restore |
| matrix persist | OP_COMPRESS_GZIP → file write | file read → OP_DECOMPRESS_GZIP | sha256_hash on load |
| semantics backup | OP_COMPRESS_GZIP → file write | file read → OP_DECOMPRESS_GZIP | sha256_hash on restore |

Compression ratio = original_size / compressed_size. Tracked per resource type for monitoring.

### Multi-task Pipeline Execution

| Resource | Function | Purpose |
|----------|----------|---------|
| conflict_matrix | Cross-pipeline conflict check | Two pipelines with overlapping resource_bitmask → serialize |
| ExecContext.resource_bitmask | Track which resources pipeline uses | Bit per domain: git, build, io, network, system |
| concurrency_control | Max 10 parallel operations | From Mayveskii/code-mode |
| edit_scope_isolation | Each pipeline owns its file set | From Mayveskii/bun (each agent edits only its crate) |
| circuit_breaker | Per-pipeline failure isolation | From Mayveskii/gh-aw-mcpg |
