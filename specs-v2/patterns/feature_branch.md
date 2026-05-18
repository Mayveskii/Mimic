# Pattern: feature_branch

Tokenized process for creating a new branch for isolated work.

---

## When to Use

Model wants to start work on a new feature, bugfix, or experiment. Intent: "create branch for X", "start work on Y".

## Goal

New branch exists at the desired starting point, working tree is clean, model is positioned on the new branch, previous branch is preserved.

## Chain (Tokenized)

```c
OpPacket packets[] = {
    {OP_GIT_STATUS,     .args = {}}},
    {OP_GIT_BRANCH,     .args = {{"name", "feat/new-feature"}, {"start_point", "main"}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", "feat/new-feature"}}}
};
```

### Step-by-Step Semantics

**Step 1: Status Check**
- `OP_GIT_STATUS`: Verify working tree state.
- If clean → proceed.
- If dirty → STOP. Ask model: "Working tree has uncommitted changes. Stash, commit, or abort?"

**Step 2: Branch Creation**
- `OP_GIT_BRANCH --name=feat/new-feature --start_point=main`.
- If branch name exists → STOP. Return "branch_already_exists".
- Branch created but HEAD not switched yet.

**Step 3: Checkout**
- `OP_GIT_CHECKOUT --target=feat/new-feature`.
- Switch to new branch.
- Verify: `git branch --show-current` == "feat/new-feature".

## Hard Constraints

- Branch name must not exist. Creates collision with existing branches.
- Working tree must be clean OR model explicitly confirms stash/commit of changes.
- `start_point` defaults to current branch if not specified.

## Invariants

- `git branch --list feat/new-feature` shows the branch.
- `git branch --show-current` == new branch name.
- Previous branch HEAD unchanged.
- Working tree identical before and after (clean state preserved).

## Result When Successful

```json
{
  "status": "success",
  "branch_name": "feat/new-feature",
  "start_point": "main",
  "previous_branch": "dev",
  "head_commit": "abc123..."
}
```

## Result When Failed

```json
{
  "status": "failure",
  "reason": "branch_already_exists" | "dirty_working_tree" | "invalid_branch_name",
  "suggestion": "Branch exists. Use git_branch_delete first, or choose different name."
}
```

## How a Model Uses This

Model says "create branch for auth feature" → Mimic checks status → if clean → creates branch → switches to it. If dirty → model decides what to do with changes. Never silently discards work.

## Energy Cost (Estimated)

- Status: tokens=1, latency=5ms → energy=5
- Branch: tokens=2, latency=10ms → energy=20
- Checkout: tokens=4, latency=20ms → energy=80
- Total: 105 token-ms

## QAC Mapping

- QAC-2: Branch creation atomic (all or nothing).
- QAC-5: No silent data loss (dirty tree check).
- QAC-7: Precise result (exact branch name confirmed).
