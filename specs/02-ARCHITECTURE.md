# ARCHITECTURE.md — Mimic

## System

Mimic is a standalone MCP server with a C-core. It is an **optional tool** — the AI-agent is fully autonomous and works without Mimic. When the agent calls Mimic, it receives help packaging an intent into a validated OpPacket chain, backed by distilled patterns and borrowed behaviors.

Two knowledge sources feed into Mimic:
- **Distillation**: production repos → best commits → mesh slots
- **Mimicry**: Mayveskii/* repos → behavior selection → implementation in Mimic

**mimic-server** is a future component — a shared knowledge hub for when multiple clients need to mix requests. Not in current scope.

## Request Flow

```
AI-agent has a task and intent (fully autonomous)
    ↓ agent CHOOSES to call Mimic MCP tool (optional)
MCP Layer (transport + middleware + hooks)
    ↓ tool → intent mapping
Orchestrator
    ↓ classify intent → query slots (if available) → build OpPacket chain
    ↓ validate: conflict matrix + energy budget + DIFC
    ↓ permission: deny → classify → budget → ok
CGO Bridge
    ↓ Go OpPacket → C OpPacket
C-Core
    ↓ ops_execute_chain()
    ↓ [OP_READ → OP_DIFF → OP_BUILD → OP_TEST → OP_COMMIT]
    ↓ each step: measured, validated, deterministic
    ↓ 2-vote verify (post-exec, if scenario requires)
result → MCP response → agent (agent decides what to do with it)
```

## Components

### MCP Layer (internal/mcp/)

Transport: stdio, SSE, HTTP.
Tool registry: MCP tools visible to the agent.
Middleware: PreToolUse/PostToolUse hooks.
Circuit breaker per tool.

MCP tools (initial set):
- `exec_chain` — execute a chain by scenario name or explicitly
- `validate` — validate a chain without executing
- `query_slots` — query relevant patterns from mesh (when mimic-server available)
- `status` — engine state, budget, stats
- `hunt` — find a pattern for a specific domain

### Orchestrator (internal/orchestrator/)

Workflow state machine. Phases:
1. **CLASSIFY** — determine task type
2. **PLAN** — build OpPacket chain (using slots if available)
3. **VALIDATE** — conflict matrix + energy budget + permission
4. **EXEC** — ops_execute_chain via CGO
5. **VERIFY** — 2-vote adversarial verify (if scenario requires)
6. **RESPOND** — result to agent

Permission pipeline: deny → classify → budget → allow.
DIFC security: label agent → label resource → check → execute → label response → filter.

### CGO Bridge (internal/cgo/)

Go ↔ C conversion. Functions: Init, Execute, ExecuteChain, ValidateChain, CheckConflict, MMapAlloc/Free/Sync.

### C-Core (core/)

Structures: OpPacket, OpCodeDef, ExecContext, ValidationResult.
Matrices: Conflict [OP_MAX × OP_MAX], Energy cost [OP_MAX × 3].
46 OpCodes declared, 5 scenarios, libbmap.a (39 symbols).

### mimic-server (FUTURE — not current scope)

Shared knowledge hub for multiple clients.
Only GET API: /slots, /patterns/{id}, /z-density, /survival/{hash}.
Distillation pipeline: repos → git blame → survival → extract → encode → slot → store.
Will be a separate repo (Mayveskii/mimic-server) when needed.

## Behaviors from Sources

| Behavior | Source | Where in Mimic | Status |
|----------|--------|---------------|--------|
| Phase graph (classify→plan→exec→verify) | Mayveskii/bun | Orchestrator | planned |
| 2-vote adversarial verify | Mayveskii/bun | Quality | planned |
| Edit scope isolation | Mayveskii/bun | Conflict matrix | planned |
| PreToolUse/PostToolUse hooks | Mayveskii/bun | MCP middleware | planned |
| Never-rules | Mayveskii/bun | Permission | planned |
| Lifetime classify (3-vote refute) | Mayveskii/bun | Orchestrator/classify | planned |
| Web search tool | Mayveskii/exa-mcp-server | Tool registry | planned |
| Web fetch tool | Mayveskii/exa-mcp-server | Tool registry | planned |
| MCP tool registration pattern | Mayveskii/exa-mcp-server | Tool registry | planned |
| Routed/unified/proxy transport | Mayveskii/gh-aw-mcpg | MCP transport | planned |
| DIFC 6-phase security | Mayveskii/gh-aw-mcpg | Security | planned |
| Circuit breaker per backend | Mayveskii/gh-aw-mcpg | MCP Layer | planned |
| OAuth PKCE | Mayveskii/gh-aw-mcpg | Auth | planned |
| WASM guards | Mayveskii/gh-aw-mcpg | Security | future |
| Tool.Def interface | Mayveskii/opencode-anomalyco- | Tool registry | planned |
| Tool loop (while(true) generator) | Mayveskii/opencode-anomalyco- | Orchestrator | planned |
| MCP transport (stdio/SSE/HTTP) | Mayveskii/opencode-anomalyco- | MCP transport | planned |
| Permission pipeline | Mayveskii/code-mode | Permission | planned |
| Denial tracking (3→circuit break) | Mayveskii/code-mode | Quality | planned |
| Auto classifier (allow/deny) | Mayveskii/code-mode | Permission | planned |
| Budget enforcement | Mayveskii/code-mode | Orchestrator | planned |
| Concurrency control (10 parallel) | Mayveskii/code-mode | Orchestrator | planned |

## Not Yet Analyzed (pending)

- Mayveskii/gastown
- Mayveskii/rustnet
- Mayveskii/netboot.xyz
- Mayveskii/git
- Mayveskii/gitingest
- Mayveskii/agency-agents
- Mayveskii/openmythos
- Mayveskii/caveman
- Mayveskii/minbpe

## Boundaries

- Mimic does NOT replace agent tools — it complements with deterministic execution
- C-core does NOT know about MCP — it knows OpPacket and conflict matrix
- Agent does NOT see OpCodes — it sees MCP tools with human-readable names
- Mimicry ≠ copying — it is selecting behavior and implementing our own
- mimic-server is FUTURE — not part of current architecture
