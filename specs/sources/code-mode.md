```yaml
repo: Mayveskii/code-mode
url: https://github.com/Mayveskii/code-mode
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of Claude Code SDK (code-mode). AI coding agent with permission pipeline,
  budget enforcement, denial tracking, auto-classifier, concurrency control, and
  structured output retry with schema-guided reprompting.

advantages:
  - id: cm_permission_pipeline
    what: Permission pipeline: deny_rules (hardcoded blocks) → ask (user confirmation) → bypass (auto-approve for safe ops) → allow_rules; 7 permission modes (default, auto, safe, etc.)
    evidence: "src/ — permission system: deny list checked first, then AI classifier, then budget check, then allow list"

  - id: cm_denial_tracking
    what: Denial tracking: 3 consecutive denies → circuit break to manual mode, 20 total denies per session → permanent fallback to manual; prevents permission fatigue
    evidence: "src/ — denial counter: consecutive_denial_count, total_denial_count; circuit break when thresholds hit"

  - id: cm_auto_classifier
    what: AI auto-classifier evaluates operation risk: safe (auto-allow) / ask (user confirm) / deny (blocked); confidence < threshold → escalate to manual ask
    evidence: "src/ — auto mode permission classifier: LLM evaluates operation description → returns risk level + confidence"

  - id: cm_budget_enforcement
    what: Budget enforcement: maxTurns (iteration limit) + maxBudgetUsd (cost limit); track cumulative spend per session; stop when either limit hit
    evidence: "src/ — budget tracking: remaining_budget = initial - Σ(cost_per_turn); stop execution when remaining ≤ 0"

  - id: cm_concurrency_control
    what: Concurrency: up to 10 tools in parallel with concurrency-safe/serial classification; serial tools queued, safe tools run concurrently; semaphore enforces limit
    evidence: "src/ — concurrency: semaphore(10), concurrency_safe flag per tool, serial tools queued sequentially"

  - id: cm_structured_output_retry
    what: Structured output retry: when model returns malformed JSON or invalid tool calls → reprompt with schema + error description → retry up to 5 times → fallback to manual escalation
    evidence: "src/ — retry logic: parse error → extract schema violation → reprompt with 'Your output was invalid because... Schema: {schema}' → retry; max_retries=5"

applications:
  - advantage_id: cm_permission_pipeline
    implemented_in: internal/orchestrator/permission.go
    mechanism: "4-stage pipeline: (1)deny_rules check (hardcoded blocks) → (2)AI auto-classify (safe/ask/deny + confidence) → (3)budget_check (cost remaining) → (4)allow_rules (session-approved ops)"
    invariant: "OP_FLAG_DANGEROUS always requires explicit allow regardless of pipeline stage. Deny rules checked first — no override possible."
    status: planned

  - advantage_id: cm_denial_tracking
    implemented_in: internal/quality/denial.go
    mechanism: "Counter per session: consecutive_denial_count++ on each deny → if ≥ 3 → circuit break to manual; total_denial_count++ → if ≥ 20 → permanent manual mode"
    invariant: "Circuit break resets only on explicit user approval. Permanent manual mode cannot be overridden within session."
    status: planned

  - advantage_id: cm_auto_classifier
    implemented_in: internal/orchestrator/permission.go
    mechanism: "LLM classifier evaluates operation: input=operation_description → output={risk: safe|ask|deny, confidence: 0-1} → if confidence < 0.7 → escalate to manual ask"
    invariant: "Auto-classifier confidence < 0.7 → always escalate to manual. Deny from classifier = hard deny. Safe only auto-approved if confidence ≥ 0.7."
    status: planned

  - advantage_id: cm_budget_enforcement
    implemented_in: internal/orchestrator/budget.go
    mechanism: "Per-session budget: remaining_usd = initial_usd - Σ(token_cost_per_turn); remaining_turns = max_turns - current_turn; stop when either ≤ 0"
    invariant: "Budget exceeded → no more executions. Agent notified of budget state before each turn. Both limits enforced independently."
    status: planned

  - advantage_id: cm_concurrency_control
    implemented_in: internal/orchestrator/concurrency.go
    mechanism: "Semaphore(10): concurrency_safe tools → acquire semaphore → run parallel; serial tools → queue → run one at a time; all slots full → wait"
    invariant: "Never more than 10 concurrent operations. Serial ops never run in parallel with each other. Safe ops can run in parallel."
    status: planned

  - advantage_id: cm_structured_output_retry
    implemented_in: internal/orchestrator/retry.go
    mechanism: "Parse model output → if invalid JSON or schema violation → extract error → reprompt: 'Invalid because: {error}. Expected schema: {schema}. Try again.' → retry (max 5) → manual escalation"
    invariant: "Max 5 retries. Each retry includes specific error description. After 5 failures → manual escalation with error history. No silent retries."
    status: planned

control:
  - advantage_id: cm_permission_pipeline
    verification: "Unit test: dangerous op without allow → denied at stage 1; safe op with high confidence → auto-approved at stage 2; budget exhausted → blocked at stage 3"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_denial_tracking
    verification: "Unit test: 3 consecutive denies → verify circuit break; 20 total denies → verify permanent manual; user approval → verify consecutive counter reset"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_auto_classifier
    verification: "Unit test: low-risk op with high confidence → auto-allow; high-risk op → auto-ask; confidence=0.5 → escalate to manual"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_budget_enforcement
    verification: "Unit test: set budget=10, execute ops costing 12 → verify stop at budget limit; maxTurns=3 → verify stop after 3 turns regardless of cost"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_concurrency_control
    verification: "Stress test: submit 15 concurrent safe ops → verify max 10 running at once; submit 2 serial ops → verify sequential execution"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_structured_output_retry
    verification: "Unit test: return malformed JSON 5 times → verify retry with error description each time → verify manual escalation after 5th; valid output on 3rd try → verify accepted"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never
```
