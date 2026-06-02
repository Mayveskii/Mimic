# ARCHITECTURE.md — Mimic

## System

Mimic is a standalone MCP server with a C-core. It is an **optional tool** — the AI-agent is fully autonomous and works without Mimic. When the agent calls Mimic, it receives help packaging an intent into a validated OpPacket chain, backed by distilled patterns and borrowed behaviors.

Two knowledge sources feed into Mimic:
- **Distillation**: production repos → best commits → mesh slots
- **Mimicry**: Mayveskii/* repos → behavior selection → implementation in Mimic

**mimic-server** is a future component — a shared knowledge hub for when multiple clients need to mix requests. Not in current scope.

> **NOTE**: Canonical specifications moved to `specs-v2/`. This file is historical. Read `specs-v2/README.md` and `specs-v2/STRUCTURE.md` first.

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

Transport: stdio, TCP, unix (per `specs-v2/c-core/RPC_FORMAT.md`).
Tool registry: MCP tools visible to the agent.
Middleware: PreToolUse/PostToolUse hooks.
Circuit breaker per tool.

MCP tools (per `specs-v2/c-core/RPC_FORMAT.md`):
- `mimic_execute` — execute a chain by scenario name or explicitly
- `mimic_validate` — validate a chain without executing
- `mimic_query` — query relevant patterns from mesh
- `mimic_status` — engine state, budget, stats

Future tools:
- `mimic_search_web` — web search
- `mimic_search_code` — code semantic search
- `mimic_checkpoint` — create checkpoint
- `mimic_resume` — resume from checkpoint
- `mimic_delegate` — delegate to subagent

### Orchestrator (internal/orchestrator/)

Workflow state machine. Phases per `specs-v2/domains/orchestrator/PROCESS.md`:
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

### C-Core (c-core/)

Structures per `specs-v2/c-core/`: OpPacket, OpCodeDef, ExecContext, ValidationResult.
Matrices: Conflict [OP_MAX × OP_MAX], Energy cost [OP_MAX × 3].
Opcode spec: `specs-v2/c-core/OPCODE_SPEC.md` (full enum 0x00-0xFF with research + self-management ops).
Build spec: `specs-v2/c-core/BUILD_CONFIG.md`.
Memory layout: `specs-v2/c-core/MEMORY_LAYOUT.md` (arena-based alternative).

### mimic-server (FUTURE — not current scope)

Shared knowledge hub for multiple clients.
Only GET API: /slots, /patterns/{id}, /z-density, /survival/{hash}.
Distillation pipeline: repos → git blame → survival → extract → encode → slot → store.
Will be a separate repo (Mayveskii/mimic-server) when needed.

## Behaviors from Sources

| Behavior | Source | Where in Mimic | Status |
|----------|--------|---------------|--------|
| Phase graph (classify→plan→exec→verify) | Mayveskii/bun | Orchestrator | specified |
| 2-vote adversarial verify | Mayveskii/bun | Quality | specified |
| Edit scope isolation | Mayveskii/bun | Conflict matrix | specified |
| PreToolUse/PostToolUse hooks | Mayveskii/bun | MCP middleware | specified |
| Never-rules | Mayveskii/bun | Permission | specified |
| Lifetime classify (3-vote refute) | Mayveskii/bun | Orchestrator/classify | specified |
| Web search tool | Mayveskii/exa-mcp-server | Tool registry | specified |
| Web fetch tool | Mayveskii/exa-mcp-server | Tool registry | specified |
| MCP tool registration pattern | Mayveskii/exa-mcp-server | Tool registry | specified |
| Routed/unified/proxy transport | Mayveskii/gh-aw-mcpg | MCP transport | specified |
| DIFC 6-phase security | Mayveskii/gh-aw-mcpg | Security | specified |
| Circuit breaker per backend | Mayveskii/gh-aw-mcpg | MCP Layer | specified |
| OAuth PKCE | Mayveskii/gh-aw-mcpg | Auth | specified |
| Tool.Def interface | Mayveskii/opencode-anomalyco- | Tool registry | specified |
| Tool loop (while(true) generator) | Mayveskii/opencode-anomalyco- | Orchestrator | specified |
| MCP transport (stdio/tcp/unix) | Mayveskii/opencode-anomalyco- | MCP transport | specified |
| Permission pipeline | Mayveskii/code-mode | Permission | specified |
| Denial tracking (3→circuit break) | Mayveskii/code-mode | Quality | specified |
| Auto classifier (allow/deny) | Mayveskii/code-mode | Permission | specified |
| Budget enforcement | Mayveskii/code-mode | Orchestrator | specified |
| Concurrency control (10 parallel) | Mayveskii/code-mode | Orchestrator | specified |
| Research hypothesis tracking | Scientific method + crewai/autogen | Research domain | NEW |
| Self-checkpoint/resume | terraform + gastown | Self-management domain | NEW |
| Budget reallocation | code-mode + hermes-agent | Self-management domain | NEW |
| Strategy pivot | crewai/openmythos | Self-management domain | NEW |
| Tool chaining | langchain | Orchestrator/RPC | NEW |
| Long-context management | vllm/ollama | Session/RAG | NEW |

## Research Capabilities (NEW)

Mimic supports long-running scientific research via `specs-v2/domains/research/`:

- **Hypothesis tracking**: falsifiable predictions with confirm/refute/refine loops
- **Experiment design**: parameterized measurements with reproducibility hashes
- **Literature ingestion**: PDF/arXiv parsing, citation graph, semantic indexing
- **Multi-session continuity**: checkpoint/resume with context preservation
- **Tool chaining**: structured output from Tool A → input to Tool B

## Self-Management Capabilities (NEW)

Mimic manages its own state via `specs-v2/domains/self-management/`:

- **Self-checkpoint**: snapshot state at configurable intervals
- **Budget reallocation**: redistribute remaining budget across subtasks
- **Strategy pivot**: switch approach after N failures, query mesh for alternatives
- **Progress self-assessment**: detect divergence and stalls
- **Context summarization**: semantic compression to stay within limits

## Boundaries

- Mimic does NOT replace agent tools — it complements with deterministic execution
- C-core does NOT know about MCP — it knows OpPacket and conflict matrix
- Agent does NOT see OpCodes — it sees MCP tools with human-readable names
- Mimicry ≠ copying — it is selecting behavior and implementing our own
- mimic-server is FUTURE — not part of current architecture
- All canonical specs are in `specs-v2/` — old specs (00-11) are historical context
