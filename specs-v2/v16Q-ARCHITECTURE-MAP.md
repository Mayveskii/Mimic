# v16Q Architecture Map — Mimicry Implementation Blueprint

**Status:** Proposed  
**Branch:** v16Q  
**Scope:** Translate v16Q specs (ARCHITECTURE, BEHAVIOR, GAP, MODELS, PLAYGROUND, ROADMAP, SEMANTICS) into concrete file structure, interfaces, and implementation waves.  
**Principle:** Mimicry, not copying. Every behavior source contributes principles; implementation is native to Mimic's C-core + Go runtime.

---

## 1. Executive Summary

v16Q transforms Mimic from a passive MCP tool into an **active patch machine**: model enters with real repo → Hunt finds patterns → Orchestrator validates → Model executes in isolated worktree → Proof collected → Skill extracted to mesh.

This document maps every v16Q spec to concrete files, interfaces, and implementation order.

---

## 2. Runtime Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              GonkaGate                                   │
│    qwen3-235b        kimi-k2.6         minimax-m2.7                    │
│         │                 │                  │                          │
│    ┌────▼────┐       ┌────▼────┐       ┌────▼────┐                     │
│    │ Session │       │ Session │       │ Session │  ← per-model,       │
│    │ Manager │       │ Manager │       │ Manager │    isolated         │
│    └────┬────┘       └────┬────┘       └────┬────┘                     │
│         │                 │                  │                          │
│         └─────────────────┼──────────────────┘                          │
│                           │                                             │
│              ┌────────────▼────────────┐                                │
│              │     SessionManager      │                                │
│              │  (Go: internal/session/)│                                │
│              └────────────┬────────────┘                                │
│                           │                                             │
│    ┌──────────────────────┼──────────────────────┐                     │
│    │                      │                      │                      │
│ ┌──▼───┐            ┌────▼────┐          ┌─────▼─────┐                │
│ │ Work │            │ Project │          │   Mesh    │                │
│ │tree  │            │  Map    │          │  Preload  │                │
│ │ Pool  │            │(SQLite) │          │ (SQLite)  │                │
│ └──┬───┘            └────┬────┘          └─────┬─────┘                │
│    │                     │                     │                       │
│ ┌──▼─────────────────────▼─────────────────────▼─────┐                 │
│ │                C-Core (core/)                      │                 │
│ │  exec_sys_file_write  exec_git_add/commit         │                 │
│ │  ops_execute_chain    conflict_matrix             │                 │
│ │  energy_matrix        rollback_state_machine      │                 │
│ └────────────────────────────────────────────────────┘                 │
│                           ▲                                             │
│              ┌────────────┴────────────┐                                │
│              │      CGO Bridge         │                                │
│              │   (internal/cgo/)       │                                │
│              └─────────────────────────┘                                │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.1 Component Table

| Component | Package / File | Behavior Source | What We Mimic | What We Do NOT Copy |
|-----------|---------------|-----------------|---------------|---------------------|
| **SessionManager** | `internal/session/manager.go` | embryo BinaryRuntime | Session lifecycle, budget locking, base_repo wiring | Go `pkg/do/` structs — reimplemented in C-core |
| **Hunt System** | `internal/hunt/` | embryo Hunt | Assess→Compress→Search→Rank pipeline | 17 Go modules → 4 focused modules |
| **Orchestrator** | `internal/orchestrator/` | bun Phase Graph | CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND | 7-stage embryo → 6-stage native |
| **Worktree Pool** | `internal/sandbox/worktree.go` | gastown ZFC | Per-session `git worktree add --detach` | gastown seance/prime → simplified session prime |
| **Projectmap** | `internal/projectmap/` | embryo projectmap | SQLite+FTS5, auto-index on WRITE | Full graph ranking → incremental update |
| **Mesh Store** | `internal/mesh/store.go` | ADR-005 TextSlot | SQLite WAL + markdown slots | gob binary → text-native |
| **RAG Engine** | `internal/mesh/query.go` | embryo Hybrid RAG | 5-signal ranking (vector+keyword+domain+survival+Z_density) | qdrant-only → local fallback |
| **Quality Gates** | `internal/quality/` | bun Two-Vote | 2/3 consensus, adversarial verify | embryo simple tally → weighted consensus |
| **Tool Registry** | `internal/mcp/tools.go` | SWE-agent ACI | 7 archetypes: view/search/edit/bash/status/commit/push | 49 ad-hoc → 7 constrained |
| **Context Compress** | `internal/rtk/` | rtk TOML | 8-stage filter, 60-90% reduction | Rust `toml_filter.rs` → Go pipeline |
| **Knowledge Graph** | `internal/graphify/` | graphify | IDF-weighted graph, AST→subgraph | Python `serve.py` → Go native |
| **Memory API** | `internal/memory/` | mem0 | `Add/Search/GetAll` with embedding + TTL | mem0 cloud API → local SQLite |
| **Checkpoint** | `internal/checkpoint/` | langgraph | Per-step snapshot, time travel, fork | PostgreSQL → SQLite |
| **Event Stream** | `internal/event/` | openhands | Every action = event, state derived from log | RabbitMQ → local JSONL |
| **Cost Tracker** | `internal/cost/` | ai-reviewer | Per-model tokens/time/cost JSONL | CSV → JSONL |
| **Normalizer** | `internal/normalize/` | ai-reviewer + SWE-agent | Raw → `Finding{}` structured | Regex + cheap model |
| **Error Classifier** | `internal/failover/` | hermes-agent | Rate limit→retry; hallucination→fallback; fatal→abort | hermes iteration budget → token budget |
| **Skill Extractor** | `internal/skill/` | hermes-agent | Trajectory→compression→`Skill{}`→mesh slot | Training pipeline → runtime extraction |
| **C-Core Engine** | `core/ops.c` | embryo BinaryRuntime | OpCode dispatch, chain execution, conflict matrix | Go `BinaryRuntime` → C native |

