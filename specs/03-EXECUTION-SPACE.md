# EXECUTION-SPACE.md — Mimic

The complete space of what agents can do through Mimic's engine. Every operation, every constraint, every dimension.

---

## Execution Space Dimensions

The agent task execution space is defined by eight dimensions:

### 1. Operation Space (what can be done)

46 OpCodes organized by domain:

| Domain | OpCodes | Agent Capabilities |
|--------|---------|-------------------|
| Memory | OP_MMAP_ALLOC/FREE/READ/WRITE/SYNC | Allocate, read, write, sync memory regions |
| I/O | OP_IO_READ/WRITE/OPEN/CLOSE/SEEK | File operations via io_uring |
| Git | OP_GIT_INIT/CLONE/FETCH/COMMIT/PUSH/DIFF/STATUS/CHECKOUT/BRANCH/MERGE/REBASE | Full git workflow |
| Build | OP_BUILD_COMPILE/LINK/TEST/DEPLOY/CLEAN | Compile, link, test, deploy, clean |
| Network | OP_NET_HTTP_GET/POST, OP_NET_TCP_CONNECT/SEND/RECV/CLOSE | HTTP requests, TCP connections |
| Process | OP_PROC_SPAWN/WAIT/KILL/SIGNAL | Spawn and manage processes |
| Utility | OP_HASH_SHA256/MD5, OP_COMPRESS/DECOMPRESS_GZIP, OP_ENCRYPT/DECRYPT_AES | Hash, compress, encrypt |
| System | OP_SYS_EXEC/ENV_GET/SET/FILE_EXISTS/DIR_CREATE/DIR_REMOVE/FILE_COPY/MOVE/DELETE | Shell exec, env, file operations |

**Currently implemented**: OP_MMAP_*, OP_GIT_*, OP_SYS_FILE_EXISTS, OP_SYS_DIR_CREATE (27 of 46)
**Not yet implemented**: OP_IO_*, OP_BUILD_*, OP_NET_*, OP_PROC_*, OP_HASH_*, OP_COMPRESS_*, OP_ENCRYPT_*, OP_SYS_EXEC/ENV_*/DIR_REMOVE/FILE_COPY/MOVE/DELETE (19 of 46)

### 2. Composition Space (how operations combine)

Operations combine into chains. Not all combinations are valid.

**Valid composition**: conflict_matrix[op1][op2] = 0
**Invalid composition**: conflict_matrix[op1][op2] = 1

Current conflict rules:
- OP_SYS_EXEC × OP_SYS_EXEC = 1 (parallel shell commands = race)
- DELETE × WRITE = 1 (write after delete = undefined behavior)
- WRITE × READ without SYNC = 1 (stale read, currently no-op — needs fix)

Maximum chain length: 1024 operations (MaxChainLength in validator.go)
Maximum total buffer: 10MB (MaxTotalBufferSize in validator.go)
Maximum parallel operations: 10 (from Mayveskii/code-mode concurrency_control)

### 3. Cost Space (what each operation costs)

Every operation has a three-dimensional cost:

| Domain | Typical cost_tokens | Typical cost_time_us | Typical cost_memory_bytes |
|--------|--------------------|-----------------------|--------------------------|
| Memory | 1-3 | 5-50 | 4096-1MB |
| Git (read) | 3-5 | 100-1000 | 0 |
| Git (write) | 5-10 | 1000-10000 | 0 |
| Build | 10-20 | 10000-60000+ | 1MB-100MB |
| Network | 2-5 | 5000-50000 | 0-1MB |
| System | 1-2 | 10-100 | 0-4096 |

Budget constraint: Σ cost_tokensᵢ ≤ budget_tokens (set by agent or config)

### 4. Safety Space (what is allowed)

Every operation has a safety_level (0-3) and flags:

| Safety Level | Meaning | Permission Required |
|-------------|---------|-------------------|
| 0 | Critical/destructive | Explicit allow + 2-vote verify |
| 1 | Potentially dangerous | Auto-classifier or explicit allow |
| 2 | Safe with side effects | Auto-allow if budget ok |
| 3 | Read-only/idempotent | Always allowed |

| Flag | Meaning |
|------|---------|
| OP_FLAG_SAFE (0x01) | Operation is safe |
| OP_FLAG_READONLY (0x02) | No side effects |
| OP_FLAG_ATOMIC (0x04) | All-or-nothing execution |
| OP_FLAG_REVERSIBLE (0x08) | Can be undone |
| OP_FLAG_NETWORK (0x10) | Requires network |
| OP_FLAG_DISK (0x20) | Touches disk |
| OP_FLAG_MEMORY (0x40) | Uses memory |
| OP_FLAG_DANGEROUS (0x80) | Always requires explicit allow |

Permission pipeline (from Mayveskii/code-mode):
```
deny_rules → classify(auto AI) → budget_check → allow_rules
```

