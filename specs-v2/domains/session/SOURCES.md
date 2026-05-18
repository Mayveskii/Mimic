# Session Domain — Sources

Where the session domain behavior comes from.

---

## code-mode

**Principles taken:**
- Budget enforcement: maxTurns + maxBudgetUsd per session.
- Iteration counting: each operation decrements budget.

**What Mimic does with them:**
Token and time budget tracked per session. Operations consume energy. Budget exhaustion halts session.

**What Mimic does NOT copy:**
- code-mode's USD-based pricing (Mimic uses abstract tokens).
- code-mode's specific turn counting.

---

## hermes-agent

**Principles taken:**
- Context compression pipeline: preflight detection + multi-pass compression + cache-aware system prompt.
- Iteration budget: consume() gates loop.

**What Mimic does with them:**
Context compressed when budget pressure detected. Stable prefix preserved (system prompt). Budget gates execution loop.

**What Mimic does NOT copy:**
- hermes-agent's Python implementation.
- hermes-agent's specific compression algorithm (Mimic uses standard gzip + semantic truncation).

---

## embryo

**Principles taken:**
- Session tracking: state tracked across operations.
- Tool loop: session state gates tool execution.

**What Mimic does with them:**
Session context tracks all operations, results, denials. Tool execution gated by session state.

---

## gastown

**Principles taken:**
- ZFC state: observable reality over cached assumptions.
- Pressure gating: system load gates dispatch.

**What Mimic does with them:**
Session state reflects observable reality. Budget pressure gates operation dispatch.

---

## Standard Practice

**Principles taken:**
- Session isolation: independent contexts.
- Snapshot/restore: save and recover state.
- Circuit breaker: failure counting.

**What Mimic does with them:**
Standard session management applied.