---

## 3. File System Map

### 3.1 New Directories (v16Q baseline)

```
internal/
├── session/           ← NEW: SessionManager, budget, lifecycle
│   ├── manager.go
│   ├── budget.go
│   ├── lifecycle.go
│   └── logger.go      (moved from current location)
├── hunt/              ← NEW: 4-module Hunt system
│   ├── assess.go      (complexity scoring)
│   ├── compress.go    (context budget trimming)
│   ├── search.go      (mesh query)
│   └── rank.go        (Z-density ranking)
├── sandbox/           ← NEW: worktree isolation
│   ├── worktree.go    (git worktree add/remove)
│   ├── proof.go       (git status, git log, ls -la)
│   └── rollback.go    (git checkout -- + git clean -fd)
├── quality/           ← EXISTING, expanded
│   ├── consensus.go   (2/3 vote)
│   ├── verify.go      (adversarial verify)
│   └── gate.go        (human-in-the-loop)
├── memory/            ← NEW: semantic memory API
│   ├── api.go         (Add, Search, GetAll)
│   ├── decay.go       (TTL, Ebbinghaus curve)
│   └── store.go       (SQLite + embedding)
├── checkpoint/        ← NEW: per-step snapshots
│   ├── snapshot.go
│   ├── travel.go      (rewind/fork)
│   └── store.go
├── event/             ← NEW: event stream
│   ├── stream.go
│   ├── action.go      (observation, action, thought, result)
│   └── replay.go
├── cost/              ← NEW: cost tracking
│   ├── tracker.go
│   ├── model.go       (per-model cost)
│   └── report.go
├── normalize/         ← NEW: raw → structured
│   ├── finding.go     (Finding struct)
│   ├── extract.go     (regex + cheap model)
│   └── aggregate.go   (severity grouping)
├── failover/          ← NEW: error classification
│   ├── classify.go    (retry vs fallback vs abort)
│   ├── retry.go
│   └── abort.go
├── skill/             ← NEW: skill extraction
│   ├── extract.go     (trajectory → Skill)
│   ├── compress.go    (pattern compression)
│   └── slot.go        (write to mesh)
├── config/            ← NEW: repo-scoped config (koanf)
│   ├── loader.go
│   ├── profile.go     (model categories)
│   └── repo.go        (.mimic/<repo>/config.yaml)
└── graphify/          ← EXISTING, expanded
     ├── ast.go         (AST extraction)
     ├── graph.go       (IDF-weighted graph)
     └── subgraph.go    (subgraph rendering)

core/
├── ops.h              ← EXISTING, expanded to v16Q opcodes
├── ops.c              ← EXISTING, rewritten: exec_sys_file_write, exec_git_*, rollback
├── dispatch.c         ← NEW: opcode → executor dispatch table
├── conflict_matrix.c  ← NEW: [256][256] uint8_t, populated from spec
├── energy_matrix.c    ← NEW: [256][3] float, populated from spec
├── rollback.c         ← NEW: 3-phase rollback, inverse mapping
├── snapshot.c         ← NEW: pre-chain state capture
├── validation.c       ← NEW: 11-step validation pipeline
├── exec_context.c     ← NEW: ExecContext lifecycle
└── bridge.h           ← EXISTING, stable C API for CGO

cmd/
├── mimic/             ← EXISTING, main entrypoint
└── playground/        ← NEW: v16Q playground harness
     └── main.go

scripts/
├── wt_create.sh       ← NEW: worktree provisioning
├── wt_destroy.sh      ← NEW: worktree cleanup
├── eval_honest.py     ← NEW: proof + consensus eval
├── distill_batch.sh   ← NEW: batch distillation top-20 repos
└── migrate_embryo.go  ← NEW: embryo 298K → TextSlot

.mimic/ (runtime, not in git)
├── slots/<domain>/    ← TextSlot markdown files
├── runs/<model>/<ts>/ ← per-run artifacts
├── <repo>/            ← repo-scoped config
│   ├── config.yaml
│   ├── personas/
│   ├── primers/
│   └── waivers/
└── runs.db            ← SQLite: events, checkpoints, memory, cost
```

