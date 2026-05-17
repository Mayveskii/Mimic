```yaml
repo: Mayveskii/gastown
url: https://github.com/Mayveskii/gastown
language: Go
status: partial
last_sync: "2025-05-17"

description: |
  Fork of gastownhall/gastown (15K stars). Multi-agent orchestration system coordinating
  20-50+ AI coding agents (Claude Code, Copilot, Codex, Gemini, OpenCode, Cursor) with
  git-backed persistent work state, 3-tier watchdog, Bors-style merge queue, and Dolt SQL.

advantages:
  - id: gt_zfc_state
    what: Zero False Confidence — tmux session is source of truth, not state files; derive state from observable reality
    evidence: "internal/witness/manager.go — IsHealthy() checks tmux; internal/polecat/manager.go — state from beads assignee, not state.json"

  - id: gt_watchdog_chain
    what: 3-tier hierarchical watchdog: Daemon(3min heartbeat, 15+ checks) → Boot(triage) → Deacon(patrol) → Witness(per-rig)
    evidence: "internal/daemon/daemon.go — heartbeat() with 15 checks; internal/boot/; internal/deacon/; internal/witness/"

  - id: gt_help_classification
    what: 6-category + 3-severity help triage with auto-routing (Emergency→Overseer, Failed→Deacon, Blocked→Mayor)
    evidence: "internal/witness/protocol.go — AssessHelp(), HelpCategory enum, HelpSeverity enum"

  - id: gt_event_driven_convoy
    what: Reactive work distribution: completion event → feedNextReadyIssue with dependency-aware dispatch (blocks, waits-for, merge-blocks)
    evidence: "internal/convoy/operations.go — CheckConvoysForIssue(), feedNextReadyIssue(), isIssueBlocked()"

  - id: gt_atomic_allocation
    what: AllocateAndAdd with flock + pending marker prevents TOCTOU name races between concurrent agents
    evidence: "internal/polecat/manager.go — AllocateAndAdd() with pool lock + <name>.pending file"

  - id: gt_rollback_on_failure
    what: Transactional resource creation: track sequence of worktree/beads/directory → cleanupOnError rolls back all
    evidence: "internal/polecat/manager.go — AddWithOptions() with cleanupOnError closure"

  - id: gt_pressure_gating
    what: Capacity-aware dispatch: checkPressure() gates refinery/polecat/dog spawning when system loaded
    evidence: "internal/daemon/daemon.go — checkPressure() calls before dispatch"

  - id: gt_lifecycle_state_machine
    what: 6-state polecat lifecycle (Working/Idle/Done/Stuck/Stalled/Zombie) with CleanupStatus tracking
    evidence: "internal/polecat/types.go — State enum, CleanupStatus enum"

  - id: gt_inter_agent_protocol
    what: 14 typed message protocol (POLECAT_DONE, HELP:, MERGED, HANDOFF, SWARM_START, DISPATCH_ATTEMPT/OK/FAIL, IDLE_PASSIVATED)
    evidence: "internal/witness/protocol.go — ProtoType enum, ClassifyMessage(); internal/witness/handlers.go"

  - id: gt_session_continuity
    what: Session recovery via seance (query past sessions) + prime (context recovery) + nudge queue poller
    evidence: "internal/nudge/; internal/session/; templates/ — CLAUDE.md, PRIME.md"

  - id: gt_plugin_system
    what: Shell-based plugins with plugin.md + run.sh (compactor-dog, stuck-agent-dog, quality-review, github-sheriff)
    evidence: "plugins/ directory with 8+ plugins"

  - id: gt_multi_agent_runtime
    what: ACP (Agent Control Protocol) abstraction: TMUX + ACP modes, Native/Subcommand/Flag invocation per agent type
    evidence: "internal/acp/; internal/mayor/manager.go — StartACP(); internal/runtime/ — ResolveRoleAgentConfig()"

applications:
  - advantage_id: gt_zfc_state
    implemented_in: internal/orchestrator/state.go
    mechanism: "State derived from actual execution reality (process alive, file exists) not cached state files"
    invariant: "No cached state that can go stale. Every state query checks observable reality."
    status: planned

  - advantage_id: gt_watchdog_chain
    implemented_in: internal/orchestrator/watchdog.go
    mechanism: "3-tier: Daemon(tick=3min, 15+ checks) → Boot(triage AI) → Deacon(patrol AI) → Witness(per-resource)"
    invariant: "Each tier has different frequency and intelligence. Higher tiers only activated when lower tiers escalate."
    status: planned

  - advantage_id: gt_help_classification
    implemented_in: internal/orchestrator/triage.go
    mechanism: "Keyword matching → 6 categories × 3 severities → auto-route to appropriate handler"
    invariant: "Emergency (P0) always escalates immediately. Medium (P2) queues for next cycle."
    status: planned

  - advantage_id: gt_event_driven_convoy
    implemented_in: internal/orchestrator/dispatch.go
    mechanism: "State transition event → check dependencies → dispatch next unblocked work item"
    invariant: "No work dispatched with unmet dependencies. Each completion triggers exactly one dependency check."
    status: planned

  - advantage_id: gt_atomic_allocation
    implemented_in: core/alloc.c
    mechanism: "flock → allocate name → create pending marker → create resource → remove marker → unlock"
    invariant: "No two agents allocate same resource. Failed allocation rolls back and releases name."
    status: planned

  - advantage_id: gt_rollback_on_failure
    implemented_in: core/rollback.c
    mechanism: "Track created resources in order → cleanupOnError reverses in reverse order"
    invariant: "Partial failure leaves no orphaned resources. Rollback is complete or explicitly partial with log."
    status: planned

  - advantage_id: gt_pressure_gating
    implemented_in: internal/orchestrator/pressure.go
    mechanism: "checkPressure(resource_type) → if high → defer dispatch → still run cleanup"
    invariant: "Dispatch deferred under pressure, but cleanup always runs. No resource leak under load."
    status: planned

  - advantage_id: gt_lifecycle_state_machine
    implemented_in: internal/orchestrator/lifecycle.go
    mechanism: "State transitions: Working→Done→Idle, Working→Stuck, Working→Stalled, Idle→Zombie"
    invariant: "Zombie detection: session alive but no worktree = zombie. Stalled: work assigned but session dead."
    status: planned

  - advantage_id: gt_inter_agent_protocol
    implemented_in: internal/mcp/protocol.go
    mechanism: "Typed messages with regex-based classification → dispatch to typed handlers"
    invariant: "Every inter-agent message has type + structured payload. Unknown types → logged + ignored."
    status: planned

  - advantage_id: gt_session_continuity
    implemented_in: internal/orchestrator/session.go
    mechanism: "Session DB + nudge queue + prime context recovery across restarts"
    invariant: "Agent state recoverable from session DB. Nudge queue prevents deadlocks."
    status: planned

  - advantage_id: gt_plugin_system
    implemented_in: internal/plugin/
    mechanism: "Plugin interface: Name() + Run(ctx) → register → execute on event"
    invariant: "Plugins cannot crash the host. Plugin errors logged but never propagate."
    status: planned

  - advantage_id: gt_multi_agent_runtime
    implemented_in: internal/mcp/runtime.go
    mechanism: "RuntimeConfig per agent type → ResolveRoleAgentConfig → BuildStartupCommand"
    invariant: "Every supported agent type has a runtime adapter. Unknown type → error, not crash."
    status: planned

control:
  - advantage_id: gt_zfc_state
    verification: "Unit test: kill process → verify state reflects 'dead' not cached 'running'"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_watchdog_chain
    verification: "Integration test: simulate stuck agent → verify Boot detects → Deacon dispatched → Witness handles"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_help_classification
    verification: "Unit test: 'EMERGENCY: disk full' → verify P0 + Emergency route; 'blocked on PR' → verify P2 + Blocked"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_event_driven_convoy
    verification: "Integration test: complete task A that blocks B → verify B auto-dispatches"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_atomic_allocation
    verification: "Stress test: 10 concurrent allocations → verify no duplicate names, no leaked pending markers"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_rollback_on_failure
    verification: "Unit test: fail at step 3 of 5 → verify steps 1-2 resources cleaned up"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_pressure_gating
    verification: "Integration test: saturate system → verify dispatch deferred + cleanup still runs"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_lifecycle_state_machine
    verification: "Unit test: simulate each state transition → verify correct next state"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_inter_agent_protocol
    verification: "Integration test: send POLECAT_DONE → verify handler fires; send HELP:emergency → verify escalation"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_session_continuity
    verification: "Integration test: crash agent mid-task → restart → verify state recovered from session DB"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_plugin_system
    verification: "Unit test: plugin that panics → verify host continues; valid plugin → verify executed"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never

  - advantage_id: gt_multi_agent_runtime
    verification: "Integration test: configure Claude runtime → verify startup; configure Codex runtime → verify different startup"
    update_trigger: "Re-analyze when gastown releases new version"
    last_verified: never
```
