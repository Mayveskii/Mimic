# Mimic

> **Execution layer for AI agents that doesn't hallucinate.**

Mimic is an MCP server with a C-core execution engine — not a wrapper around bash, but a deterministic runtime that validates every operation before it runs, measures cost, and rolls back on failure.

---

## For Sponsors: Why This Matters

AI agents waste billions of tokens on trial-and-error. 

- **Claude Code** retries failed commands ad-hoc
- **Cursor** executes without validation
- **Every agent** treats tools like `git`, `make`, `docker` as black boxes

**Mimic changes the equation:**

| Without Mimic | With Mimic |
|---------------|------------|
| Model guesses arguments → crashes → retries → **$$$ burned** | Schema validates before run → zero retries → **30-50% token savings** |
| `git log` returns 5000 lines → 25K tokens | RTK compression → 250 tokens → **95% context saved** |
| Agent calls `rm -rf` by accident | Conflict matrix blocks destructive sequences |
| No offline knowledge | Mesh of distilled production patterns — works without internet |

**Result**: Agents execute faster, cheaper, safer. And every execution improves the knowledge base.

---

## Architecture

```
 ┌─ AI Agent (Claude, GPT, Kimi — autonomous, optional to use Mimic)
 │
 │  JSON-RPC over stdio / TCP
 ↓
 ┌─────────────────────────────────┐
 │ MCP Server (Go)                 │  Tool routing, web search (Exa), mesh query
 │ • 48 tools available            │
 │ • JSON Schema for every arg     │
 └──────────┬──────────────────────┘
            │
            ▼
 ┌─────────────────────────────────┐
 │ Orchestrator (Go)               │  6-phase pipeline:
 │                                 │  1. Classify intent
 │                                 │  2. Plan → OpPacket chain
 │                                 │  3. Validate (conflict + budget + permission)
 │                                 │  4. Execute via CGO
 │                                 │  5. Verify (2-vote adversarial)
 │                                 │  6. Respond + compress
 └──────────┬──────────────────────┘
            │
            ▼
 ┌─────────────────────────────────┐
 │ C-Core (C)                      │  96 OpCodes, real syscalls
 │ • ops_execute_chain()           │  • stat(), open(), git, make, curl
 │ • Conflict matrix [96×96]       │  • Measured latency per op
 │ • Energy cost [tokens, μs, bytes│  • Rollback on failure
 └──────────┬──────────────────────┘
            │
            ▼
         Linux
```

---

## Two Knowledge Sources (Proven Patterns)

### 1. Distillation — Production Code That Survived

Takes 90+ production repos (etcd, k8s, go-ethereum, Redis, nginx...):
```
git blame → survival index = surviving_lines / total_added
survival ≥ 0.7  → mesh slot (proven pattern)
survival < 0.1  → discard
```

**Offline. Local `.gob` files. No API calls.**

### 2. Mimicry — Best Behaviors from Top Repos

Analyzes 16 Mayveskii/* repos and selects HOW to implement:

| Source Repo | Behavior | Applied In Mimic |
|-------------|----------|-----------------|
| bun | Phase graph + 2-vote verify | Orchestrator + Quality |
| rtk | Token compression pipeline | RTK compression (95% reduction) |
| exa-mcp-server | Web search + rate limiting | Exa integration |
| graphify | IDF-weighted graph search | Mesh query |

**Not copying code. Selecting approach.**

---

## Tools (48 Total)

**System**: file ops, dir ops, env, exec  
**Build**: compile, link, test, deploy, clean  
**Git**: status, diff, add, commit, branch, checkout  
**Network**: HTTP GET/POST, TCP  
**Mesh**: query, execute_pattern, auto_apply, status  
**ProjectMap**: index, query_symbol, search_text, synthesize  
**Exa**: search_web, fetch_content, deep_research  
**Plan**: generate_validated_plan  

Every tool has JSON Schema, cost metrics, safety level.

---

## Quick Start

```bash
# 1. Build (C-core + Go)
make build

# 2. Configure
cp .env.example .env
# Set EXA_API_KEY for web search (optional — mesh works offline)

# 3. Run
./bin/mimic serve              # stdio MCP (for opencode, Claude Code)
./bin/mimic serve --tcp :1337  # TCP mode (for remote agents)

# 4. Test
make test          # Go + C tests
make check         # lint + test + semantics
```

---

## Roadmap: From Passive Tool to Autonomous Agent

| Stage | What | When |
|-------|------|------|
| **0 — Passive** ✅ | Agent asks → Mimic executes | Now |
| **1 — Proactive** | Suggests next steps before being asked | v0.4 |
| **2 — Planning** | Generates multi-step plans with checkpoints | v0.5 |
| **3 — Generative** | Proposes new patterns from session logs | v0.6 |
| **4 — Autonomous** | Self-directed learning, no human intervention | v0.7 |

**Collective Intelligence**: When multiple agents use Mimic, their mesh graphs merge. The more agents use it, the smarter it gets — for everyone.

---

## Stats

- **48 MCP tools**
- **18 domain mesh graphs** (offline knowledge)
- **90+ production repos** in distillation pipeline
- **95% token reduction** via RTK compression
- **MIT License**

---

## Read More

- `specs/01-AGENTS.md` — Rules for agents working on Mimic
- `specs/02-ARCHITECTURE.md` — Components, flows, boundaries
- `docs/adr/` — Every non-trivial decision (11 ADRs)
- `docs/architecture/ROADMAP_AUTONOMY.md` — 4-stage autonomy plan
