# Roadmap to Autonomy — Mimic

## Vision

Current: Mimic is a **passive tool** — agent asks, Mimic executes.

Target: Mimic is an **active participant** — observes, anticipates, suggests, learns.

Inspiration: GiT paper shows that a universal interface (text tokens) + joint training across tasks leads to emergence. Mimic applies the same principle to agent operations.

## Stages

### Stage 0: Passive Tool (NOW — v0.3)

Agent calls tool → Mimic executes → returns result.

**Capabilities:**
- 45 deterministic tools (C-core)
- Mesh lookup for proven patterns
- ActionBytes decode → agent decides
- ProjectMap indexes workspace
- TCP server for multi-client

**Limitations:**
- No initiative — waits for explicit query
- No cross-domain reasoning
- No learning from agent sessions
- Static knowledge (mesh doesn't grow)

### Stage 1: Proactive Suggestions (v0.4)

Mimic observes agent's context and **suggests** next steps before being asked.

**Implementation:**
```
1. Workspace event stream
   - Agent reads file → ProjectMap notes file context
   - Agent edits file → Mesh suggests patterns for that domain
   
2. Background mesh query
   - On every tool call, extract domain keywords
   - Pre-query mesh for related patterns
   - Offer suggestion in response metadata
   
3. MCP notification channel
   - New method: `notifications/suggest`
   - Send proactive hints to agent
```

**Example:**
```
Agent: SYS_FILE_READ path=raft.go
Mimic (suggest): "This looks like Raft consensus.
                   Related: leader election (sim=0.81),
                   log replication (sim=0.76).
                   Call MESH_QUERY domain=raft?"
```

**New tools:**
- `MESH_WATCH` — register interest in domain
- `CONTEXT_SUGGEST` — get proactive suggestions

---

### Stage 2: Compositional Planning (v0.5)

Mimic **generates** multi-step plans, not just single operations.

**Implementation:**
```
1. LLM-based plan generation
   - Input: task + workspace context + available tools
   - Output: ordered OpPacket chain
   
2. Validation before execution
   - Conflict matrix check
   - Budget estimation (token + time + memory)
   - Rollback strategy
   
3. Execution with checkpoints
   - After each step: verify invariant
   - On failure: rollback to last checkpoint
   - On success: update mesh with new pattern
```

**Example:**
```
Agent: "Refactor error handling in this package"
Mimic (plan):
  1. PROJECT_MAP_QUERY_SYMBOL type=func returns=error
  2. SYS_FILE_READ each file
  3. MESH_QUERY "go error wrapping pattern"
  4. FILE_EDIT apply wrapping (per function)
  5. BUILD_TEST verify
  6. (on success) save as new mesh slot
```

**New tools:**
- `PLAN_GENERATE` — generate validated plan
- `PLAN_EXECUTE` — execute with checkpoint rollback
- `PLAN_FEEDBACK` — mark plan success/failure

---

### Stage 3: Self-Improving Mesh (v0.6)

Mimic **learns** from every interaction, growing its knowledge base.

**Implementation:**
```
1. Session logging
   - Every tool call → log with context, result, success/failure
   - Store in `data/sessions/YYYY-MM-DD.log`
   
2. Pattern extraction
   - Successful plan → extract invariant + actions
   - Embed invariant → int8[384]
   - Store as new slot in workspace graph
   
3. Survival tracking
   - Track how often new slot is used
   - If usage < threshold after 30 days → demote to cold
   - If agent edits slot → update embedding
   
4. Cross-agent sharing (future)
   - Opt-in: anonymized slots uploaded to mimic-server
   - Community mesh: all agents benefit from collective learning
```

**Example:**
```
After 10 sessions with PostgreSQL:
- Extracted: "pgx connection pool with retry"
- New slot added: domain=database, sim threshold=0.6
- Next agent asks "connect to postgres" → gets this pattern
```

**New tools:**
- `MESH_LEARN` — convert last N interactions to slot
- `MESH_FORGET` — remove low-usage slot
- `SESSION_EXPORT` — share patterns (opt-in)

---

### Stage 4: Autonomous Agent (v0.7+)

Mimic operates **independently** with high-level goals, not step-by-step instructions.

**Implementation:**
```
1. Goal decomposition
   - Input: "Make this service production-ready"
   - Decompose: observability + error handling + config + docs
   
2. Parallel execution
   - Multiple sub-agents (goroutines) work on sub-goals
   - Shared workspace graph for coordination
   
3. Self-monitoring
   - Track own performance (latency, accuracy, cost)
   - If mesh query slow → trigger index rebuild
   - If pattern accuracy low → request agent feedback
   
4. Meta-learning
   - Learn which domains need more slots
   - Trigger distillation for underrepresented domains
   - Adapt similarity thresholds per domain
```

**Example:**
```
Human: "Deploy this"
Mimic:
  - Checks Dockerfile (no health check → adds it)
  - Checks monitoring (no metrics → adds Prometheus)
  - Checks tests (failing → fixes)
  - Builds container
  - Deploys with blue/green
  - Verifies health endpoint
  - Reports: "Done. Changes: ..."
```

---

## Technical Prerequisites by Stage

| Prerequisite | Stage 1 | Stage 2 | Stage 3 | Stage 4 |
|-------------|---------|---------|---------|---------|
| Text-native mesh | ✓ | ✓ | ✓ | ✓ |
| Cross-domain edges | | ✓ | ✓ | ✓ |
| HNSW index | | ✓ | ✓ | ✓ |
| Plan validation (C-core) | | ✓ | ✓ | ✓ |
| Session logging | | | ✓ | ✓ |
| Pattern extraction | | | ✓ | ✓ |
| Multi-agent coordination | | | | ✓ |
| Meta-learning loop | | | | ✓ |

## Comparison with GiT Principles

| GiT | Mimic Autonomy |
|-----|---------------|
| Universal text tokens | Markdown-native slots + JSON-RPC tools |
| Multi-task joint training | Cross-domain mesh edges + session learning |
| Auto-regressive generation | LLM plan generation + OpPacket validation |
| Zero-shot generalization | Proactive suggestions from context |
| All tasks benefit each other | Successful plans become reusable slots |

## Milestones

- **v0.4 (2 weeks):** Proactive suggestions, text-native slots
- **v0.5 (1 month):** Compositional planning, HNSW index
- **v0.6 (2 months):** Self-improving mesh, session learning
- **v0.7 (3+ months):** Autonomous agent, meta-learning

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Mesh grows too fast | RAM explosion | Tiered storage (hot/warm/cold) |
| Autonomy makes wrong decisions | Production damage | 2-vote gate, dry-run mode |
| Session data sensitivity | Privacy leak | Anonymization, local-only default |
| Complexity explosion | Unmaintainable | ADR for every stage, rollback plan |
