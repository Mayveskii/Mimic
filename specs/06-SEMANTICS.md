# SEMANTICS.md — Mimic

Every function: name | input | output | invariant | source

> **NOTE**: Canonical function specs moved to `specs-v2/c-core/OPCODE_SPEC.md` and `specs-v2/domains/*/PROCESS.md`. This file is historical.

---

## C-Core: ops.c

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| ops_init | void | int (0=ok) | Cannot init twice | ops.c | ✅ Implemented |
| ops_shutdown | void | void | Only after init | ops.c | ✅ Implemented |
| ops_register | OpCodeDef* | int (0=ok) | opcode < OP_MAX, registry not full | ops.c | ✅ Implemented |
| ops_get_definition | OpCode | const OpCodeDef* | NULL if opcode ≥ OP_MAX | ops.c | ✅ Implemented |
| ops_execute | OpPacketEx* | int (0=ok) | Measures latency_ns via CLOCK_MONOTONIC | ops.c | ✅ Implemented |
| ops_execute_chain | OpPacketEx[], count, ExecContext* | int (0=ok) | validate_chain passed BEFORE first exec | ops.c | ✅ Implemented |
| ops_validate_chain | OpPacketEx[], count, ExecContext* | ValidationResult | 9-step pipeline, O(n²) conflict check | ops.c | ✅ Implemented |
| ops_check_conflict | OpCode, OpCode | bool | invalid opcodes → conflict | ops.c | ✅ Implemented |
| ops_calculate_action | OpPacketEx[], count | float | S = Σ(cost_tokens) | ops.c | ✅ Implemented |
| ops_get_time_ns | void | uint64_t | CLOCK_MONOTONIC | ops.c | ✅ Implemented |
| ops_opcode_to_string | OpCode | const char* | "UNKNOWN" if opcode ≥ OP_MAX | ops.c | ✅ Implemented |
| ops_string_to_opcode | const char* | OpCode | OP_NOP if not found | ops.c | ✅ Implemented |
| ops_packet_init | OpPacketEx*, OpCode | void | zeroed, fd = -1 | ops.c | ✅ Implemented |
| ops_packet_set_string | OpPacketEx*, key, value | void | arg_count < MAX_ARGS=16 | ops.c | ✅ Implemented |
| ops_packet_set_int | OpPacketEx*, key, int64 | void | arg_count < MAX_ARGS=16 | ops.c | ✅ Implemented |
| ops_mmap_alloc | size_t | void* (NULL=fail) | MAP_PRIVATE\|MAP_ANONYMOUS | ops.c | ✅ Implemented |
| ops_mmap_free | void*, size_t | int (0=ok) | ptr ≠ NULL, size ≠ 0 | ops.c | ✅ Implemented |
| ops_mmap_sync | void*, size_t | int (0=ok) | MS_SYNC | ops.c | ✅ Implemented |
| ops_register_builtins | void | void | Registers all 91 OpCodes | ops.c | ✅ Implemented |
| ops_rollback_chain | OpPacketEx[], failed_index, ExecContext* | int | 3-phase: inverse → cleanup → hash verify | ops.c | ✅ Implemented |
| ops_create_backup | path, backup_path_buf, buf_size | int | copies to .mimic/backups/ | ops.c | ✅ Implemented |
| ops_best_effort_cleanup | OpPacketEx*, ExecContext* | void | domain-specific per ROLLBACK_SPEC.md | ops.c | ✅ Implemented |
| ops_compute_state_hash | ExecContext* | uint64_t | FNV-1a over FDs + mmap regions | ops.c | ✅ Implemented |

---

## C-Core: New functions (from specs-v2)

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| session_has_explicit_allow | OpCode, chain_id | bool | default deny | ops.c | ✅ Stub (session layer override point) |
| session_has_2vote_verify | chain_id | bool | default deny | ops.c | ✅ Stub (session layer override point) |
| arg_value_string | OpPacketEx*, key | const char* | NULL if not found | ops.c | ✅ Internal |
| arg_value_int | OpPacketEx*, key, default | int64 | default if not found | ops.c | ✅ Internal |
| arg_value_bool | OpPacketEx*, key, default | bool | default if not found | ops.c | ✅ Internal |
| init_conflict_matrix | void | void | 15 rules (5 self + 10 cross-domain) | ops.c | ✅ Internal |
| init_energy_costs | void | void | 91 entries (cost_tokens, cost_time_us, cost_mem) | ops.c | ✅ Internal |
| register_op | OpCode, name, desc, exec, ... | void | fills OpCodeDef | ops.c | ✅ Internal |

