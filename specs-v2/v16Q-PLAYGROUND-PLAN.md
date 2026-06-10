# v16Q Playground Preparation Plan

**Status: Awaiting approval**
**Goal: Runnable multi-model eval playground with real slots, real repos, real proof.**

---

## Phase 1: Data & Slots (2-3 days)

### 1.1 Distillation Run
- Run `make distill` on top-20 repos from `mimicrya/repos-manifest.yaml`
- Filter: survival index ≥ 0.7, Z-density ≥ 0.5
- Target: ~500 high-quality slots per domain (go, rust, python)

### 1.2 TextSlot Migration (ADR-005)
- Convert gob slots → markdown-native TextSlot format
- Schema: frontmatter (domain, layer, state_hash, survival, z_density) + body
- Store in `.mimic/slots/<domain>/`

### 1.3 Mesh Registry Load
- Load slots into mesh registry (SQLite + embedding)
- Verify query: `MESH_QUERY domain=rust, layer=error_handling` returns ranked results
- Build local fallback index (cosine similarity on int8[384])

## Phase 2: Sandbox Setup (1-2 days)

### 2.1 Subject Repos
- Primary: `Mayveskii/rtk` (Rust, 115 files, cargo test passes)
- Secondary: `etcd-io/etcd` (Go, distributed systems patterns)
- Tertiary: `ethereum/go-ethereum` (Go, consensus/security patterns)
- Clone all to `/opt/mimic-sandbox/subjects/`

### 2.2 Worktree Provisioning
- Script: `scripts/wt_create.sh <repo> <model>` → `git worktree add --detached`
- 3 worktrees per repo: `<repo>-qwen`, `<repo>-kimi`, `<repo>-minimax`
- Cleanup: `scripts/wt_destroy.sh` → `git worktree remove + prune`

### 2.3 Repo-Scoped Config
- `.mimic/rtk/config.yaml` — model profiles, categories, budget
- `.mimic/rtk/personas/` — domain-specific reviewers (rust, safety, perf)
- `.mimic/rtk/primers/` — project conventions (error handling, async patterns)
- `.mimic/rtk/waivers/` — known false positives

## Phase 3: Model Integration (1 day)

### 3.1 GonkaGate Verification
- Verify proxy connectivity: qwen, kimi, minimax
- Test sequential execution: model A finishes → model B starts
- Confirm no rate limiting at current volume

### 3.2 Execution Harness
- `cmd/playground/main.go` — SessionManager per model
- Sequential run with semaphore (1 active model)
- Timeout: 10 min per model per task

## Phase 4: Eval Harness (2 days)

### 4.1 Honest Eval Script
- `scripts/eval_honest.py`
- After each execution: `git status`, `git log --oneline -5`, `ls -la`
- Output: JSON with proof artifacts paths
- Rule: no eval without proof

### 4.2 Normalization
- Raw model output → `Finding{file, line, summary, severity, confidence}`
- Regex fallback + cheap model (qwen-mini) for extraction
- Store: `.mimic/runs/<model>/<timestamp>/findings.json`

### 4.3 Consensus Aggregation
- 3 models → normalize → compare
- 2/3 agreement = preliminary fact
- Disagreement → flag for human review

### 4.4 Report Generation
- `report.md`: aggregated findings, stats, cost
- `agent_handoff.md`: context for human reviewer
- `run-log.jsonl`: tokens, time, cost per step

## Deliverables

| Artifact | Location | Purpose |
|----------|----------|---------|
| Mesh slots | `.mimic/slots/` | Queryable patterns for hunt |
| Subject repos | `/opt/mimic-sandbox/subjects/` | Real code to patch |
| Worktrees | `/opt/mimic-sandbox/subjects/<repo>-<model>/` | Isolated execution |
| Config | `.mimic/<repo>/` | Repo-scoped behavior |
| Eval harness | `scripts/eval_honest.py` | Proof + consensus |
| Reports | `.mimic/runs/` | Auditable artifacts |

## Approval Required

- [ ] Approve Phase 1: which 20 repos to distill?
- [ ] Approve Phase 2: which 3 subject repos?
- [ ] Approve Phase 3: GonkaGate proxy confirmed working?
- [ ] Approve Phase 4: consensus threshold (2/3 or strict 3/3)?

**After approval: I prepare the playground. You write the code.**

---

*Branch: v16Q*