Denial tracking: 3 consecutive denies → circuit break → manual mode

### 5. Knowledge Space (what patterns are available)

Two sources feed the knowledge space:

**Distillation** (repos-manifest.yaml): 90+ production repos → git blame → survival index → mesh slots
- Current status: all pending (no distillation run yet)
- When available: agent can query slots by domain, layer, state hash
- Access: query_slots MCP tool → si_query_domain/si_query_domain_layer

**Mimicry** (behavior-sources.yaml): 16 Mayveskii/* repos → behavior selection
- Current status: 6 repos partially analyzed (bun, exa-mcp-server, gh-aw-mcpg, opencode, code-mode, embryo)
- 10 repos not yet analyzed
- When available: orchestrator uses borrowed behaviors to build chains

### 6. Indexing Space (workspace context)

Mimic indexes the agent's workspace for fast context retrieval:

| Index Layer | Content | Update Trigger | Query Method |
|------------|---------|----------------|--------------|
| tree | File/directory structure | OP_SYS_FILE_DELETE, OP_SYS_DIR_CREATE/REMOVE | si_query_domain_layer("workspace", "tree") |
| symbols | Function/type/variable definitions | OP_BUILD_COMPILE (parse output) | si_query_domain_layer("workspace", "symbols") |
| deps | Dependency graph | OP_BUILD_COMPILE, OP_IO_READ (package files) | si_query_domain_layer("workspace", "deps") |
| git_state | Branch, diff, stash | OP_GIT_COMMIT, OP_GIT_CHECKOUT, OP_GIT_MERGE | si_query_domain_layer("git", "state") |

**Invariant**: index is stale after any WRITE without re-index. Staleness detected by snapshot_diff (compare current vs indexed state).

### 7. Compression Space (data at rest)

All mesh slot data is compressed at rest:

| Resource | Compression | Verification | Ratio Tracking |
|----------|------------|--------------|----------------|
| bmap slots | OP_COMPRESS_GZIP on write | sha256_hash check on read | compression_ratio = original / compressed |
| session state | OP_COMPRESS_GZIP on persist | sha256_hash on restore | per-session |
| conflict/energy matrices | OP_COMPRESS_GZIP on disk | sha256_hash on load | on matrix update |

**Invariant**: no uncompressed data in bmap. Every read verifies hash before decompress.

### 8. Pipeline Space (multi-task execution)

Multiple independent pipelines can execute concurrently:

```
Pipeline isolation rules:
- Each pipeline owns its resource scope (file set, domain)
- Shared resources (git index, build cache) → serialized access
- Conflict check: resource_bitmask overlap between pipelines → serialize
- No overlap → parallel execution (up to 10 concurrent)
```

| Pipeline Type | Isolation Level | Serialization Points |
|--------------|----------------|---------------------|
| Read-only queries | Full parallel | None |
| Git operations (same repo) | Serialized | git index lock |
| Build operations (different targets) | Parallel per target | Link step (shared output) |
| Build operations (same target) | Serialized | Compile output |
| Network requests | Full parallel | Rate limits |
| File writes (different files) | Parallel | None |
| File writes (same file) | Serialized | File lock |

---

## Execution Flow Through the Space

```
Agent calls Mimic MCP tool
    ↓
CLASSIFY phase
    - What domain? (git, build, io, network, system, mixed)
    - What safety level required?
    - Is a slot available for this domain? (query si_query_domain)
    - Context IN: session budget_remaining, denial_count, workspace index
    ↓
PLAN phase
    - Build OpPacket chain from scenario template OR custom chain
    - If slot available: binary_rag(query, domain) → top-k patterns → incorporate
    - Apply borrowed behaviors (phase graph, permission, hooks)
    - Context IN: classified_intent + workspace index + RAG results
    ↓
VALIDATE phase
    - Check conflict_matrix for all pairs → no conflicts
    - Sum energy costs → within budget
    - Check safety levels → permissions obtained
    - Check invariants (from scenario definition)
    - Context IN: planned_chain + estimated_cost
    ↓
EXEC phase
    - ops_execute_chain via CGO
    - Each OpPacket executed sequentially
    - Latency measured per operation
    - Retry if retry_count > 0
    - On failure: rollback per scenario definition
    - Context IN: validated_chain + ExecContext (open_fds, mmap_regions)
    ↓
VERIFY phase (if required)
    - 2-vote adversarial verify
    - executor_A and executor_B independently check result
    - If consensus: pass
    - If disagreement: tiebreak or manual escalation
    - Context IN: execution_result + invariant_checks
    ↓
RESPOND phase
    - Result + metrics to agent
    - Update workspace index if WRITE operations occurred
    - Compress and store result in mesh if pattern is new
    - Context OUT: result + metrics + slot_suggestions
    - Agent decides what to do with result
```

---

## Task Types and Their Execution Patterns

### Type 1: Read-only queries

```
Example: "show me the diff" / "what's the git status?"

Chain: single OP_GIT_DIFF or OP_GIT_STATUS
Safety: READONLY, safety_level=3
Cost: low (1-5 tokens, <1ms)
2-vote: no
Rollback: N/A
```

### Type 2: Safe mutations

```
Example: "create a directory" / "add a file"

Chain: OP_SYS_DIR_CREATE or OP_SYS_FILE_COPY
Safety: ATOMIC, safety_level=2
Cost: low (2-5 tokens, <100ms)
2-vote: no
Rollback: reverse operation (DIR_REMOVE, FILE_DELETE)
```

### Type 3: Git workflows

```
Example: "commit safely" / "create feature branch" / "hotfix"

Chain: scenario template (atomic_commit, feature_branch, hotfix)
Safety: ATOMIC + DISK, safety_level=1-2
Cost: medium (4-16 tokens, 5-30ms)
2-vote: yes for merges/hotfixes, no for standard commits
Rollback: checkout to previous state
```

### Type 4: Build pipelines

```
Example: "build and test" / "compile this target"

Chain: OP_BUILD_COMPILE → OP_BUILD_TEST
Safety: ATOMIC + DISK + MEMORY, safety_level=1-2
Cost: high (15-30 tokens, 10-120s)
2-vote: no
Rollback: OP_BUILD_CLEAN
```

### Type 5: Deployments

```
Example: "build, test, tag, push"

Chain: safe_deploy scenario
Safety: DANGEROUS + NETWORK + DISK, safety_level=0
Cost: very high (25+ tokens, 30s+)
2-vote: yes (deploy is irreversible)
Rollback: delete tag, revert commit
```

### Type 6: Pattern application

```
Example: "find a pattern for X and apply it"

Chain: OP_MMAP_READ (slot) → inv_find_similar (check) → OP_BUILD_COMPILE → OP_BUILD_TEST
Safety: depends on pattern, safety_level=1-2
Cost: high (20+ tokens, 30s+)
2-vote: yes (applying external pattern to codebase)
Rollback: revert to previous state if test fails
```

### Type 7: Parallel operations

```
Example: "build these 3 crates in parallel"

Chain: [OP_BUILD_COMPILE(a) || OP_BUILD_COMPILE(b) || OP_BUILD_COMPILE(c)] → OP_BUILD_LINK
Safety: ATOMIC per shard, conflict_matrix must show no cross-dependencies
Cost: parallel = max(shard_costs), not sum
2-vote: no
Rollback: clean all shards if any fails
```

### Type 8: Context retrieval (RAG)

```
Example: "how does etcd handle RAFT apply?" / "find patterns for consensus"

Chain: binary_rag(query, domain) → int8_quantize → batch_cosine_int8 → top-k slots
Safety: READONLY, safety_level=3
Cost: low (2-5 tokens, <10ms for indexed data)
2-vote: no
Rollback: N/A (read-only)
Output: ranked slot list with survival index + Z-density per slot
```

### Type 9: Multi-pipeline orchestration

```
Example: "build backend + build frontend + run integration tests"

Chain: [pipeline_A: BUILD_COMPILE+TEST(backend)] || [pipeline_B: BUILD_COMPILE(frontend)] → [pipeline_C: BUILD_TEST(integration)]
Safety: ATOMIC per pipeline, cross-pipeline conflict check
Cost: parallel phase = max(A, B), then C
2-vote: yes on integration test phase
Rollback: clean all pipelines if integration fails
```

### Type 10: Workspace self-build

```
Example: "build Mimic's own specs from workspace analysis"

Chain:
  [0] OP_IO_READ     args: {path: "core/*.c"}        → read source files
  [1] OP_HASH_SHA256  args: {data: source}            → hash for change detection
  [2] Internal: parse functions → compare with SEMANTICS.md
  [3] OP_IO_WRITE     args: {path: "SEMANTICS.md"}    → update if changed
  [4] OP_COMPRESS_GZIP args: {data: old_semantics}    → compress previous version

Safety: DISK + MEMORY, safety_level=1
Cost: medium (10-20 tokens, 1-5s)
2-vote: no
Rollback: restore from compressed previous version
```

---

## How Mimic Runs Locally

```
Agent process (opencode, cursor, any MCP client)
    ↓ stdio or http://localhost:PORT
Mimic process (single binary, started by agent or manually)
    ├── MCP Layer: receives tool calls
    ├── Orchestrator: classifies, plans, validates
    ├── CGO Bridge: translates to C
    └── C-Core: executes chains deterministically

No cloud. No external services. No network required (unless agent operation itself needs network).
All state is local: mesh slots in bmap files, matrices in data/matrices/.
Agent is fully autonomous — Mimic is an optional tool it calls when useful.
```