---

## C-Core: mmap_ops.c, git_ops.c, git_scenarios.c

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| mmap_ops_register_all | void | void | Registers OP_MMAP_* executors | mmap_ops.c | ✅ Part of ops_register_builtins |
| git_ops_register_all | void | void | Registers OP_GIT_* executors | git_ops.c | ✅ Part of ops_register_builtins |
| scenario_atomic_commit | path, message | ScenarioResult | status+diff+commit atomically | git_scenarios.c | ⏳ Not implemented (needs scenario layer) |
| scenario_safe_merge | source, target | ScenarioResult | Fast-forward only | git_scenarios.c | ⏳ Not implemented |
| scenario_feature_branch | name | ScenarioResult | Create branch without switching | git_scenarios.c | ⏳ Not implemented |
| scenario_hotfix | name, target | ScenarioResult | branch + commit + merge into target | git_scenarios.c | ⏳ Not implemented |
| scenario_ci_diff_check | base, head | ScenarioResult | diff --check, no whitespace errors | git_scenarios.c | ⏳ Not implemented |

---

## C-Core: libbmap.a

**Sources exist** in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`. 39 functions.

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| bmap_open | path | bmap_t* | NULL if not exists | libbmap.a | ⏳ Not implemented |
| bmap_close | bmap_t* | void | — | libbmap.a | ⏳ Not implemented |
| bmap_read_cell | bmap_t*, cell_id | cell_data | — | libbmap.a | ⏳ Not implemented |
| bmap_write_cell | bmap_t*, cell_id, data | int | — | libbmap.a | ⏳ Not implemented |
| bmap_write | bmap_t* | int | — | libbmap.a | ⏳ Not implemented |
| bmap_free_cell | bmap_t*, cell_id | int | — | libbmap.a | ⏳ Not implemented |
| bmap_cell_serialized_size | cell | size_t | — | libbmap.a | ⏳ Not implemented |
| si_create | void | slot_index_t* | — | libbmap.a | ⏳ Not implemented |
| si_destroy | slot_index_t* | void | — | libbmap.a | ⏳ Not implemented |
| si_insert | slot_index_t*, slot | int | — | libbmap.a | ⏳ Not implemented |
| si_query_domain | slot_index_t*, domain | result_set | — | libbmap.a | ⏳ Not implemented |
| si_query_domain_layer | slot_index_t*, domain, layer | result_set | — | libbmap.a | ⏳ Not implemented |
| si_query_state_hash | slot_index_t*, hash | result_set | — | libbmap.a | ⏳ Not implemented |
| si_build_from_bmap | bmap_t* | slot_index_t* | — | libbmap.a | ⏳ Not implemented |
| si_result_free | result_set | void | — | libbmap.a | ⏳ Not implemented |
| inv_create | void | invariant_t* | — | libbmap.a | ⏳ Not implemented |
| inv_destroy | invariant_t* | void | — | libbmap.a | ⏳ Not implemented |
| inv_add | invariant_t*, condition | int | — | libbmap.a | ⏳ Not implemented |
| inv_find_similar | invariant_t*, condition, threshold | result_set | — | libbmap.a | ⏳ Not implemented |
| inv_load | path | invariant_t* | — | libbmap.a | ⏳ Not implemented |
| inv_save | invariant_t*, path | int | — | libbmap.a | ⏳ Not implemented |
| inv_dedup_check | invariant_t*, condition | bool | — | libbmap.a | ⏳ Not implemented |
| gnk_compute | bmap_t*, domain | gnk_result | — | libbmap.a | ⏳ Not implemented |
| gnk_score_domains | bmap_t* | gnk_result* | — | libbmap.a | ⏳ Not implemented |
| gnk_result_free | gnk_result* | void | — | libbmap.a | ⏳ Not implemented |
| snapshot_build | bmap_t* | snapshot_t* | — | libbmap.a | ⏳ Not implemented |
| snapshot_load | path | snapshot_t* | — | libbmap.a | ⏳ Not implemented |
| snapshot_write | snapshot_t*, path | int | — | libbmap.a | ⏳ Not implemented |
| snapshot_sign | snapshot_t*, key | int | — | libbmap.a | ⏳ Not implemented |
| snapshot_diff | snapshot_t*, snapshot_t* | diff_result | — | libbmap.a | ⏳ Not implemented |
| snapshot_diff_free | diff_result | void | — | libbmap.a | ⏳ Not implemented |
| snapshot_free | snapshot_t* | void | — | libbmap.a | ⏳ Not implemented |
| layer_walk | bmap_t*, layer | walk_result | — | libbmap.a | ⏳ Not implemented |
| drift_detect | bmap_t*, snapshot_t* | drift_result | — | libbmap.a | ⏳ Not implemented |
| cosine_f32 | float[], float[], n | float | [-1, 1] | libbmap.a | ⏳ Not implemented |
| cosine_int8 | int8[], int8[], n | float | [-1, 1] | libbmap.a | ⏳ Not implemented |
| batch_cosine_int8 | int8[][], int8[], n, batch | float[] | [-1, 1] each | libbmap.a | ⏳ Not implemented |
| int8_quantize | float[], n | int8[] | — | libbmap.a | ⏳ Not implemented |
| sha256_hash | data, len | hash[32] | — | libbmap.a | ⏳ Not implemented |
| z_density_compute | bmap_t*, domain | float | >= 0 | libbmap.a | ⏳ Not implemented |

---

## CGO Bridge (internal/cgo/)

All Go-side code is stub. User writes code. I review against specs-v2/.

**Status: ⏳ CGO bridge not yet implemented.**

---

## New OpCodes (Research + Self-Management)

Added to `specs-v2/c-core/OPCODE_SPEC.md`:

| OpCode | Description | Safety | Status |
|--------|-------------|--------|--------|
| OP_RESEARCH_HYPOTHESIS_CREATE | Create falsifiable hypothesis | SAFE | ✅ Stub registered |
| OP_RESEARCH_HYPOTHESIS_LOAD | Load existing hypothesis | READONLY | ✅ Stub registered |
| OP_RESEARCH_HYPOTHESIS_INFERENCE | Confirm/refute/refine | SAFE | ✅ Stub registered |
| OP_RESEARCH_EXPERIMENT_RUN | Execute experiment | DANGEROUS | ✅ Stub registered |
| OP_RESEARCH_RESULT_STORE | Store result | DANGEROUS | ✅ Stub registered |
| OP_RESEARCH_STATISTICAL_TEST | Compute significance/CI | READONLY | ✅ Stub registered |
| OP_RESEARCH_LITERATURE_FETCH | Download paper | NETWORK | ✅ Stub registered |
| OP_RESEARCH_LITERATURE_PARSE | Extract structured info | SAFE | ✅ Stub registered |
| OP_RESEARCH_LITERATURE_INDEX | Index as mesh slot | DANGEROUS | ✅ Stub registered |
| OP_RESEARCH_CITATION_LINK | Link cited papers | SAFE | ✅ Stub registered |
| OP_RESEARCH_LITERATURE_EMBED | Generate embedding | MEMORY | ✅ Stub registered |
| OP_RESEARCH_PROGRESS_STORE | Store key findings | DANGEROUS | ✅ Stub registered |
| OP_RESEARCH_CONTEXT_SUMMARIZE | Semantic compression | SAFE | ✅ Stub registered |
| OP_SELF_CHECKPOINT_CREATE | Snapshot current state | DISK | ✅ Stub registered |
| OP_SELF_CHECKPOINT_RESTORE | Load from snapshot | SAFE | ✅ Stub registered |
| OP_SELF_BUDGET_REALLOCATE | Rebalance budget | READONLY | ✅ Stub registered |
| OP_SELF_STRATEGY_PIVOT | Switch approach | SAFE | ✅ Stub registered |
| OP_SELF_PROGRESS_ASSESS | Compare actual vs planned | READONLY | ✅ Stub registered |
| OP_SELF_CONTEXT_SUMMARIZE | Semantic compression | SAFE | ✅ Stub registered |

---

## Completeness Summary

| Module | Total Functions | Implemented | Stubs | Pending |
|--------|----------------|-------------|-------|---------|
| ops.c core | 23 | 23 | 0 | 0 |
| ops.c executors (I/O, System, Build, Git basics, Network basics, Process basics) | 35 | 35 | 0 | 0 |
| ops.c stubs (research, self-mgmt, advanced git/network) | 33 | 0 | 33 | 0 |
| libbmap.a | 39 | 0 | 0 | 39 |
| CGO bridge | 6 | 0 | 0 | 6 |
| git_scenarios | 5 | 0 | 0 | 5 |
| **Total** | **141** | **58** | **33** | **50** |

---

## Go Layer: Exa API (internal/tool/exa/)

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| exa.LoadConfigFromEnv | — | Config | all env vars parsed, defaults applied | exa-mcp-server | ✅ Implemented |
| exa.Config.Disabled | — | bool | APIKey empty → true | exa-mcp-server | ✅ Implemented |
| exa.NewClient | Config | *Client / nil | cfg.Disabled → nil | exa-mcp-server | ✅ Implemented |
| Client.Search | query, numResults, type | *SearchResponse / error | numResults ∈ [1,100], query non-empty | exa-mcp-server | ✅ Implemented |
| Client.Fetch | urls[], maxChars | *ContentsResponse / error | len(urls) ≤ 100 | exa-mcp-server | ✅ Implemented |
| Client.post | path, body | []byte / error | retry ≤ RetryMax, backoff = base*2^attempt | exa-mcp-server | ✅ Implemented |

## Go Layer: Exa MCP Handlers (internal/mcp/)

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| NewExaHandler | exa.Config | *ExaHandler | client nil if disabled | exa-mcp-server | ✅ Implemented |
| ExaHandler.HandleExaSearch | args map | JSON-RPC result | returns isError=true if disabled or invalid params | exa-mcp-server | ✅ Implemented |
| ExaHandler.HandleExaFetch | args map | JSON-RPC result | urls array required; maxChars optional | exa-mcp-server | ✅ Implemented |
| ExaHandler.HandleMimicResearch | args map | JSON-RPC result | topic required; depth ∈ {shallow,deep} | exa-mcp-server | ✅ Implemented |
| ExaHandler.disabled | — | error result | message explains missing EXA_API_KEY | exa-mcp-server | ✅ Implemented |

## Go Layer: RTK Compression (internal/rtk/)

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| Compress | output string, ContentType, Config | string | result ≤ MaxLines, ≤ MaxChars if set | rtk-ai | ✅ Implemented |
| DetectContentType | sample string | ContentType | never empty | rtk-ai | ✅ Implemented |
| stripAnsi | text | string | removes all \x1b\[[0-9;]*m | rtk-ai | ✅ Implemented |
| collapseBlanks | text | string | 3+ newlines → 1 | rtk-ai | ✅ Implemented |
| headTailSplit | lines[], Config | string | HeadLines + TailLines ≤ MaxLines | rtk-ai | ✅ Implemented |

## Configuration (specs/11-CONFIGURATION.md)

| Variable | Type | Scope | Invariant | Source | Status |
|----------|------|-------|-----------|--------|--------|
| MIMIC_PORT | uint16 | env, runtime | >0 && <65535 | go-service-template-rest | ✅ Documented |
| MIMIC_BUDGET_TOKENS | uint64 | env, session | >0 && ≤10B | code-mode | ✅ Documented |
| MIMIC_RTK_ENABLED | bool | env, runtime | true/false | rtk-ai | ✅ Documented |
| EXA_API_KEY | string | env, secret | len==0 or ≥20 | exa-mcp-server | ✅ Documented |
| EXA_BASE_URL | string | env, runtime | valid HTTPS URL | exa-mcp-server | ✅ Documented |
| MIMIC_AUTO_RESEARCH | bool | env, runtime | true/false | ADR-0009 | ✅ Documented |

---

## Completeness Summary

| Module | Total Functions | Implemented | Stubs | Pending |
|--------|----------------|-------------|-------|---------|
| ops.c core | 23 | 23 | 0 | 0 |
| ops.c executors (I/O, System, Build, Git basics, Network basics, Process basics) | 35 | 35 | 0 | 0 |
| ops.c stubs (research, self-mgmt, advanced git/network) | 33 | 0 | 33 | 0 |
| Exa client + handlers | 9 | 9 | 0 | 0 |
| RTK compression | 4 | 4 | 0 | 0 |
| libbmap.a | 39 | 0 | 0 | 39 |
| CGO bridge | 6 | 0 | 0 | 6 |
| git_scenarios | 5 | 0 | 0 | 5 |
| **Total** | **154** | **71** | **33** | **50** |

Core execution engine: **COMPLETE** (91 OpCodes registered, validation, rollback, tests passing)
Exa + RTK: **COMPLETE** (13 functions implemented, tested, documented)
Storage layer (bmap): **PENDING** (next priority)
CGO bridge: **PENDING** (after bmap)

---

## Three Phases (v0.3–v0.5) — GiT Analogy

### Phase A: Text-Native Mesh (ADR-005)
Direct analogy to GiT: universal text representation instead of domain-specific modules.

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| TextSlot.ToMarkdown | TextSlot | []byte markdown | Self-contained, round-trip parseable | text_slot.go | ✅ Implemented |
| ParseTextSlot | []byte markdown | *TextSlot | All fields populated or zero values | text_slot.go | ✅ Implemented |
| LoadAllTextSlots | dir path | []*TextSlot | Walk .md files recursively | text_slot.go | ✅ Implemented |

**Semantics**: Every slot is a markdown document with `# Slot: <id>` header, `## Invariant`, `## Actions`, `## Cross-Domain Links`, `## Embedding` (base64 int8[384]), `## Metadata`. LLM can read mesh directly. `6-20×` smaller than gob (1.9 GB → ~200 MB).