### 3.2 Modified Existing Files

| File | Change |
|------|--------|
| `internal/mesh/text_slot.go` | Add `Store()` method (SQLite persistence) |
| `internal/mesh/query.go` | Add 5-signal hybrid ranking |
| `internal/projectmap/projectmap.go` | Add auto-index hook on WRITE |
| `internal/mcp/tool_schemas.go` | Collapse 49 → 7 archetypes |
| `internal/mcp/mcp.go` | Add `isError` flag, fix JSON-RPC error format |
| `internal/orchestrator/plan.go` | Wire to 6-stage pipeline, add budget gates |
| `internal/cgo/cgo.go` | Add `SetBaseRepo(worktreePath)` |
| `core/ops.c` | Implement declared-but-missing opcodes; add rollback hooks |
| `Makefile` | Add `make playground`, `make distill-batch` |

---

## 4. Interface Contracts

### 4.1 Go Interfaces (Orchestration Layer)

```go
// internal/orchestrator/pipeline.go

package orchestrator

type Stage interface {
    Name() string
    Process(ctx *SessionContext, input []byte) (output []byte, err error)
}

// 6-stage pipeline: STATE → MESH → DIRECT → CLASSIFY → EXEC → FLYWHEEL → RESPOND
type Pipeline struct {
    Stages []Stage
    Budget *Budget
}

func (p *Pipeline) Run(ctx *SessionContext, intent string) (*Result, error)
```

```go
// internal/hunt/hunt.go

type Hunter interface {
    Assess(repoPath string) (*ComplexityScore, error)
    Compress(context []byte, maxTokens int) ([]byte, error)
    Search(query string, domain string) ([]*mesh.TextSlot, error)
    Rank(slots []*mesh.TextSlot) ([]*mesh.TextSlot, error)
}
```

```go
// internal/session/manager.go

type SessionManager interface {
    Init(repoPath string, modelID string) (*SessionContext, error)
    Execute(ctx *SessionContext, intent string) (*Result, error)
    Finalize(ctx *SessionContext) (*Artifacts, error)
    Destroy(ctx *SessionContext) error
}

type SessionContext struct {
    ID            string
    ModelID       string
    WorktreePath  string
    BaseSHA       string
    Budget        Budget
    DenialCount   int
    CircuitBroken bool
    MeshPreload   []*mesh.TextSlot
}
```

```go
// internal/sandbox/worktree.go

type Sandbox interface {
    Provision(repoPath string, modelID string) (worktreePath string, err error)
    Destroy(worktreePath string) error
    CollectProof(worktreePath string) (*Proof, error)
    Rollback(worktreePath string) error
}
```

```go
// internal/memory/api.go

type Memory interface {
    Add(fact string, embedding [384]int8, ttl time.Duration) error
    Search(query string, topK int) ([]*MemoryEntry, error)
    GetAll() ([]*MemoryEntry, error)
    Decay(now time.Time) error
}
```

```go
// internal/checkpoint/snapshot.go

type Checkpoint interface {
    Save(ctx *SessionContext, step int) (checkpointID string, error)
    Load(checkpointID string) (*SessionContext, error)
    Fork(checkpointID string, newIntent string) (*SessionContext, error)
}
```

```go
// internal/normalize/finding.go

type Finding struct {
    Source     string  // model ID
    File       string
    Line       int
    Summary    string
    Details    string
    Severity   Severity // MustFix / Major / Review / Consider
    Confidence float64  // 0.0..1.0
}
```

### 4.2 C Structs (Execution Layer)

