# Orchestrator Domain — Sources

Where the orchestrator domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- Pipeline: state→mesh→DIRECT→classify→exec→flywheel→respond. Each stage enriches context.
- Rollback: failed stage receives error context, does not restart from beginning.
- Context flow: cumulative, forward-only. Later phases see everything from earlier phases.

**What Mimic does with them:**
6-phase pipeline with explicit rollback paths. Context accumulates per session.

**What Mimic does NOT copy:**
- Embryo's specific pipeline implementation (Go channels, goroutines).
- Embryo's flywheel optimization (Mimic uses simpler feedback loop).

---

## bun (PR #30412)

**Principles taken:**
- Phase graph: 6 phases with gate transitions. No skips.
- 2-vote adversarial verify: two independent verifiers for critical operations.
- Edit scope isolation: resource_bitmask per operation prevents cross-contamination.
- Pre/post hooks: PreToolUse/PostToolUse middleware chain.
- Never-rules: hardcoded deny set, no override.

**What Mimic does with them:**
All 6 phases mandatory. Never-rules enforced before every operation. 2-vote on critical ops. Hooks on every tool call.

**What Mimic does NOT copy:**
- Bun's crate-based architecture.
- Bun's specific hook registration API.

---

## gastown

**Principles taken:**
- ZFC state: observable reality over cached assumptions.
- Event-driven convoy: completion event triggers dependency-aware dispatch.
- Pressure gating: system load gates dispatch.
- Help classification: triage incoming requests by severity.

**What Mimic does with them:**
Every phase checks observable reality. Completion events trigger next tasks. Pressure gates prevent overload.

**What Mimic does NOT copy:**
- Gastown's actor-based concurrency model.
- Gastown's specific event bus implementation.

---

## graphify

**Principles taken:**
- AST extraction: two-pass (structural + call-graph) with confidence labels.
- IDF-weighted search: exact > prefix > substring with gap-ratio cutoff.
- Hub-throttled traversal: skip high-degree hubs as transit.

**What Mimic does with them:**
Pattern retrieval uses 5-signal ranking. High-confidence patterns preferred. Hub nodes skipped in traversal.

**What Mimic does NOT copy:**
- Graphify's specific AST parser.
- Graphify's graph database backend.

---

## code-mode

**Principles taken:**
- Budget enforcement: maxTurns + maxBudgetUsd per session.
- Concurrency control: up to 10 tools in parallel.

**What Mimic does with them:**
Energy budget enforced. Parallel pipelines limited to 10.

---

## hermes-agent

**Principles taken:**
- Closed learning loop: agent creates skills from experience.
- Context compression pipeline: preflight detection + multi-pass compression.

**What Mimic does with them:**
Novel patterns recorded as feedback. Context compressed automatically when budget pressure detected.

---

## Standard Patterns

**Principles taken:**
- Circuit breaker: 3 failures → open circuit.
- Retry with exponential backoff.
- Timeout enforcement.

**What Mimic does with them:**
Standard reliability patterns applied to orchestration.
