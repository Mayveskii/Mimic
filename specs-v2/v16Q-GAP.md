# v16Q Gap: Real-World Patch Machine

## The Gap

Mimic currently executes what a model asks. It does not **find** problems,
does not **validate** fixes, does not **remember** successes, and cannot
**prove** that anything real happened.

The gap is the distance between "passive tool" and **neighbor-from-the-git-desk**:
an agent that looks at real code, finds real issues, proposes real patches,
and gets them **merged**.

## Success Metric: MERGE COUNT FEEL

- **MERGE COUNT**: Number of PRs generated through Mimic that are merged into real repositories.
- **FEEL**: Patches are indistinguishable from human junior-dev work — relevant, tested, contextual, documented.

No tables. No synthetic scores. Only `git log --grep="Mimic"` and maintainer approvals.

## Current State (As-Is)

| What | State |
|------|-------|
| Sandbox | Broken — writes to CWD, no isolation |
| Hunt | Missing — model does not know what to fix |
| Pipeline | Stub — no validation, no rollback |
| Memory | None — each session starts from zero |
| Tools | 49 bloated — models hallucinate names |
| Proof | None — no honest eval possible |
| Slots | 298K embryo data — not migrated, not queryable |
| Distillation | Scripts exist — not run, not curated |

## Desired State (To-Be)

1. Model enters playground with real repo (e.g., `rtk`, `etcd`, `go-ethereum`)
2. Hunt system finds: "this pattern is fragile, 3 production repos fixed it this way"
3. Orchestrator validates: conflict matrix clear, budget sufficient, invariants hold
4. Model executes: writes patch in isolated worktree
5. Proof collected: `git diff`, `cargo test`, `git log`
6. Human reviews: patch is real, test passes, style matches
7. PR submitted, maintainer merges
8. Mimic extracts skill: "pattern X fix Y" → mesh slot with survival tracking

## Gap Closure Checklist

- [ ] Playground with mesh slots (production patterns, survival ≥ 0.7)
- [ ] Worktree sandbox per model with git proof
- [ ] Hunt system: assess → compress → search → rank
- [ ] Orchestrator pipeline: 6-stage with rollback
- [ ] 7 ACI tools instead of 49 bloated stubs
- [ ] Multi-model eval: qwen + kimi + minimax, consensus 2/3
- [ ] Cost tracking: tokens, time, dollars per run
- [ ] Human review gate: no PR without human approval
- [ ] Skill extraction: successful trajectory → mesh slot

## Why This Matters

> Future findings will connect people worldwide to solve tasks for progress at zero cost.

The only way to get there is **merge count**. Every merged patch is a proof
that Mimic amplified a model into a productive engineer. Every rejection is
a lesson that improves the mesh.

No merges = no progress. Merges = the path.

---

*Branch: v16Q*
*Owner: user writes code. Assistant documents gap and prepares playground.*