```c
// core/ops.h — v16Q additions

typedef struct {
    OpCode opcode;
    char name[OP_NAME_LEN];
    char description[OP_DESC_LEN];
    int (*execute)(OpPacket* packet);
    OpCode inverse_opcode;           // NEW: rollback support
    int (*inverse_execute)(OpPacket* original, OpPacket* inverse_result); // NEW
    uint32_t required_flags;
    uint32_t forbidden_flags;
    float cost_tokens;
    float cost_time_us;
    float cost_memory_bytes;
    uint8_t safety_level;            // 0=critical, 4=safe
    bool is_reversible;
    bool is_atomic;
    uint64_t resource_bitmask;       // NEW: conflict detection
} OpCodeDef;

typedef struct {
    bool is_valid;
    uint32_t error_code;
    char error_msg[256];
    uint32_t invalid_op_index;
    uint32_t conflict_op_pair[2];
    float total_energy;
    float estimated_latency_us;
} ValidationResult;

typedef struct {
    uint32_t context_id;
    uint64_t start_time_ns;
    int32_t* open_fds;
    size_t fd_count;
    void** mmap_regions;
    size_t* mmap_sizes;
    size_t mmap_count;
    uint32_t current_op_id;
    uint32_t total_ops;
    uint32_t success_count;
    uint32_t error_count;
    void* pre_state_blob;
    size_t pre_state_size;
    uint64_t pre_state_hash;
    uint64_t resource_bitmask;
    uint8_t conflict_status;
    uint64_t session_budget_tokens;
    uint64_t session_budget_time_ms;
    uint8_t denial_count;
    uint8_t circuit_broken;
} ExecContext;
```

### 4.3 CGO Bridge Contract

```go
// internal/cgo/bridge.go

package cgo

// SetBaseRepo wires C-core to per-session worktree.
// Must be called once per SessionContext.Init.
func SetBaseRepo(path string) error

// ExecuteChain runs validated packets in C-core.
// Returns results, energy used, latency per step.
func ExecuteChain(packets []Packet) (*ChainResult, error)

// ValidateChain checks 11 validation steps before execution.
func ValidateChain(packets []Packet) (*ValidationResult, error)

// RollbackChain triggers 3-phase rollback on failure.
func RollbackChain(packets []Packet, failedIndex int) error
```

---

## 5. Data Flow

```
Model prompt ("fix race in handler.go")
    ↓
SessionManager.Init()
    → validateGitReality() → HEAD SHA, dirty check
    → provisionWorktree() → git worktree add --detach
    → bootstrapProjectmap() → SQLite FTS5 scan
    → preloadMesh() → domain=go, survival≥0.7
    → loadRepoConfig() → .mimic/<repo>/config.yaml
    → lockBudget() → max_tokens=100K, max_iterations fixed
    → setBaseRepo() → C-core wired to worktree
    ↓
Hunt.Assess() → complexity score (files, deps, history)
Hunt.Compress() → trim context to budget
Hunt.Search() → mesh query: "race condition go handler"
Hunt.Rank() → Z-density × survival_index × similarity
    ↓
Orchestrator.Pipeline.Run()
    Stage 1: STATE → load session state, repo map
    Stage 2: MESH → inject top-ranked slots as context
    Stage 3: DIRECT → sim≥0.85? bypass model, return slot
    Stage 4: CLASSIFY → intent → domain → process name
    Stage 5: PLAN → build OpPacket chain
    Stage 6: VALIDATE → 11-step validation
        → conflict_matrix check
        → energy budget check
        → permission check (never-rules)
    Stage 7: EXEC → cgo.ExecuteChain()
        → per-step: exec opcode → checkpoint.Save() → observe
    Stage 8: VERIFY → 2-vote (if critical)
    Stage 9: RESPOND → result + metrics + traceability
    ↓
Auto-index on WRITE hook
    → exec_sys_file_write → projectmap.IndexFile(path)
    ↓
SessionManager.Finalize()
    → CollectProof() → git status, git log, ls -la
    → NormalizeOutput() → raw → []Finding
    → WriteArtifacts() → .mimic/runs/<model>/<ts>/
    → ExtractSkills() → successful trajectory → mesh slot
    → DestroyWorktree() → git worktree remove
    ↓
Model receives result
    → structured findings + proof artifacts + cost report
```

---

## 6. Behavior-to-Component Mapping

### Tier 1: Core Architecture (embryo)

| Behavior | Component | File | Principle |
|----------|-----------|------|-----------|
| BinaryRuntime (OpCodes 0x10-0x41) | C-Core Engine | `core/ops.c`, `core/dispatch.c` | Opcode chain execution with session tracking |
| Orchestrator pipeline (7-stage) | Pipeline | `internal/orchestrator/pipeline.go` | Chain of responsibility: each stage = interface |
| Hunt system (17 modules) | Hunter | `internal/hunt/assess.go`…`rank.go` | 4 modules: assess→compress→search→rank |
| Mesh agent cascade (3-tier) | Mesh Query | `internal/mesh/query.go` | local→medium→top escalation with cost tracking |
| Hybrid RAG (5-signal) | RAG Engine | `internal/mesh/query.go` | vector+keyword+domain+survival+Z_density |
| Projectmap SQLite + auto-index | Projectmap | `internal/projectmap/projectmap.go` | FTS5 index updated after every WRITE hook |
| Survival index (git blame) | Distillation | `scripts/distill_batch.sh` | `surviving_lines / total_lines` per commit |
| Mapstore slots + invariants | Mesh Store | `internal/mesh/store.go` | Every slot has ≥1 invariant, queryable |

