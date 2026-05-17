```yaml
repo: Mayveskii/bun
url: https://github.com/Mayveskii/bun
language: Zig, C++, JavaScript
status: partial
last_sync: "2025-05-17"

description: |
  Fork of Oven-sh/bun for PR #30412. Contains 26 workflow files for
  Zig→Rust porting with ~170 parallel agents. Phase-gated orchestration,
  2-vote adversarial verification, edit scope isolation, and lifetime
  classification with 3-vote refute + 12% random audit sampling.

advantages:
  - id: bun_phase_graph
    what: Phase graph orchestration with gate transitions between 6 phases (CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND); no phase skips, rollback on failure
    evidence: "PR #30412: .github/workflows/phase-*.yml (26 files) — phase-classify.yml, phase-plan.yml, phase-validate.yml, phase-exec.yml, phase-verify.yml, phase-respond.yml"

  - id: bun_two_vote
    what: 2-vote adversarial verification — two independent verifiers with different models/prompts + tiebreak mechanism for critical operations
    evidence: "PR #30412: verify workflows with verify-A/verify-B pattern, tiebreak.yml for disagreement resolution"

  - id: bun_edit_scope_isolation
    what: Each agent edits only its assigned crate — conflict matrix prevents cross-contamination; resource_bitmask per operation detects overlap
    evidence: "PR #30412: scope rules in workflow files, 'each agent edits ONLY its assigned crate', shard-to-crate mapping in matrix strategy"

  - id: bun_pre_post_hooks
    what: PreToolUse/PostToolUse middleware chain: pre-bash-zig-build.js validates build commands, post-edit-zig-format.js auto-formats after edits
    evidence: "PR #30412: settings.json PreToolUse/PostToolUse hooks, pre-bash-zig-build.js, post-edit-zig-format.js"

  - id: bun_never_rules
    what: Never-rules: never git reset/checkout/restore/stash/rebase, never edit .zig (only .rs), never Box::leak, never re-gate completed phase
    evidence: "PR #30412: workflow rules across all 26 files — deny_rules section in each workflow"

  - id: bun_lifetime_classify
    what: Operation lifecycle classification using 3-vote refute: 3 independent classifiers vote on operation type + 12% random sample audit for quality
    evidence: "PR #30412: lifetime-classify workflow with 3 classifier jobs + audit-sample job"

  - id: bun_parallel_agents
    what: ~170 agents running simultaneously with scoped isolation; matrix strategy splits work into shards per crate; each shard has independent context + scope
    evidence: "PR #30412: workflow matrix strategy with shard assignments, strategy.matrix.shard=[0..169], concurrency groups per shard"

applications:
  - advantage_id: bun_phase_graph
    implemented_in: internal/orchestrator/classify.go, plan.go, validate.go, exec.go, verify.go, respond.go
    mechanism: "6-phase state machine: CLASSIFY(intent analysis) → PLAN(strategy) → VALIDATE(pre-flight checks) → EXEC(implementation) → VERIFY(adversarial) → RESPOND(output); each phase gates entry to next; rollback on any failure"
    invariant: "No EXEC without passed VALIDATE. No phase skips. Failed phase → rollback to last checkpoint, not restart."
    status: planned

  - advantage_id: bun_two_vote
    implemented_in: internal/quality/verify.go
    mechanism: "Two independent verifiers (different model/prompt) vote pass/fail → if consensus → accept; if disagree → tiebreak (third verifier or human)"
    invariant: "Critical operations (git push, deploy, encrypt) always undergo 2-vote. Single verifier pass = insufficient for critical ops."
    status: planned

  - advantage_id: bun_edit_scope_isolation
    implemented_in: core/conflict_matrix (resource_bitmask per operation)
    mechanism: "Each agent assigned resource_bitmask(crates) → before write: check bitmask overlap with in-flight writes → overlap → serialize, no overlap → parallel"
    invariant: "No two pipelines write to same resource simultaneously. Bitmask overlap detection is O(1) via AND operation."
    status: planned

  - advantage_id: bun_pre_post_hooks
    implemented_in: internal/mcp/middleware.go
    mechanism: "Middleware chain: PreToolUse(block dangerous, validate commands) → execute tool → PostToolUse(auto-format, validate output); hooks configured per tool type"
    invariant: "OP_FLAG_DANGEROUS always blocked by PreToolUse unless explicit allow. PostToolUse runs even if tool partially fails."
    status: planned

  - advantage_id: bun_never_rules
    implemented_in: internal/orchestrator/permission.go
    mechanism: "Hardcoded deny_rules set: {git reset, git checkout, git restore, git stash, git rebase, .zig edits, Box::leak, re-gate} → checked before every operation"
    invariant: "Never-rules cannot be overridden by any permission mode. Violation → hard stop, not warning."
    status: planned

  - advantage_id: bun_lifetime_classify
    implemented_in: internal/orchestrator/classify.go
    mechanism: "3-vote refute classification: 3 independent classifiers vote on operation type → majority wins → 12% random sample sent to audit queue for quality verification"
    invariant: "Classification confidence ≥ 2/3 votes required. Audit sample rate = 12%. Audit failure → re-classify entire batch."
    status: planned

  - advantage_id: bun_parallel_agents
    implemented_in: internal/orchestrator/concurrency.go
    mechanism: "Matrix strategy: work split into shards per crate → each shard = agent with isolated context + scope → concurrency group per shard → max 10 concurrent pipelines"
    invariant: "resource_bitmask overlap → serialize. No overlap → parallel. Max 10 concurrent pipelines. Each pipeline has independent scope."
    status: planned

control:
  - advantage_id: bun_phase_graph
    verification: "Integration test: submit chain → verify all 6 phases executed in order; skip VALIDATE → verify blocked"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_two_vote
    verification: "Unit test: submit critical op → verify 2 independent checks; force disagreement → verify tiebreak path"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_edit_scope_isolation
    verification: "Unit test: two pipelines with overlapping crates → conflict detected; non-overlapping crates → parallel execution"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_pre_post_hooks
    verification: "Integration test: submit dangerous op → verify blocked by PreToolUse; submit safe edit → verify PostToolUse auto-format runs"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_never_rules
    verification: "Unit test: submit 'git reset' → verify denied regardless of permission mode; submit 'git commit' → verify allowed"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_lifetime_classify
    verification: "Unit test: submit ambiguous intent → verify 3-vote classification; verify 12% audit sampling rate over 100 operations"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never

  - advantage_id: bun_parallel_agents
    verification: "Stress test: 170 shards → verify max 10 concurrent pipelines; verify no scope leakage between shards"
    update_trigger: "Re-analyze when bun PR #30412 updates or merges"
    last_verified: never
```
