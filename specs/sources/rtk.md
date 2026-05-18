```yaml
repo: Mayveskii/rtk
url: https://github.com/Mayveskii/rtk
language: Rust
status: partial
last_sync: "2025-05-17"

description: |
  Fork of rtk-ai/rtk (49K stars). CLI proxy that reduces LLM token consumption by 60-90%
  on common dev commands. 8-stage TOML filter pipeline, 59 built-in filters, language-aware
  code stripping, SQLite token tracking, 13 AI agent hook integrations, trust-on-first-use
  security model.

advantages:
  - id: rtk_toml_pipeline
    what: 8-stage TOML filter DSL — strip_ansi, replace(regex+backrefs), match_output(unless guard), strip/keep_lines, truncate_lines_at, head/tail_lines, max_lines, on_empty
    evidence: "src/core/toml_filter.rs — CompiledFilter struct, apply_filter() 8-stage pipeline"

  - id: rtk_stream_filter
    what: Line-by-line StreamFilter trait — feed_line(line)→Option<String>, flush()→String, on_exit(code,raw)→Option<String>
    evidence: "src/core/stream.rs — StreamFilter trait, BlockStreamFilter<H>, LineStreamFilter<H>"

  - id: rtk_code_filter
    what: Language-aware code stripping — MinimalFilter (comments+blanks, 20-40% savings), AggressiveFilter (imports+signatures only, 60-90%), smart_truncate (importance-weighted)
    evidence: "src/core/filter.rs — Language enum, FilterStrategy trait, MinimalFilter, AggressiveFilter, smart_truncate()"

  - id: rtk_command_rewrite
    what: Static rule table mapping regex→rtk_cmd + rewrite_command() stripping sudo/env/git-opts + exclude_commands + transparent_prefixes
    evidence: "src/discover/registry.rs — classify_command(), rewrite_command(); src/discover/rules.rs — RULES table"

  - id: rtk_trust_model
    what: Trust-on-first-use for project-local .rtk/filters.toml — Trusted/Untrusted/ContentChanged/EnvOverride states
    evidence: "src/hooks/trust.rs — TrustStatus enum, check_trust()"

  - id: rtk_token_tracking
    what: SQLite WAL tracker with estimate_tokens(~4 chars/token), 90-day auto-cleanup, per-project GLOB queries
    evidence: "src/core/tracking.rs — Tracker struct, estimate_tokens(), get_summary_filtered()"

  - id: rtk_discovery
    what: Scan session JSONL for unfiltered commands → suggest corrections; detect repeated CLI mistakes → write .claude/rules/cli-corrections.md
    evidence: "src/discover/mod.rs — scan_sessions(); src/learn/detector.rs — detect_corrections()"

  - id: rtk_tee_failure
    what: On filtered command failure → save full unfiltered output to ~/.local/share/rtk/tee/ for debugging without re-execution
    evidence: "src/core/tee.rs — save_tee_output()"

  - id: rtk_agent_hooks
    what: Init scripts for 13 AI agents — writes hook scripts, CLAUDE.md, patches settings.json, creates .rtk/filters.toml
    evidence: "src/hooks/init.rs — init_agent(), AGENTS constant"

applications:
  - advantage_id: rtk_toml_pipeline
    implemented_in: internal/rtk/pipeline.go
    mechanism: "Parse TOML filter definition → build 8-stage pipeline → apply(input) → filtered output"
    invariant: "Each stage receives output of previous stage. Order is fixed. No stage is skippable."
    status: planned

  - advantage_id: rtk_stream_filter
    implemented_in: core/stream_filter.c
    mechanism: "C callback interface: rtk_filter_feed_line(handle, line, out) → rtk_filter_flush(handle, out) → rtk_filter_on_exit(handle, code, raw, out)"
    invariant: "feed_line never blocks. flush returns all buffered data. on_exit generates summary if applicable."
    status: planned

  - advantage_id: rtk_code_filter
    implemented_in: core/code_filter.c
    mechanism: "Detect language by extension → select FilterStrategy → apply(minimal or aggressive) → smart_truncate if overflow"
    invariant: "MinimalFilter never removes code. AggressiveFilter keeps all imports and signatures. smart_truncate prioritizes pub/export lines."
    status: planned

  - advantage_id: rtk_command_rewrite
    implemented_in: internal/rtk/classify.go
    mechanism: "Match command against RULES regex table → rewrite to rtk equivalent → check exclude list → check transparent prefix list"
    invariant: "Excluded commands never rewritten. Transparent prefixes (docker exec, etc.) stripped before matching, re-prepended after."
    status: planned

  - advantage_id: rtk_trust_model
    implemented_in: internal/rtk/trust.go
    mechanism: "Load project .rtk/filters.toml → check trust status → if Untrusted/ContentChanged → block + warn → if EnvOverride → allow"
    invariant: "Project-local filters never execute without trust. ContentChanged always blocks until re-reviewed."
    status: planned

  - advantage_id: rtk_token_tracking
    implemented_in: internal/rtk/tracking.go
    mechanism: "SQLite WAL → INSERT per command → estimate_tokens(len/4) → auto-DELETE after 90 days → GLOB query per project"
    invariant: "Token estimates within ±20% of tiktoken. Auto-cleanup never removes data < 90 days."
    status: planned

  - advantage_id: rtk_discovery
    implemented_in: internal/rtk/discover.go
    mechanism: "Scan session JSONL → classify each command against RULES → report unfiltered → detect correction patterns → write rule file"
    invariant: "Discovery is read-only — never modifies sessions. Written rules require human review."
    status: planned

  - advantage_id: rtk_tee_failure
    implemented_in: internal/rtk/tee.go
    mechanism: "If exit_code != 0 → save raw stdout/stderr to ~/.local/share/rtk/tee/<timestamp>.log → print hint path"
    invariant: "Tee files are append-only, never overwritten. Path printed exactly once per failure."
    status: planned

  - advantage_id: rtk_agent_hooks
    implemented_in: internal/rtk/hooks.go
    mechanism: "Select agent type → write hook script to agent config dir → patch settings.json → write CLAUDE.md/AGENTS.md → create .rtk/filters.toml"
    invariant: "Hook scripts are idempotent — running twice produces same result. Settings.json patches are additive."
    status: planned

control:
  - advantage_id: rtk_toml_pipeline
    verification: "Unit test: TOML with all 8 stages → verify each stage output matches expected"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_stream_filter
    verification: "Integration test: stream 1000 lines through C filter → verify same output as Rust StreamFilter"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_code_filter
    verification: "Unit test: Rust file with comments → MinimalFilter strips comments, AggressiveFilter keeps signatures only"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_command_rewrite
    verification: "Unit test: 'sudo git log' → 'rtk git log'; 'docker exec mycontainer cat /etc/hosts' → 'docker exec mycontainer rtk cat /etc/hosts'"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_trust_model
    verification: "Unit test: Untrusted filter → blocked; ContentChanged filter → blocked; RTK_TRUST_ALL_FILTERS=1 → allowed"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_token_tracking
    verification: "Integration test: insert 100 commands → query summary → verify totals ±5%"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_discovery
    verification: "Integration test: session with 10 unfiltered commands → verify all 10 reported"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_tee_failure
    verification: "Unit test: failing command → verify tee file exists with full output"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never

  - advantage_id: rtk_agent_hooks
    verification: "Integration test: rtk init claude → verify CLAUDE.md exists, settings.json patched, .rtk/filters.toml created"
    update_trigger: "Re-analyze when rtk releases new version"
    last_verified: never
```
