# Required Data Pool — Identified by 3-Tier Battlefield Benchmark

> **Date:** 2026-05-18  
> **Test:** 8 tasks across 3 tiers, kimi k2.6 via OpenRouter  
> **Cost:** $0.010 / $2.00 budget (0.5% spent)  
> **Result:** 4 pass, 2 partial, 2 fail

---

## Executive Summary

The benchmark revealed **exactly what data is missing** for Mimic to handle real-world tasks. Not guesses — measured gaps based on actual model behavior.

| Tier | Status | What Worked | What's Missing |
|------|--------|-------------|----------------|
| **1 — Simple** | ✅ 4/4 PASS | JSON Schema prevents argument collisions | Nothing critical |
| **2 — Medium** | ❌ 0/2 FAIL | Model uses correct single tools | **Tool chaining / decomposition metadata** |
| **3 — Complex** | ⚠️ 0/2 PARTIAL | Model reads files | **Multi-turn loop, context compression rules, project archetypes** |

---

## Detailed Findings

### Finding 1: Tier 2 Complete Failure — No Decomposition

**What happened:**
- Task: "Check git status, then list recent commits, then read Makefile"
- Model did: Called `GIT_STATUS` once
- Model should have: Called `GIT_STATUS` → `GIT_LOG` → `IO_READ` sequentially

**Root cause:**
The model (kimi k2.6) does **not** automatically decompose multi-step tasks in a single API call. It makes one tool call and expects the caller to feed the result back for the next decision.

**This is expected behavior for LLMs.** They don't internally loop — they need external orchestration.

**Required data to fix:**

```yaml
# mesh_slots/task_decomposition_patterns.json
{
  "domain": "task_orchestration",
  "pattern": {
    "trigger": "analyze_repo_state",
    "steps": [
      {"tool": "GIT_STATUS", "why": "current_state"},
      {"tool": "GIT_LOG", "args": {"limit": 5}, "why": "history"},
      {"tool": "IO_READ", "args": {"path": "Makefile"}, "why": "build_system"}
    ],
    "dependencies": "sequential",
    "output": "summary"
  },
  "source": "bun#30412_phase_graph",
  "survival_index": 0.92
}
```

**Community contribution:** Add 50+ task decomposition patterns to mesh covering:
- Repository analysis (git status → log → read config)
- Build verification (compile → test → check binary)
- Debugging (reproduce → log → analyze → fix)
- Refactoring (scan → identify → modify → verify)

---

### Finding 2: Tier 3 Partial — No Actionable Output

**What happened:**
- Task: "Analyze core/ops.c error handling and suggest 3 improvements"
- Model did: Read the file (IO_READ), then stopped
- Model should have: Read → analyze patterns → suggest concrete changes

**Root cause:**
Two issues:
1. **No multi-turn loop:** After reading file, model didn't get result back to continue analysis
2. **No "actionable output" pattern in mesh:** Model doesn't know what "good" analysis looks like

**Required data to fix:**

```yaml
# mesh_slots/code_review_patterns.json
{
  "domain": "code_review",
  "pattern": {
    "trigger": "improve_error_handling",
    "template": "Review {file} for {aspect}. Identify 3 specific improvements: (1) ..., (2) ..., (3) ...",
    "examples": [
      {
        "input": "core/ops.c error handling",
        "output": "1. Add context to ERR_EXEC_FAIL (include errno)... 2. Log rollback failures... 3. Validate inverse opcodes before rollback..."
      }
    ]
  },
  "source": "hermes-agent_closed_learning_loop",
  "survival_index": 0.88
}
```

**Community contribution:** Add 100+ code analysis templates with input/output examples.

---

### Finding 3: RTK Compression Never Triggered

**What happened:**
- No task produced output >5KB
- RTK compression wasn't stress-tested

**Required data to fix:**

```yaml
# mesh_slots/rtk_test_vectors.json
{
  "domain": "compression",
  "pattern": {
    "trigger": "large_output",
    "thresholds": {
      "code": 50,
      "log": 100,
      "json": 20,
      "diff": 30
    },
    "test_vectors": [
      {"type": "git_log", "lines": 500, "expected_reduction": "95%"},
      {"type": "build_output", "lines": 200, "expected_reduction": "90%"},
      {"type": "json_array", "items": 1000, "expected_reduction": "80%"}
    ]
  }
}
```

**Community contribution:** Provide real large outputs from your projects for RTK stress testing.

---

## Priority Matrix for Community

| Priority | Data Needed | Effort | Impact | Who Can Help |
|----------|-------------|--------|--------|--------------|
| 🔴 **P0** | Task decomposition patterns | Medium | **High** — unlocks Tier 2 | Anyone with build/debug workflows |
| 🔴 **P0** | Multi-turn orchestrator loop | Hard | **High** — core feature | Go developers |
| 🟡 **P1** | Code review templates (input/output) | Medium | Medium | Senior developers |
| 🟡 **P1** | RTK test vectors (>5KB outputs) | Low | Medium | Anyone with large repos |
| 🟢 **P2** | Language-specific slots (Go, Rust, Python) | Medium | Low | Language specialists |
| 🟢 **P2** | Project archetype configs | Low | Low | DevOps/SRE |

---

## What We Fixed During Benchmark

| Bug | Location | Fix | Impact |
|-----|----------|-----|--------|
| SYS_FILE_EXISTS returned empty result | `core/ops.c:895` | Write `exists: true/false` to `packet->result` | Tier 1 now 4/4 PASS |
| JSON Schema field name | `test/battlefield/three_tier_benchmark.py` | `inputSchema` → `parameters` for OpenAI compatibility | Arguments now correct |

---

## Technical Metrics (Raw)

```json
{
  "total_tasks": 8,
  "tier_1": {"pass": 4, "fail": 0, "partial": 0},
  "tier_2": {"pass": 0, "fail": 2, "partial": 0},
  "tier_3": {"pass": 0, "fail": 0, "partial": 2},
  "total_cost_usd": 0.01,
  "total_tokens": 16038,
  "tokens_per_dollar": 1606370,
  "argument_collisions": 0,
  "avg_latency_ms": 8612
}
```

---

## Call to Action

**If you want to contribute:**

1. **Easiest:** Add your build/test workflow as a decomposition pattern
   ```bash
   # Example: your daily workflow
   git status → git log → cat Makefile → make test
   ```
   → Open PR to `data/seeds/community-workflows.json`

2. **Medium:** Share a code review you did recently
   → We'll distill it into a mesh slot

3. **Hardest:** Implement multi-turn loop in `internal/orchestrator/`
   → See `docs/adr/` for architecture decisions

**All contributions measured.** Every pattern gets survival index. Only proven patterns survive.

---

*This document is generated by the battlefield benchmark. Re-run with:*
```bash
python3 test/battlefield/three_tier_benchmark.py
```
