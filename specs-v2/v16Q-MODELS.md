# v16Q Model Integration Specification

## GonkaGate Proxy Configuration

| Parameter | Value |
|-----------|-------|
| Endpoint | GonkaGate proxy |
| Models | `qwen/qwen3-235b`, `moonshotai/kimi-k2.6`, `minimaxai/minimax-m2.7` |
| Execution | Sequential (1 API key) |
| Max tokens | 100,000 per call |
| Rate limiting | None observed |

## Model Categories

| Category | Role | GonkaGate Mapping |
|----------|------|-------------------|
| `local` | Fast, cheap, first attempt | `qwen/qwen3-235b` |
| `medium` | Balanced quality/cost | `moonshotai/kimi-k2.6` |
| `top` | Best quality, verification | `minimaxai/minimax-m2.7` |

## Multi-Model Consensus Protocol

1. **Execute** same task on all 3 models in isolated worktrees
2. **Normalize** raw output → structured `Finding{}` per model
3. **Compare** findings across models
4. **Rule**: 2/3 agreement = preliminary fact
5. **Disagreement** = flag for human review
6. **Proof** mandatory: `git status`, `git log --oneline -5`, `ls -la`

## Per-Model Worktree Isolation

```
/opt/mimic-sandbox/subjects/rtk/
├── rtk-qwen/     # worktree for qwen
├── rtk-kimi/     # worktree for kimi
└── rtk-minimax/  # worktree for minimax
```

- Same baseline SHA (HEAD)
- Independent execution
- Independent proof collection
- No shared mutable state

## Fuzzy Matching & Sanitization

| Issue | Mitigation |
|-------|------------|
| `<think>` tags (minimax) | Strip before parsing |
| Markdown fences | Extract JSON from ```json blocks |
| Truncated tool names (`SYS_ILE_WRITE`) | Levenshtein match against registry |
| Typos in paths (`mimic-andbox`) | Prefix validation + closest match |
| Trailing slash (`EISDIR`) | Normalize path before mkdir/open |

## Cost Tracking

```jsonl
{"persona_id":"qwen","model":"qwen3-235b","tokens_in":4200,"tokens_out":890,"time_ms":3400,"cost_usd":0.0012}
{"persona_id":"kimi","model":"kimi-k2.6","tokens_in":4200,"tokens_out":1200,"time_ms":2800,"cost_usd":0.0021}
```

- JSONL append to `.mimic/runs/<target>/run-log.jsonl`
- Per-model and aggregated totals in final report

---

*Branch: v16Q*
*Child branches: v16Q-qq (qwen solo tuning), v16Q-qz (consensus thresholds)*
