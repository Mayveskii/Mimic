# ACTION-CHAINS.md — Mimic

Complete map of every action chain (OpPacket sequence) that Mimic MUST support, derived from binary-mesh proven capabilities. Every chain references its binary-mesh source, the Mimic OpCode sequence, current spec coverage, and gap status.

---

## 1. Chain Catalog

### Legend

- **Source**: binary-mesh endpoint or package that proves this capability
- **OpPacket Chain**: sequence of OpCodes in Mimic C-core
- **Spec Coverage**: which Mimic spec documents this (if any)
- **Gap**: EXISTS (spec+code), SPECIFIED (spec only), MISSING (neither)

---

## 2. File Operations

### CHAIN-01: Read File

```
Source: /api/files/read (binary-mesh)
Chain: OP_SYS_FILE_EXISTS → OP_IO_READ
OpPacket:
  [0] OP_SYS_FILE_EXISTS  {path: "/path/to/file"}
  [1] OP_IO_READ           {path: "/path/to/file", offset: 0, limit: 2000}
Invariants: file must exist before read
Spec: 07-RESOURCES.md (OP_SYS_FILE_EXISTS), 05-BEHAVIOR.md (#14 workspace indexing)
Gap: SPECIFIED — OP_IO_READ executor missing (0/46 I/O ops implemented)
QAC: QAC-2 (precondition: file exists), QAC-3 (energy: 1.0 tokens, 10us)
```

### CHAIN-02: Write File

```
Source: /api/files/write (binary-mesh)
Chain: OP_SYS_FILE_EXISTS → [OP_SYS_DIR_CREATE] → OP_IO_WRITE
OpPacket:
  [0] OP_SYS_FILE_EXISTS  {path: "/path/to/file"}
  [1] OP_SYS_DIR_CREATE   {path: "/path/to/"} (if parent missing)
  [2] OP_IO_WRITE          {path: "/path/to/file", content: "...", create: true}
Invariants: parent dir exists, write within buffer limit (1MB per QAC-3)
Spec: 07-RESOURCES.md (OP_SYS_DIR_CREATE, OP_IO_WRITE)
Gap: MISSING — OP_IO_WRITE executor missing
QAC: QAC-2, QAC-3, QAC-4 (WRITE x READ conflict if concurrent)
```

### CHAIN-03: Create Directory

```
Source: /api/files/mkdir (binary-mesh)
Chain: OP_SYS_FILE_EXISTS → OP_SYS_DIR_CREATE
OpPacket:
  [0] OP_SYS_FILE_EXISTS  {path: "/path/to/dir"}
  [1] OP_SYS_DIR_CREATE   {path: "/path/to/dir"}
Invariants: dir must not already exist (or idempotent mkdir -p)
Spec: 07-RESOURCES.md, ops.c (op_sys_dir_create_execute EXISTS)
Gap: EXISTS — fully implemented
```

### CHAIN-04: Delete File

```
Source: /api/files/delete (binary-mesh)
Chain: OP_SYS_FILE_EXISTS → OP_SYS_FILE_DELETE
OpPacket:
  [0] OP_SYS_FILE_EXISTS   {path: "/path/to/file"}
  [1] OP_SYS_FILE_DELETE   {path: "/path/to/file"}
Invariants: file exists, DELETE x WRITE conflict (BEHAVIOR.md #2)
Spec: 07-RESOURCES.md (OP_SYS_FILE_DELETE)
Gap: SPECIFIED — executor missing
QAC: QAC-4 (DELETE x WRITE = 1 in conflict_matrix)
```

### CHAIN-05: Rename/Move File

```
Source: /api/files/rename (binary-mesh)
Chain: OP_SYS_FILE_EXISTS → OP_SYS_FILE_MOVE
OpPacket:
  [0] OP_SYS_FILE_EXISTS   {path: "/old/path"}
  [1] OP_SYS_FILE_MOVE     {source: "/old/path", dest: "/new/path"}
Invariants: source exists, dest parent exists, no overwrite without flag
Spec: 07-RESOURCES.md (OP_SYS_FILE_MOVE)
Gap: SPECIFIED — executor missing
```

---

## 3. Git Operations

### CHAIN-06: Atomic Commit

```
Source: git_scenarios.c (atomic_commit scenario), /api/commits (binary-mesh)
Chain: OP_GIT_STATUS → OP_GIT_DIFF → OP_GIT_ADD → OP_GIT_COMMIT
OpPacket:
  [0] OP_GIT_STATUS    {path: "."}
  [1] OP_GIT_DIFF      {staged: true}
  [2] OP_GIT_ADD        {files: ["file1.go", "file2.go"]}
  [3] OP_GIT_COMMIT     {message: "feat: add new feature"}
Invariants: status clean after commit, no conflicts in diff
Spec: 07-RESOURCES.md (git_scenarios.c), 04-SCENARIOS.md
Gap: EXISTS — fully implemented in git_scenarios.c
QAC: QAC-2 (status must be clean), QAC-13 (Revert detection on revert commits)
```