### Tier 2: Agent Behavior (hermes / gastown / bun / rtk / graphify)

| Behavior | Component | File | Principle |
|----------|-----------|------|-----------|
| Closed learning loop | Skill Extractor | `internal/skill/extract.go` | Skills from experience, self-improvement |
| Context compression pipeline | RTK Filter | `internal/rtk/pipeline.go` | 8-stage filter, 60-90% reduction |
| Iteration budget | Budget | `internal/session/budget.go` | `Budget.Consume()` gates loop, exhausted → stop |
| Error classifier + failover | Failover | `internal/failover/classify.go` | Rate limit→retry; hallucination→fallback; fatal→abort |
| Trajectory compression | Skill Compress | `internal/skill/compress.go` | Record tool-call trajectories, compress for slot |
| Tool guardrails | Tool Registry | `internal/mcp/tools.go` | Per-turn reset, deny/ask/allow pipeline |
| ZFC state | Sandbox | `internal/sandbox/proof.go` | Observable reality as source of truth |
| Watchdog chain (3-tier) | Quality Gates | `internal/quality/gate.go` | Human-in-the-loop for critical ops |
| Session continuity | Memory | `internal/memory/api.go` | Session prime + nudge queue |
| Rollback on failure | Sandbox | `internal/sandbox/rollback.go` | Transactional cleanupOnError |
| Phase graph | Orchestrator | `internal/orchestrator/pipeline.go` | CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND |
| Two-vote verify | Quality | `internal/quality/consensus.go` | Two independent verifiers + tiebreak |
| Never-rules | Validation | `core/validation.c` | `git reset/checkout/rebase` blocked without emergency flag |
| Token compression | RTK | `internal/rtk/pipeline.go` | TOML filter pipeline |
| IDF-weighted graph search | Graphify | `internal/graphify/graph.go` | Knowledge graph, AST → graph → subgraph |

### Tier 3: Review & Orchestration (ai-reviewer)

| Behavior | Component | File | Principle |
|----------|-----------|------|-----------|
| Multi-persona pipeline | Normalizer | `internal/normalize/aggregate.go` | Pre-run → reviewers → normalize → waive → aggregate |
| Repo-scoped artifacts | Artifacts | `internal/session/lifecycle.go` | `.mimic/runs/<model>/<ts>/` structure |
| Structured findings | Normalizer | `internal/normalize/finding.go` | `Finding{source,file,line,summary,severity,confidence}` |
| Model categories | Config | `internal/config/profile.go` | `fastest_good/balanced/best_code/frontier_best` |
| Artifact-driven runs | Session Manager | `internal/session/manager.go` | Every run saves prompts, raw, findings, reports |
| Context eval | Budget | `internal/session/budget.go` | Pre-flight token counting, abort if over budget |
| Waiver evaluation | Normalizer | `internal/normalize/aggregate.go` | Location match → LLM confirmation → suppress |

### Tier 4: External Pioneers

| Behavior | Component | File | Principle |
|----------|-----------|------|-----------|
| Repo map (tree-sitter) | Graphify | `internal/graphify/ast.go` | Tree-sitter graph, PageRank-ranked symbols |
| Conventions file | Config | `internal/config/repo.go` | `.mimic/conventions.md` persistent instructions |
| Auto-commit | Sandbox | `internal/sandbox/proof.go` | Every change = git commit with sensible message |
| Trajectory recording | Event Stream | `internal/event/action.go` | Structured episodes: observation, action, thought, result |
| ACI (constrained tools) | Tool Registry | `internal/mcp/tools.go` | view, search, edit, bash with strict schemas |
| Demonstrations | Skill Extractor | `internal/skill/extract.go` | Few-shot from recorded trajectories |
| Semantic memory API | Memory | `internal/memory/api.go` | `add/search/get_all` with embedding-based retrieval |
| Memory decay | Memory | `internal/memory/decay.go` | TTL, Ebbinghaus forgetting curve |
| Event stream | Event | `internal/event/stream.go` | Every action = event, state derived from log |
| Sandboxed runtime | Sandbox | `internal/sandbox/worktree.go` | Worktree isolation (alternative: Docker) |
| Checkpointing | Checkpoint | `internal/checkpoint/snapshot.go` | Per-step snapshots, time travel, fork |
| Persistence | Checkpoint | `internal/checkpoint/store.go` | State in SQLite, survives process crash |
| Context providers | Hunt | `internal/hunt/search.go` | Pluggable context sources |
| Forgetting curve | Memory | `internal/memory/decay.go` | 16pp better recall than Mem0 on LoCoMo |

