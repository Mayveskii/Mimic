# Pattern: atomic_commit

Tokenized process for safely committing changes to a git repository.

---

## Intent

"Save my work to the repository."  
"Commit these files."  
"Checkpoint my changes."

---

## Tokenized Process

```
[1] OP_GIT_STATUS   {path: "."}
    → Check working tree: modified, staged, untracked, conflicts.
    → If conflicts exist → STOP. Return "conflicts detected, resolve first."

[2] OP_GIT_DIFF     {staged: false}
    → Show what would be committed (all modified files).
    → Return diff to model for review.
    → Model must confirm "yes, commit these changes."

[3] OP_GIT_ADD      {files: ["file1.go", "file2.go"]}  
    → Stage only explicitly named files.
    → Never stage implicitly. Never `git add .`.

[4] OP_GIT_COMMIT   {message: "model-provided description"}
    → Create commit.
    → Return commit hash.
```

---

## Invariants

- After [4], `git status` shows clean working tree.
- Files staged in [3] = files in commit [4]. No hidden additions.
- Diff reviewed in [2] = diff in commit [4]. No surprise changes.
- Commit message is non-empty and describes changes.

---

## Never-Rules (Hard Constraints)

- NEVER `git add .` without explicit file list.
- NEVER commit if conflicts exist.
- NEVER commit without reviewing diff.
- NEVER empty commit message.

---

## Rollback

If [4] fails:
- Reset staged changes: `git reset HEAD`.
- Working tree restored to state before [3].
- Model receives: "commit failed, staged changes reset, working tree clean."

---

## Parameters

| Parameter | Required | Description |
|---|---|---|
| files | yes | Array of file paths to commit |
| message | yes | Commit message describing changes |
| path | no | Repository path (default: ".") |

---

## Energy Cost

| Operation | cost_tokens | cost_time_us | cost_memory_bytes |
|---|---|---|---|
| OP_GIT_STATUS | 3.0 | 5000 | 0 |
| OP_GIT_DIFF | 3.0 | 8000 | 0 |
| OP_GIT_ADD | 2.0 | 3000 | 0 |
| OP_GIT_COMMIT | 5.0 | 10000 | 0 |
| **Total** | **13.0** | **26000** | **0** |

# UNCERTAIN: actual costs need measurement on real repos

---

## Result

```
status: "success"
commit_hash: "abc123..."
files_committed: ["file1.go", "file2.go"]
message: "feat: add new endpoint"
clean_tree: true
```

---

## Source Behaviors

- **embryo:** BinaryRuntime provides OpCode-based git execution.
- **bun:** Phase graph with gate transitions. Never-rules enforce safety.
- **gastown:** ZFC state — status derived from observable reality, not cache.
- **anti-pattern AP-05/AP-24:** Validate structure stability before any mutation.
