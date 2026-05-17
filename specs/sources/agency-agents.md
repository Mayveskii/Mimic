```yaml
repo: Mayveskii/agency-agents
url: https://github.com/Mayveskii/agency-agents
language: Markdown/Shell
status: partial
last_sync: "2025-05-17"

description: |
  Fork of VRSEN/agency-swarm. AI agent framework where agents are defined as Markdown
  documents with skill sets, division categorization, integration multiplexing, strategy
  playbooks, handoff templates, and workflow process patterns. Shell-based orchestration
  with convert.sh for agent-to-code pipeline.

advantages:
  - id: aa_agent_as_markdown
    what: Agent definition = Markdown file with structured sections (role, skills, constraints, examples); convert.sh transforms MD → executable agent code
    evidence: "agents/ — *.md agent definitions with role/skills/constraints sections; scripts/convert.sh — MD→code transformation pipeline"

  - id: aa_division_categorization
    what: Agents categorized into divisions (orchestrator, specialist, worker) with skill overlap detection and conflict resolution
    evidence: "agents/ — division folders (orchestrator/, specialist/, worker/); division categorization in agent frontmatter"

  - id: aa_integration_multiplexer
    what: Integration multiplexer: single agent interface → route to multiple API backends (OpenAI, Anthropic, local); scripts/convert.sh generates adapter code
    evidence: "scripts/convert.sh — generates integration adapter from agent MD; integrations/ — backend adapter templates"

  - id: aa_strategy_playbooks
    what: Strategy playbooks: predefined decision trees for common agent workflows (research, code, review) with branching conditions
    evidence: "strategy/ — playbook definitions with decision tree structures and branching logic"

  - id: aa_handoff_templates
    what: Handoff templates: structured context transfer between agents with required fields (state, pending, blockers, next_action)
    evidence: "templates/handoff/ — handoff.md templates with required field definitions; handoff validation in convert.sh"

  - id: aa_workflow_process_pattern
    what: Workflow process pattern: define steps → assign agents → track progress → handle failures → escalate; repeatable across projects
    evidence: "workflows/ — process pattern definitions; scripts/convert.sh — workflow→orchestration code generation"

applications:
  - advantage_id: aa_agent_as_markdown
    implemented_in: internal/orchestrator/agent_def.go
    mechanism: "Parse agent MD: extract role/skills/constraints sections → validate required fields → generate AgentConfig struct → register in agent registry"
    invariant: "Every agent definition must have role + at least 1 skill. Missing required section → parse error with line number."
    status: planned

  - advantage_id: aa_division_categorization
    implemented_in: internal/orchestrator/division.go
    mechanism: "Agent frontmatter division field → categorize into orchestrator/specialist/worker → detect skill overlap → resolve conflicts (specialist wins)"
    invariant: "Every agent belongs to exactly one division. Overlapping skills → specialist division takes precedence. No orphaned agents."
    status: planned

  - advantage_id: aa_integration_multiplexer
    implemented_in: internal/mcp/multiplexer.go
    mechanism: "AgentConfig.backends[] → multiplexer route: pick backend by availability + cost → generate adapter → call backend → normalize response"
    invariant: "At least one backend must be available per agent. Backend failure → fallback to next in list. All backends down → error, not hang."
    status: planned

  - advantage_id: aa_strategy_playbooks
    implemented_in: internal/orchestrator/playbook.go
    mechanism: "Load playbook → evaluate decision tree → branch on conditions → assign agent steps → track progress per branch"
    invariant: "Every branch has defined termination condition. Infinite loops detected after max_depth. Branch selection = deterministic given same inputs."
    status: planned

  - advantage_id: aa_handoff_templates
    implemented_in: internal/orchestrator/handoff.go
    mechanism: "Agent A → fill handoff template (state, pending, blockers, next_action) → validate required fields → Agent B receives structured context"
    invariant: "Handoff without required fields → rejected. Agent B starts with full context from Agent A. No state loss in transfer."
    status: planned

  - advantage_id: aa_workflow_process_pattern
    implemented_in: internal/orchestrator/workflow.go
    mechanism: "Define workflow steps → assign agents per step → track step status → failure → retry/escalate → completion → next step"
    invariant: "Every step has assigned agent and timeout. Failed step → retry (max 3) then escalate. Step order enforced."
    status: planned

control:
  - advantage_id: aa_agent_as_markdown
    verification: "Unit test: valid MD → verify AgentConfig generated; MD missing role → verify parse error"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never

  - advantage_id: aa_division_categorization
    verification: "Unit test: agent with overlapping skills in worker + specialist → verify specialist wins"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never

  - advantage_id: aa_integration_multiplexer
    verification: "Integration test: 3 backends, fail first 2 → verify 3rd backend used; all down → verify error returned"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never

  - advantage_id: aa_strategy_playbooks
    verification: "Integration test: research playbook → verify decision tree branches correctly; infinite loop condition → verify max_depth termination"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never

  - advantage_id: aa_handoff_templates
    verification: "Unit test: handoff with all required fields → verify accepted; handoff missing blockers → verify rejected"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never

  - advantage_id: aa_workflow_process_pattern
    verification: "Integration test: 3-step workflow, fail step 2 → verify retry → verify escalation after 3 retries"
    update_trigger: "Re-analyze when agency-agents releases new version"
    last_verified: never
```
