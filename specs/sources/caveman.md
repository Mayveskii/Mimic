```yaml
repo: Mayveskii/caveman
url: https://github.com/Mayveskii/caveman
language: Python
status: partial
last_sync: "2025-05-17"

description: |
  Fork of caveman-ai/caveman. LLM-driven code context management with intensity levels,
  auto-clarity boundaries, file type detection, sensitive path protection, validate-then-accept
  output validation, and LLM-powered compression. Designed to prepare code context within
  token budgets while preserving critical information.

advantages:
  - id: cm_intensity_level_config
    what: 4 intensity levels (low/medium/high/max) control how aggressively context is compressed; higher intensity = more compression + more risk of information loss
    evidence: "SKILL.md — intensity level definitions; compress.py — intensity parameter controls compression strategy selection"

  - id: cm_auto_clarity_boundary
    what: Auto-clarity boundary: when compressed context would lose critical clarity markers → stop compression + emit warning; preserves readability threshold
    evidence: "SKILL.md — clarity boundary rules; compress.py — clarity check before finalizing compression output"

  - id: cm_file_type_detection
    what: File type detection via content + magic bytes, not extension alone; detect.py classifies files into text/code/data/binary with confidence scores
    evidence: "caveman/detect.py — detect_file_type() with magic byte checks + content heuristics; confidence score per classification"

  - id: cm_sensitive_path_protection
    what: Sensitive path protection: detect.py is_sensitive_path() identifies .env, credentials, keys, secrets → auto-exclude from context; never include in LLM prompt
    evidence: "caveman/detect.py — is_sensitive_path() with regex patterns for .env, *key*, *secret*, *credential*, *token* files"

  - id: cm_validate_then_accept
    what: Validate-then-accept: LLM output must pass structural validation before use; validate.py checks syntax, completeness, and format constraints
    evidence: "caveman/validate.py — validate_output() with structural checks; rejected output → retry with clarified constraints"

  - id: cm_llm_driven_compression
    what: LLM-driven compression: call_claude() compresses code context within token budget; prompt includes budget + priority signals; LLM decides what to keep/drop
    evidence: "caveman/compress.py — call_claude() with token budget parameter; compression prompt with priority markers"

applications:
  - advantage_id: cm_intensity_level_config
    implemented_in: internal/orchestrator/intensity.go
    mechanism: "Config intensity: low→minimal compression, medium→moderate, high→aggressive, max→maximum compression; each level maps to compression parameter set"
    invariant: "Intensity level is enum (4 values). Higher intensity never includes more content than lower. Max intensity = token budget hard limit."
    status: planned

  - advantage_id: cm_auto_clarity_boundary
    implemented_in: internal/quality/clarity.go
    mechanism: "Before finalizing compression: compute clarity_score(compressed) → if < threshold → undo last compression step → emit warning with lost content summary"
    invariant: "Compressed output clarity_score ≥ threshold. If unreachable at given intensity → escalate to next higher intensity."
    status: planned

  - advantage_id: cm_file_type_detection
    implemented_in: internal/orchestrator/detect.go
    mechanism: "Magic bytes first 512 bytes → classify text/code/data/binary → confidence score → if confidence < 0.7 → fallback to extension heuristic"
    invariant: "Binary files never included in LLM context. Confidence < 0.7 → flagged for review. Extension-only detection = last resort."
    status: planned

  - advantage_id: cm_sensitive_path_protection
    implemented_in: internal/orchestrator/sensitive.go
    mechanism: "Path matching against sensitive patterns: .env, *key*, *secret*, *credential*, *token*, *password* → auto-exclude + log exclusion"
    invariant: "Sensitive paths NEVER included in LLM context. Exclusion logged with reason. No override for sensitive path exclusion."
    status: planned

  - advantage_id: cm_validate_then_accept
    implemented_in: internal/quality/validate.go
    mechanism: "LLM output → structural validation (syntax, completeness, format) → pass → accept; fail → retry with clarified constraints (max 3 retries)"
    invariant: "No LLM output used without validation pass. 3 validation failures → manual escalation. Validation includes syntax check."
    status: planned

  - advantage_id: cm_llm_driven_compression
    implemented_in: internal/orchestrator/compress.go
    mechanism: "Build prompt with token_budget + priority signals → call LLM → receive compressed context → validate → accept or retry"
    invariant: "Compressed output ≤ token_budget. Priority-marked content always preserved. LLM compression is advisory, validation is mandatory."
    status: planned

control:
  - advantage_id: cm_intensity_level_config
    verification: "Unit test: same input at low vs max intensity → verify max output ≤ low output in token count"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never

  - advantage_id: cm_auto_clarity_boundary
    verification: "Unit test: compression that drops clarity below threshold → verify rollback + warning emitted"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never

  - advantage_id: cm_file_type_detection
    verification: "Unit test: .py file with binary header → verify magic-byte detection; .txt extension with code content → verify code classification"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never

  - advantage_id: cm_sensitive_path_protection
    verification: "Unit test: .env file → verify excluded; credentials.json → verify excluded; README.md → verify included"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never

  - advantage_id: cm_validate_then_accept
    verification: "Unit test: malformed LLM output → verify rejected; valid output → verify accepted; 3 malformed → verify escalation"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never

  - advantage_id: cm_llm_driven_compression
    verification: "Integration test: 50K tokens with budget=10K → verify compressed output ≤ 10K + priority content preserved"
    update_trigger: "Re-analyze when caveman releases new version"
    last_verified: never
```
