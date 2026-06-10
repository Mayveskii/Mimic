# v16Q Architecture

## Runtime Components

```
┌─────────────────────────────────────────────────────────────┐
│                         GonkaGate                            │
│  qwen3-235b    kimi-k2.6    minimax-m2.7                   │
└────────┬─────────────────────┬──────────────────┬───────────┘
         │                     │                  │
    ┌────▼────┐           ┌────▼────┐       ┌────▼────┐
    │  Model  │           │  Model  │       │  Model  │
    │ Session │           │ Session │       │ Session │
    └────┬────┘           └────┬────┘       └────┬────┘
         │                     │                  │
         └─────────────────────┼──────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   SessionManager    │
                    │  (Go runtime)       │
                    └──────────┬──────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
    ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
    │ Worktree│          │ Project │          │  Mesh   │
    │  (git)  │          │  Map    │          │ Preload │
    └────┬────┘          │(SQLite) │          └────┬────┘
         │               └────┬────┘               │
         │                    │                    │
    ┌────▼────────────────────▼────────────────────▼────┐
    │                  C-Core (ops.c)                   │
    │  exec_sys_file_write  exec_git_add/commit        │
    │  ops_execute_chain  conflict_matrix              │
    └───────────────────────────────────────────────────┘
```

## Session Lifecycle

1. **Init**
   - `validateGitReality()` → HEAD SHA, dirty check
   - `provisionWorktree()` → `git worktree add --detach`
   - `bootstrapProjectmap()` → full SQLite FTS5 scan
   - `preloadMesh()` → domain-specific slots
   - `loadRepoConfig()` → `.mimic/<repo>/config.yaml`
   - `lockBudget()` → max_tokens, max_iterations fixed
   - `setBaseRepo()` → C-core wired to worktree

2. **Execute**
   - `Hunt.Assess()` → complexity score
   - `Hunt.Compress()` → context within budget
   - `Hunt.Search()` → mesh query
   - `Hunt.Rank()` → Z-density ranking
   - `Pipeline.Run()` → 6-stage execution
   - `CheckInvariants()` → before every opcode
   - `Checkpoint.Save()` → after every step

3. **Finalize**
   - `CollectProof()` → git status, git log, ls -la
   - `NormalizeOutput()` → raw → structured findings
   - `WriteArtifacts()` → run directory with all files
   - `ExtractSkills()` → successful trajectory → mesh slot
   - `DestroyWorktree()` → `git worktree remove`

## Data Flow

```
Model prompt
    → SessionManager
        → Context assembly (repo map + mesh slots + memory + primers)
        → Tool call generation
        → C-core execution (with invariant check)
        → Git state update (auto-index on WRITE)
        → Checkpoint save
        → Observation return
    → Model receives result
```

## Isolation Model

| Layer | Mechanism |
|-------|-----------|
| Process | Per-session goroutine |
| Filesystem | Git worktree `--detach` |
| Git state | Independent HEAD, index, working tree |
| C-core | `base_repo` set per session |
| Network | GonkaGate proxy (sequential) |

---

*Branch: v16Q*
