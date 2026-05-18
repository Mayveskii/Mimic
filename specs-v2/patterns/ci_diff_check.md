# Pattern: ci_diff_check

Tokenized pre-commit validation for diff cleanliness.

---

## When to Use

Automatically invoked before `atomic_commit` and `safe_merge`. Model does not call directly. Pre-condition check.

## Goal

Diff contains no trailing whitespace, no tab/space mixing, no conflict markers, no missing newline at EOF.

## Chain (Tokenized)

```c
OpPacket packets[] = {
    {OP_GIT_DIFF,       .args = {{"cached", "false"}, {"pathspec", ""}}},
    // Parsing is done by the diff executor, not a separate opcode
};
```

Actually, diff check is a parser on `OP_GIT_DIFF` output, not a separate opcode. The `OP_GIT_DIFF` executor performs these checks automatically and sets `result_code` accordingly.

### Step-by-Step Semantics

**Step 1: Get Diff**
- `OP_GIT_DIFF --cached=false`: Get unstaged diff.
- If `--cached=true`: Get staged diff.

**Step 2: Parse (Internal to Executor)**
- Scan diff output line by line.
- Detect patterns:
  - `<<<<<<<` / `=======` / `>>>>>>>` → conflict_marker violation.
  - Trailing whitespace on added lines → trailing_whitespace violation.
  - Tab character after space on added lines → tab_space_mix violation.
  - File ending without `\n` → missing_newline violation.

**Step 3: Report**
- If violations found → return failure with violation list.
- If clean → return success with empty violations.

## Hard Constraints

- Any conflict marker = immediate failure. No exceptions.
- Trailing whitespace on ADDED lines only (not removed lines) = failure.
- Check applied to EVERY file in diff.

## Invariants

- If check passes → diff is clean for commit/merge.
- If check fails → exact locations reported (file, line, type).

## Result When Successful

```json
{
  "status": "success",
  "violations": [],
  "files_checked": 3
}
```

## Result When Failed

```json
{
  "status": "failure",
  "violations": [
    {"file": "main.go", "line": 42, "type": "trailing_whitespace", "column": 80},
    {"file": "utils.go", "line": 15, "type": "conflict_marker", "column": 1}
  ],
  "auto_fix_available": true
}
```

## Auto-Fix (Optional)

If `auto_fix_available` and model approves:
- Trailing whitespace: `sed -i 's/[[:space:]]*$//' <file>`.
- Missing newline: `echo >> <file>`.
- Tab/space mix: convert tabs to 4 spaces (or project-configured indent).
- Conflict markers: NEVER auto-fix. Manual resolution required.

## Energy Cost (Estimated)

- Diff: tokens=1, latency=5ms → energy=5
- Parse: tokens=0.5, latency=1ms → energy=0.5
- Total: 5.5 token-ms

## QAC Mapping

- QAC-4 (State Validation): Diff validated before commit.
- QAC-7 (Artifact Precision): Exact violation locations reported.
- QAC-9 (Wholeness): Every file in diff checked.