---

## 7. Gap Closure Matrix

| Gap | Component | File | Validation |
|-----|-----------|------|------------|
| Sandbox broken → no isolation | Sandbox | `internal/sandbox/worktree.go` | `git worktree add --detach` verified per model |
| Hunt missing | Hunter | `internal/hunt/*.go` | Find real patterns in rtk repo, Z-density > 0.5 |
| Pipeline stub → no validation | Orchestrator | `internal/orchestrator/pipeline.go` | 6-stage executes, no EXEC without VALIDATE |
| Memory none | Memory | `internal/memory/*.go` | Session resumes with previous context |
| Tools bloated → 49 | Tool Registry | `internal/mcp/tools.go` | 7 archetypes, models stop hallucinating names |
| Proof none | Sandbox | `internal/sandbox/proof.go` | `git status` + `git log` + `ls -la` mandatory |
| Slots 298K not queryable | Mesh Store | `internal/mesh/store.go` | `MESH_QUERY domain=go` returns ranked results |
| Distillation not run | Scripts | `scripts/distill_batch.sh` | Top-20 repos distilled, survival≥0.7 |
| Auto-index missing | Projectmap | `internal/projectmap/projectmap.go` | Write hook triggers re-index |
| Worktree pool never called | CGO Bridge | `internal/cgo/bridge.go` | `SetBaseRepo()` called per session |
| MCP returns error not isError | MCP Server | `internal/mcp/mcp.go` | JSON-RPC response has `isError` flag |
| C-core base repo never set | C-Core | `core/ops.c` | `g_base_repo` set from Go via CGO |
| Budget none | Session | `internal/session/budget.go` | Budget.Consume() gates every iteration |
| Cost tracking none | Cost | `internal/cost/*.go` | JSONL `run-log.jsonl` per model per run |
| Normalization none | Normalizer | `internal/normalize/*.go` | Raw model output → []Finding |
| Consensus S1 only | Quality | `internal/quality/consensus.go` | S2-S4 through GonkaGate, 2/3 vote |
| Checkpoint none | Checkpoint | `internal/checkpoint/*.go` | Rewind to any step, fork from checkpoint |
| Event stream none | Event | `internal/event/*.go` | Reconstruct session from JSONL events |
| Repo graph flat | Graphify | `internal/graphify/ast.go` | Tree-sitter graph with PageRank |
| Context eval none | Budget | `internal/session/budget.go` | Pre-flight abort if context > budget |
| Skill extraction none | Skill | `internal/skill/*.go` | Successful run → mesh slot with survival tracking |
| ACI none | Tool Registry | `internal/mcp/tools.go` | Constrained tool schemas |
| Repo config global | Config | `internal/config/repo.go` | `.mimic/<repo>/config.yaml` loaded per session |
| Conventions none | Config | `internal/config/repo.go` | `.mimic/conventions.md` versioned with code |

---

## 8. Implementation Waves

### Wave 0: Foundation (Unblocks Everything)

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| Config system | `internal/config/*.go` | go-service-template-rest | koanf layered config, repo-scoped |
| C-core stable API | `core/bridge.h`, `core/ops.h` | embryo | Stable C API for CGO |
| CGO bridge v2 | `internal/cgo/bridge.go` | embryo | SetBaseRepo, ExecuteChain, ValidateChain, RollbackChain |
| Tool archetypes | `internal/mcp/tools.go` | SWE-agent ACI | 7 tools: view, search, edit, bash, status, commit, push |

**Blocked by:** Nothing  
**Blocks:** All other waves  
**Duration:** 3-4 days  
**Test:** `make test` passes, `make check` passes

---

