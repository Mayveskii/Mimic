# MIMIC-MASTER-SPEC.md — Semantic Specification of Desired Behavior

What Mimic IS. What Mimic DOES. Why it works. No invented numbers. No timelines. Pure semantics from sources.

---

## 1. Mimic Is a Model Amplifier

A weak model + Mimic = a strong model. Not because Mimic thinks, but because Mimic carries proven behavior in tokenized processes that any model can follow step-by-step without thinking.

The model expresses intent. Mimic translates intent into a validated chain of action tokens. Each token = one proven operation with measured cost, checked conflicts, and verified result. The model does not need to know HOW to commit, build, search, or deploy — Mimic provides the tokenized process. The model follows.

This is how a laptop becomes a laboratory: the intelligence is in the mesh slots (surviving patterns from production), the conflict matrix (enforced prohibitions), the energy matrix (measured costs). The model is the conductor. Mimic is the orchestra with pre-written scores.

---

## 2. Tokenization of Processes — The Core Mechanism

Every operation Mimic performs is an OpPacket — a token of action. An OpPacket contains:
- opcode: which proven operation to execute
- args: key-value parameters
- flags: safety properties (SAFE, READONLY, ATOMIC, REVERSIBLE, NETWORK, DISK, MEMORY, DANGEROUS)
- buffer: data payload
- timeout_us: maximum execution time
- retry_count: number of retries on failure

A chain of OpPackets = a tokenized process. Instead of the model reasoning about "how to commit safely", it receives the tokenized process: OP_GIT_STATUS → OP_GIT_DIFF → OP_GIT_ADD → OP_GIT_COMMIT. Each token is proven, each transition is validated, each cost is measured.

Tokenized processes come from two sources:
- **Named scenarios**: pre-built chains proven in production (atomic_commit, safe_merge, feature_branch, hotfix, ci_diff_check, build_and_test, safe_deploy, search_and_apply, parallel_build_shards)
- **Custom chains**: the model provides explicit OpPacket sequences; Mimic validates them before execution

Tokenized knowledge comes from mesh slots:
- Each slot is indexed by domain:layer:modality:pattern_name
- A slot carries survival_index (proven by git blame), z_density (knowledge density), polarity (POSITIVE/NEGATIVE/COUNTER), invariants, and extraction_hash
- Any model can retrieve slots through linear lookup (O(1), no intelligence needed), keyword lookup (minimal intelligence), or semantic search (full intelligence with survival-weighted ranking)

---

## 3. How Any Model Uses Mimic

The model calls a Mimic MCP tool with an intent. Mimic executes a 6-phase pipeline:

### CLASSIFY
Determine what the model wants: domain, safety level, task type.
- Named scenario match or custom chain identification
- 3-vote refute classification from bun/PR#30412: three independent classifiers vote on the operation type; majority wins; 12% random sample sent to audit queue
- Zero False Confidence from gastown: derive state from observable reality (process alive, file exists), not cached state files
- Help triage from gastown: 6-category × 3-severity routing (Emergency→Overseer, Failed→Deacon, Blocked→Mayor)
- Pressure gating from gastown: checkPressure() gates dispatch under system load; cleanup always runs even when dispatch deferred

