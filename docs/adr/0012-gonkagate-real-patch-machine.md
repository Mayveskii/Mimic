# ADR-0012: GonkaGate Integration & Real Patch Machine for v16Q

## Decision

Use GonkaGate as the default OpenAI-compatible LLM provider for Mimic v16Q, wire it through a hot-reload configuration and a `local → medium → top` model cascade, and replace the critical stubs in `session.Manager.Execute()`, `orchestrator.PlanValidateExecStage`, `orchestrator.VerifyStage`, and `GeneratePlanFromGoal()` so that Mimic can generate, verify, and report real patches.

## Why (formal)

Mimic's value proposition is **MERGE COUNT FEEL**: PRs generated via Mimic that are merged into real repos and feel like human junior-dev work. Waves 0-4 built all supporting infrastructure (sandbox, mesh, memory, consensus, graphify, cost), but the **brain is missing** — no LLM is actually invoked during execution. The behavior matrix (`specs-v2/v16Q-BEHAVIOR.md`) and model specification (`specs-v2/v16Q-MODELS.md`) already mandate:

- `local/medium/top` model cascade from `Mayveskii/embryo`.
- Multi-model consensus (2/3) from `Mayveskii/bun` two-vote verify.
- Per-model worktree isolation from `specs-v2/v16Q-MODELS.md`.
- Hot-reload config from `go-service-template-rest` koanf pattern and existing `RepoConfig` personas/primers/waivers.

GonkaGate is the only provider known to be online (`project_context_main/bridge/config/models.yaml`: status online 2026-05-31), cost-measured (~$0.000353/1M tokens), and providing all three required models (`qwen/qwen3-235b`, `moonshotai/kimi-k2.6`, `minimaxai/minimax-m2.7`) under a single API key.

## Measured

| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| `session.Manager.Execute()` calls LLM | no | yes | +real inference |
| `orchestrator.PlanValidateExecStage` executes real ops | no (`/tmp/test`) | yes | +real execution |
| `orchestrator.VerifyStage` performs 2-vote verify | no | yes | +verification |
| Config change requires recompile | yes | no | +hot-reload |
| E2E patch generation on `rtk` | none | ≥1 passing patch | +MERGE FEEL signal |
| Cost/value tracked per finding | cost only | cost + estimated value | +$$ signal |

## Invariant

Affected invariants from `project_context_main/PHILOSOPHY/CORE_INVARIANTS.md`:

- **6. Работа На Результат** — token budget is measured per real call and recorded in `run-log.jsonl`.
- **7. Safety: reversible, verified, isolated, tracked, reproducible** — every write happens in a git worktree; proof collection (`git status`, `git log`, `ls -la`) is mandatory; critical ops require 2-vote verify.
- **2. Mesh = Бинарная Память Самого Mimic** — model consumes mesh as health pack and uses it, not verifies it.

Verification:
- `go test ./...` green.
- `cmd/playground` run on `rtk` produces real `git diff`, passing `BUILD_TEST`, and `report.md`.
- `scripts/eval_honest.py` proof collection succeeds for every E2E run.

## Alternatives

1. **Native Kimi/Moonshot API** — rejected: direct API returns 401 Invalid Authentication (`project_context_main/MEMORY/BRIDGE_SETUP.md`); GonkaGate is the working proxy.
2. **OpenRouter** — rejected: not currently configured or funded in this project; GonkaGate key and models are already known and documented.
3. **Local vLLM/ollama** — rejected: requires separate GPU inference setup and embedding service; adds operational complexity before MVP.
4. **Rewrite C-core first** — rejected: `references/git-execution-engine.md` already decided to use git plumbing as execution substrate; rewriting C-core is parallel work, not a blocker.

## Consilium

Not a model vote. Decision driven by:
- Founder invariant: use known working provider (`docs/gonkagate/`).
- Founder's requirement: no time estimates, serious scope, best implementation at every level.
- Existing bridge config: `gonka-qwen3-235b` is default routing.

## Test

- Unit: `internal/model/gonkagate_test.go` with local HTTP test server.
- Unit: `internal/config/watch_test.go` for hot-reload.
- Integration: `internal/orchestrator/pipeline_test.go` with mock model caller.
- E2E: `cmd/playground` on `rtk` repo.
- CI: `make check` green.

## Artifact precision

- Source behaviors: `mesh_agent_cascade`, `two_vote_verify`, `iteration_budget`, `error_classifier_failover`, `tool_guardrails`, `closed_learning_loop` from `mimicrya/behavior-sources.yaml`.
- Survival index of source behaviors: high (all from production or founder-verified repos).
- Invariant coverage: safety, budget, consensus, isolation.
- Extraction reproducibility: GonkaGate API is OpenAI-compatible and already tested in `test/battlefield/gonkagate_e2e_test.py`.
