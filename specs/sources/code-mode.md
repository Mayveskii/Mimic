```yaml
repo: Mayveskii/code-mode
url: https://github.com/Mayveskii/code-mode
language: TypeScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of Claude Code SDK (code-mode). AI coding agent with permission pipeline,
  budget enforcement, denial tracking, auto-classifier, and concurrency control.

advantages:
  - id: cm_permission_pipeline
    what: Permission pipeline: deny → ask → bypass → allow rules, 7 permission modes
    evidence: "src/ — permission system implementation"

  - id: cm_denial_tracking
    what: Denial tracking: 3 consecutive denies → circuit break, 20 total → fallback to manual
    evidence: "src/ — denial counter and circuit break logic"

  - id: cm_auto_classifier
    what: AI auto-classifier decides allow/deny based on operation risk level
    evidence: "src/ — auto mode permission classifier"

  - id: cm_budget_enforcement
    what: Budget enforcement: maxTurns, maxBudgetUsd — execution cost limits
    evidence: "src/ — budget tracking and enforcement"

  - id: cm_concurrency_control
    what: Concurrency: up to 10 tools in parallel (concurrency-safe vs serial)
    evidence: "src/ — concurrency implementation, 10 parallel limit"

  - id: cm_structured_output_retry
    what: Structured output retry: limit=5 retries for malformed model outputs
    evidence: "src/ — retry logic for structured output parsing"

applications:
  - advantage_id: cm_permission_pipeline
    implemented_in: internal/orchestrator/permission.go
    mechanism: "4-stage pipeline: deny_rules → classify(auto AI) → budget_check → allow_rules"
    invariant: "OP_FLAG_DANGEROUS always requires explicit allow regardless of pipeline stage."
    status: planned

  - advantage_id: cm_denial_tracking
    implemented_in: internal/quality/denial.go
    mechanism: "Counter per session: 3 consecutive denies or 20 total → circuit break to manual mode"
    invariant: "Circuit break resets only on explicit user approval."
    status: planned

  - advantage_id: cm_auto_classifier
    implemented_in: internal/orchestrator/permission.go
    mechanism: "AI classifier evaluates operation: safe/ask/deny → routes to permission pipeline"
    invariant: "Auto-classifier confidence < threshold → escalate to manual ask."
    status: planned

  - advantage_id: cm_budget_enforcement
    implemented_in: internal/orchestrator/budget.go
    mechanism: "Per-session budget: remaining = initial - Σ(cost_tokens). Stop when remaining ≤ 0."
    invariant: "Budget exceeded → no more executions. Agent notified of budget state."
    status: planned

  - advantage_id: cm_concurrency_control
    implemented_in: internal/orchestrator/concurrency.go
    mechanism: "Semaphore: max 10 concurrent operations. Concurrency-safe ops parallel, serial ops queued."
    invariant: "Never more than 10 concurrent operations. Serial ops never parallel."
    status: planned

  - advantage_id: cm_structured_output_retry
    implemented_in: internal/orchestrator/retry.go
    mechanism: "If model returns malformed chain → retry parse up to 5 times → fallback to manual"
    invariant: "Max 5 retries. After 5 failures → manual escalation."
    status: planned

control:
  - advantage_id: cm_permission_pipeline
    verification: "Unit test: dangerous op without allow → denied; with allow → passed"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_denial_tracking
    verification: "Unit test: 3 consecutive denies → verify circuit break triggered"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_auto_classifier
    verification: "Unit test: low-risk op → auto-allow; high-risk op → auto-ask"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_budget_enforcement
    verification: "Unit test: set budget=10, execute ops costing 12 → verify stop at budget limit"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_concurrency_control
    verification: "Stress test: submit 15 concurrent ops → verify max 10 running at once"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never

  - advantage_id: cm_structured_output_retry
    verification: "Unit test: return malformed output 5 times → verify manual escalation"
    update_trigger: "Re-analyze when code-mode releases new version"
    last_verified: never
```
