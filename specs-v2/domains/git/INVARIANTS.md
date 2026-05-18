# Git Domain — Invariants

Rules that MUST hold for every process in the git domain. Violation of any invariant = process failure, regardless of individual step success.

---

## GINV-01: No Commit Without Prior Diff Review

**What it prevents:** Surprise changes entering the repository.

**What it requires:** Before OP_GIT_COMMIT, OP_GIT_DIFF or OP_GIT_STATUS must have executed successfully in the same chain, and its output must have been processed (not just executed).

**Source of this rule:**
- AP-05 (context injection without stability): buildMessages corruption from unreviewed changes.
- bun PR #30412: Never-rules enforce review before commit.
- gastown ZFC: observable reality over cached assumptions.

**Consequence of violation:** Commit REJECTED at validation. Model receives "diff not reviewed" error.

---

## GINV-02: Working Tree Clean After Commit

**What it prevents:** Partial commits, lingering changes, false confidence.

**What it requires:** After OP_GIT_COMMIT succeeds, OP_GIT_STATUS must show "nothing to commit, working tree clean". If not, the commit was partial.

**Source of this rule:**
- AP-15 (partial rollback): incomplete cleanup leads to orphaned state.
- embryo BinaryRuntime: session tracking ensures clean state.

**Consequence of violation:** VERIFY phase marks commit as PARTIAL. Model receives warning: "commit created but N files remain uncommitted".

---

## GINV-03: Fast-Forward Only for Merge

**What it prevents:** Merge commits, history divergence, revert complexity.

**What it requires:** OP_GIT_MERGE with `ff_only=false` is REJECTED unless explicitly requested by model AND 2-vote verified.

**Source of this rule:**
- bun PR #30412: phase graph enforces clean history.
- gonka history: revert of merge commits is harder than revert of single commits.

**Consequence of violation:** Merge REJECTED at validation. Model receives "non-FF merge requires 2-vote".

---

## GINV-04: No Push Without Explicit Confirmation

**What it prevents:** Accidental publication of incomplete or incorrect code.

**What it requires:** OP_GIT_PUSH has `OP_FLAG_DANGEROUS`. Validation requires explicit model allow. Auto-classifier cannot auto-allow push.

**Source of this rule:**
- AP-13 (override on deny): dangerous ops executed despite safety rules.
- hermes-agent: credential pool rotation prevents hardcoded keys but confirms before use.

**Consequence of violation:** Push REJECTED at validation with `ERR_PERMISSION_DENY`.

---

## GINV-05: No Checkout With Uncommitted Changes

**What it prevents:** Silent data loss from overwritten working tree changes.

**What it requires:** Before OP_GIT_CHECKOUT, OP_GIT_STATUS must show clean working tree, OR model must explicitly confirm stash/reset of changes.

**Source of this rule:**
- AP-20 (regenerate seed without validation): state changes without checking impact.
- gastown rollback: any operation that could lose state must validate first.

**Consequence of violation:** Checkout REJECTED. Model receives "uncommitted changes detected" with options: stash, commit, abort.

---

## GINV-06: Branch Name Unique Within Repo

**What it prevents:** Branch collisions, ambiguous refs.

**What it requires:** OP_GIT_BRANCH validates that name does not exist locally or remotely before creation.

**Source of this rule:**
- gastown atomic allocation: collision check prevents race conditions (flock-like).
- embryo tool registry: unique identifiers prevent duplicate tool registrations.

**Consequence of violation:** Branch creation REJECTED with "branch_already_exists".

---

## GINV-07: Fetch Before Remote-Based Decisions

**What it prevents:** Decisions based on stale remote state.

**What it requires:** Before OP_GIT_MERGE or OP_GIT_PUSH that references remote branches, OP_GIT_FETCH must have executed in the same chain or within the last 60 seconds.

**Source of this rule:**
- AP-11 (stale cached state): decisions based on stale data → wrong actions.
- gastown ZFC: always verify from observable reality.

**Consequence of violation:** Operation REJECTED with "remote state stale, fetch first".

---

## GINV-08: Rebase Never Automatic

**What it prevents:** History rewriting without explicit intent.

**What it requires:** OP_GIT_REBASE always requires explicit model confirmation. Auto-classifier flags rebase as CRITICAL (safety level 0). No chain may contain rebase without manual allow.

**Source of this rule:**
- bun PR #30412: rebase in never-rules set.
- binary-mesh history: revert of rebase is destructive and complex.

**Consequence of violation:** Rebase REJECTED with `ERR_PERMISSION_DENY`. Denial logged.

---

## GINV-09: Commit Message Non-Empty and Descriptive

**What it prevents:** Uninformative history, useless `git log`.

**What it requires:** OP_GIT_COMMIT `message` argument must be non-empty and at least 10 characters. Empty or "temp" messages are REJECTED.

**Source of this rule:**
- embryo session tracking: context accumulation requires meaningful commit messages for traceability.
- Quality heuristic: good commit messages correlate with pattern survival.

**Consequence of violation:** Commit REJECTED with "message too short or empty".

---

## GINV-10: Every Git Op Records Session Context

**What it prevents:** Unattributed changes, audit gaps.

**What it requires:** Every git operation appends to session context: timestamp, opcode, repo path, result. This is automatic via OP_SESS_CONTEXT_APPEND.

**Source of this rule:**
- bun PR #30412: session enrichment on every tool use.
- AP-21 (missing node ID): batches without attribution → incorrect tracking.

**Consequence of violation:** Missing context is detected at RESPOND phase. Warning logged. Session may be incomplete for future queries.
