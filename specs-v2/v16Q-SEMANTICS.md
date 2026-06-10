# v16Q Semantic Diff

Full as-is vs to-be specification for Mimic baseline playground and multi-model honest evaluation.
Derived from: embryo, ai-reviewer, aider, swe-agent, mem0, openhands, langgraph, hermes-agent, gastown, rtk, graphify.

---

## 1. As-Is (Current Mimic State)

| Component | Status | What Exists | What Is Broken |
|-----------|--------|-------------|----------------|
| C-core opcodes | partial | 96 opcodes declared, ~40 implemented, ~50 stubs | `g_base_repo` never set → worktree pool returns early |
| MCP server | partial | JSON-RPC handler, 49 tools, CGO bridge | Returns JSON-RPC error instead of `isError` flag |
| Projectmap | partial | SQLite+FTS5, regex parser | No auto-index on WRITE hook |
| Mesh/RAG | partial | Qdrant + local fallback | No 5-signal hybrid, no survival weight |
| Session logger | partial | JSONL logging | ExtractPatterns = stub |
| Orchestrator | stub | `internal/orchestrator/loop.go` structs | Not wired to MCP, no pipeline execution |
| Hunt system | missing | None | Not spec'd, not implemented |
| Quality | empty | Package exists | 2-vote verify not implemented |
| Worktree sandbox | broken | `ops_set_base_repo` defined | Never called from Go → execution falls back to CWD |
| Tool registry | bloated | 35-49 tools | No MUS audit, models overloaded |

---

## 2. To-Be (v16Q Baseline)

### 2.1 Sandbox & Execution Runtime

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| Per-session worktree provisioning | embryo + gastown ZFC | Go `SessionManager` creates worktree via `git worktree add --detach`. Not static C-core pool. |
| Git reality check on init | gastown ZFC | `git status --porcelain`, `git rev-parse HEAD`, `git log --oneline -1` before any tool call. Baseline SHA fixed for session. |
| Auto-index on WRITE | embryo `pkg/projectmap/` | Hook in `exec_sys_file_write` → `projectmap.IndexFile(path)` after `fclose`. |
| C-core base repo wiring | embryo BinaryRuntime | `cgo.SetBaseRepo(worktreePath)` called once per session. |
| Rollback on failure | embryo + gastown | `git checkout -- .` + `git clean -fd` on error. 3-phase rollback in C-core. |
| Artifact-driven runs | ai-reviewer | Every run writes to `.mimic/runs/<model>/<timestamp>/`: `prompt.md`, `raw.md`, `tool_calls.json`, `findings.json`, `git_log.txt`, `ls_la.txt`. |
| Proof collection | User requirement | `git status`, `git log --oneline -5`, `ls -la` mandatory after execution. No eval without proof. |

### 2.2 Pipeline & Orchestration

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| 6-stage pipeline | embryo `pkg/orchestrator/` | `state→mesh→DIRECT→classify→exec→flywheel→respond`. Each stage = Go interface with `Process()`. Chain of responsibility. |
| Hunt system (minimal) | embryo `pkg/hunt/` | 4 modules: `Assess` (complexity), `Compress` (context budget), `Search` (mesh query), `Rank` (Z-density). |
| Iteration budget | hermes-agent | `Budget.Consume()` gates tool loop. Exhausted → grace call → stop. Max iterations locked at session start. |
| Error classifier + failover | hermes-agent | Rate limit → retry; hallucination → fallback; unrecoverable → abort. |
| Conflict matrix | embryo + bun | `[N×N]` matrix defining incompatible operations. Checked before exec. |
| Energy cost matrix | embryo | `[N×3]` matrix: tokens, time_us, memory_bytes. Budget check before exec. |
| Plan validation | embryo + bun | `ops_validate_chain`: 11 steps. No EXEC without passed VALIDATE. |

### 2.3 Multi-Model & GonkaGate

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| Model categories | ai-reviewer + embryo cascade | `local`/`medium`/`top` abstract categories mapped to concrete models via config profile. GonkaGate proxy handles routing. |
| Per-model worktree isolation | User requirement | Each model gets isolated worktree. Sequential execution (1 API key). |
| Normalization (raw → structured) | ai-reviewer + SWE-agent | Raw model output → `Finding{file, line, summary, severity, confidence}` via regex + cheap model fallback. |
| Multi-model consensus | bun `two_vote_verify` | 3 models → normalize → aggregate → 2/3 vote = preliminary fact. Disagreement = flag for human. |
| Cost tracking | ai-reviewer | Per-model: tokens in/out, wall time, estimated cost. JSONL `run-log.jsonl`. |
| Context eval | ai-reviewer `--context-eval` | Pre-flight token counting. Abort if context > budget. `--dry-run` mode available. |
| Fuzzy tool matching | Mimic fix | `SYS_RAW_FILE_WRITE` → `SYS_FILE_WRITE`. Parser handles `<think>` tags, markdown fences, typos. |

