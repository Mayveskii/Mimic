```yaml
repo: Mayveskii/bun
url: https://github.com/Mayveskii/bun
language: Zig, C++, JavaScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of Oven-sh/bun for PR #30412. Contains 26 workflow files for
  Zig→Rust porting with ~170 parallel agents.

advantages:
  - id: bun_phase_graph
    what: Phase graph orchestration with gate transitions between phases
    evidence: "PR #30412: .github/workflows/phase-*.yml (26 files)"

  - id: bun_two_vote
    what: 2-vote adversarial verification — two independent verifiers + tiebreak
    evidence: "PR #30412: verify workflows, 2-vote pattern in workflow rules"

  - id: bun_edit_scope_isolation
    what: Each agent edits only its assigned crate — no cross-contamination
    evidence: "PR #30412: scope rules in workflow files, 'each agent edits ONLY its assigned crate'"

  - id: bun_pre_post_hooks
    what: PreToolUse/PostToolUse middleware chain for blocking dangerous operations and auto-formatting
    evidence: "PR #30412: settings.json PreToolUse/PostToolUse hooks, pre-bash-zig-build.js, post-edit-zig-format.js"

  - id: bun_never_rules
    what: Never-rules: never git destructive, never edit .zig, never Box::leak, never re-gate
    evidence: "PR #30412: workflow rules across all 26 files"

  - id: bun_lifetime_classify
    what: Operation lifecycle classification using 3-vote refute + 12% random sample
    evidence: "PR #30412: lifetime-classify workflow"

  - id: bun_parallel_agents
    what: ~170 agents running simultaneously with scoped isolation
    evidence: "PR #30412: workflow matrix strategy, shard assignments"

applications:
  - advantage_id: bun_phase_graph
    implemented_in: internal/orchestrator/classify.go, plan.go, validate.go, exec.go, verify.go, respond.go
    mechanism: "6-phase state machine: CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND with rollback on failure"
    invariant: "No EXEC without passed VALIDATE. No phase skips."
    status: planned

  - advantage_id: bun_two_vote
    implemented_in: internal/quality/verify.go
    mechanism: "vote(executor_A, executor_B) → consensus or tiebreak"
    invariant: "Critical operations (git push, deploy, encrypt) always undergo 2-vote."
    status: planned

  - advantage_id: bun_edit_scope_isolation
    implemented_in: core/conflict_matrix (resource_bitmask per operation)
    mechanism: "Conflict matrix extended with resource scope — overlapping scopes → conflict"
    invariant: "No two pipelines write to same resource simultaneously."
    status: planned

  - advantage_id: bun_pre_post_hooks
    implemented_in: internal/mcp/middleware.go
    mechanism: "Middleware chain: PreToolUse (block dangerous) → execute → PostToolUse (auto-format)"
    invariant: "OP_FLAG_DANGEROUS always blocked by PreToolUse unless explicit allow."
    status: planned

  - advantage_id: bun_never_rules
    implemented_in: internal/orchestrator/permission.go
    mechanism: "Hardcoded deny_rules: never git reset/checkout/restore/stash/rebase, never delete protected"
    invariant: "Never-rules cannot be overridden by any permission mode."
    status: planned

  - advantage_id: bun_lifetime_classify
    implemented_in: internal/orchestrator/classify.go
    mechanism: "3-vote refute classification: 3 independent classifiers vote, majority wins, 12% random audit"
    invariant: "Classification confidence ≥ 2/3 votes required."
    status: planned

  - advantage_id: bun_parallel_agents
    implemented_in: internal/orchestrator/concurrency.go
    mechanism: "Up to 10 concurrent pipelines with scope isolation and conflict matrix check"
    invariant: "resource_bitmask overlap → serialize. No overlap → parallel."
    status: planned

control:
  - advantage_id: bun_phase_graph
    verification: "Integration test: submit chain → verify all 6 phases executed in order"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_two_vote
    verification: "Unit test: submit critical op → verify 2 independent checks + tiebreak path"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_edit_scope_isolation
    verification: "Unit test: two pipelines with overlapping files → conflict detected"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_pre_post_hooks
    verification: "Integration test: submit dangerous op → verify blocked by PreToolUse"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_never_rules
    verification: "Unit test: submit 'git reset' → verify denied regardless of permission mode"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_lifetime_classify
    verification: "Unit test: submit ambiguous intent → verify 3-vote classification"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_parallel_agents
    verification: "Stress test: 10 concurrent pipelines → verify isolation + no conflicts"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never
```
