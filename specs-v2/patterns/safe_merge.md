# Pattern: safe_merge

Tokenized process for merging one branch into another with fast-forward guarantee.

---

## When to Use

Model wants to integrate changes from one branch into another. Any intent like "merge feature into main", "bring dev up to date", "integrate PR" maps here.

## Goal

Target branch receives all commits from source branch, with a linear history. No merge commits. No conflicts introduced by the merge itself. If conflicts exist, they are detected BEFORE merge and reported to model.

## Chain (Tokenized)

```c
OpPacket packets[] = {
    {OP_GIT_FETCH,      .args = {{"remote", "origin"}}},
    {OP_GIT_DIFF,       .args = {{"base", "main"}, {"head", "feature"}}},
    {OP_GIT_MERGE,      .args = {{"source", "feature"}, {"ff_only", "true"}}}
};
```

### Step-by-Step Semantics

**Step 1: Fetch**
- `OP_GIT_FETCH`: Retrieve latest refs from `origin`.
- If network error → classify as retryable/permanent (see error classification spec).
- Result: remote refs updated locally.

**Step 2: Diff Preview**
- `OP_GIT_DIFF --base=main --head=feature`: Show what would change.
- If diff contains conflict markers (`<<<<<<<`) → STOP. Return failure before merge.
- Model receives diff output for review.

**Step 3: Merge (Fast-Forward Only)**
- `OP_GIT_MERGE --source=feature --ff_only=true`.
- If not fast-forward → STOP. Return "not_fast_forward".
- If success → target branch HEAD now equals source branch HEAD.

## Hard Constraints

- `ff_only=true` is mandatory. Non-FF merge = REJECTED.
- If `git diff --check` finds trailing whitespace or conflict markers → STOP before merge.
- Source branch must be specified explicitly. No "merge whatever is current".

## Invariants

- After merge, `git log --oneline target` shows linear history.
- `git merge-base target source` == source HEAD (FF verification).
- Working tree is clean after merge.

## Result When Successful

```json
{
  "status": "success",
  "merge_type": "fast_forward",
  "target_branch": "main",
  "source_branch": "feature",
  "commits_merged": 5,
  "new_head": "a1b2c3d..."
}
```

## Result When Failed

```json
{
  "status": "failure",
  "reason": "not_fast_forward" | "conflicts_in_preview" | "fetch_failed",
  "suggestion": "Rebase feature onto main first, then retry merge."
}
```

## How a Model Uses This

Model says "merge feature into main" → Mimic returns 3-step process. If not FF → model gets "not fast forward, rebase first?". Model decides. No automatic force-merge.

## Energy Cost (Estimated)

- Fetch: tokens=5, latency=500ms → energy=2500
- Diff: tokens=1, latency=5ms → energy=5
- Merge: tokens=4, latency=50ms → energy=200
- Total: 2705 token-ms

## QAC Mapping

- QAC-2 (Invariant Coverage): FF guarantee, linear history.
- QAC-7 (Artifact Precision): Diff preview ensures no surprise changes.
- QAC-10 (Conflict Detection): Conflicts caught before merge.