### 2.4 Memory & Long-Term State

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| Repo graph map | Aider | Tree-sitter graph of all symbols, PageRank-ranked by dependency. Not flat file list. |
| Semantic memory API | Mem0 | `memory.Add(fact)`, `memory.Search(query)`, `memory.GetAll()`. SQLite+embedding backend. TTL and decay (YourMemory/Ebbinghaus). |
| Trajectory artifact | SWE-agent | Structured episode: `{observation, action, thought, result}`. Replayable. Used for demonstrations and skill extraction. |
| Event stream | OpenHands | Every action = event. State derived from event log. Can reconstruct session from events. |
| Checkpointing | LangGraph | Snapshot state after each step. Time travel: rewind to any checkpoint. Fork from checkpoint. |
| Session prime | gastown | `seance` (query past sessions) + `prime` (context recovery) + `nudge queue` (deferred tasks). New session continues old, not starts from zero. |
| Invariant registry | embryo `pkg/mapstore/` | Each slot/domain has ≥1 invariant. Checked before exec. `inv_create`, `invariant_add`, `find_similar`. |
| Skill extraction | hermes-agent | Successful trajectories → compression → `Skill{trigger, action, success_rate}` → stored as mesh slot. |

### 2.5 Interface & Tools

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| ACI (Agent-Computer Interface) | SWE-agent | Constrained tool set: `view`, `search`, `edit`, `bash` with strict schemas. Not 49 ad-hoc tools. |
| Tool audit → 7 archetypes | User requirement | MUS: frequency × irreplaceability × latency. Collapse 49 tools → 7 archetypes. Hide stubs. |
| Repo-scoped config | ai-reviewer | `.mimic/<repo>/config.yaml` + `personas/` + `primers/` + `waivers/`. Loaded per session. |
| Conventions file | Aider + Cursor + Cline | `.mimic/conventions.md` in repo root. Persistent instructions, versioned with code. |
| Context primers / concepts | ai-reviewer | Deterministic lookup: `Mimic context primers <repo> --files X --concepts Y`. Zero-AI. |
| MCP discovery & lifecycle | Claude Desktop / OpenHands | Scan `mcpServers` in config. Spawn via `os/exec` stdio pipes. JSON-RPC handshake. Manage process lifecycle in Go. |
| Context compression | rtk + hermes-agent | TOML filter pipeline or tiktoken truncation. 60-90% reduction. Pre-flight detection. |
| Knowledge graph | graphify | IDF-weighted graph search. AST extraction → graph → subgraph rendering. |

### 2.6 Findings & Community

| Requirement | Source | Implementation |
|-------------|--------|----------------|
| Structured findings | ai-reviewer | `Finding{source, file, line, summary, details, severity, confidence}`. Aggregated by severity: Must Fix / Major / Review / Consider. |
| Waiver evaluation | ai-reviewer | Location match → LLM confirmation → suppress. Waived findings listed separately. |
| Community mesh (future) | User vision | Open knowledge sharing. Anonymized slots upload (opt-in). Federation across agents. Zero-cost access to distilled patterns. |
| Human-in-the-loop | gastown + bun | Critical ops require human approval. 2-vote verify for destructive operations. |

---

## 3. Component Diff Matrix

| Component | As-Is | To-Be (v16Q) | Delta |
|-----------|-------|--------------|-------|
| Sandbox | Broken static pool | Dynamic per-session worktree | **Rewrite** |
| Auto-index | None | Hook on WRITE | **Add** |
| Pipeline | Stub structs | 6-stage wired | **Implement** |
| Hunt | Missing | 4-module minimal | **Implement** |
| Budget | None | Locked per session | **Add** |
| Error classifier | None | retry/fallback/abort | **Add** |
| Normalization | None | Raw → structured JSON | **Add** |
| Multi-model | S1 consensus only | S1-S4 + 2/3 vote | **Extend** |
| Repo map | Flat SQLite | Graph-ranked symbols | **Upgrade** |
| Memory | Logs only | Semantic memory API | **Add** |
| Trajectory | Logs only | Structured replayable | **Upgrade** |
| Checkpoint | None | Per-step snapshots | **Add** |
| Tools | 49 bloated | 7 ACI archetypes | **Collapse** |
| Config | Global | Repo-scoped + profiles | **Restructure** |
| Context eval | None | Pre-flight token check | **Add** |
| Findings | Raw text | Structured + aggregated | **Upgrade** |

---

## 4. GonkaGate Integration

### 4.1 Models
- `qwen/qwen3-235b` — primary (local tier equivalent)
- `moonshotai/kimi-k2.6` — medium tier
- `minimaxai/minimax-m2.7` — top tier / verification

### 4.2 Proxy Configuration
- Endpoint: GonkaGate proxy
- Single API key → sequential execution
- No rate limiting observed
- `max_tokens = 100000` per call

### 4.3 Per-Model Isolation
- Each model gets own worktree: `rtk-qwen`, `rtk-kimi`, `rtk-minimax`
- Baseline: same HEAD SHA
- Proof collected per model
- Consensus: 2/3 agreement required for fact status

---

## 5. Filesystem Context (Always Remember)

- `/home/cisco/mimic/specs-v2/` — domain-based semantic specs (this file lives here)
- `/home/cisco/mimic/specs/` — legacy numbered specs (00-11)
- `/home/cisco/mimic/project_context_main/` — project context (EVIDENCE, MEMORY, PHILOSOPHY, TODO)
- `/home/cisco/mimic/mimicrya/behavior-sources.yaml` — 123 behaviors from 20+ repos
- `/home/cisco/mimic/data/mesh/` — slot storage, graphs, registry

---

*Branch: v16Q*
*Next: v16Q-qq (qwen-specific experiments), v16Q-qz (multi-model consensus tuning)*
