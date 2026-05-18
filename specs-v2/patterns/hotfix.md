# Pattern: hotfix

Tokenized process for emergency fix straight to production/main.

---

## When to Use

Critical bug needs immediate fix. No time for feature branch workflow. Intent: "hotfix for X", "critical fix now".

## Goal

Fix created on dedicated branch, committed, merged into target (main/production), branch cleaned up. Every step verified. 2-vote on merge into production.

## Chain (Tokenized)

```c
OpPacket packets[] = {
    {OP_GIT_STATUS,     .args = {}}},
    {OP_GIT_BRANCH,     .args = {{"name", "hotfix/critical-fix"}, {"start_point", "main"}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", "hotfix/critical-fix"}}},
    // Model applies fix here (via IO write ops)
    {OP_GIT_ADD,        .args = {{"paths", "file1.go,file2.go"}}},
    {OP_GIT_COMMIT,     .args = {{"message", "hotfix: resolve critical issue"}}},
    {OP_GIT_CHECKOUT,   .args = {{"target", "main"}}},
    {OP_GIT_MERGE,      .args = {{"source", "hotfix/critical-fix"}, {"ff_only", "true"}}},
    {OP_GIT_BRANCH,     .args = {{"name", "hotfix/critical-fix"}, {"delete", "true"}}}
};
```

### Step-by-Step Semantics

**Step 1-3: Create and Switch to Hotfix Branch**
- Same as feature_branch pattern.
- Branch created from target (main/production).

**Step 4: Model Applies Fix**
- Model writes fix files via IO domain.
- Files are in working tree of hotfix branch.

**Step 5: Stage**
- `OP_GIT_ADD`: Stage fix files.

**Step 6: Commit**
- `OP_GIT_COMMIT`: Commit with descriptive message.
- Commit hash recorded.

**Step 7: Switch to Target**
- `OP_GIT_CHECKOUT --target=main`.

**Step 8: Merge with 2-Vote**
- `OP_GIT_MERGE --source=hotfix/critical-fix --ff_only=true`.
- **CRITICAL**: Before merge, 2-vote verification required.
  - Verifier A: Diff correctness (fix addresses the stated issue).
  - Verifier B: Safety check (no unintended changes in diff).
  - Both pass → merge proceeds.
  - Either fails → merge aborted, hotfix branch preserved for revision.

**Step 9: Delete Hotfix Branch**
- `OP_GIT_BRANCH --delete=true`: Local cleanup.

## Hard Constraints

- 2-vote verification MANDATORY for merge into production branch.
- Fast-forward only (no merge commits for hotfixes).
- Target branch MUST be specified explicitly.
- Hotfix branch deleted after successful merge.

## Invariants

- Fix is on dedicated branch before entering target.
- Target branch receives clean FF merge.
- Working tree clean after every checkout.
- 2-vote verification recorded in session log.

## Result When Successful

```json
{
  "status": "success",
  "hotfix_branch": "hotfix/critical-fix",
  "commits": ["hotfix: resolve critical issue"],
  "target": "main",
  "merge_hash": "abc123...",
  "verification": {
    "verifier_a": "pass",
    "verifier_b": "pass"
  }
}
```

## Result When Failed

```json
{
  "status": "failure",
  "reason": "merge_not_clean" | "verification_rejected" | "target_not_specified",
  "rollback": "hotfix branch preserved at origin, main unchanged",
  "verification": {
    "verifier_a": "fail",
    "verifier_b": "pass",
    "failure_reason": "Diff includes unrelated changes in config.go"
  }
}
```

## How a Model Uses This

Model says "hotfix for authentication bypass" → Mimic creates hotfix branch → model writes fix → commit → Mimic requires 2-vote on merge → model reviews verifiers' findings → approves or aborts. No unverified code reaches production.

## Energy Cost (Estimated)

- Status + branch + checkout + add + commit + checkout + merge + delete
- Total tokens: ~20
- Total latency: ~200ms
- Total energy: ~4000 token-ms
- 2-vote overhead: +2x merge energy = +8000 token-ms (verification is expensive but mandatory)

## QAC Mapping

- QAC-1 (2-Vote): Mandatory on production merge.
- QAC-2 (Invariants): FF merge, branch cleanup.
- QAC-10 (Conflict Detection): Diff checked before merge.
- QAC-12 (Never-Rules): No direct commit to main.