### Phase B: Qdrant-Primary + Cross-Domain Edges

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| MeshRegistry.Query | queryText, topK, embedFn, floatEmbedFn, qdrantClient | *MeshResult | Qdrant first, local fallback | query.go | ✅ Implemented |
| MeshRegistry.TraverseEdges | startID, maxDepth | []SlotLink | BFS over Slot.Links, no cycles | query.go | ✅ Implemented |

**Semantics**: `Query()` now hits qdrant HNSW index first (~10 ms, 180 K vectors). If `< 3` results, falls back to local int8 brute force (textSlots or gob Maps). `TraverseEdges()` enables cross-domain emergence: `go-error-handling` ─[similar_to]→ `k8s-health-check`.

### Phase C: Generative Chains

| Function | Input | Output | Invariant | Source | Status |
|----------|-------|--------|-----------|--------|--------|
| ValidatePlan | *Plan, budget | error | Conflict matrix + C-core packet check + budget | plan.go | ✅ Implemented |
| GeneratePlanFromGoal | goal, []ToolSchema | *Plan | Stub: will call LLM in production | plan.go | ⏳ Stub |
| Logger.Log | LogEntry | error | Append JSONL atomically | logger.go | ✅ Implemented |
| ExtractPatterns | path string | []Pattern | Group by goal, retention filter | logger.go | ⏳ Stub |
| PlanHandler.HandleGeneratePlan | args map | map result | Delegates to GeneratePlanFromGoal + ValidatePlan | plan_handler.go | ✅ Implemented |

**Semantics**: LLM generates multi-step `OpPacket` chain → Mimic validates via C-core (`ValidatePlan` checks conflict matrix, energy cost, packet serialization) → execute with checkpoints → on success, log session → `ExtractPatterns` converts to new `TextSlot`. Self-improving mesh without human curation.

---

**GiT ↔ Mimic mapping**:
- GiT text tokens → TextSlot markdown
- GiT multi-task joint training → Cross-domain edges (SlotLink)
- GiT auto-regressive generation → ValidatePlan + generative chains
- GiT zero-shot → Mesh.AutoApply with adaptive threshold