### CHAIN-07: Safe Merge (FF-Only)

```
Source: git_scenarios.c (safe_merge scenario)
Chain: OP_GIT_FETCH → OP_GIT_DIFF → OP_GIT_MERGE (ff-only)
OpPacket:
  [0] OP_GIT_FETCH     {remote: "origin"}
  [1] OP_GIT_DIFF       {base: "main", head: "origin/main"}
  [2] OP_GIT_MERGE       {source: "origin/main", ff_only: true}
Invariants: fast-forward only, no merge commits
Spec: 07-RESOURCES.md (git_scenarios.c)
Gap: EXISTS — fully implemented
QAC: QAC-4 (MERGE x COMMIT conflict), QAC-9 (never force merge → AP-13)
```

### CHAIN-08: Feature Branch

```
Source: git_scenarios.c (feature_branch scenario)
Chain: OP_GIT_BRANCH → OP_GIT_CHECKOUT
OpPacket:
  [0] OP_GIT_BRANCH     {name: "feat/new-feature"}
  [1] OP_GIT_CHECKOUT    {ref: "feat/new-feature"}
Invariants: branch must not exist
Spec: 07-RESOURCES.md
Gap: EXISTS
```

### CHAIN-09: Hotfix Branch

```
Source: git_scenarios.c (hotfix scenario)
Chain: OP_GIT_BRANCH → OP_GIT_COMMIT → OP_GIT_CHECKOUT(target) → OP_GIT_MERGE
OpPacket:
  [0] OP_GIT_BRANCH       {name: "hotfix/critical-fix"}
  [1] OP_GIT_COMMIT        {message: "fix: critical security issue"}
  [2] OP_GIT_CHECKOUT      {ref: "main"}
  [3] OP_GIT_MERGE         {source: "hotfix/critical-fix"}
Invariants: must merge back into target branch
Spec: 07-RESOURCES.md
Gap: EXISTS
```

### CHAIN-10: CI Diff Check

```
Source: git_scenarios.c (ci_diff_check scenario)
Chain: OP_GIT_DIFF → validate_output
OpPacket:
  [0] OP_GIT_DIFF        {base: "main", head: "HEAD"}
  [1] OP_HASH_SHA256      {data: diff_output} (integrity check)
Invariants: no trailing whitespace, no merge conflict markers
Spec: 07-RESOURCES.md
Gap: EXISTS (diff) + MISSING (validation logic not in ops.c)
```

### CHAIN-11: Git Push (Safe)

```
Source: /api/commits (binary-mesh), /api/commits/action
Chain: OP_GIT_STATUS → OP_GIT_PUSH
OpPacket:
  [0] OP_GIT_STATUS       {path: "."}
  [1] OP_GIT_PUSH         {remote: "origin", branch: "main"}
Invariants: working tree clean, branch up to date with remote
Spec: 07-RESOURCES.md (OP_GIT_PUSH)
Gap: SPECIFIED — push executor exists in git_ops.c
QAC: QAC-4 (PUSH x CHECKOUT conflict), QAC-9 (never force push → AP-13)
```

---

## 4. Build & Compile

### CHAIN-12: Compile Project

```
Source: /api/checkpoints/run (binary-mesh: c6_govet), /api/exec
Chain: OP_BUILD_COMPILE → OP_BUILD_TEST
OpPacket:
  [0] OP_BUILD_COMPILE   {command: "go build ./...", cwd: "/path/to/project"}
  [1] OP_BUILD_TEST       {command: "go test ./...", cwd: "/path/to/project"}
Invariants: compile succeeds before test runs
Spec: 07-RESOURCES.md (OP_BUILD_COMPILE, OP_BUILD_TEST)
Gap: SPECIFIED — executors missing (0/5 Build ops implemented)
QAC: QAC-3 (energy: 5.0 tokens, ~1500us estimated), QAC-4 (COMPILE x CHECKOUT conflict)
```

### CHAIN-13: Build Clean

```
Source: /api/exec (binary-mesh)
Chain: OP_BUILD_CLEAN
OpPacket:
  [0] OP_BUILD_CLEAN      {command: "go clean -cache", cwd: "/path/to/project"}
Gap: SPECIFIED — executor missing
```

### CHAIN-14: Deploy

