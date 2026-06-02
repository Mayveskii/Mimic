```yaml
repo: Mayveskii/hermes-agent
url: https://github.com/Mayveskii/hermes-agent
language: Python
status: partial
last_sync: "2025-05-17"

description: |
  Fork of NousResearch/hermes-agent. Self-improving AI agent with closed learning loop,
  50+ tools, 7 terminal backends, context compression, trajectory recording, and
  multi-platform messaging gateway. 3479 files, production-grade.

advantages:
  - id: ha_closed_learning_loop
    what: Agent creates skills from experience, self-improves during use, periodic memory/skill nudge via turn counter
    evidence: "agent/curator.py — background review; agent/memory_manager.py — build_memory_context_block; tools/skill_manager_tool.py — skill CRUD; agent/conversation_loop.py — _turns_since_memory nudge"

  - id: ha_context_compression
    what: Preflight token detection + multi-pass compression + stable system prompt prefix for cache reuse
    evidence: "agent/context_compressor.py — threshold_tokens check; agent/conversation_compression.py — compress; agent/conversation_loop.py — preflight compression, _cached_system_prompt"

  - id: ha_iteration_budget
    what: IterationBudget.consume() gates loop — budget exhausted → one grace call → stop
    evidence: "agent/iteration_budget.py — IterationBudget struct with consume/remaining; agent/conversation_loop.py — budget check"

  - id: ha_tool_guardrails
    what: Per-turn reset of guardrails, deny/ask/allow pipeline, OP_FLAG_DANGEROUS always blocked, halt decisions
    evidence: "agent/tool_guardrails.py — reset_for_turn; tools/approval.py — approval modes; agent/conversation_loop.py — _tool_guardrails"

  - id: ha_delegation
    what: Spawn isolated subagents with _current_task_id for cross-agent file state registry
    evidence: "tools/delegate_tool.py — delegate_task; agent/conversation_loop.py — effective_task_id, _current_task_id"

  - id: ha_trajectory_compression
    what: Record tool-call trajectories + compress for training next-gen models
    evidence: "agent/trajectory.py — has_incomplete_scratchpad; trajectory_compressor.py — compression pipeline"

  - id: ha_memory_nudge
    what: Turn-based nudge: _turns_since_memory >= nudge_interval → trigger memory review; rehydrate from history on fresh agent
    evidence: "agent/conversation_loop.py — _should_review_memory, _turns_since_memory, prior_user_turns hydration"

  - id: ha_skill_provenance
    what: ContextVar tracks skill write origin (background review vs foreground tool) to prevent uncontrolled mutations
    evidence: "tools/skill_provenance.py — set_current_write_origin; agent/conversation_loop.py — _memory_write_origin binding"

  - id: ha_streaming_health
    what: 90s stale-stream detection + 60s read timeout prevent hangs on zombie provider connections
    evidence: "agent/conversation_loop.py — _interruptible_streaming_api_call, _cleanup_dead_connections"

  - id: ha_error_classifier_failover
    what: classify_api_error → FailoverReason → retry with backoff vs fallback provider vs abort
    evidence: "agent/error_classifier.py — FailoverReason enum; agent/conversation_loop.py — failover logic, _try_activate_fallback"

  - id: ha_message_sanitization
    what: Surrogate repair, non-ASCII strip, tool call argument repair, role alternation fix before every API call
    evidence: "agent/message_sanitization.py — _sanitize_surrogates, _repair_tool_call_arguments, _strip_images_from_messages"

  - id: ha_prompt_caching
    what: Stable system prompt prefix + Anthropic cache_control breakpoints for ~75% input token savings
    evidence: "agent/prompt_caching.py — apply_anthropic_cache_control; agent/conversation_loop.py — _cached_system_prompt, _use_prompt_caching"

  - id: ha_credential_pool
    what: Multi-key rotation with env fallback — credential_sources.py + credential_pool.py
    evidence: "agent/credential_pool.py — pool management; agent/credential_sources.py — env var fallback"

  - id: ha_multi_transport
    what: Transport abstraction: Anthropic native, OpenAI chat completions, Bedrock, Codex responses, Gemini native
    evidence: "agent/transports/ — anthropic.py, chat_completions.py, bedrock.py, codex.py, gemini_native.py"

applications:
  - advantage_id: ha_closed_learning_loop
    implemented_in: internal/orchestrator/learning.go
    mechanism: "Turn counter → nudge interval → trigger curator review → create/improve skill"
    invariant: "Skills only created after ≥1 successful execution. Skill provenance tracked per write."
    status: planned

  - advantage_id: ha_context_compression
    implemented_in: internal/orchestrator/compress.go
    mechanism: "estimate_tokens → if ≥ threshold → compress(protect_first_n, protect_last_n) → rebuild cache prefix"
    invariant: "System prompt prefix byte-stable after compression. First N + last N messages never compressed."
    status: planned

  - advantage_id: ha_iteration_budget
    implemented_in: internal/orchestrator/budget.go
    mechanism: "IterationBudget{max, used} → consume() returns false when exhausted → grace call → stop"
    invariant: "Budget exceeded → no more tool executions. Grace call is exactly one extra iteration."
    status: planned

  - advantage_id: ha_tool_guardrails
    implemented_in: internal/orchestrator/guardrails.go
    mechanism: "reset_for_turn() → check each tool call against deny/ask/allow → halt if dangerous"
    invariant: "OP_FLAG_DANGEROUS always requires explicit allow. Guardrails reset every turn."
    status: planned

  - advantage_id: ha_delegation
    implemented_in: internal/orchestrator/delegate.go
    mechanism: "delegate_task(task_id, prompt) → spawn subagent with isolated context → file state registry"
    invariant: "Each delegation gets unique task_id. Parent can track child state via registry."
    status: planned

  - advantage_id: ha_trajectory_compression
    implemented_in: internal/quality/trajectory.go
    mechanism: "Record tool calls + results → compress with role reduction → output for training"
    invariant: "Trajectory includes all tool calls in order. Compression preserves decision points."
    status: planned

  - advantage_id: ha_memory_nudge
    implemented_in: internal/orchestrator/memory.go
    mechanism: "Increment turns_since_memory → if ≥ nudge_interval → trigger review → reset counter"
    invariant: "Nudge fires at most once per interval. Fresh agents rehydrate counter from history."
    status: planned

  - advantage_id: ha_skill_provenance
    implemented_in: internal/tool/provenance.go
    mechanism: "ContextVar write_origin = 'background_review' | 'foreground_tool' → tag each skill write"
    invariant: "Background review skills tagged differently from foreground tool skills."
    status: planned

  - advantage_id: ha_streaming_health
    implemented_in: internal/mcp/health.go
    mechanism: "Stale stream detector: 90s no delta → abort + cleanup_dead_connections on next turn"
    invariant: "No API call hangs beyond 90s without delta. Dead connections cleaned before next request."
    status: planned

  - advantage_id: ha_error_classifier_failover
    implemented_in: internal/mcp/failover.go
    mechanism: "classify_api_error(error) → FailoverReason → jittered_backoff retry or activate_fallback"
    invariant: "Rate limit errors → backoff + retry. Auth errors → no retry. Server errors → fallback provider."
    status: planned

  - advantage_id: ha_message_sanitization
    implemented_in: internal/quality/sanitize.go
    mechanism: "Before every API call: sanitize surrogates → repair tool call args → fix role alternation"
    invariant: "No lone surrogates (U+D800-DFFF) in API payloads. Tool call arguments always valid JSON."
    status: planned

  - advantage_id: ha_prompt_caching
    implemented_in: internal/orchestrator/caching.go
    mechanism: "Build system prompt once per session → cache → inject cache_control breakpoints on system + last 3 msgs"
    invariant: "System prompt bytes identical across turns. Cache breakpoints on fixed positions."
    status: planned

  - advantage_id: ha_credential_pool
    implemented_in: internal/mcp/credentials.go
    mechanism: "Pool of API keys → rotate on rate limit → fallback to env var if pool empty"
    invariant: "No unauthenticated API calls in production. Rate-limited keys temporarily removed from pool."
    status: planned

  - advantage_id: ha_multi_transport
    implemented_in: internal/mcp/transport.go
    mechanism: "Transport interface with Anthropic/OpenAI/Bedrock/Codex/Gemini implementations"
    invariant: "Every provider implements same Transport interface. Provider switch = config change only."
    status: planned

control:
  - advantage_id: ha_closed_learning_loop
    verification: "Integration test: execute task → verify skill created → verify skill improved on reuse"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_context_compression
    verification: "Unit test: generate conversation > threshold → verify compression reduces tokens below threshold"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_iteration_budget
    verification: "Unit test: set budget=3 → verify loop stops after 3 iterations + 1 grace call"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_tool_guardrails
    verification: "Unit test: submit dangerous op → verify blocked; submit safe op → verify allowed"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_delegation
    verification: "Integration test: delegate task → verify subagent completes with isolated state"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_trajectory_compression
    verification: "Unit test: record trajectory → compress → verify decision points preserved"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_memory_nudge
    verification: "Integration test: nudge_interval=3 → verify review triggered every 3 turns"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_skill_provenance
    verification: "Unit test: background review creates skill → verify provenance tag = background_review"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_streaming_health
    verification: "Integration test: simulate 90s stale stream → verify abort + connection cleanup"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_error_classifier_failover
    verification: "Unit test: rate limit error → verify backoff; auth error → verify no retry; server error → verify fallback"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_message_sanitization
    verification: "Unit test: inject surrogates → verify stripped; inject bad JSON in tool_call → verify repaired"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_prompt_caching
    verification: "Integration test: 2 consecutive turns → verify system prompt bytes identical → verify cache hit"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_credential_pool
    verification: "Unit test: exhaust key via rate limit → verify rotation → verify env fallback"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never

  - advantage_id: ha_multi_transport
    verification: "Integration test: switch provider config → verify same tool call succeeds via new transport"
    update_trigger: "Re-analyze when hermes-agent releases new version"
    last_verified: never
```
