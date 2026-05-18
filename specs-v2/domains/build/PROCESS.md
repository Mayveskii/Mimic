# Build Domain — Compile, Test, Deploy

How a model builds, tests, and deploys code through Mimic.

---

## What This Domain Does

Any model needs to compile code, run tests, and sometimes deploy. The build domain provides tokenized processes for these operations. The model does not need to know build systems (make, cargo, go build, npm, etc.) — it expresses intent, Mimic provides the proven process.

The domain covers: compile, link, test, deploy, clean.

---

## Processes

### build_and_test

**When to use:**  
Model wants to verify that code compiles and passes tests. "Build this project" or "Run tests."

**Goal:**  
Code compiles successfully. Tests run and pass. Compilation failures detected before test execution.

**Chain (semantically):**

1. Compile (compile).
   - Determine build system from project files (go.mod → go, Cargo.toml → cargo, Makefile → make, package.json → npm).
   - Run appropriate compile command.
   - Capture stdout/stderr.
   - If compile fails → stop, return error to model, do not run tests.

2. Test (test).
   - If compile succeeds → run tests.
   - Capture test output, failures, coverage.
   - Timeout enforced (default 300s, configurable).

**Hard constraints:**
- Tests NEVER run if compilation fails.
- Compilation must produce zero errors. Warnings reported but do not block.
- Test timeout must be set and enforced.

**Invariants:**
- compile_success → test_may_run.
- compile_failure → test_never_runs.
- All test failures reported to model with file/line/function.

**Result when successful:**
```
status: "success"
compile: {
  command: "go build ./...",
  duration_ms: 12400,
  errors: 0,
  warnings: 3
}
test: {
  command: "go test ./...",
  duration_ms: 45800,
  passed: 47,
  failed: 0,
  skipped: 2,
  coverage: 0.73
}
```

**Result when failed:**
```
status: "failure"
stage: "compile" | "test"
error: "undefined: SomeFunc in main.go:42" | "TestFoo: expected 42, got 0"
compile_output: <full output>
```

**How a model uses this:**  
Model says "build and test" → Mimic detects go.mod → runs go build → if success → runs go test → returns results. If build fails → model gets error output, not test results.

---

### safe_deploy

**When to use:**  
Model wants to deploy to production or staging. High-risk operation.

**Goal:**  
Build, test, tag version, push to remote, deploy with health check.

**Chain (semantically):**

1. Compile release build (compile with release flags).
2. Run full test suite (test).
3. Create git tag (tag).
4. Push tag to remote (push).
5. Deploy to target host (deploy).
6. Health check (verify deployment is running).

**Hard constraints:**
- Deploy ONLY after all tests pass.
- Tag follows semantic versioning.
- 2-vote verification required before deploy step.
- No deploy without explicit model confirmation (safety level 0).

**Invariants:**
- All tests pass → tag created → push succeeds → deploy.
- Any failure at any step → rollback to previous state.
- Health check confirms deployment is serving requests.
- Deploy keys NEVER hardcoded in source (use credential pool from hermes-agent).

**Result when successful:**
```
status: "success"
version: "v1.2.3"
tag_hash: "abc123..."
deploy_target: "192.168.111.25:2022"
health_check: "pass"
```

**Result when failed:**
```
status: "failure"
stage: "compile" | "test" | "tag" | "push" | "deploy" | "health_check"
rollback: "tag deleted, deployment reverted to previous version"
```

**How a model uses this:**  
Model says "deploy to production" → Mimic returns full process with 2-vote requirement on deploy step → model approves each step. If test fails → deploy stops before tag. Model never accidentally deploys broken code.

---

### parallel_build_shards

**When to use:**  
Model wants to build multiple independent targets (e.g., 3 crates, frontend + backend).

**Goal:**  
Build multiple targets in parallel where safe, respecting dependencies.

**Chain (semantically):**

1. Detect shards and dependencies.
   - Parse project structure for independent build targets.
   - Build dependency graph.

2. Check cross-shard conflicts.
   - conflict_matrix[shard_A] × [shard_B] = 0 → parallel.
   - conflict = 1 → serialize.

3. Execute builds.
   - Parallel shards: compile concurrently.
   - Link step: after all shards compiled, link shared outputs.

**Hard constraints:**
- No two shards write to same output directory simultaneously.
- Shared dependency compilation → serialize or use lock.

**Invariants:**
- Each shard builds independently.
- Link step waits for all shards.
- Total time = max(shard_time) + link_time, not sum.

**Result when successful:**
```
status: "success"
shards: [
  {name: "backend", status: "success", duration_ms: 12300},
  {name: "frontend", status: "success", duration_ms: 8900}
]
link: {status: "success", duration_ms: 1200}
total_duration_ms: 13500  # max(shard) + link
```

**Result when failed:**
```
status: "failure"
failed_shard: "backend"
error: "compilation error in api.go:42"
other_shards: "cleaned"
```

**How a model uses this:**  
Model says "build all crates" → Mimic detects 3 crates with no cross-dependencies → builds in parallel → links after all succeed. Model gets combined result.

---

## Principles From Sources

### bun

**Principles taken:**
- Phase-d-build-queue: build only after plan validation.
- Phase-d-crate-shard: ~170 agents building shards in parallel with scoped isolation.
- Edit scope isolation: each shard edits only its own output directory.

**What Mimic does with them:**
Build processes validated before execution. Parallel shards with conflict detection. Resource bitmask isolation per shard.

### code-mode

**Principles taken:**
- Budget enforcement: maxTurns + maxBudgetUsd per session.
- Concurrency control: up to 10 tools in parallel.

**What Mimic does with them:**
Build processes consume energy budget. Parallel build shards limited to 10 concurrent.

### hermes-agent

**Principles taken:**
- Error classifier: build errors classified as syntax, logic, or environment.
- Iteration budget: limited retries for flaky tests.

**What Mimic does with them:**
Build errors classified before retry decision. Flaky test retries limited by budget.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "build" |
| layer | "process" |
| modality | "code" |
| pattern_name | "build_and_test" / "safe_deploy" / "parallel_build_shards" |
| invariants | ["compile_before_test", "tag_after_test", "2vote_before_deploy", "no_hardcoded_keys"] |

---

## Cross-Domain Conflicts

Build domain conflicts with:
- **git domain**: cannot checkout during compilation (working tree changes mid-build).
- **io domain**: cannot write source files during compilation.
- **network domain**: fetch dependencies conflicts with other network operations.