```
Source: /api/exec, deploy_cycle (binary-mesh MCP)
Chain: OP_GIT_PUSH → OP_BUILD_COMPILE → OP_BUILD_DEPLOY
OpPacket:
  [0] OP_GIT_PUSH         {remote: "origin", branch: "main"}
  [1] OP_BUILD_COMPILE     {command: "go build -o bin/mimic ./cmd/mimic"}
  [2] OP_BUILD_DEPLOY      {command: "systemctl restart mimic", host: "192.168.111.25:2022"}
Invariants: compile succeeds before deploy, health check after deploy
Spec: 07-RESOURCES.md (OP_BUILD_DEPLOY)
Gap: MISSING — OP_BUILD_DEPLOY executor missing, SSH not in OpCode set
QAC: QAC-3 (energy: 8.0 tokens, ~5000us), QAC-9 (AP-07: never hardcode deploy keys)
```

---

## 5. Mesh & Knowledge Operations

### CHAIN-15: Mesh Query

```
Source: /api/mesh/query (binary-mesh)
Chain: OP_HASH_SHA256 → OP_DECOMPRESS_GZIP → [internal: si_query_domain_layer]
OpPacket:
  [0] OP_HASH_SHA256      {data: query_text} (hash for cache lookup)
  [1] OP_DECOMPRESS_GZIP   {data: cached_slot} (if cache hit)
  [internal: si_query_domain_layer(domain, layer)]
Invariants: RAG without survival signal = unverified (BEHAVIOR.md #12)
Spec: 07-RESOURCES.md (binary_rag), 05-BEHAVIOR.md (#12 Binary RAG)
Gap: SPECIFIED — libbmap.a functions exist but no .c source
QAC: QAC-1 (SI from git blame), QAC-5 (Z-density > 0)
```

### CHAIN-16: Mesh Enrich

```
Source: /api/mesh/enrich (binary-mesh)
Chain: OP_IO_READ → OP_HASH_SHA256 → OP_COMPRESS_GZIP → [internal: si_insert]
OpPacket:
  [0] OP_IO_READ           {path: "data/new_patterns.bmap"}
  [1] OP_HASH_SHA256       {data: patterns} (integrity before insert)
  [2] OP_COMPRESS_GZIP      {data: patterns} (compress before storage)
  [internal: si_insert(domain, layer, modality, pattern_name)]
Invariants: every inserted slot has >= 1 invariant (QAC-2)
Spec: 07-RESOURCES.md (slot_write), 09-DISTILLATION-ARTIFACTS.md
Gap: SPECIFIED — si_insert in libbmap.a (no .c source)
QAC: QAC-2, QAC-7, QAC-9 (polarity: POSITIVE for new slots)
```

### CHAIN-17: Mesh Distill

```
Source: /api/mesh/distill (binary-mesh), cmd/distill (planned)
Chain: OP_GIT_CLONE → [internal: git blame] → [internal: survival_index] → OP_COMPRESS_GZIP → [internal: si_insert]
OpPacket:
  [0] OP_GIT_CLONE        {url: "https://github.com/etcd-io/etcd", path: "/tmp/distill/etcd"}
  [internal: git blame → survival_index computation]
  [internal: extract functions with SI >= 0.7]
  [1] OP_COMPRESS_GZIP     {data: artifact_protobuf}
  [internal: si_insert(domain, layer, modality, pattern_name)]
Invariants: survival_index >= 0.7 for slot candidate, < 0.1 = discard
Spec: 09-DISTILLATION-ARTIFACTS.md (Section 4), 05-BEHAVIOR.md (#4)
Gap: MISSING — distillation pipeline not implemented (cmd/distill/main.go planned)
QAC: QAC-1 (SI from git blame), QAC-10 (temporal consistency), QAC-13 (revert detection)
```

### CHAIN-18: Mesh Reload

```
Source: /api/mesh/reload (binary-mesh)
Chain: OP_HASH_SHA256 → OP_DECOMPRESS_GZIP → [internal: bmap re-read + si_rebuild]
OpPacket:
  [0] OP_HASH_SHA256       {data: current_bmap} (verify on-disk hash)
  [1] OP_DECOMPRESS_GZIP   {data: bmap_data} (decompress for verification)
  [internal: bmap re-read from disk + si_build_from_bmap]
Invariants: hash matches stored hash, re-index after reload
Spec: 07-RESOURCES.md (snapshot_diff, drift_detect)
Gap: SPECIFIED — libbmap.a functions exist
QAC: QAC-8 (integrity), QAC-10 (re-validate SI)
```

### CHAIN-19: Add Slot (Manual)

```
Source: /api/mesh/add-slot (binary-mesh)
Chain: OP_HASH_SHA256 → OP_COMPRESS_GZIP → [internal: si_insert]
OpPacket:
  [0] OP_HASH_SHA256       {data: slot_data}
  [1] OP_COMPRESS_GZIP      {data: slot_data}
  [internal: si_insert with polarity, counter_pattern_id for NEGATIVE]
Invariants: slot has >= 1 invariant (QAC-2), NEGATIVE must have counter_pattern (QAC-9)
Spec: 07-RESOURCES.md, 10-QUALITY-GATES.md (QAC-9)
Gap: SPECIFIED — needs polarity field in slot schema
QAC: QAC-2, QAC-7, QAC-9
```

