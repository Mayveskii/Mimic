# Security Domain — Permissions, Never-Rules, DIFC

How Mimic enforces safety boundaries on all operations.

---

## What This Domain Does

Security is not an afterthought. Every operation passes through a permission pipeline before execution. Hard rules (never-rules) cannot be overridden. DIFC ensures information flows only in authorized directions. The model cannot accidentally perform dangerous operations — they are blocked by design.

The domain covers: permission pipeline, never-rules, DIFC, circuit breaker, credential management.

---

## Processes

### permission_pipeline

**When to use:**  
Before ANY operation execution. Every tool call passes through this pipeline.

**Goal:**  
Determine whether an operation is allowed, denied, or requires explicit confirmation.

**Chain (semantically):**

1. **Deny rules** (hardcoded blocks).
   - Check against never-rules set.
   - Match → hard stop, no override.

2. **Auto-classify** (AI risk assessment).
   - Classifier evaluates operation description.
   - Returns: {risk: safe|ask|deny, confidence: 0-1}.
   - Confidence < 0.7 → escalate to manual ask.

3. **Budget check**.
   - Verify operation cost ≤ remaining budget.
   - Exceeds → block, notify model.

4. **Allow rules** (session-approved ops).
   - Check if operation is in session allow-list.
   - In list → auto-approve.
   - Not in list → ask model.

**Hard constraints:**
- Never-rules cannot be overridden by ANY permission mode.
- OP_FLAG_DANGEROUS always requires explicit allow.
- Deny rules checked FIRST — no bypass.

**Invariants:**
- Every operation checked against deny rules.
- Every operation has classification result logged.
- Every deny increments denial counters.

**Result:**
```
decision: "allow" | "ask" | "deny"
reason: "never_rule: git_reset" | "auto_classify: safe" | "budget_exceeded"
confidence: 0.92  # for auto-classify decisions
```

---

### never_rules

**When to use:**  
Hardcoded rules that apply to ALL operations, ALL models, ALL sessions.

**Goal:**  
Prevent catastrophic operations regardless of permissions.

**Rule set (from bun):**
- Never `git reset`.
- Never `git checkout` used destructively.
- Never `git restore`.
- Never `git stash`.
- Never `git rebase`.
- Never edit files outside workspace.
- Never expose credentials in output.
- Never delete `.git` directory.

**Hard constraints:**
- Never-rules are HARD CODED. Cannot be changed by model, by config, by any permission.
- Violation → hard stop, operation rejected, model notified, logged.

**Invariants:**
- Rule set is immutable at runtime.
- Every violation is logged with context.

---

### difc_pipeline

**When to use:**  
When Mimic serves multiple agents or connects to mimic-server.

**Goal:**  
Ensure information flows only from higher clearance to lower. Never reverse.

**Chain (semantically):**

1. Label agent with clearance level.
2. Label resource with classification level.
3. Check: clearance ≥ classification?
4. If pass → execute operation.
5. Label response with output classification.
6. Filter response: remove information agent must not see.

**Hard constraints:**
- Information flows only from ≥ clearance to ≤ clearance.
- Phase 3 denial → operation NEVER executed.
- No bypass of clearance check.

**Invariants:**
- Response never contains information above agent clearance.
- Filtered fields logged (what was removed and why).

---

## Principles From Sources

### bun

**Principles taken:**
- Never-rules: hardcoded deny set, no override.
- Phase graph: operations live inside gated pipeline.

**What Mimic does with them:**
Exact never-rules. Exact phase gating.

### code-mode

**Principles taken:**
- Permission pipeline: deny → classify → budget → allow.
- Denial tracking: 3 consecutive → circuit break.

**What Mimic does with them:**
Exact pipeline order. Exact thresholds.

### gh-aw-mcpg

**Principles taken:**
- DIFC 6-phase: label → check → execute → label response → filter.
- Circuit breaker: per-backend 3-state with 30s probe.
- OAuth PKCE: authorization with code_verifier.

**What Mimic does with them:**
DIFC for multi-agent. Circuit breaker per tool/resource. PKCE for auth.

### hermes-agent

**Principles taken:**
- Credential pool: multi-key rotation, env fallback.
- Tool guardrails: per-turn reset, OP_FLAG_DANGEROUS blocked.

**What Mimic does with them:**
Credentials never in source. Dangerous flags blocked by default.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "security" |
| layer | "process" |
| modality | "code" |
| pattern_name | "permission_pipeline" / "never_rules" / "difc_pipeline" |

---

## Cross-Domain Conflicts

Security domain applies to ALL other domains. No operation exempt.