### PLAN
Build the OpPacket chain from scenario template or custom input.
- If mesh slots available for the classified domain: query slots, check preconditions (BEHAVIOR.md #10 mimicry control), incorporate matching patterns into the chain
- RAG pipeline from embryo/pkg/rag: 5-signal hybrid — vector_similarity + keyword_match + domain_filter + survival_index + z_density — weighted combination [0.3, 0.2, 0.2, 0.15, 0.15]
- IDF-weighted search from graphify: three-tier precedence (exact 1000x > prefix 100x > substring 1x) + source-file bonus + gap-ratio seed cutoff
- Hub-throttled traversal from graphify: BFS/DFS with p99 degree hub threshold — hubs skipped as transit unless they are seeds
- Context compression from hermes-agent: preflight token detection + multi-pass compression + stable system prompt prefix for cache reuse (~75% input token savings)
- Memory nudge from hermes-agent: turns_since_memory counter triggers periodic review at configurable interval
- Capability detection from rustnet: eBPF capability check with ordered fallback (CAP_BPF+CAP_PERFMON → CAP_SYS_ADMIN → nosuid detection)
- Platform abstraction from rustnet: DegradationReason enum for actionable diagnostics, not just "failed"

### VALIDATE
Check the planned chain before ANY execution.
- Pairwise conflict check: conflict_matrix[op1][op2] for every pair in the chain
- Energy budget check: sum of cost_tokens × cost_time_us ≤ budget_remaining
- Permission pipeline from code-mode: deny_rules → classify(auto AI) → budget_check → allow_rules
- Never-rules from bun: {git reset, git checkout, git restore, git stash, git rebase, .zig edits, Box::leak, re-gate completed phase} — cannot be overridden by any permission mode; violation = hard stop
- Tool guardrails from hermes-agent: per-turn reset, deny/ask/allow pipeline, OP_FLAG_DANGEROUS always blocked unless explicit allow
- Denial tracking from code-mode: 3 consecutive denies → circuit break to manual mode; 20 total denies per session → permanent manual mode
- Trust-on-first-use from rtk: project-local .rtk/filters.toml trust states (Trusted/Untrusted/ContentChanged/EnvOverride)
- Sensitive path protection from caveman: .env, credentials, keys, secrets, tokens — auto-excluded, never included in LLM context
- Schema validation from gitingest: typed IngestionQuery + CloneConfig — validated data models for all tool I/O
- Auto-clarity boundary from caveman: when compressed context would lose critical clarity markers → stop compression + emit warning
- Validate-then-accept from caveman: LLM output must pass structural validation before use; rejected → retry with clarified constraints (max 3)

### EXEC
Execute the validated chain deterministically.
- ops_execute_chain from embryo/pkg/do: sequential execution of OpPackets via C-core, latency measured per operation with CLOCK_MONOTONIC
- Rollback on failure from gastown: track created resources in order → cleanupOnError reverses in reverse order; partial failure leaves no orphaned resources
- Atomic allocation from gastown: AllocateAndAdd with flock + pending marker prevents TOCTOU races between concurrent agents
- Structured output retry from code-mode: malformed model output → reprompt with schema + error description → retry (max 5) → manual escalation
- Streaming health from hermes-agent: 90s stale-stream detection + 60s read timeout — prevent hangs on zombie provider connections
- Error classification from hermes-agent: classify_api_error → FailoverReason → retry with backoff vs fallback provider vs abort
- Output truncation from opencode: 2000 lines / 50KB hard limit; overflow → temp file; agent reads overflow with offset/limit
- Immutable-after-set from rustnet: first-writer-wins for identity fields, latest-wins for transient state — prevents attribution conflicts
- Per-protocol merge from rustnet: identity→first-wins, dialog→latest-wins, TLS→prefer complete over partial
- DPI detection chain from rustnet: ordered protocol detection with port-gated fast-path → signature fallback, 17 application protocols
- QUIC fragment reassembly from rustnet: 5 strategies with increasing desperation (cache→contiguous→fragment→reconstruction→greedy)
- Connection state machine from rustnet: TCP RFC 793 + QUIC with progressive staleness indicators (white→yellow→red)
- Command rewrite from rtk: static rule table mapping regex→rtk_cmd, stripping sudo/env/git-opts, exclude_commands, transparent_prefixes
- Tee on failure from rtk: filtered command fails → save full unfiltered output to ~/.local/share/rtk/tee/ for debugging
- Discovery learning from rtk: scan session JSONL for unfiltered commands → suggest corrections; detect repeated CLI mistakes → write .claude/rules/cli-corrections.md
- Agent hooks from rtk: init scripts for 13 AI agents (Claude, Copilot, Cursor, Gemini, Codex, Windsurf, Cline, OpenCode, etc.)

### VERIFY
Post-execution verification when scenario requires it.
- 2-vote adversarial verify from bun: two independent verifiers (different model/prompt) vote pass/fail → consensus → accept; disagreement → tiebreak (third verifier or human)
- Critical operations (git push, deploy, encrypt, economic) ALWAYS undergo 2-vote
- Convergence stability check from openmythos: spectral radius validation (eigenvalue check) for iterative convergence guarantee
- Adapted for Mimic: if chain result converges (delta below threshold across verification steps) → accept early; if divergence → rollback

### RESPOND
Return result to model with full traceability.
- Status, artifacts, metrics (latency per op, total tokens, memory peak, energy cost)
- Validation report (which invariants passed/failed/skipped)
- Pattern references (which mesh slots used, with survival_index + z_density)
- Rollback status (if failed, was state fully restored)
- Every decision references its source: commit, SuperInvariant, behavior-source repo

---

## 4. Knowledge From Distillation

Production repositories → git blame → survival index → mesh slots.

Distillation pipeline behavior:
1. Clone repository at pinned commit
2. Run git blame on every source file
3. Compute survival_index = surviving_lines / total_lines_added per commit
4. Extract functions where survival_index ≥ 0.7
5. For every revert commit → create NEGATIVE artifact (polarity=NEGATIVE)
6. For every NEGATIVE → find or create COUNTER artifact (polarity=COUNTER)
7. Encode as protobuf with: source_repo, source_commit, domain, layer, modality, pattern, invariants, survival_index, z_density, polarity, counter_pattern_id, extraction_hash
8. Compress with OP_COMPRESS_GZIP, verify with OP_HASH_SHA256
9. Write to bmap cell, index by domain:layer:modality:pattern_name

Quality gates per artifact (all must pass before storage):
- QAC-1: survival_index computed from git blame
- QAC-2: at least 1 verifiable invariant attached
- QAC-5: z_density computed from actual slot data
- QAC-7: artifact_precision = survival_index × invariant_coverage × extraction_reproducibility; all three > 0
- QAC-9: every NEGATIVE artifact links to a COUNTER
- QAC-10: survival_index re-validated when blame data changes
- QAC-12: extraction_hash present and matches current extractor
- QAC-13: 100% of explicit revert commits become NEGATIVE artifacts

Distillation sources tracked in mimicrya/repos-manifest.yaml (90+ production repos, all status=pending).

---

## 5. Knowledge From Mimicry

Mayveskii/* repositories → behavior selection → implementation in Mimic.

Mimicry is NOT copying. It is:
1. Observe behavior in source repository
2. Identify preconditions that make this behavior valid
3. Identify invariants this behavior maintains
4. Implement our own version that satisfies same preconditions and invariants
5. Verify our version passes same quality gates

21 source repositories with 146 behaviors extracted. Full catalog:

### embryo (10 behaviors)
- binary_runtime: BinaryRuntime with OpCodes 0x10-0x41, tool loop, session, inference pool, hooks → core/ops.c
- orchestrator_pipeline: 7-stage state→mesh→DIRECT→classify→exec→flywheel→respond with rollback → internal/orchestrator/
- mesh_agent_cascade: 3-tier inference (local→medium→top) with cost-aware escalation → internal/orchestrator/cascade.go
- hunt_system: 17-module assess→compress→search→rank by z_density → internal/tool/tools_hunt.go
- mesh_graph_slots: domain/layer/state_hash indexed slots → core/bmap
- hybrid_rag: 5-signal RAG + qdrant → internal/orchestrator/rag.go
- projectmap_sqlite: SQLite FTS5, auto-index on WRITE → internal/orchestrator/projectmap.go
- survival_index: git blame → surviving/total → data/extraction/
- mapstore_slots_invariants: every slot has ≥1 invariant → core/bmap
- mcp_tools: 45+ MCP tools, JSON-RPC, circuit breaker → internal/mcp/ + internal/tool/

### hermes-agent (15 behaviors)
- closed_learning_loop: agent creates skills from experience, self-improves, periodic memory/skill nudge
- context_compression_pipeline: preflight detection + multi-pass compression + stable system prompt prefix
- iteration_budget: IterationBudget.consume() gates loop; budget exhausted → grace call → stop
- tool_guardrails: per-turn reset, deny/ask/allow pipeline, OP_FLAG_DANGEROUS always blocked
- delegation_subagents: spawn isolated subagents with file state registry across agents
- trajectory_compression: record tool-call trajectories + compress for training
- memory_nudge_turns: turn counter triggers memory review at configurable interval
- skill_provenance_tracking: ContextVar tracks skill write origin (background vs foreground)
- streaming_health_check: 90s stale-stream detection + 60s read timeout
- error_classifier_failover: classify_api_error → retry with backoff vs fallback vs abort
- message_sanitization: surrogate repair, non-ASCII strip, tool call argument repair, role alternation fix
- prompt_caching: stable system prompt prefix + cache_control breakpoints (~75% token savings)
- credential_pool: multi-key rotation with env fallback
- multi_transport: Anthropic/OpenAI/Bedrock/Codex/Gemini transport adapters

### gastown (12 behaviors)
- zfc_state: Zero False Confidence — observable reality is source of truth, not state files
- watchdog_chain: 3-tier Daemon(3min)→Boot(triage)→Deacon(patrol)→Witness(per-rig)
- help_classification_routing: 6-category × 3-severity triage with auto-routing
- event_driven_convoy: completion event → feedNextReadyIssue with dependency-aware dispatch
- atomic_allocation: AllocateAndAdd with flock + pending marker → no TOCTOU
- rollback_on_failure: track sequence → cleanupOnError reverses all
- pressure_gating: checkPressure() gates dispatch; cleanup always runs
- lifecycle_state_machine: 6-state (Working/Idle/Done/Stuck/Stalled/Zombie) with cleanup tracking
- inter_agent_protocol: 14 typed messages (POLECAT_DONE, HELP:, MERGED, HANDOFF, etc.)
- session_continuity: seance + prime + nudge queue for recovery
- plugin_system: shell plugins with plugin.md + run.sh
- multi_agent_runtime: ACP abstraction with TMUX/ACP modes per agent type

### bun (7 behaviors)
- phase_graph: CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND with gates
- two_vote_verify: 2 independent verifiers + tiebreak for critical ops
- edit_scope_isolation: resource_bitmask per operation; overlap → serialize
- pre_post_hooks: PreToolUse/PostToolUse middleware chain
- never_rules: hardcoded deny set, no override possible
- lifetime_classify: 3-vote refute + 12% random audit sample
- parallel_agents: ~170 agents with scoped isolation via conflict matrix

### code-mode (6 behaviors)
- permission_pipeline: deny_rules → classify(auto) → budget_check → allow_rules
- denial_tracking: 3 consecutive → circuit break; 20 total → permanent manual
- auto_classifier: AI evaluates risk → safe/ask/deny → confidence threshold for escalation
- budget_enforcement: per-session maxTurns + maxBudgetUsd
- concurrency_control: up to 10 parallel, concurrency-safe vs serial classification
- structured_output_retry: max 5 retries for malformed output → manual escalation

### rustnet (10 behaviors)
- dpi_detection_chain: ordered protocol detection with port-gated fast-path → signature fallback
- quic_fragment_reassembly: 5-strategy data extraction with increasing desperation
- connection_state_machine: TCP RFC 793 + QUIC with progressive staleness (white→yellow→red)
- immutable_after_set: first-writer-wins for identity, latest-wins for transient
- dpi_info_merge: per-protocol merge strategy (identity→first, dialog→latest, TLS→complete over partial)
- structured_filter: Vim/fzf keyword:value pairs AND-combined with regex support
- platform_abstraction_trait: DegradationReason enum for actionable diagnostics
- capability_detection: eBPF capability check with ordered fallback
- concurrent_pipeline: DashMap + crossbeam channels + snapshot isolation
- sandbox_multiplatform: Landlock/Seatbelt/Job Objects

### netboot.xyz (4 behaviors)
- template_driven_generation: Jinja2 templates per OS with consistent schema
- task_pipeline_ansible: ordered pipeline of composable tasks
- multi_target_build_matrix: Legacy BIOS, UEFI, ARM64, RPi4, Secure Boot from single source
- config_driven_registry: all OS URLs/versions/checksums centralized in defaults/main.yml

### exa-mcp-server (5 behaviors)
- web_search_tool: MCP tool: query → structured results
- web_fetch_tool: MCP tool: urls → markdown with metadata
- mcp_tool_pattern: 3-step registration: define(schema) → validate(input) → execute(logic)
- content_extraction: URL content extraction with PDF, YouTube, subpage crawling
- rate_limit_management: API key rotation and rate limit tracking

### git (6 behaviors)
- command_dispatch: cmd_struct array mapping names to function pointers — tool router
- content_addressable_store: SHA-1 hash-addressed object store — tool result cache
- pack_delta_compression: delta compression for efficient serialization
- hook_middleware: pre/post hooks via run-command.c — middleware pattern
- strbuf_dynamic_string: safe dynamic string buffer with add/release/grow — C utility
- perf_tracing: trace2 region enter/leave for performance profiling

### gitingest (6 behaviors)
- ingestion_pipeline: 4-stage parse_query → clone_repo → ingest_query → format_output
- schema_validation: Pydantic typed+validated data models for tool I/O
- file_tree_traversal: walk with gitignore + include/exclude patterns
- async_timeout_bounded: @async_timeout decorator for resource-bounded execution
- query_url_parsing: disambiguate URL patterns, resolve branches/commits/tags
- token_counting_estimation: estimate token count for cost estimation before LLM call

### awesome-mcp-servers (2 behaviors — reference)
- mcp_ecosystem_map: 2245 MCP servers across 50 categories
- mcp_server_categorization: 50-category taxonomy for MCP capabilities

### agency-agents (6 behaviors)
- agent_as_markdown: self-contained .md per agent with identity/mission/capabilities/workflow/metrics
- division_categorization: 15 domain directories organizing 120+ agents
- integration_multiplexer: single source → generate adapters for 10+ AI tools
- strategy_playbooks: phase 0-6 project lifecycle (Discovery→Strategy→Foundation→Build→Hardening→Launch→Operate)
- handoff_templates: inter-agent context passing templates
- workflow_process_pattern: each agent defines step-by-step workflow

### gh-aw-mcpg (6 behaviors)
- mcp_gateway_routed: /mcp/{serverID} → route to specific backend
- mcp_gateway_unified: /mcp → fan-out to all healthy backends
- difc_security: 6-phase DIFC pipeline: label → check → execute → label response → filter
- circuit_breaker: per-backend 3-state (closed→open→half-open) with 30s half-open probe
- oauth_pkce: OAuth 2.0 with PKCE for client authorization
- wasm_guards: WASM sandbox for isolated policy evaluation

### openmythos (6 behaviors)
- recurrent_depth_execution: same block applied N times with state injection → retry/refinement loop
- moe_routing: sparse expert selection via top-K gating → dynamic tool selection
- depth_lora_adaptation: per-loop LoRA adapters that specialize per iteration
- prelude_recurrent_coda: 3-stage: pre-processing → iterative core → post-processing
- convergence_stability_check: spectral radius validation for convergence guarantee
- adaptive_computation_time: ACT threshold for halting → adaptive execution budget

### caveman (6 behaviors)
- intensity_level_config: 4 levels of output compression (low/medium/high/max)
- auto_clarity_boundary: stop compression when critical clarity markers would be lost
- file_type_detection: content + magic bytes, not extension alone
- sensitive_path_protection: .env, credentials, keys → auto-exclude
- validate_then_accept: structural validation before accepting LLM output
- llm_driven_compression: invoke LLM as sub-tool for context compression within budget

### opencode (6 behaviors)
- tool_registry: Tool.Def interface with id, description, zod parameters, execute → ExecuteResult
- tool_loop: while(true) generator: API call → stream → execute tools → check continuation
- mcp_transport: StreamableHTTP + SSE + Stdio transports
- codesearch_tool: semantic code search via Exa MCP endpoint
- output_truncation: 2000 lines / 50KB limit, overflow → temp file with offset/limit reading
- permission_system: per-tool allow/deny rules, session-scoped

### minbpe (4 behaviors)
- progressive_inheritance: Tokenizer→Basic→Regex→GPT4 progressive capability via vtable pattern
- pair_merge_compression: iterative BPE: find most frequent pair → merge into new token
- versioned_parsing_rules: GPT2 vs GPT4 split patterns — versioned/configurable parsing
- special_token_registration: reserved tokens with separate handling

### rtk (9 behaviors)
- toml_filter_pipeline: 8-stage TOML filter DSL with fixed stage order
- stream_filter_trait: line-by-line StreamFilter with feed_line/flush/on_exit callbacks
- language_aware_code_filter: MinimalFilter + AggressiveFilter + smart_truncate
- command_classification_rewrite: regex→rtk_cmd mapping + rewrite_command() stripping
- trust_on_first_use: .rtk/filters.toml trust states per project
- token_tracking_sqlite: SQLite WAL with estimate_tokens(~4 chars/token), 90-day cleanup
- discovery_learning: scan JSONL for unfiltered commands, suggest corrections
- tee_on_failure: save full unfiltered output on filtered command failure
- multi_agent_hooks: init scripts for 13 AI agents

### graphify (10 behaviors)
- ast_extraction_engine: 30+ language tree-sitter extractors, two-pass with EXTRACTED/INFERRED/AMBIGUOUS
- idf_weighted_search: three-tier precedence (exact 1000x > prefix 100x > substring 1x) + gap-ratio cutoff
- hub_throttled_traversal: BFS/DFS with p99 hub threshold, hubs skipped as transit
- subgraph_to_text: render within token_budget, seeds first, sort by degree, truncate with count
- three_pass_dedup: exact → MinHash+JW(92%) → LLM tiebreaker, Union-Find merge
- leiden_community_detection: Leiden with Louvain fallback, seed=42, oversized split, cohesion scoring
- mcp_stdio_server: 10 tools + 6 resources + mtime hot-reload
- ssrf_security_layer: private-IP block, metadata block, DNS rebinding guard, streaming with caps
- graph_diff_analysis: structural diff between snapshots
- incremental_graph_merge: repo_tag:: prefix + append-only merge

### vllm (4 behaviors)
- measured_optimization_decision: every perf change shows measured before/after on real hardware (PR#36)
- paged_memory_management: block-level virtual memory for KV cache
- continuous_batching_scheduler: preemptive scheduling with priority + recompute
- kernel_fusion_via_compile: torch.compile for single-kernel fusion

### go-service-template-rest (10 behaviors)
- strict_server_interface: OpenAPI-first → oapi-codegen → typed request/response per endpoint
- probe_health_interface: Name + Check(ctx) for health/readiness per subsystem
- koanf_layered_config: defaults → YAML → env overrides + unknown-key validation + per-stage budgets
- middleware_stack: Correlation → OTel → SecurityHeaders → AccessLog → BodyLimit → FramingGuard → Recover
- rfc7807_errors: Problem Details JSON (type/title/status/detail/request_id) for structured errors
- repository_pattern_sqlc: Querier interface + sqlc-generated + domain records
- bootstrap_lifecycle: phased startup with budgets (config=10s, probe=15s, telemetry=2s, total=30s)
- drain_mode: atomic.Bool draining flag → /health/ready returns 503 → existing requests complete
- network_policy_enforcement: startup-time egress allow/deny lists
- multi_agent_skills: 30+ skills + 16 subagent TOML definitions

---

## 6. C-Core Runtime — Desired Behavior

The C-core is the deterministic execution engine. It receives validated OpPacket chains via CGO bridge, executes them sequentially, measures latency, handles retries.

Structures (from ops.h, SEMANTICS.md):
- OpPacket: opcode, args[16], arg_count, buffer, buffer_size, flags, timeout_us, retry_count, result_code, latency_ns
- OpCodeDef: opcode, name, cost_tokens, cost_time_us, cost_memory_bytes, safety_level, flags
- ExecContext: open file descriptors, mmap regions, success/error counts, resource_bitmask
- ValidationResult: valid, conflict_count, conflict_pairs[], budget_remaining

Functions (from SEMANTICS.md):
- ops_init/ops_shutdown: single init, single shutdown, no double-init
- ops_register/ops_get_definition: register OpCode definitions, lookup by code
- ops_execute: single OpPacket execution, measures latency_ns via CLOCK_MONOTONIC
- ops_execute_chain: sequential chain execution, validates BEFORE first exec
- ops_validate_chain: O(n²) pairwise conflict check
- ops_check_conflict: pairwise conflict query
- ops_calculate_action: S = Σ(cost_tokens × cost_time_us)
- ops_mmap_alloc/ops_mmap_free/ops_mmap_sync: MAP_PRIVATE|MAP_ANONYMOUS memory management

46 OpCodes organized by domain:
- Memory (5): OP_MMAP_ALLOC/FREE/READ/WRITE/SYNC — all implemented
- Git (11): OP_GIT_INIT/CLONE/FETCH/COMMIT/PUSH/DIFF/STATUS/CHECKOUT/BRANCH/MERGE/REBASE — all implemented
- I/O (5): OP_IO_READ/WRITE/OPEN/CLOSE/SEEK — not implemented
- Build (5): OP_BUILD_COMPILE/LINK/TEST/DEPLOY/CLEAN — not implemented
- Network (6): OP_NET_HTTP_GET/POST, OP_NET_TCP_CONNECT/SEND/RECV/CLOSE — not implemented
- Process (4): OP_PROC_SPAWN/WAIT/KILL/SIGNAL — not implemented
- Utility (6): OP_HASH_SHA256/MD5, OP_COMPRESS/DECOMPRESS_GZIP, OP_ENCRYPT/DECRYPT_AES — not implemented
- System (9): OP_SYS_EXEC/ENV_GET/ENV_SET/FILE_EXISTS/DIR_CREATE/DIR_REMOVE/FILE_COPY/MOVE/DELETE — 2 of 9 implemented

5 named scenarios in C-core:
- atomic_commit: OP_GIT_STATUS → OP_GIT_DIFF → OP_GIT_ADD → OP_GIT_COMMIT
- safe_merge: OP_GIT_FETCH → OP_GIT_DIFF → OP_GIT_MERGE (ff-only)
- feature_branch: OP_GIT_BRANCH → OP_GIT_CHECKOUT
- hotfix: OP_GIT_BRANCH → OP_GIT_COMMIT → OP_GIT_CHECKOUT(target) → OP_GIT_MERGE
- ci_diff_check: OP_GIT_DIFF → validate output

libbmap.a: 39 symbols (bmap_*, si_*, inv_*, gnk_*, snapshot_*, layer_walk, drift_detect, cosine_*, int8_quantize, sha256_hash, z_density_compute) — no .c source files exist.

---

## 7. Conflict Matrix — Desired Behavior

The conflict matrix defines which OpCodes CANNOT coexist in the same chain or in concurrent pipelines.

Documented conflict rules from sources:
1. OP_SYS_EXEC × OP_SYS_EXEC = 1 (parallel shell commands = race condition, from ops.c)
2. DELETE × WRITE = 1 (write after delete = undefined behavior, from validator.go)
3. WRITE × READ without SYNC = 1 (stale read, from validator.go — currently no-op, needs fix)
4. Same-domain operations with overlapping resources = 1 (edit scope isolation, from bun)
5. resource_bitmask overlap between concurrent pipelines → serialize (from bun)
6. goroutine-per-epoch prune × Store/Retrieve under mutex = 1 (from gonka d9c46cae0)
7. buildMessages mutation × buildMessages read = 1 (from binary-mesh c9e83c3, 11 revisions to stabilize)
8. toolloop 429 iteration × budget check = 1 (from binary-mesh 675601e)
9. secrets × source code = 1 (from binary-mesh ff5ff82)
10. concurrent shared state without lock = 1 (SuperInvariant, 1936 members)
11. concurrent access without synchronization = 1 (SuperInvariant, 6 members)

Each conflict rule references: op1, op2, conflict_level, source_commit or SuperInvariant, condition.

Desired behavior: conflict_matrix[op1][op2] = 1 → these ops NEVER execute together in the same chain. In concurrent pipelines: resource_bitmask overlap → serialize. No override. No exception.

---

## 8. Energy Cost Matrix — Desired Behavior

Every OpCode has a three-dimensional cost: cost_tokens, cost_time_us, cost_memory_bytes.

Measured values (from ops_register_builtins in ops.c):
| OpCode | cost_tokens | cost_time_us | cost_memory_bytes |
|--------|------------|-------------|-------------------|
| OP_NOP | 0.0 | 0.01 | 0.0 |
| OP_SYS_FILE_EXISTS | 1.0 | 10.0 | 0.0 |
| OP_SYS_DIR_CREATE | 2.0 | 50.0 | 4096.0 |

Remaining 43 OpCodes: use estimated defaults in Go energy_cost_matrix.go until measured.

Desired behavior: before EXEC, sum of cost_tokens × cost_time_us for all operations in chain must be ≤ budget_remaining. If actual measured cost during execution exceeds 1.2× estimated → abort chain, rollback, record overestimate.

---

## 9. Anti-Patterns — Desired Behavior

30 anti-patterns documented (AP-01 through AP-30). Each = NEGATIVE artifact in mesh. Every NEGATIVE links to COUNTER artifact.

5 anti-pattern domains:
1. resource_cleanup_under_lock (AP-01, AP-02, AP-15) — from gonka, binary-mesh
2. context_structure_stability (AP-05, AP-24, AP-23) — from binary-mesh (11 buildMessages revisions)
3. input_validation_before_io (AP-04, AP-09, AP-10, AP-21) — from gonka, bitcoin, envoy
4. idempotent_close_cleanup (AP-03, AP-22) — from gonka
5. economic_invariant_enforcement (AP-09, AP-10, AP-21, AP-28) — from gonka, bitcoin

Meta-invariant uniting all 30: no_side_effect_without_prior_validation. Every historical failure traces to a violation. Implementation: ops_validate_chain called BEFORE ops_execute_chain.

---

## 10. Measured Data — Constants From Production

From binary-mesh enricher:
- API calls: 72603
- Cycles: 1765
- PRs scanned: 33755
- Slots created: 32949
- Efficiency: 45.4%
- Throughput: 18.67 slots/cycle
- PR yield: 97.6%

From binary-mesh RAG:
- AST symbols: 38829
- KW terms: 117230
- MD entries: 6367
- Comments: 4979
- Qdrant points: 180684
- Signal ratio: 92.4%

From binary-mesh quality matrix:
- L0 Stability: 8000/10000 (5585 samples)
- L4 Usefulness: 2100/10000 (23 samples — under-sampled)
- L6 Reuse: 1931/10000 (428 samples)
- L8 Latency: 8005/10000 (4319 samples)
- L9 Completion: 7072/10000 (50 samples — under-sampled)
- Composite: 4784/10000

From historical reverts:
- binary-mesh: 4 explicit (buildMessages, shell exploration flip-flop, max_tokens, workspace snapshot)
- gonka: 1 (regenerate seed)

From binary-mesh context:
- Estimated tokens: 19635 / 32768 max = 59.9% pressure
- Messages: 34
- Tokens per message: 577.5

Top SuperInvariants (from 148121 slots across 8 domains):
- "incorrect logic must be replaced to satisfy postcondition" — 57724 members
- "nullable reference must be guarded before dereference" — 9409 members
- "new functionality requires explicit dependency import" — 12146 members
- "behavioral change requires test coverage" — 5910 members
- "missing precondition check must be added before operation" — 2491 members
- "user input requires validation before processing" — 2231 members
- "concurrent shared state requires lock before access" — 1936 members
- "error condition must be surfaced not swallowed" — 1964 members
- "fallible external call requires retry with backoff" — 1943 members
- "arithmetic operation requires overflow guard" — 1140 members
- "collection access requires length verification" — 1090 members
- "operation boundary requires observability signal" — 819 members
- "long-running operation requires cancellation boundary" — 590 members
- "concurrent access requires synchronization" — 6 members

---

## 11. Data Integrity — Desired Behavior

All data at rest is compressed and verified:
- Every bmap slot write: OP_COMPRESS_GZIP before storage
- Every bmap slot read: OP_HASH_SHA256 verification BEFORE decompression
- Every session snapshot: OP_COMPRESS_GZIP at rest, OP_HASH_SHA256 on restore
- Every conflict/energy matrix: OP_COMPRESS_GZIP on disk, OP_HASH_SHA256 on load

Compression ratio tracked. Alert if ratio drops below 1.5 (data may be corrupted or already compressed).

Desired behavior: no uncompressed data in bmap. No unverified data from bmap. Ever.

---

## 12. Session Management — Desired Behavior

Every agent conversation creates a session. Session carries:
- Session ID, budget remaining (tokens + time), operations executed
- Denial history (3 consecutive → circuit break; 20 total → permanent manual mode)
- Context flow (cumulative, forward-only, except rollback signals)
- Compression at rest, hash verified on restore

---

## 13. Workspace Indexing — Desired Behavior

Mimic maintains a local index of the agent's workspace:
- tree: file/directory structure
- symbols: function/type/variable definitions
- deps: dependency graph
- git_state: branch, diff, stash

Index update triggers:
- WRITE operations → mark stale, re-index on next query
- BUILD_COMPILE → parse output for symbols/deps
- GIT_COMMIT/CHECKOUT/MERGE → update git_state

From embryo/pkg/projectmap: SQLite FTS5, updated after every WRITE opcode.
From graphify: AST extraction for 30+ languages, IDF-weighted search, hub-throttled traversal.

Desired behavior: index always current or explicitly stale. No query on stale index without re-index.

---

## 14. Multi-Pipeline Execution — Desired Behavior

Multiple independent intents execute concurrently:
- Cross-pipeline conflict check: resource_bitmask overlap → serialize
- No overlap → parallel (up to 10 concurrent pipelines from code-mode)
- Each pipeline follows full 6-phase sequence independently
- Shared resources (git index, build cache) → explicit locks
- Pipeline failure → that pipeline rolls back independently

From bun: ~170 parallel agents with scoped isolation via conflict matrix.
From gastown: event-driven convoy — completion triggers dependency-aware dispatch.
From rustnet: concurrent pipeline with DashMap + crossbeam + snapshot isolation.

---

## 15. DIFC Security — Desired Behavior

When Mimic serves multiple agents or connects to mimic-server:
6-phase DIFC from gh-aw-mcpg:
1. Label agent: determine clearance level
2. Label resource: determine classification level
3. Check: clearance ≥ classification?
4. Execute: if pass
5. Label response: determine output classification
6. Filter output: remove what agent must not see

Desired behavior: information flows only from ≥ clearance to ≤ clearance. Never the reverse.

---

## 16. Desired Outcome

When Mimic is built:

An agent says "commit these files safely" → Mimic provides a tokenized process: OP_GIT_STATUS → OP_GIT_DIFF → OP_GIT_ADD → OP_GIT_COMMIT. Each token is proven, each transition validated, each cost measured. The model follows the score; the model does not compose.

An agent says "build and test" → Mimic provides: OP_BUILD_COMPILE → OP_BUILD_TEST. Compilation must succeed before test runs. Costs measured. Index updated.

An agent says "find patterns for distributed consensus" → Mimic queries mesh slots from etcd/cockroach/k8s with survival indices and z_density. The model reads the ranked patterns; the model does not guess.

An agent says "apply this pattern to my codebase" → Mimic validates preconditions (BEHAVIOR.md #10), executes with 2-vote verify, rolls back if test fails. The model trusts the validation; the model does not improvise.

An agent says "deploy to production" → Mimic builds, tests, tags, pushes, deploys, health-checks — with 2-vote on every critical step. The model trusts the process; the model does not skip steps.

Every result is deterministic. Every cost is measured. Every pattern is proven. Every failure is surfaced.