---

## 6. RAG Operations

### CHAIN-20: RAG Search

```
Source: /api/rag/search (binary-mesh)
Chain: OP_HASH_SHA256 → [internal: int8_quantize → batch_cosine_int8 → top-k → survival+Z-density boost]
OpPacket:
  [0] OP_HASH_SHA256       {data: query} (cache lookup)
  [internal: int8_quantize(query)]
  [internal: si_query_domain(domain)]
  [internal: batch_cosine_int8(query_vec, candidates)]
  [internal: rank by survival_index + z_density]
Invariants: RAG without survival signal = unverified (BEHAVIOR.md #12)
Spec: 07-RESOURCES.md (binary_rag flow), 05-BEHAVIOR.md (#12)
Gap: SPECIFIED — libbmap.a functions exist
QAC: QAC-1, QAC-5, QAC-7 (precision > 0 for deep cache results)
```

### CHAIN-21: RAG Index

```
Source: /api/rag/index (binary-mesh)
Chain: OP_IO_READ → OP_HASH_SHA256 → [internal: 5-signal index]
OpPacket:
  [0] OP_IO_READ           {path: "workspace/file.go"}
  [1] OP_HASH_SHA256       {data: file_content} (cache key)
  [internal: 5-signal index: semantic, ast, kw, md, comment]
Invariants: index is stale after WRITE without re-index (BEHAVIOR.md #11)
Spec: 07-RESOURCES.md (workspace indexing), 05-BEHAVIOR.md (#11)
Gap: SPECIFIED — RAG indexing logic exists in binary-mesh
QAC: QAC-8 (AST parse success), QAC-10 (re-index on file change)
Measured: 38829 AST symbols, 117230 KW terms, 6367 MD entries, 4979 comments, 180684 qdrant points
```

---

## 7. Analysis Operations

### CHAIN-22: Code Quality Audit

```
Source: /api/analysis/code-quality (binary-mesh)
Chain: OP_IO_READ → [internal: AST + unused exports + architecture violations]
OpPacket:
  [0] OP_IO_READ           {path: "pkg/orchestrator/orchestrator.go"}
  [internal: AST extraction, unused export detection, architecture violation check]
Invariants: every finding has severity and impact score
Spec: Not in Mimic specs
Gap: MISSING — no Mimic spec for code quality audit
QAC: QAC-2 (findings as invariants), QAC-5 (Z-density of findings)
```

### CHAIN-23: Project Inference

```
Source: /api/analysis/project-inference (binary-mesh)
Chain: OP_IO_READ → [internal: classify app/vendor/framework]
OpPacket:
  [0] OP_IO_READ           {path: "go.mod"}
  [internal: classify by Dir + Imports (not just ImportPath)]
Invariants: framework bugs flagged as unfixable, app bugs as fixable
Spec: Not in Mimic specs
Gap: MISSING — no Mimic spec for project classification
Binary-mesh history: 9be7bdc → 924f44a → cd162f3 → 9f04abb (4 iterations to get right)
QAC: QAC-4 (framework/app classification affects conflict rules)
```

### CHAIN-24: Doc vs Code Comparison

```
Source: /api/analysis/doc-comparison (binary-mesh)
Chain: OP_IO_READ(doc) → OP_IO_READ(code) → [internal: compare signatures]
OpPacket:
  [0] OP_IO_READ           {path: "README.md"}
  [1] OP_IO_READ           {path: "pkg/agent/agent.go"}
  [internal: compare documented APIs vs actual function signatures]
Invariants: every documented API must match actual signature
Spec: Not in Mimic specs
Gap: MISSING
QAC: QAC-6 (consistency), QAC-8 (TEXT modality integrity)
```

### CHAIN-25: Full File Analysis

```
Source: /api/analysis/file (binary-mesh)
Chain: OP_IO_READ → [internal: symbols + callers + callees + issues]
OpPacket:
  [0] OP_IO_READ           {path: "pkg/orchestrator/orchestrator.go"}
  [internal: extract symbols, callers, callees, issues]
Invariants: every symbol has at least one caller or is an entry point
Spec: Not in Mimic specs
Gap: MISSING
QAC: QAC-2 (symbol extraction as invariant)
```

---

## 8. Project Map Operations

### CHAIN-26: Project Map Lookup

