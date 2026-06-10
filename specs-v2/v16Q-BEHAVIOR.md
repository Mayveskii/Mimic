# v16Q Behavior Sources Matrix

Complete behavior inventory for Mimic v16Q baseline.
All sources are behavior donors — not code to copy, but principles to mimic.

---

## Tier 1: Core Architecture (Mimic's DNA)

| Repo | Behavior | Status in Mimic | What to Mimic |
|------|----------|-----------------|---------------|
| **Mayveskii/embryo** | BinaryRuntime (OpCodes 0x10-0x41) | partial (C-core derived) | Opcode chain execution, session tracking, inference pool |
| **Mayveskii/embryo** | Orchestrator pipeline (7-stage) | planned | `state→mesh→DIRECT→classify→exec→flywheel→respond` + rollback |
| **Mayveskii/embryo** | Hunt system (17 modules) | missing | `assess→compress→search→rank` by Z-density |
| **Mayveskii/embryo** | Mesh agent cascade (3-tier) | planned | `local→medium→top` escalation with cost tracking |
| **Mayveskii/embryo** | Hybrid RAG (5-signal) | partial | `vector+keyword+domain+survival+Z_density` weighted ranking |
| **Mayveskii/embryo** | Projectmap SQLite + auto-index | partial | FTS5 index updated after every WRITE |
| **Mayveskii/embryo** | Survival index (git blame) | planned | `surviving_lines / total_lines` per commit |
| **Mayveskii/embryo** | Mapstore slots + invariants | planned | Every slot has ≥1 invariant |

## Tier 2: Agent Behavior (Production Practices)

| Repo | Behavior | Status in Mimic | What to Mimic |
|------|----------|-----------------|---------------|
| **Mayveskii/hermes-agent** | Closed learning loop | planned | Skills from experience, self-improvement |
| **Mayveskii/hermes-agent** | Context compression pipeline | planned | Preflight detection + multi-pass compression |
| **Mayveskii/hermes-agent** | Iteration budget | planned | `Budget.consume()` gates loop, exhausted → stop |
| **Mayveskii/hermes-agent** | Error classifier + failover | planned | classify API error → retry vs fallback vs abort |
| **Mayveskii/hermes-agent** | Trajectory compression | planned | Record tool-call trajectories, compress for training |
| **Mayveskii/hermes-agent** | Tool guardrails | planned | Per-turn reset, deny/ask/allow pipeline |
| **Mayveskii/gastown** | ZFC state | planned | Observable reality as source of truth (tmux/git/fs) |
| **Mayveskii/gastown** | Watchdog chain (3-tier) | planned | Daemon→Boot→Deacon→Witness heartbeat |
| **Mayveskii/gastown** | Session continuity (seance/prime/nudge) | planned | Query past sessions, recover context, deferred tasks |
| **Mayveskii/gastown** | Rollback on failure | planned | Transactional resource creation, cleanupOnError |
| **Mayveskii/bun** | Phase graph | planned | `CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND` |
| **Mayveskii/bun** | Two-vote verify | planned | Two independent verifiers + tiebreak |
| **Mayveskii/bun** | Never-rules | planned | `never git reset/checkout/rebase/re-gate` |
| **Mayveskii/rtk** | Token compression (TOML pipeline) | planned | 8-stage filter, 60-90% reduction |
| **Mayveskii/graphify** | IDF-weighted graph search | planned | Knowledge graph, AST → graph → subgraph |

## Tier 3: Review & Orchestration (ai-reviewer)

| Repo | Behavior | Status in Mimic | What to Mimic |
|------|----------|-----------------|---------------|
| **mv-core/ai-reviewer** | Multi-persona pipeline | reference | Pre-run explainers → reviewers → normalize → waive → aggregate → post-run |
| **mv-core/ai-reviewer** | Repo-scoped artifacts | reference | `.ai-review/<repo>/personas/primers/waivers` |
| **mv-core/ai-reviewer** | Structured findings | reference | `Finding{source,file,line,summary,severity,confidence}` |
| **mv-core/ai-reviewer** | Model categories | reference | `fastest_good/balanced/best_code/frontier_best` |
| **mv-core/ai-reviewer** | Artifact-driven runs | reference | Every run saves prompts, raw, findings, reports |
| **mv-core/ai-reviewer** | Context eval | reference | Pre-flight token counting, CSV export |
| **mv-core/ai-reviewer** | Waiver evaluation | reference | Location match → LLM confirmation → suppress |

## Tier 4: External Pioneers (New for v16Q)

| Repo | Behavior | Status in Mimic | What to Mimic |
|------|----------|-----------------|---------------|
| **Aider-AI/aider** | Repo map | missing | Tree-sitter graph, PageRank-ranked symbols |
| **Aider-AI/aider** | Conventions file | missing | `.aider.chat.md` — persistent project instructions |
| **Aider-AI/aider** | Auto-commit | missing | Every change = git commit with sensible message |
| **SWE-agent/SWE-agent** | Trajectory recording | missing | Structured episodes: `{observation,action,thought,result}` |
| **SWE-agent/SWE-agent** | ACI (Agent-Computer Interface) | missing | Constrained tools: `view`, `search`, `edit` with strict schemas |
| **SWE-agent/SWE-agent** | Demonstrations | missing | Few-shot from recorded trajectories |
| **mem0ai/mem0** | Semantic memory API | missing | `add/search/get_all` with embedding-based retrieval |
| **mem0ai/mem0** | Memory decay | missing | TTL, Ebbinghaus forgetting curve |
| **OpenHands/OpenHands** | Event stream | missing | Every action = event. State derived from event log |
| **OpenHands/OpenHands** | Sandboxed runtime | missing | Docker-based isolation (alternative to worktree) |
| **langchain-ai/langgraph** | Checkpointing | missing | Per-step state snapshots, time travel, fork |
| **langchain-ai/langgraph** | Persistence | missing | State in SQLite/Postgres, survives process crash |
| **Continue.dev** | Context providers | missing | Pluggable context sources: open, debugger, problems, folder |
| **sachitrafa/YourMemory** | Forgetting curve | missing | 16pp better recall than Mem0 on LoCoMo |

---

## Behavior Count

- **Tier 1 (embryo)**: 8 behaviors, 1 partial, 7 planned/missing
- **Tier 2 (hermes/gastown/bun/rtk/graphify)**: 14 behaviors, all planned/missing
- **Tier 3 (ai-reviewer)**: 7 behaviors, all reference (not yet in behavior-sources.yaml)
- **Tier 4 (external)**: 13 behaviors, all missing

**Total: 42 behaviors** to mimic, distill, or integrate for v16Q baseline.

---

*Branch: v16Q*
