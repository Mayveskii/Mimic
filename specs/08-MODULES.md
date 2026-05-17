# MODULES.md — Mimic

Per-module documentation: what each module does, what resources it contains, how it connects to others.

---

## core/ — C-Core Engine

### What it does
Deterministic execution engine. Receives validated OpPacket chains, executes them sequentially, measures latency, handles retries.

### Resources
- ops.c/ops.h: core engine (init, register, validate, execute)
- git_ops.c: OP_GIT_* executors
- git_scenarios.c: scenario chains (atomic_commit, safe_merge, feature_branch, hotfix, ci_diff_check)
- git_search_ops.c: ⚠️ broken — opcode collision 0x50-0x5F with NET
- mmap_ops.c: OP_MMAP_* executors
- libbmap.a: 39 symbols for mesh storage (no .c sources — needs rewrite)

### Connections
- Called by: internal/cgo/ (via CGO)
- Calls: nothing external (self-contained)
- Data: reads/writes bmap files, git repos, filesystem

### Current state
- 27 of 46 OpCodes have executors
- 5 scenarios implemented
- libbmap.a has no source code
- git_search_ops.c has namespace collision

---

## internal/mcp/ — MCP Server

### What it does
MCP protocol server. Handles JSON-RPC over stdio/SSE/HTTP. Routes tool calls to orchestrator.

### Resources (planned)
- transport.go: stdio, SSE, StreamableHTTP transports (from Mayveskii/opencode-anomalyco-)
- server.go: JSON-RPC handler, connection management
- middleware.go: PreToolUse/PostToolUse hooks (from Mayveskii/bun)
- circuit_breaker.go: per-tool circuit breaker (from Mayveskii/gh-aw-mcpg)

### Connections
- Called by: external MCP clients (any AI-agent)
- Calls: internal/tool/ (tool registry), internal/orchestrator/ (execution)
- Data: session state, tool call logs

### Current state
- Not implemented
- Behaviors identified from: opencode-anomalyco- (transport), bun (hooks), gh-aw-mcpg (circuit breaker)

---

## internal/tool/ — MCP Tool Registry

### What it does
Defines MCP tools that agents can call. Each tool has: name, description, parameters, execute function.

### Resources (planned)
- registry.go: tool registration and lookup (from Mayveskii/opencode-anomalyco- Tool.Def)
- tools_exec_chain.go: execute a scenario or custom chain
- tools_validate.go: dry-run validation
- tools_query_slots.go: query mesh slots (when available)
- tools_status.go: engine state, budget, stats
- tools_hunt.go: find patterns for a domain (from Mayveskii/embryo pkg/hunt/)

### Connections
- Called by: internal/mcp/ (routes tool calls here)
- Calls: internal/orchestrator/ (to execute), internal/cgo/ (to reach C-core)
- Data: tool definitions, execution results

### Current state
- Not implemented
- Behaviors identified from: opencode-anomalyco- (Tool.Def), exa-mcp-server (tool pattern), embryo (hunt system)

---

## internal/cgo/ — CGO Bridge

### What it does
Translates Go types to C types and back. Thin wrapper around C-core functions.

### Resources (exists in c-core/cbridge/, needs migration)
- cgo_wrapper.go: Init, Execute, ExecuteChain, ValidateChain, CheckConflict, MMap*
- validator.go: Go-side pre-flight validation (rules, conflict rules)
- helpers.go: builder pattern for OpPacket construction

### Connections
- Called by: internal/tool/ and internal/orchestrator/
- Calls: core/ (C functions via CGO)
- Data: OpPacket conversion, ValidationResult

### Current state
- Exists in binary-mesh c-core/cbridge/
- Needs migration to internal/cgo/
- Known bugs: memory leak in ExecuteChain buffer handling, race condition on gMutex

---

## internal/orchestrator/ — Workflow Engine

### What it does
Phase graph: CLASSIFY → PLAN → VALIDATE → EXEC → VERIFY → RESPOND. Builds OpPacket chains from agent intent, validates them, routes to execution. Manages context flow between phases.

### Resources (planned)
- classify.go: determine task type and domain (from Mayveskii/bun lifetime_classify)
- planner.go: build OpPacket chain from scenario template or custom input
- validator.go: coordinate Go-side + C-side validation
- permission.go: deny → classify → budget → allow pipeline (from Mayveskii/code-mode)
- executor.go: route to CGO ExecuteChain
- verifier.go: 2-vote adversarial verify (from Mayveskii/bun)
- budget.go: token/time tracking and enforcement (from Mayveskii/code-mode)
- concurrency.go: up to 10 parallel operations (from Mayveskii/code-mode)
- context_flow.go: context passing between phases (cumulative, forward-flowing)
- rag.go: binary RAG — int8_quantize → batch_cosine_int8 → top-k → survival + Z-density boost (from Mayveskii/embryo pkg/rag/)
- pipeline.go: multi-task pipeline orchestration with conflict check and isolation
- indexer.go: workspace indexing — si_insert on startup + after writes, staleness detection via snapshot_diff