```
Source: /api/project-map/lookup (binary-mesh)
Chain: OP_IO_READ → [internal: SQLite lookup]
OpPacket:
  [0] OP_IO_READ           {path: ".project-map.json"} (index file)
  [internal: SQLite FTS5 lookup by symbol/package/route]
Invariants: index is stale after WRITE without re-index
Spec: 07-RESOURCES.md (emb_projectmap_sqlite), 05-BEHAVIOR.md (#11)
Gap: SPECIFIED — projectmap concept exists but not linked to OpPacket chain
QAC: QAC-5 (index quality), QAC-10 (stale detection)
Binary-mesh measured: 38829 symbols, 0 call edges (project_map_stats — needs population)
```

### CHAIN-27: Project Map Build

```
Source: /api/project-map/build (binary-mesh)
Chain: OP_IO_READ(workspace) → [internal: AST extraction + SQLite FTS5 insert]
OpPacket:
  [0] OP_IO_READ           {path: "entire workspace"}
  [internal: AST extraction for all .go files, insert into SQLite]
Invariants: build triggered after file WRITE operations
Spec: 05-BEHAVIOR.md (#11)
Gap: SPECIFIED — but no OpPacket chain for triggering build
QAC: QAC-3 (build energy cost), QAC-10 (re-build on change)
```

### CHAIN-28: Project Map Flow (Call Graph)

```
Source: /api/project-map/flow (binary-mesh)
Chain: [internal: trace callers/callees through graph]
OpPacket: (no I/O — pure graph traversal)
  [internal: trace function call flow through project map graph]
Invariants: every call edge is either EXTRACTED or INFERRED (graphify spec)
Spec: Not in Mimic specs directly (graphify has this)
Gap: MISSING — Mimic has no call graph traversal spec
QAC: QAC-5 (Z-density of graph edges), QAC-11 (cross-domain edges)
```

---

## 9. Chat & Solve Operations

### CHAIN-29: Chat (Single Turn)

```
Source: /api/chat (binary-mesh)
Chain: [internal: context_assemble] → [internal: inference_call] → [internal: tool_loop]
OpPacket: (orchestration chain, not direct OpCode)
  [internal: assemble context (mesh stats + quality + buckets)]
  [internal: callInference with budget + timeout]
  [internal: parse tool calls → execute → feed back]
Invariants: iteration budget (BEHAVIOR.md #7 from code-mode), 90s stale-stream (hermes-agent)
Spec: 05-BEHAVIOR.md (#6 Phase Transitions, #7 Permission Pipeline)
Gap: SPECIFIED — but no OpPacket chain for chat orchestration
QAC: QAC-3 (inference energy), QAC-6 (error classification), AP-06 (undifferentiated retry)
Binary-mesh context efficiency: 19635 tokens / 32768 max = 59.9% pressure
```

### CHAIN-30: Solve (Autonomous Fix)

```
Source: /api/solve (binary-mesh)
Chain: [CLASSIFY] → [PLAN] → [VALIDATE] → [EXEC] → [VERIFY] → [RESPOND]
OpPacket: (full pipeline, each phase may generate sub-chains)
  Phase CLASSIFY: OP_IO_READ(context) → [internal: classify intent + domain]
  Phase PLAN: [internal: mesh_query → build OpPacket chain]
  Phase VALIDATE: ops_validate_chain → check conflict_matrix + energy
  Phase EXEC: ops_execute_chain → measure latency
  Phase VERIFY: [internal: 2-vote verify] → check invariants
  Phase RESPOND: OP_COMPRESS_GZIP(result) → [internal: store feedback]
Invariants: no EXEC without VALIDATE pass, 2-vote for critical ops
Spec: 05-BEHAVIOR.md (#6, #8), 04-SCENARIOS.md
Gap: SPECIFIED — pipeline phases defined but not linked to concrete OpPacket chains
QAC: QAC-1 through QAC-13 (all gates apply)
```

### CHAIN-31: Dual Solve (A/B Comparison)

```
Source: /api/solve (binary-mesh dual_solve mode)
Chain: Solve(Approach A) ∥ Solve(Approach B) → [internal: compare effectiveness]
OpPacket: (two parallel solve pipelines)
  [0] fork: pipeline_A + pipeline_B
  [1] pipeline_A: CHAIN-30 (approach A)
  [2] pipeline_B: CHAIN-30 (approach B)
  [3] compare: score_A vs score_B → winner
Invariants: conflict_matrix[pipeline_A] × [pipeline_B] = all 0 (BEHAVIOR.md #14)
Spec: 05-BEHAVIOR.md (#14 Multi-task Pipeline Execution)
Gap: SPECIFIED — but no dual_solve OpPacket chain in specs
QAC: QAC-4 (cross-pipeline conflict), QAC-3 (energy for two pipelines)
Measured: 0% patch rate, avg 31 min latency — needs spec improvement
```

---

## 10. Survival & Distillation

### CHAIN-32: Survival Index Computation

