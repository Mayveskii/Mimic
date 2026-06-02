# SCENARIOS.md — Mimic

Execution patterns for each scenario. Every scenario = a named OpPacket chain with invariants, costs, and rollback strategy.

---

## Scenario Structure

Each scenario is defined by:
- **Name**: human-readable, used as MCP tool argument
- **Intent**: what the agent wants to accomplish
- **OpPacket chain**: ordered sequence of operations
- **Invariants**: conditions that must hold before, during, and after execution
- **Energy cost**: total tokens + estimated latency
- **Rollback**: what happens if a step fails
- **2-vote verify**: whether post-exec verification is required
- **Source**: where this pattern came from (behavior source or distillation)

---

## Implemented Scenarios (from git_scenarios.c)

### atomic_commit

```
Intent: "commit these files safely without breaking anything"

Chain:
  [0] OP_GIT_STATUS   args: {path: "."}           → check working tree state
  [1] OP_GIT_DIFF     args: {staged: true}         → verify what will be committed
  [2] OP_GIT_ADD      args: {files: [...]}         → stage specified files
  [3] OP_GIT_COMMIT   args: {message: "..."}       → create commit

Invariants:
  - Working tree must not have conflicts before commit
  - Commit message must not be empty
  - Files must exist before staging

Energy cost: ~8.0 tokens, ~15000μs

Rollback: if OP_GIT_COMMIT fails → OP_GIT_CHECKOUT to previous state

2-vote verify: no (standard commit)

Source: Mayveskii/embryo (pkg/do/BinaryRuntime)
```

### safe_merge

```
Intent: "merge source into target without force-push or merge conflicts"

Chain:
  [0] OP_GIT_FETCH    args: {remote: "origin"}     → get latest remote refs
  [1] OP_GIT_DIFF     args: {base: source, head: target} → preview merge
  [2] OP_GIT_MERGE    args: {source: "...", target: "..."} → fast-forward only

Invariants:
  - Fast-forward only (no merge commits)
  - No conflicts in diff preview
  - Source branch must exist

Energy cost: ~12.0 tokens, ~25000μs

Rollback: if merge fails → OP_GIT_CHECKOUT to previous HEAD

2-vote verify: yes (merge changes history)

Source: Mayveskii/embryo (pkg/do/BinaryRuntime)
```

### feature_branch

```
Intent: "create a new feature branch"

Chain:
  [0] OP_GIT_BRANCH   args: {name: "..."}          → create branch
  [1] OP_GIT_CHECKOUT args: {ref: "..."}            → switch to branch

Invariants:
  - Branch must not already exist
  - Working tree must be clean before checkout

Energy cost: ~4.0 tokens, ~5000μs

Rollback: if checkout fails → checkout back to previous branch

2-vote verify: no

Source: Mayveskii/embryo (pkg/do/BinaryRuntime)
```

### hotfix

```
Intent: "create hotfix branch, commit, and merge into target"

Chain:
  [0] OP_GIT_BRANCH   args: {name: "hotfix/..."}    → create hotfix branch
  [1] OP_GIT_COMMIT   args: {message: "hotfix: ..."} → commit fix
  [2] OP_GIT_CHECKOUT args: {ref: target}            → switch to target
  [3] OP_GIT_MERGE    args: {source: "hotfix/..."}   → merge into target

Invariants:
  - Hotfix branch name must be unique
  - Must merge back into specified target (e.g., main)
  - Merge must be fast-forward or clean

Energy cost: ~16.0 tokens, ~30000μs

Rollback: if merge fails → checkout back, delete hotfix branch

2-vote verify: yes (hotfix modifies production branch)

Source: Mayveskii/embryo (pkg/do/BinaryRuntime)
```

### ci_diff_check

```
Intent: "check for whitespace errors and formatting issues"

Chain:
  [0] OP_GIT_DIFF     args: {base: "...", head: "..."} → get diff
  [1] Internal: parse diff output for:
      - Trailing whitespace
      - Tab/space mixing
      - Merge conflict markers (<<<<<<, =======, >>>>>>>)
      - No newline at end of file

Invariants:
  - Diff must contain no whitespace errors
  - No conflict markers

Energy cost: ~3.0 tokens, ~2000μs

Rollback: N/A (read-only operation)

2-vote verify: no

Source: Mayveskii/embryo (pkg/do/BinaryRuntime)
```

