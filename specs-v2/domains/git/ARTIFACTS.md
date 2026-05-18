# Git Domain — Artifacts

How git processes are stored as mesh slots.

---

## Slot Structure for Git Patterns

Every git pattern stored in the mesh follows the SLOT_SCHEMA.md with these domain-specific values:

| Field | Value |
|---|---|
| domain | `DOMAIN_GIT` (0) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | One of: `atomic_commit`, `safe_merge`, `feature_branch`, `hotfix`, `ci_diff_check` |
| pattern_code | Serialized OpPacket chain (binary) + semantic description (text) |
| invariants | `["status_clean_after_commit", "diff_reviewed_before_commit", "ff_only_merge", "branch_unique", "fetch_before_remote_ops", "no_rebase_auto"]` |
| polarity | `POLARITY_POSITIVE` (0) for valid patterns, `POLARITY_NEGATIVE` (1) for anti-patterns with `counter_slot_id` |
| survival_index | From git blame on source repos (embryo, binary-mesh, gonka) |
| z_density | Computed from usage frequency + survival + text compression ratio |
| extraction_hash | sha256 of extraction tool + parameters + source content |

---

## Pattern Codes

### atomic_commit

```c
OpPacket chain[4] = {
    {OP_GIT_STATUS, .args = {{"path", "."}}},
    {OP_GIT_DIFF,   .args = {{"cached", "false"}}},
    {OP_GIT_ADD,    .args = {{"paths", ""}}},  // paths filled at runtime
    {OP_GIT_COMMIT, .args = {{"message", ""}}}  // message filled at runtime
};
```

Text description:
```
Step 1: Check working tree state. Ensure clean or model confirms intent.
Step 2: Review diff. Report violations (conflict markers, trailing whitespace).
Step 3: Stage explicitly named files. Never implicit staging.
Step 4: Create commit with non-empty descriptive message.
Verify: git status shows clean working tree.
```

### safe_merge

```c
OpPacket chain[3] = {
    {OP_GIT_FETCH,  .args = {{"remote", "origin"}}},
    {OP_GIT_DIFF,   .args = {{"base", ""}, {"head", ""}}},  // filled at runtime
    {OP_GIT_MERGE,  .args = {{"source", ""}, {"ff_only", "true"}}}
};
```

### feature_branch

```c
OpPacket chain[3] = {
    {OP_GIT_STATUS,     .args = {}}},
    {OP_GIT_BRANCH,     .args = {{"name", ""}, {"start_point", ""}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", ""}}}
};
```

### hotfix

```c
OpPacket chain[9] = {
    {OP_GIT_STATUS,     .args = {}}},
    {OP_GIT_BRANCH,     .args = {{"name", "hotfix/"}, {"start_point", ""}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", "hotfix/"}}},
    // Model applies fix via IO domain here
    {OP_GIT_ADD,        .args = {{"paths", ""}}},
    {OP_GIT_COMMIT,     .args = {{"message", ""}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", ""}}},
    {OP_GIT_MERGE,      .args = {{"source", "hotfix/"}, {"ff_only", "true"}}},
    {OP_GIT_BRANCH,     .args = {{"name", "hotfix/"}, {"delete", "true"}}}
};
```

### ci_diff_check

```c
OpPacket chain[1] = {
    {OP_GIT_DIFF, .args = {{"cached", "false"}, {"check", "true"}}}
};
// Executor performs internal validation on diff output
```

---

## Invariant Hashes

Each git invariant has a canonical string representation, hashed for mesh indexing:

| Invariant | Canonical String | xxHash64 |
|---|---|---|
| status_clean_after_commit | `git_status_clean_after_commit_v1` | # UNCERTAIN: compute at build time |
| diff_reviewed_before_commit | `git_diff_reviewed_before_commit_v1` | # UNCERTAIN |
| ff_only_merge | `git_ff_only_merge_v1` | # UNCERTAIN |
| branch_unique | `git_branch_name_unique_v1` | # UNCERTAIN |
| fetch_before_remote_ops | `git_fetch_before_remote_decisions_v1` | # UNCERTAIN |
| no_rebase_auto | `git_no_rebase_automatic_v1` | # UNCERTAIN |

---

## Anti-Pattern Slots

Git domain NEGATIVE slots document failure modes:

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-05 (context injection without stability) | `git_context_injection_unstable` | `atomic_commit` |
| AP-11 (stale cached state) | `git_stale_cached_state` | `safe_merge` |
| AP-15 (partial rollback) | `git_partial_rollback` | `safe_merge` |
| AP-16 (flip-flop decision) | `git_flip_flop_decision` | `feature_branch` |

---

## Retrieval Path

Git patterns are retrieved via:
1. **Linear**: `si_query_exact(DOMAIN_GIT, LAYER_PROCESS, MODALITY_CODE, "atomic_commit")`
2. **Keyword**: `si_query_invariant(invariant_hash("git_status_clean_after_commit_v1"))` → all git commit patterns.
3. **Semantic**: Query "how do I safely commit changes?" → domain=git, layer=process → reranked by survival_index.

Survival index source: git blame on `embryo/pkg/do/`, `binary-mesh/c-core/git_*.c`, `gonka-ai/*/commit history`.
