# v16Q Roadmap

## Branch Topology

```
v16Q                          ← baseline spec + core architecture
├── v16Q-qq                   ← qwen-specific experiments
├── v16Q-qz                   ← qz-model tuning (multi-model consensus)
├── v16Q-qa                   ← aider/ai-reviewer integration
├── v16Q-mem                  ← mem0 + memory layer experiments
└── v16Q-swe                  ← SWE-agent ACI + trajectory mode
```

## v16Q (Current)

- [x] Semantic diff (as-is vs to-be)
- [x] Behavior sources matrix (42 behaviors)
- [x] Architecture specification
- [x] Model integration spec (GonkaGate)
- [x] Vision document
- [ ] Implementation: SessionManager
- [ ] Implementation: Hunt system (4 modules)
- [ ] Implementation: Pipeline wiring
- [ ] Implementation: Worktree isolation fix
- [ ] Implementation: Auto-index hook

## v16Q-qq (Qwen Solo Tuning)

- Optimize prompts for `qwen/qwen3-235b`
- Single-model baseline before multi-model
- Measure: tokens/call, latency, patch quality
- Target: real patches on rtk test subject

## v16Q-qz (Consensus Thresholds)

- Multi-model consensus tuning
- 2/3 vote threshold calibration
- Disagreement resolution rules
- Cost vs accuracy tradeoff measurement

## v16Q-qa (Review Mode)

- ai-reviewer personas integration
- Repo-scoped `.mimic/<repo>/personas/`
- Pre-run explainer → reviewer → aggregate pipeline
- Apply to Mimic's own PRs

## v16Q-mem (Memory Experiments)

- Semantic memory API (`add/search/decay`)
- Session prime + nudge queue
- Trajectory recording + replay
- Skill extraction from successful runs

## v16Q-swe (ACI Mode)

- Constrained tool set (view/search/edit/bash)
- Trajectory demonstrations
- SWE-bench style evaluation
- Docker sandbox option (alternative to worktree)

---

*Branch: v16Q*
