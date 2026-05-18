# Git Domain — Tokenized Git Workflows

What a model can accomplish through Mimic in the git domain.

---

## What This Domain Does

Every git operation a model needs — from simple status checks to multi-step safe merges — is available as a tokenized process. The model does not need to know git internals. It expresses intent, Mimic provides the proven process, the model follows step by step.

The domain covers: status, diff, branch, checkout, commit, merge, fetch, push, rebase, clone.

---

## Processes

### atomic_commit

**When to use:**  
The model wants to save changes to the local repository. Any intent like "commit these files", "save my work", "checkpoint my changes" maps here.

**Goal:**  
Create a commit that contains exactly the changes the model intended, with a clean working tree afterward, no conflicts, no surprises.

**Chain (semantically):**

1. Check working tree state (status) — what files are modified, staged, untracked.
2. Review what will be committed (diff staged) — exact changes that will enter the commit.
3. Stage specified files (add) — only files the model explicitly selected.
4. Create commit (commit) — with a message that describes the changes.

**Hard constraints:**
- Never commit without reviewing diff first. The model must see what it is committing.
- Never stage files implicitly. Each staged file must be explicitly named.
- Commit message must not be empty. It must describe what changed.
- Never use `git add .` without explicit model confirmation.
- Never commit if conflicts exist in working tree.

**Invariants:**
- After commit, `git status` shows clean working tree (nothing to commit).
- Files staged = files in commit. No hidden additions.
- Diff reviewed before commit = diff in commit. No surprise changes.
- Commit hash is returned to model for reference.

**Result when successful:**
```
status: "success"
commit_hash: "abc123..."
files_committed: ["file1.go", "file2.go"]
message: "feature: add new endpoint"
clean_tree: true
```

**Result when failed:**
```
status: "failure"
reason: "conflicts_in_working_tree" | "empty_commit_message" | "nothing_to_commit"
state_restored: true  # working tree is exactly as before the attempt
```

**How a model uses this:**  
Model says "commit files X and Y with message Z" → Mimic classifies as git domain, atomic_commit process → Mimic returns the chain above (status→diff→add→commit) → model follows step by step. If status shows conflicts → process stops before diff, model is notified. If diff is approved → proceed to add and commit.

---

### safe_merge

**When to use:**  
The model wants to merge one branch into another without risk of merge commits or force-push.

**Goal:**  
Fast-forward merge only. No merge commits. No conflicts. Clean history.

**Chain (semantically):**

1. Fetch remote changes (fetch) — get latest refs from origin.
2. Preview merge (diff base vs head) — what would change if merged.
3. Merge with fast-forward only (merge --ff-only).

**Hard constraints:**
- Fast-forward only. If merge would create a merge commit → reject, notify model.
- No merge if conflicts exist in diff preview.
- Source branch must exist.

**Invariants:**
- After merge, target branch HEAD is direct descendant of source branch HEAD.
- No merge commit created.
- Working tree is clean after merge.

**Result when successful:**
```
status: "success"
merge_type: "fast_forward"
new_head: "def456..."
commits_merged: 3
```

**Result when failed:**
```
status: "failure"
reason: "not_fast_forward" | "conflicts_in_preview" | "source_branch_missing"
rollback: "checkout to previous HEAD"
```

**How a model uses this:**  
Model says "merge feature branch into main" → Mimic returns safe_merge process. If not fast-forward → model gets "not fast forward, here is diff, what do you want to do?" Model decides: rebase first, or abort. No automatic force-merge.

---

### feature_branch

**When to use:**  
Model wants to create a new branch for isolated work.

**Goal:**  
New branch exists, working tree is clean, model is on the new branch.

**Chain (semantically):**

1. Create branch (branch).
2. Switch to branch (checkout).

**Hard constraints:**
- Branch name must not already exist.
- Working tree must be clean before checkout (or changes must be stashed with model confirmation).

**Invariants:**
- Branch exists after creation.
- HEAD points to new branch.
- Previous branch is preserved.

**Result when successful:**
```
status: "success"
branch_name: "feat/new-feature"
previous_branch: "main"
```

**Result when failed:**
```
status: "failure"
reason: "branch_already_exists" | "dirty_working_tree"
```

**How a model uses this:**  
Model says "create branch for feature X" → Mimic returns process. If tree dirty → model gets " stash changes first?" Model decides. No silent data loss.

---

### hotfix

**When to use:**  
Critical fix needs to go straight to production/main branch.

**Goal:**  
Fix created on dedicated branch, committed, merged back into target (main/production), branch deleted after merge.

**Chain (semantically):**

1. Create hotfix branch from target (branch).
2. Apply fix and commit (commit).
3. Switch to target (checkout).
4. Merge hotfix branch (merge --ff-only).
5. Delete hotfix branch (branch -d).

**Hard constraints:**
- Hotfix branch name must be unique.
- Merge must be fast-forward or clean.
- Target branch must be specified explicitly.

**Invariants:**
- Fix is on dedicated branch before entering target.
- Target branch receives clean merge.
- Hotfix branch deleted after successful merge.
- 2-vote verification required before merge into production branch.

**Result when successful:**
```
status: "success"
hotfix_branch: "hotfix/critical-fix"
commits: ["fix: resolve security issue"]
target: "main"
merge_hash: "ghi789..."
```

