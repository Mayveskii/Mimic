# Build Domain — Sources

Where the build domain behavior comes from.

---

## bun (PR #30412)

**Principles taken:**
- Phase-d-build-queue: build only after plan validation.
- Phase-d-crate-shard: ~170 agents building shards in parallel with scoped isolation.
- Edit scope isolation: each shard edits only its own output directory.
- 2-vote before deploy: critical operations verified independently.

**What Mimic does with them:**
Build processes validated before execution. Parallel shards with conflict detection. Resource bitmask isolation per shard. 2-vote on deploy.

**What Mimic does NOT copy:**
- Bun's crate system (Mimic is language-agnostic).
- Bun's specific build agent orchestration.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- pkg/hunt/: 17-module hunt pipeline includes build artifact assessment.
- pkg/checkpoints/: C6-C10 gates include build and test checkpoints.
- pkg/projectmap/: SQLite-based project navigation indexes build outputs.

**What Mimic does with them:**
Build quality assessed via hunt pipeline. Build/test checkpoints gate progression. Project map indexes build artifacts for symbol lookup.

**What Mimic does NOT copy:**
- Embryo's Go-specific build commands (Mimic detects build system).
- Embryo's Makefile conventions.

---

## code-mode (Mayveskii/code-mode)

**Principles taken:**
- Budget enforcement: maxTurns + maxBudgetUsd per session.
- Concurrency control: up to 10 tools in parallel.

**What Mimic does with them:**
Build processes consume energy budget. Parallel build shards limited to 10 concurrent.

---

## hermes-agent

**Principles taken:**
- Error classifier: build errors classified as syntax, logic, or environment.
- Iteration budget: limited retries for flaky tests.
- Credential pool: SSH keys and deploy tokens managed centrally, never hardcoded.

**What Mimic does with them:**
Build errors classified before retry decision. Flaky test retries limited by budget. Deploy uses credential pool.

**What Mimic does NOT copy:**
- Hermes-agent's Python virtualenv handling.
- Hermes-agent's specific CI integration.

---

## graphify

**Principles taken:**
- AST extraction: two-pass structural + call-graph analysis.
- Confidence labels: every extracted pattern has confidence score.

**What Mimic does with them:**
Build output (compiled binaries) can be analyzed via graphify for pattern extraction. High-confidence patterns from successful builds are preferred.

---

## caveman

**Principles taken:**
- Sensitive path protection: .env, credentials in build output → auto-detect and block.
- File type detection: magic bytes identify build artifacts.

**What Mimic does with them:**
Deploy scripts scanned for credential patterns. Build artifacts typed for appropriate handling.

---

## rustnet

**Principles taken:**
- Sandbox: Landlock on Linux, Seatbelt on macOS, Job Objects on Windows.
- Process isolation: spawned build processes cannot escape workspace.

**What Mimic does with them:**
Build and test processes run in sandbox. Network access blocked unless explicitly enabled (for integration tests).

**What Mimic does NOT copy:**
- Rustnet's eBPF integration (Mimic uses Landlock/Seatbelt).
- Rustnet's sandbox policy language.

---

## Standard Practice

**Principles taken:**
- Semantic versioning: MAJOR.MINOR.PATCH for releases.
- Clean separation of source and build output.
- Test coverage measurement.

**What Mimic does with them:**
Version tags validated. Build output isolated. Coverage reported in test results.
