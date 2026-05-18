# Build Domain — Invariants

Rules that MUST hold for every process in the build domain.

---

## BINV-01: Compile Before Test

**What it prevents:** Testing uncompilable code, wasting time on broken builds.

**What it requires:** OP_BUILD_TEST MUST NOT execute until OP_BUILD_COMPILE returns success (result_code == 0) in the same chain.

**Source of this rule:**
- Common sense: tests need compiled binaries.
- bun PR #30412: build after plan validation, test after build.

**Consequence of violation:** Validation REJECTS chain if test appears before compile. If compile fails and test is later in chain, execution stops at compile, test never runs.

---

## BINV-02: Zero Compile Errors

**What it prevents:** False confidence from "green" builds with unresolved errors.

**What it requires:** OP_BUILD_COMPILE must produce zero errors. Warnings are logged but do not fail. Any error = build failure.

**Source of this rule:**
- Standard engineering practice.
- embryo hunt system: only patterns from successful builds are indexed.

**Consequence of violation:** Compile returns failure. Model receives full compiler output.

---

## BINV-03: Deploy Only After All Tests Pass

**What it prevents:** Deployment of broken code.

**What it requires:** OP_BUILD_DEPLOY requires that OP_BUILD_TEST in the same chain (or verified recent run) passed with zero failures.

**Source of this rule:**
- bun PR #30412: deploy is safety level 0 (CRITICAL).
- gastown: rollback on failure prevents bad deploy.

**Consequence of violation:** Deploy REJECTED at validation with "tests not passed".

---

## BINV-04: Test Timeout Enforced

**What it prevents:** Infinite hangs in test suites, wasted budget.

**What it requires:** OP_BUILD_TEST MUST have timeout_ms set. Default: 300000ms (5 minutes). Max: 3600000ms (1 hour).

**Source of this rule:**
- AP-14 (infinite wait): no timeout on streaming → hangs forever.
- AP-30 (missing cancellation boundary): long operation blocks forever.

**Consequence of violation:** Test returns `ERR_TIMEOUT`. Model receives partial results + "timeout exceeded".

---

## BINV-05: Build Output Isolation

**What it prevents:** Cross-build contamination, race conditions.

**What it requires:** Each build target MUST write to its own output directory. Shared output directories require explicit serialization.

**Source of this rule:**
- bun PR #30412: edit scope isolation per shard.
- embryo mesh: resource bitmask prevents shared write access.

**Consequence of violation:** Conflict detected at validation (CR-02: Build Output Directory Lock). Parallel builds serialized.

---

## BINV-06: No Hardcoded Credentials in Build

**What it prevents:** Secret leakage via build scripts, deploy configs.

**What it requires:** OP_BUILD_DEPLOY MUST use credential pool (hermes-agent pattern), not hardcoded keys. Build scripts scanned for patterns like `password=`, `api_key=`, `token=`.

**Source of this rule:**
- AP-07 (hardcoded secrets): API keys committed to source → leaked.
- caveman: sensitive path protection.

**Consequence of violation:** Deploy REJECTED with "hardcoded credential detected in build script".

---

## BINV-07: Clean Preserves Source

**What it prevents:** Accidental deletion of source files.

**What it requires:** OP_BUILD_CLEAN MUST only delete files in designated output directories. Source files are NEVER deleted by build clean.

**Source of this rule:**
- AP-15 (partial rollback): incomplete cleanup.
- gastown ZFC: observable check — verify what clean targets.

**Consequence of violation:** Clean operation REJECTED if target path contains source files.

---

## BINV-08: Semantic Versioning for Deploy Tags

**What it prevents:** Arbitrary version strings, deployment confusion.

**What it requires:** OP_BUILD_DEPLOY creates a git tag. Tag MUST follow semantic versioning (MAJOR.MINOR.PATCH) or project-specific scheme validated against history.

**Source of this rule:**
- Standard release engineering.
- gonka history: tag conventions tracked for rollback.

**Consequence of violation:** Tag creation REJECTED with "invalid version format".

---

## BINV-09: Parallel Build Conflict Detection

**What it prevents:** Concurrent compilation of shared dependencies.

**What it requires:** Before parallel build dispatch, resource_bitmask overlap check. Any overlap → serialize.

**Source of this rule:**
- bun PR #30412: ~170 agents building shards with scoped isolation.
- CONFLICT_RULES.md CR-02.

**Consequence of violation:** Parallel builds serialized automatically. No corruption.

---

## BINV-10: Build System Auto-Detection

**What it prevents:** Wrong build command, wasted execution.

**What it requires:** OP_BUILD_COMPILE auto-detects build system from project files. If ambiguous, model must specify. No guesswork.

**Source of this rule:**
- embryo projectmap: file index enables quick build system detection.
- hermes-agent: error classification helps identify wrong build command.

**Consequence of violation:** If build system unknown, compile returns "unknown_build_system" with detected candidates.