**Result when failed:**
```
status: "failure"
reason: "merge_not_clean" | "target_not_specified"
rollback: "delete hotfix branch, restore target to previous HEAD"
```

**How a model uses this:**  
Model says "hotfix for bug X into main" → Mimic returns full process with 2-vote requirement on merge step. Model sees "this is production branch, merge requires verification" → model approves or aborts.

---

### ci_diff_check

**When to use:**  
Before any commit or merge, check that diff is clean — no whitespace errors, no conflict markers.

**Goal:**  
Diff contains no trailing whitespace, no tab/space mixing, no conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`), no missing newline at EOF.

**Chain (semantically):**

1. Get diff (diff).
2. Parse diff output for: trailing whitespace, tab/space mixing, conflict markers, missing newline.

**Hard constraints:**
- Any conflict marker = immediate failure, stop process.
- Any trailing whitespace on added lines = failure.

**Invariants:**
- If check passes → diff is clean for commit/merge.
- If check fails → exact locations of violations reported to model.

**Result when successful:**
```
status: "success"
violations: []
```

**Result when failed:**
```
status: "failure"
violations: [
  {file: "main.go", line: 42, type: "trailing_whitespace"},
  {file: "utils.go", line: 15, type: "conflict_marker"}
]
```

**How a model uses this:**  
Automatically invoked before atomic_commit and safe_merge. Model does not call this directly — it is a pre-condition check. If fails → model gets violations list with fix suggestions.

---

## Principles From Sources

### embryo

**Principles taken:**
- OpCode-based execution: every git operation is a token (OpPacket), not arbitrary command.
- BinaryRuntime: git operations chained deterministically, no ad-hoc shell commands.
- Session tracking: git state tracked across operations, no stale assumptions.

**What Mimic does with them:**
Every git process = chain of OpPackets with defined order and invariants. No "git commit -am" shortcuts. Each step explicit.

**What Mimic does NOT copy:**
Embryo's Go implementation details. Only the pattern: tokenization + deterministic chaining.

### bun (PR #30412)

**Principles taken:**
- Phase graph: git operations live inside CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND.
- Never-rules: {git reset, git checkout used for destructive purpose, git restore, git stash, git rebase} = hard deny.
- 2-vote verification: merge into production branch requires two independent verifications.
- Edit scope isolation: git operations on one repo do not interfere with operations on another.

**What Mimic does with them:**
Never-rules are enforced before any git process. 2-vote on critical operations. Resource bitmask prevents cross-repo contamination.

**What Mimic does NOT copy:**
Bun's crate-based scope isolation (Mimic uses domain/resource-based isolation, not crates).

### gastown

**Principles taken:**
- Zero False Confidence: never trust cached git state ("branch is clean") — always verify with actual `git status`.
- Rollback on failure: if any git operation fails, all changes from current process are rolled back.
- Atomic allocation: git branch names allocated with collision check (flock-like) to prevent race conditions.

**What Mimic does with them:**
Every git process starts with fresh status check. Failed process rolls back to pre-operation state. Branch creation checks for existing names.

### hermes-agent

**Principles taken:**
- Streaming health: git operations on remote repos have timeout (no infinite hangs).
- Error classification: git errors classified as retryable (network) vs permanent (conflict) vs auth.
- Tool guardrails: OP_FLAG_DANGEROUS on destructive git operations (force-push, reset) always blocked.

**What Mimic does with them:**
Git fetch/push have timeouts. Errors classified before retry decision. Dangerous flags require explicit model approval.

### anti-patterns (from bin-mesh and gonka history)

**Principles taken:**
- Context injection without structure stability (AP-05/AP-24): git operations must not mutate shared state (like buildMessages) without guaranteeing structure is stable.
- Flip-flop decisions (AP-16): never revert a git decision within minutes without evidence. Every revert requires justification.
- Stale cached state (AP-11): never trust that "branch is up to date" without explicit fetch.

**What Mimic does with them:**
Every git operation validates state from observable reality, not cache. Reverts are tracked as NEGATIVE artifacts with counter_pattern. Fetch is mandatory before merge decisions.

---

## Artifact Storage

How git processes become mesh slots:

| Field | Value |
|-------|-------|
| domain | "git" |
| layer | "process" |
| modality | "code" |
| pattern_name | "atomic_commit" / "safe_merge" / "feature_branch" / "hotfix" / "ci_diff_check" |
| pattern_code | semantic chain description |
| invariants | ["status_clean_after_commit", "diff_reviewed_before_commit", "ff_only_merge", "branch_unique"]
| survival_index | from git blame on source repos (embryo, binary-mesh) |
| z_density | computed from usage frequency + survival |
| polarity | POSITIVE (valid processes) or NEGATIVE/COUNTER (for anti-patterns) |
| extraction_hash | sha256 of extraction tool + parameters |

---

## Cross-Domain Conflicts

Git domain conflicts with:
- **build domain**: cannot compile during git checkout (working tree changes mid-compile).
- **io domain**: cannot write files during git operations on same path.
- **network domain**: fetch/push conflicts with other network operations on same remote.

Conflict rule: if git operation holds working tree lock → all write operations on same repo serialized.