### Wave 1: Sandbox + Execution Runtime

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| Worktree provisioning | `internal/sandbox/worktree.go` | gastown ZFC | `git worktree add --detach` per model |
| Proof collection | `internal/sandbox/proof.go` | User requirement | `git status`, `git log`, `ls -la` mandatory |
| Rollback | `internal/sandbox/rollback.go` | embryo + gastown | `git checkout --` + `git clean -fd` on error |
| Auto-index hook | `internal/projectmap/projectmap.go` | embryo | `IndexFile(path)` after every WRITE |
| C-core base repo wiring | `internal/cgo/bridge.go` | embryo | `cgo.SetBaseRepo(worktreePath)` per session |
| C-core exec context | `core/exec_context.c` | embryo | Per-session ExecContext with resource bitmask |
| C-core validation | `core/validation.c` | bun | 11-step validation pipeline |
| C-core conflict matrix | `core/conflict_matrix.c` | embryo + bun | [256][256] populated from spec |
| C-core energy matrix | `core/energy_matrix.c` | embryo | [256][3] populated from spec |
| C-core rollback | `core/rollback.c` | embryo + gastown | 3-phase rollback with inverse mapping |

**Blocked by:** Wave 0  
**Blocks:** Wave 2  
**Duration:** 4-5 days  
**Test:** S1-S4 scenarios pass in isolated worktrees

---

### Wave 2: Pipeline + Orchestration

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| Session manager | `internal/session/manager.go` | embryo | Init → Execute → Finalize → Destroy lifecycle |
| Budget | `internal/session/budget.go` | hermes-agent | Locked per session, `Consume()` gates loop |
| 6-stage pipeline | `internal/orchestrator/pipeline.go` | bun | STATE→MESH→DIRECT→CLASSIFY→EXEC→VERIFY→RESPOND |
| Hunt system | `internal/hunt/*.go` | embryo | 4 modules: assess, compress, search, rank |
| Error classifier | `internal/failover/*.go` | hermes-agent | retry / fallback / abort |
| Cost tracker | `internal/cost/*.go` | ai-reviewer | Per-model JSONL `run-log.jsonl` |
| Normalizer | `internal/normalize/*.go` | ai-reviewer + SWE-agent | Raw → []Finding |

**Blocked by:** Wave 1  
**Blocks:** Wave 3  
**Duration:** 5-6 days  
**Test:** Full pipeline on rtk repo, proof collected, cost tracked

---

### Wave 3: Multi-Model + Quality

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| GonkaGate harness | `cmd/playground/main.go` | ai-reviewer | Sequential execution, 10min timeout |
| Per-model worktree | `internal/sandbox/worktree.go` | User requirement | `rtk-qwen`, `rtk-kimi`, `rtk-minimax` |
| Consensus aggregation | `internal/quality/consensus.go` | bun | 2/3 vote = preliminary fact |
| Fuzzy matching | `internal/mcp/mcp.go` | Mimic fix | `SYS_RAW_FILE_WRITE` → `SYS_FILE_WRITE`, `<think>` strip |
| Eval harness | `scripts/eval_honest.py` | User requirement | Proof + consensus JSON |
| Report generation | `internal/cost/report.go` | ai-reviewer | `report.md` + `agent_handoff.md` |

**Blocked by:** Wave 2  
**Blocks:** Wave 4  
**Duration:** 3-4 days  
**Test:** S1-S4 through all 3 models, 2/3 consensus, reports generated

---

### Wave 4: Memory + Long-Term State

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| Semantic memory API | `internal/memory/*.go` | mem0 | `Add`, `Search`, `GetAll` with SQLite+embedding |
| Memory decay | `internal/memory/decay.go` | YourMemory | TTL, Ebbinghaus curve |
| Event stream | `internal/event/*.go` | openhands | Every action = event, state from log |
| Checkpointing | `internal/checkpoint/*.go` | langgraph | Per-step snapshot, time travel, fork |
| Repo graph map | `internal/graphify/ast.go` | Aider | Tree-sitter graph, PageRank-ranked |
| Session prime | `internal/session/lifecycle.go` | gastown | Query past sessions, recover context |
| Trajectory artifact | `internal/event/action.go` | SWE-agent | Structured episode, replayable |
| Skill extraction | `internal/skill/*.go` | hermes-agent | Success → `Skill{}` → mesh slot |
| Invariant registry | `internal/mesh/store.go` | embryo | `inv_create`, `invariant_add`, `find_similar` |

**Blocked by:** Wave 3  
**Blocks:** Wave 5  
**Duration:** 5-6 days  
**Test:** Session resumes with memory, checkpoint rewind works, skill extracted after success

---

### Wave 5: Distillation + Mesh Population

| Task | File | Behavior Source | Deliverable |
|------|------|-----------------|-------------|
| Batch distillation | `scripts/distill_batch.sh` | embryo | Top-20 repos, survival≥0.7, Z-density≥0.5 |
| TextSlot migration | `scripts/migrate_embryo.go` | ADR-005 | 298K embryo → markdown-native slots |
| Mesh registry load | `internal/mesh/store.go` | embryo | SQLite + embedding index, queryable |
| Context compression | `internal/rtk/pipeline.go` | rtk | 8-stage filter, 60-90% reduction |
| Knowledge graph | `internal/graphify/*.go` | graphify | IDF-weighted graph, subgraph rendering |
| Conventions file | `internal/config/repo.go` | Aider | `.mimic/conventions.md` per repo |
| Context primers | `internal/config/repo.go` | ai-reviewer | Deterministic lookup, zero-AI |

