# Build Domain — Artifacts

How build processes are stored as mesh slots.

---

## Slot Structure for Build Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_BUILD` (1) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `build_and_test` / `safe_deploy` / `parallel_build_shards` |
| pattern_code | OpPacket chain + semantic description |
| invariants | `["compile_before_test", "zero_compile_errors", "deploy_after_tests", "test_timeout_enforced", "output_isolation", "no_hardcoded_creds", "clean_preserves_source", "semver_tags", "parallel_conflict_check", "build_system_detected"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | From build success rate across source repos |
| z_density | Computed from build output compression + pattern frequency |

---

## Pattern Codes

### build_and_test

```c
OpPacket chain[3] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", "go.mod"}}},  // build system detection
    {OP_BUILD_COMPILE,    .args = {{"target", "./..."}}},
    {OP_BUILD_TEST,       .args = {{"filter", ""}, {"timeout_ms", "300000"}}}
};
```

Text description:
```
Step 1: Detect build system from project files (go.mod, Cargo.toml, Makefile, package.json).
Step 2: Compile with zero-error requirement. Warnings logged. Failure stops chain.
Step 3: Run tests with timeout. Return pass/fail counts and coverage.
Verify: compile success before test execution.
```

### safe_deploy

```c
OpPacket chain[6] = {
    {OP_BUILD_COMPILE,    .args = {{"target", "release"}}},
    {OP_BUILD_TEST,       .args = {{"filter", ""}, {"timeout_ms", "600000"}}},
    // 2-vote verification point (critical)
    {OP_GIT_TAG,          .args = {{"version", ""}, {"message", ""}}},
    {OP_GIT_PUSH,         .args = {{"remote", "origin"}, {"refspec", ""}}},
    {OP_BUILD_DEPLOY,     .args = {{"target", ""}, {"version", ""}}},
    {OP_NET_HTTP_GET,     .args = {{"url", ""}, {"timeout_ms", "30000"}}}  // health check
};
```

### parallel_build_shards

```c
OpPacket chain[3] = {
    {OP_ORCH_CLASSIFY,    .args = {{"intent", "detect build shards"}}},
    {OP_ORCH_PLAN,        .args = {{"domain", "build"}, {"scenario", "parallel_shards"}}},
    {OP_ORCH_EXEC,        .args = {{"parallel", "true"}, {"max_concurrent", "10"}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-02 (goroutine-per-item under mutex) | `build_goroutine_mutex_deadlock` | `parallel_build_shards` |
| AP-07 (hardcoded secrets) | `build_hardcoded_deploy_keys` | `safe_deploy` |
| AP-17 (blocking save) | `build_blocking_json_save` | `build_and_test` |

---

## Retrieval Path

Build patterns retrieved via:
1. Linear: exact match on pattern name.
2. Keyword: invariant_hash queries for "compile_before_test", "test_timeout_enforced".
3. Semantic: "how do I build and test this project?" → domain=build.

Survival index: build success rate from repos like `go build` on golang/go, `cargo test` on rust-lang/rust.