```
Source: /api/survival (binary-mesh)
Chain: OP_GIT_CLONE → [internal: git blame] → [internal: compute survival]
OpPacket:
  [0] OP_GIT_CLONE        {url: "repo", path: "/tmp/survival/repo"}
  [internal: for each commit: git blame → surviving_lines / total_lines_added]
Invariants: survival >= 0.7 → slot candidate, < 0.1 → discard
Spec: 05-BEHAVIOR.md (#4), 09-DISTILLATION-ARTIFACTS.md
Gap: SPECIFIED — but no concrete implementation (pkg/survival/ exists in embryo)
QAC: QAC-1 (SI from git blame, not guessed), QAC-10 (temporal consistency), QAC-13 (revert detection)
```

### CHAIN-33: Distillation Pipeline

```
Source: /api/mesh/distill (binary-mesh), cmd/distill (planned)
Chain: OP_GIT_CLONE → OP_GIT_DIFF → [internal: blame] → [internal: extract] → OP_COMPRESS_GZIP → [internal: si_insert]
OpPacket:
  [0] OP_GIT_CLONE        {url: "etcd-io/etcd", commit: "abc123"}
  [1] OP_GIT_DIFF          {base: "HEAD~100", head: "HEAD"} (for recent changes)
  [internal: git blame per file → survival_index]
  [internal: extract functions with SI >= 0.7]
  [internal: extract REVERT commits → NEGATIVE artifacts (QAC-13)]
  [2] OP_COMPRESS_GZIP     {data: artifact_protobuf}
  [internal: si_insert(domain, layer, modality, pattern_name)]
Invariants: same commit + same tool → same bmap (09-DISTILLATION-ARTIFACTS.md #8)
Spec: 09-DISTILLATION-ARTIFACTS.md (Section 4)
Gap: MISSING — distillation pipeline not implemented
QAC: QAC-1 through QAC-13 (all quality gates apply per 10-QUALITY-GATES.md #3)
```

---

## 11. Quality & Checkpoint Operations

### CHAIN-34: Quality Matrix Query

```
Source: /api/quality (binary-mesh)
Chain: [internal: query quality state from recorder]
OpPacket: (no I/O — pure state query)
Invariants: every MCP handler records quality metrics (binary-mesh: b49511d)
Spec: Not in Mimic specs
Gap: MISSING — no quality matrix spec in Mimic
Current binary-mesh scores: L0=8000, L4=2100, L6=1931, L8=8005, L9=7072, composite=4784
QAC: All 13 QACs map to quality axes (10-QUALITY-GATES.md #4)
```

### CHAIN-35: Checkpoint Run

```
Source: /api/checkpoints/run (binary-mesh)
Chain: [internal: go_vet → pattern_scan → AST_scan → duplicate_check → transport_coupling → fix_verify]
OpPacket: (orchestration of multiple sub-checks)
  [0] OP_BUILD_COMPILE     {command: "go vet ./..."}
  [1] OP_IO_READ           {path: "workspace"} → [internal: pattern_scan]
  [2] OP_IO_READ           {path: "workspace"} → [internal: AST_scan]
  [3] OP_IO_READ           {path: "workspace"} → [internal: duplicate_check]
  [4] OP_IO_READ           {path: "workspace"} → [internal: transport_coupling]
  [5] OP_BUILD_COMPILE     {command: "go build"} → [internal: fix_verify]
Invariants: each checkpoint must pass before next
Spec: Not in Mimic specs
Gap: MISSING — no checkpoint spec
QAC: QAC-3 (compile energy), QAC-2 (pattern detection as invariant)
```

---

## 12. Observer & Attack Operations

### CHAIN-36: Observer Watch

```
Source: /api/observer (binary-mesh)
Chain: [internal: start observation] → [internal: monitor progress] → [internal: report findings]
OpPacket: (orchestration, not direct OpCode)
Invariants: observer tracks all hunt/solve operations
Spec: Not in Mimic specs
Gap: MISSING
QAC: QAC-2 (observation as invariant), QAC-5 (Z-density of findings)
```

### CHAIN-37: Hunt (Pattern Search)

```
Source: /api/corean/attack (binary-mesh)
Chain: [assess] → [compress] → [search(mesh)] → [rank(Z-density)] → [report]
OpPacket: (17-module hunt pipeline from embryo)
  [0] OP_IO_READ           {path: "target_codebase"}
  [internal: assess(complexity)]
  [internal: compress(context)]
  [internal: search(mesh for patterns)]
  [internal: rank(by Z-density)]
Invariants: hunt always returns ranked results with Z-density scores
Spec: Not in Mimic specs (embryo pkg/hunt/ has 17 files)
Gap: MISSING — hunt system not spec'd in Mimic
QAC: QAC-1 (SI of hunt results), QAC-5 (Z-density ranking)
```

---

## 13. Context & Session Operations