---

## Planned Scenarios (from behavior sources)

### build_and_test

```
Intent: "compile, run tests, report results"

Chain:
  [0] OP_BUILD_COMPILE  args: {target: "...", flags: [...]}
  [1] OP_BUILD_TEST     args: {target: "...", filter: "..."}

Invariants:
  - Compilation must succeed before test
  - Test timeout must be set

Energy cost: ~15.0 tokens, ~60000μs+

Rollback: OP_BUILD_CLEAN if compile fails

2-vote verify: no

Source: Mayveskii/bun (cargo check + bun bd test pattern)
```

### safe_deploy

```
Intent: "build, test, tag, push"

Chain:
  [0] OP_BUILD_COMPILE  args: {target: "release"}
  [1] OP_BUILD_TEST     args: {target: "release"}
  [2] OP_GIT_TAG        args: {name: "v..."}
  [3] OP_GIT_PUSH       args: {remote: "origin", tag: "v..."}

Invariants:
  - All tests pass before tag
  - Tag follows semver
  - Push only after local tag created

Energy cost: ~25.0 tokens, ~120000μs+

Rollback: delete tag if push fails

2-vote verify: yes (deploy is irreversible)

Source: Mayveskii/bun (phase-d-build-queue pattern)
```

### search_and_apply

```
Intent: "find a pattern in mesh and apply it to current context"

Chain:
  [0] OP_MMAP_READ     args: {slot_id: "..."}      → read slot from bmap
  [1] Internal: inv_find_similar(args: {invariant, context}) → check preconditions
  [2] OP_BUILD_COMPILE  args: {target: "..."}       → build with pattern
  [3] OP_BUILD_TEST     args: {target: "..."}       → verify pattern works

Invariants:
  - Slot invariant must match current context (BEHAVIOR.md #10 mimicry control)
  - If invariant fails → degrade to next_lower_tier

Energy cost: ~20.0 tokens, ~80000μs+

Rollback: if test fails → revert to previous state

2-vote verify: yes (applying external pattern to codebase)

Source: Mayveskii/embryo (pkg/hunt/ meshscan pattern)
```

### parallel_build_shards

```
Intent: "build multiple crates/packages in parallel"

Chain (concurrent):
  [0a] OP_BUILD_COMPILE args: {target: "crate-a"}
  [0b] OP_BUILD_COMPILE args: {target: "crate-b"}
  [0c] OP_BUILD_COMPILE args: {target: "crate-c"}
  [1]  OP_BUILD_LINK    args: {targets: ["crate-a", "crate-b", "crate-c"]}

Invariants:
  - No cross-dependencies between shards (conflict_matrix check)
  - Each shard edits only its own scope (edit scope isolation)

Energy cost: ~30.0 tokens, ~40000μs (parallel)

Rollback: if any shard fails → clean all

2-vote verify: no (build only)

Source: Mayveskii/bun (phase-d-crate-shard pattern, ~170 agents simultaneously)
```

---

## Scenario Lifecycle

Each scenario follows the phase graph from BEHAVIOR.md #6:

```
CLASSIFY → determine which scenario matches agent intent
PLAN → build OpPacket chain from scenario template
VALIDATE → conflict matrix + energy budget + permission
EXEC → ops_execute_chain via CGO
VERIFY → 2-vote if required by scenario
RESPOND → result to agent
```

Scenarios are registered in the orchestrator and can be:
- Called by name: `exec_chain("atomic_commit", {files: [...], message: "..."})`
- Composed: `safe_deploy` = `build_and_test` + `safe_deploy` suffix
- Custom: agent provides explicit OpPacket chain without a named scenario

---

## Scenario Registry Update Process

Scenarios are updated per GitHub best practices:

1. **New scenario** → PR to dev with scenario definition (chain + invariants + costs)
2. **Cost update** → measure actual latency/tokens → update energy_cost_matrix → ADR if change > 20%
3. **Invariant change** → always requires ADR (why invariant changed)
4. **Distillation update** → new patterns from repos-manifest.yaml → new scenarios → PR
5. **Behavior source update** → new behaviors from behavior-sources.yaml → new scenarios → PR