**Blocked by:** Wave 4  
**Blocks:** Nothing (production-ready)  
**Duration:** 4-5 days  
**Test:** Mesh query returns ranked results, distillation produces 500 slots/domain, rtk compression measured

---

## 9. Dependency Graph

```
Wave 0 (Foundation)
    ├── Config system ─────┐
    ├── C-core stable API ─┼──→ Wave 1 (Sandbox)
    ├── CGO bridge v2 ─────┘       ├── Worktree provisioning
    └── Tool archetypes              ├── Proof + rollback
                                     ├── Auto-index hook
                                     └── C-core context/validation/conflict/energy/rollback
                                              │
                                              ▼
                                     Wave 2 (Pipeline)
                                          ├── Session manager
                                          ├── Budget
                                          ├── 6-stage pipeline
                                          ├── Hunt system
                                          ├── Error classifier
                                          ├── Cost tracker
                                          └── Normalizer
                                               │
                                               ▼
                                     Wave 3 (Multi-Model)
                                          ├── GonkaGate harness
                                          ├── Per-model worktree
                                          ├── Consensus aggregation
                                          ├── Fuzzy matching
                                          ├── Eval harness
                                          └── Report generation
                                               │
                                               ▼
                                     Wave 4 (Memory)
                                          ├── Semantic memory API
                                          ├── Memory decay
                                          ├── Event stream
                                          ├── Checkpointing
                                          ├── Repo graph map
                                          ├── Session prime
                                          ├── Trajectory artifact
                                          ├── Skill extraction
                                          └── Invariant registry
                                               │
                                               ▼
                                     Wave 5 (Distillation)
                                          ├── Batch distillation
                                          ├── TextSlot migration
                                          ├── Mesh registry load
                                          ├── Context compression
                                          ├── Knowledge graph
                                          ├── Conventions file
                                          └── Context primers
```

---

## 10. Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| C-core rewrite breaks existing CGO | High | Keep `internal/cgo/cgo.go` stable, add `bridge.go` as v2 API, dual-compile |
| 298K embryo slots too large for SQLite | Medium | Migrate subset (top survival 0.7), keep rest as cold storage |
| Tree-sitter graph too slow | Medium | Incremental update on WRITE, not full rebuild |
| GonkaGate rate limits at scale | Medium | Sequential execution (1 API key), backoff, cache |
| Multi-model consensus never agrees | Medium | Lower threshold to 1/3 for exploratory, 2/3 for facts |
| Worktree pool exhausts disk | High | Auto-cleanup after session, max 3 worktrees per repo |
| Budget gate too aggressive | Medium | Configurable thresholds per repo profile |
| 7 tools too limiting | Low | Hide stubs, not delete. Enable per-repo tool extensions |

---

## 11. Validation Criteria

v16Q baseline is complete when:

1. `make playground` runs S1-S4 on rtk with proof collection.
2. 3 models (qwen, kimi, minimax) execute sequentially, each in isolated worktree.
3. 2/3 consensus produces preliminary facts for S1.
4. Cost tracking JSONL exists per run.
5. `MESH_QUERY domain=go` returns ≥10 ranked slots.
6. Session resume: new session continues old context (memory API).
7. Checkpoint rewind: can rewind S4 to step 2 and fork.
8. Skill extraction: successful S4 trajectory → mesh slot with survival tracking.
9. `make check` passes (lint + test + semantics-check).
10. No P0 blockers remain.

---

## 12. Artifact Precision

| Artifact | Survival Index | Invariant Coverage | Reproducibility |
|----------|---------------|--------------------|-----------------|
| TextSlot migration | ≥0.7 | 100% (all slots have ≥1 invariant) | git blame → identical survival |
| C-core validation | 1.0 (new code) | 11 steps = 100% | same input → same result |
| Tool archetypes | ≥0.9 (7 tools cover 95% calls) | never-rules enforced | same intent → same tool |
| Hunt ranking | ≥0.7 (Z-density calibrated) | cross-domain links | same query → same top-5 |
| Consensus | ≥0.67 (2/3 threshold) | all findings structured | same input → same consensus |

---

*Branch: v16Q*  
*Next: Review → Approve → Wave 0 implementation*  
*Mimicry principle: Every behavior is borrowed, every implementation is native.*