### CHAIN-38: Context Create

```
Source: /api/context/create (binary-mesh)
Chain: [internal: create session with budget + domain + direction]
OpPacket: (session management)
Invariants: every session has budget, denial tracker, context flow
Spec: 08-MODULES.md (internal/session/), 05-BEHAVIOR.md (#13 Context Flow)
Gap: SPECIFIED — session concept exists but not linked to OpPacket
QAC: QAC-3 (session budget), AP-19 (toolloop eating iterations)
```

### CHAIN-39: Context Enrichment

```
Source: /api/mesh/enrich (binary-mesh)
Chain: [internal: mesh_query → enrich task context]
OpPacket: (pure knowledge operation)
  [internal: mesh_query(task_description)]
  [internal: if DIRECT HIT → return pattern]
  [internal: if ENRICHMENT → expand context with related slots]
  [internal: if MISS → flag for distillation]
Invariants: DIRECT HIT <1μs, ENRICHMENT <100ms
Spec: 07-RESOURCES.md, 05-BEHAVIOR.md (#12 Binary RAG)
Gap: SPECIFIED — but no concrete OpPacket chain
QAC: QAC-5 (Z-density of enriched context), QAC-7 (precision of DIRECT HIT)
```

---

## 14. Exec Operations

### CHAIN-40: Shell Exec

```
Source: /api/exec (binary-mesh)
Chain: OP_SYS_EXEC
OpPacket:
  [0] OP_SYS_EXEC         {command: "go test ./...", cwd: "/path", timeout: 30000}
Invariants: OP_SYS_EXEC x OP_SYS_EXEC = 1 (conflict, BEHAVIOR.md #2)
Spec: 07-RESOURCES.md (OP_SYS_EXEC), 05-BEHAVIOR.md (#2 conflict matrix)
Gap: SPECIFIED — executor partially exists
QAC: QAC-3 (exec energy), QAC-4 (EXEC x EXEC conflict), AP-18 (no grep-based scans)
```

---

## 15. Goroutine Management

### CHAIN-41: Goroutine Add

```
Source: /api/goroutines/add (binary-mesh)
Chain: [internal: spawn background goroutine with ID + function]
OpPacket: (process management)
  [0] OP_PROC_SPAWN        {command: "background_task", async: true}
Invariants: goroutine has ID, can be tracked/updated/deleted
Spec: Not in Mimic specs (binary-mesh has /api/goroutines)
Gap: MISSING — no goroutine management spec in Mimic
QAC: QAC-4 (goroutine conflicts), QAC-14 (multi-task from bun)
```

---

## 16. LSP Operations (NEW — not in binary-mesh)

### CHAIN-42: Go to Definition

```
Source: LSP protocol (not in binary-mesh — new for Mimic)
Chain: OP_IO_READ → [internal: project_map_lookup(symbol)] → [internal: return location]
OpPacket:
  [0] OP_IO_READ           {path: "current_file.go"}
  [internal: project_map_lookup(symbol_name, mode="symbol")]
Invariants: symbol must exist in project map, O(1) lookup
Spec: Not in Mimic specs
Gap: MISSING — LSP operations not spec'd
QAC: QAC-5 (index quality), QAC-10 (stale index → wrong definition)
```

### CHAIN-43: Find References

```
Source: LSP protocol (not in binary-mesh)
Chain: [internal: project_map_lookup(symbol, mode="callers")]
OpPacket: (pure graph traversal)
  [internal: project_map_lookup(symbol, mode="callers")]
Invariants: all callers listed, INFERRED callers flagged
Spec: Not in Mimic specs
Gap: MISSING
QAC: QAC-5, QAC-11 (cross-domain callers)
```

### CHAIN-44: Hover/Type Info

```
Source: LSP protocol (not in binary-mesh)
Chain: [internal: project_map_lookup(symbol, mode="symbol")] → [internal: return type + doc]
OpPacket: (pure lookup)
Invariants: type information available for all indexed symbols
Spec: Not in Mimic specs
Gap: MISSING
QAC: QAC-8 (TEXT modality integrity)
```

---

## 17. Deploy Operations

### CHAIN-45: Full Deploy Cycle

```
Source: deploy_cycle MCP tool (binary-mesh)
Chain: OP_GIT_ADD → OP_GIT_COMMIT → OP_GIT_PUSH → OP_BUILD_COMPILE → OP_BUILD_DEPLOY
OpPacket:
  [0] OP_GIT_ADD           {files: "."}
  [1] OP_GIT_COMMIT        {message: "deploy: production update"}
  [2] OP_GIT_PUSH           {remote: "origin", branch: "main"}
  [3] OP_BUILD_COMPILE      {command: "go build -o bin/mimic ./cmd/mimic"}
  [4] OP_BUILD_DEPLOY       {command: "ssh root@host 'cd /root/mimic && git pull && go build && systemctl restart mimic'"}
Invariants: compile succeeds before deploy, health check after deploy
Spec: Not in Mimic specs
Gap: MISSING — no deploy spec
QAC: QAC-3 (deploy energy), QAC-4 (PUSH x CHECKOUT conflict), AP-07 (no hardcoded SSH keys)
Binary-mesh equivalent: deploy_cycle tool (commit → push → pull → build → restart)
```

