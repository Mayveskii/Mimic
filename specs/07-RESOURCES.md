# RESOURCES.md — Mimic

Complete map of every resource in Mimic: what it is, what it does for agent operations, how it translates to the OpPacket language.

> **NOTE**: Canonical specifications moved to `specs-v2/`. This file is historical. Read `specs-v2/c-core/OPCODE_SPEC.md` and `specs-v2/domains/*/ARTIFACTS.md` for current specs.

## Resource Translation Principle

Every agent operation (bash command, git operation, file read, build step) is translated into an OpPacket — a structured, validated, deterministic unit. The agent never sees OpCodes; it sees MCP tools with human-readable names.

```
Agent intent: "commit these files safely"
    ↓
Mimic translation:
    OpPacket[0]: OP_GIT_STATUS  (args: {path: "."})
    OpPacket[1]: OP_GIT_DIFF    (args: {staged: true})
    OpPacket[2]: OP_GIT_ADD     (args: {files: [...]})
    OpPacket[3]: OP_GIT_COMMIT  (args: {message: "..."})
    ↓
Validation: conflict_matrix[0..3] = all 0, energy_budget = 8.0 tokens, latency ≈ 15000μs
    ↓
Execution: ops_execute_chain(packets, 4, ctx)
    ↓
Result: {success: true, commit: "abc123", latency_ns: 12400000}
```

---

## C-Core Resources (c-core/)

### ops.c / ops.h — Core Engine

| Resource | Purpose for Agent Operations | OpPacket Translation |
|----------|------------------------------|---------------------|
| ops_init | Initialize engine before any operation | Called once at startup, no OpPacket |
| ops_register | Register OpCode definition | Internal |
| ops_execute | Execute single OpPacket | Agent tool call → single OpPacket |
| ops_execute_chain | Execute validated sequence | Agent tool call → scenario → multiple OpPackets |
| ops_validate_chain | Validate without executing | Agent `validate` tool → ValidationResult |
| ops_check_conflict | Check if two OpCodes conflict | Internal validation |
| ops_calculate_action | Compute action S = Σ(cost_tokens × cost_time_us) | Internal: choose cheapest chain |
| ops_packet_init | Create OpPacket with auto-incremented ID | Internal |
| ops_packet_set_string | Set string argument | Internal |
| ops_packet_set_int | Set integer argument | Internal |
| ops_mmap_alloc | Allocate memory-mapped region | Agent: large buffers |
| ops_mmap_free | Free mmap region | Cleanup |
| ops_mmap_sync | Sync mmap to disk | Agent: durability |

### git_ops.c — Git Operations

| OpCode | Agent Operation | OpPacket Args | Result |
|--------|----------------|---------------|--------|
| OP_GIT_INIT | Initialize repo | {path: "."} | repo created |
| OP_GIT_CLONE | Clone repository | {url: "...", path: "..."} | repo cloned |
| OP_GIT_FETCH | Fetch remote | {remote: "origin"} | refs updated |
| OP_GIT_COMMIT | Commit staged | {message: "..."} | commit hash |
| OP_GIT_PUSH | Push to remote | {remote: "origin", branch: "main"} | refs pushed |
| OP_GIT_DIFF | Show differences | {staged: true} or {base: "...", head: "..."} | diff output |
| OP_GIT_STATUS | Working tree status | {path: "."} | status output |
| OP_GIT_CHECKOUT | Switch branch | {ref: "..."} | HEAD moved |
| OP_GIT_BRANCH | Create/list branches | {name: "..."} or {list: true} | branch info |
| OP_GIT_MERGE | Merge branches | {source: "...", target: "..."} | merge result |
| OP_GIT_REBASE | Rebase onto branch | {onto: "..."} | rebase result |
| OP_GIT_TAG | Create annotated tag | {name: "v...", message: "..."} | tag created |
| OP_GIT_RESET | Reset (emergency) | {mode: "soft/hard", target: "..."} | NEVER-RULE: requires explicit flag |

> Full opcode spec: `specs-v2/c-core/OPCODE_SPEC.md`

### Scenarios

| Scenario | Intent | OpPacket Chain | Source |
|----------|--------|----------------|--------|
| atomic_commit | "commit safely" | STATUS → DIFF → ADD → COMMIT | `specs-v2/patterns/atomic_commit.md` |
| safe_merge | "merge without force-push" | FETCH → DIFF → MERGE (ff-only) | `specs-v2/patterns/safe_merge.md` |
| feature_branch | "create feature branch" | BRANCH → CHECKOUT | `specs-v2/patterns/feature_branch.md` |
| hotfix | "hotfix and merge" | BRANCH → COMMIT → CHECKOUT(target) → MERGE | `specs-v2/patterns/hotfix.md` |
| ci_diff_check | "check formatting" | DIFF(base,head) → parse for whitespace errors | `specs-v2/patterns/ci_diff_check.md` |
| build_and_test | "compile and test" | BUILD_COMPILE → BUILD_TEST | `specs-v2/patterns/build_and_test.md` |

### libbmap.a — Binary Map Storage

39 functions for mesh storage. Sources exist in `/home/cisco/findings/fck_sleep/binary-mesh/c-core/`.

| Function Group | Purpose |
|----------------|---------|
| bmap_* | Store/retrieve mesh slots |
| slot_index_* | Index slots by domain/layer/hash |
| invariant_* | Store/check preconditions |
| gnk_scorer_* | Score domains by quality |
| snapshot_* | Point-in-time mesh state |
| matrix_* | Walk layers, detect drift |
| math_* | Vector similarity (cosine) |
| hash_* | Integrity (sha256) |
| z_density_* | Slot quality metric |

---

## Environment Configuration

All tunable parameters: `specs-v2/c-core/ENV_CONFIG.md`

Key parameters for research:
- `MIMIC_MAX_CONTEXT_SIZE_MB` = 64 (default, raised from 1)
- `MIMIC_MAX_RESEARCH_CONTEXT_MB` = 256
- `MIMIC_CHECKPOINT_INTERVAL_MINUTES` = 30
- `MIMIC_RESEARCH_MODE` = false (enable for research semantics)
- `MIMIC_SELF_CHECKPOINT_ENABLED` = true
- `MIMIC_STRATEGY_PIVOT_THRESHOLD` = 3

---

## Build Configuration

Compile-time flags: `specs-v2/c-core/BUILD_CONFIG.md`

New feature flags:
- `MIMIC_ENABLE_RESEARCH_DOMAIN` = 1 (hypothesis, experiments, literature)
- `MIMIC_ENABLE_SELF_MANAGEMENT` = 1 (checkpoint, pivot, budget reallocate)

---

## RPC Format

MCP JSON-RPC + binary mesh wire: `specs-v2/c-core/RPC_FORMAT.md`

---

## Mesh Exchange

Cross-node slot serialization: `specs-v2/c-core/MESH_EXCHANGE.md`

---

## Memory Layout

Arena-based alternative to fixed-size OpPacket: `specs-v2/c-core/MEMORY_LAYOUT.md`

---

## Quality Gates

QAC-1..13, measured thresholds, anti-patterns: `specs-v2/domains/quality/` and `specs/sources/anti-patterns.md`