### Connections
- Called by: internal/tool/ (on tool call)
- Calls: internal/cgo/ (execute chains), internal/quality/ (verify)
- Data: phase state, chain templates, budget counters, context flow, RAG results

### Current state
- Not implemented
- Behaviors identified from: bun (phase graph, 2-vote, hooks), code-mode (permission, budget, concurrency), embryo (pipeline, RAG, projectmap, hunt)

---

## internal/session/ — Agent Sessions

### What it does
Track agent session state: budget remaining, operations executed, denial history, context flow.

### Resources (planned)
- session.go: session lifecycle (create, track, expire)
- budget_tracker.go: per-session budget accounting
- denial_tracker.go: 3 consecutive → circuit break (from Mayveskii/code-mode)
- context_store.go: cumulative context per session (classified_intent → planned_chain → validated_chain → execution_result → verify_result → response)
- compress.go: session state compressed at rest via OP_COMPRESS_GZIP, verified via sha256_hash on restore

### Connections
- Called by: internal/orchestrator/ (check budget, record denials, pass context)
- Calls: nothing
- Data: session state, budget, denial history, context flow, compressed snapshots

### Current state
- Not implemented
- Behaviors identified from: code-mode (denial tracking, budget enforcement), embryo (session from pkg/do/BinaryRuntime)

---

## internal/quality/ — Verification

### What it does
Post-execution verification: 2-vote adversarial verify, drift detection, invariant checking, compression integrity.

### Resources (planned)
- verify.go: 2-vote mechanism with tiebreak (from Mayveskii/bun)
- invariant_check.go: verify slot invariants in current context (from libbmap.a inv_*)
- drift_detect.go: compare mesh state vs snapshot (from libbmap.a drift_detect)
- compress_verify.go: sha256_hash check on every decompressed slot read (from OP_HASH_SHA256 + OP_DECOMPRESS_GZIP)

### Connections
- Called by: internal/orchestrator/ (VERIFY phase)
- Calls: internal/cgo/ (inv_find_similar, drift_detect via bmap)
- Data: verify results, invariant matches, drift reports, compression integrity checks

### Current state
- Not implemented
- Behaviors identified from: bun (2-vote), libbmap.a (invariants, drift), ops.h (OP_HASH_SHA256, OP_COMPRESS_GZIP)

---

## mimicrya/ — Knowledge Manifests

### What it does
Two YAML files that track the two knowledge sources: behavior sources (mimicry) and production repos (distillation).

### Resources
- behavior-sources.yaml: 16 Mayveskii/* repos with extracted behaviors and status
- repos-manifest.yaml: 90+ production repos with distillation status and Z-density

### Connections
- Read by: internal/orchestrator/ (classify: what behaviors are available)
- Read by: data/extraction/ (distillation: which repos to process)
- Updated by: manual (behavior sources), distill workflow (repos)

### Current state
- behavior-sources.yaml: 6 repos partially analyzed, 10 pending
- repos-manifest.yaml: all repos status=pending

---

## data/ — Extraction and Storage

### What it does
Scripts for distillation, seed data, pre-built matrices, workspace index storage.

### Resources (planned)
- extraction/: distillation scripts (git clone → blame → survival → extract → encode → slot)
- seeds/: initial mesh slots from first distillation run
- matrices/: pre-built conflict_matrix and energy_cost_matrix files
- indexes/: workspace index bmap files (tree, symbols, deps, git_state layers)
- rag/: pre-computed int8 vectors for binary RAG over slots

### Connections
- Called by: make distill (CI or manual)
- Writes to: mimicrya/repos-manifest.yaml (update status/slots/z_density)
- Reads: mimicrya/repos-manifest.yaml (which repos to process)
- Reads/Writes: data/indexes/ (workspace index), data/rag/ (vectors)

### Current state
- Not implemented
- Pattern from: embryo (pkg/survival/ git blame pipeline), embryo (pkg/projectmap/ SQLite navigation), embryo (pkg/rag/ hybrid RAG)

---

## cmd/mimic/ — Entrypoint

### What it does
Main entry point. Parses args, initializes C-core, starts MCP server.

### Resources (planned)
- main.go: init → serve → shutdown

### Connections
- Calls: internal/mcp/ (start server), internal/cgo/ (Init C-core)

### Current state
- Not implemented