---

## 18. Missing Chains Summary

| # | Chain | Binary-mesh Source | Mimic Gap | Priority |
|---|-------|-------------------|-----------|----------|
| 01 | Read File | /api/files/read | OP_IO_READ executor | P0 |
| 02 | Write File | /api/files/write | OP_IO_WRITE executor | P0 |
| 04 | Delete File | /api/files/delete | OP_SYS_FILE_DELETE executor | P0 |
| 05 | Rename File | /api/files/rename | OP_SYS_FILE_MOVE executor | P0 |
| 12 | Compile Project | /api/checkpoints | OP_BUILD_COMPILE executor | P0 |
| 14 | Deploy | deploy_cycle | OP_BUILD_DEPLOY + SSH | P1 |
| 22 | Code Quality Audit | /api/analysis/code-quality | No spec | P1 |
| 23 | Project Inference | /api/analysis/project-inference | No spec | P1 |
| 24 | Doc vs Code | /api/analysis/doc-comparison | No spec | P1 |
| 25 | Full File Analysis | /api/analysis/file | No spec | P1 |
| 28 | Call Graph Flow | /api/project-map/flow | No spec | P1 |
| 33 | Distillation Pipeline | /api/mesh/distill | Not implemented | P0 |
| 34 | Quality Matrix | /api/quality | No spec | P1 |
| 35 | Checkpoint Run | /api/checkpoints/run | No spec | P1 |
| 36 | Observer Watch | /api/observer | No spec | P2 |
| 37 | Hunt System | /api/corean/attack | No spec | P1 |
| 41 | Goroutine Management | /api/goroutines | No spec | P2 |
| 42 | Go to Definition | LSP (new) | No spec | P1 |
| 43 | Find References | LSP (new) | No spec | P1 |
| 44 | Hover/Type Info | LSP (new) | No spec | P2 |
| 45 | Full Deploy Cycle | deploy_cycle | No spec | P1 |

---

## 19. OpCode Implementation Gap

| Domain | OpCodes | Implemented | Missing | Coverage |
|--------|---------|-------------|---------|----------|
| Memory | 5 | 5 | 0 | 100% |
| Git | 11 | 11 | 0 | 100% |
| I/O | 5 | 0 | 5 | 0% ← P0 blocker |
| Build | 5 | 0 | 5 | 0% ← P0 blocker |
| Network | 6 | 0 | 6 | 0% ← P1 |
| Process | 4 | 0 | 4 | 0% ← P1 |
| Utility | 6 | 0 | 6 | 0% ← P2 |
| System | 9 | 2 | 7 | 22% ← P0 |
| **Total** | **46** | **18** | **28** | **39.1%** |

To reach 80% coverage: need 37/46 = 19 more executors. Priority: I/O (5) + System (7) + Build (5) + Process (2) = 19.

---

## 20. Action Chain → Quality Gate Mapping

Every chain must pass QAC gates before execution:

| Chain Type | Required QAC Gates | Failure Mode if Skipped |
|-----------|-------------------|------------------------|
| File I/O (01-05) | QAC-2 (exists), QAC-4 (conflict), QAC-3 (energy) | AP-04: unvalidated input, AP-05: context corruption |
| Git (06-11) | QAC-4 (conflict), QAC-9 (never-rules), QAC-13 (revert) | AP-13: override deny, AP-16: flip-flop decision |
| Build (12-14) | QAC-3 (energy), QAC-4 (conflict) | AP-17: blocking startup, AP-18: grep-based scan |
| Mesh (15-19) | QAC-1 (SI), QAC-2 (invariant), QAC-5 (Z-density), QAC-7 (precision), QAC-9 (polarity) | AP-11: stale cache, AP-29: swallowed error |
| RAG (20-21) | QAC-1 (SI), QAC-5 (Z-density), QAC-10 (temporal) | AP-11: stale state, AP-23: token overflow |
| Analysis (22-25) | QAC-2 (findings as invariants), QAC-8 (integrity) | AP-27: flaky without fix |
| Solve (29-31) | All 13 QACs | All 30 anti-patterns |
| Deploy (45) | QAC-3 (energy), QAC-7 (no secrets), QAC-9 (counter-patterns) | AP-07: hardcoded keys |
| LSP (42-44) | QAC-5 (index quality), QAC-10 (stale index) | AP-11: stale cached state |
