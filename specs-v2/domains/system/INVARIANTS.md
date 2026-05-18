# System Domain — Invariants

Rules that MUST hold for every process in the system domain.

---

## SINV-01: Exec Command Validated

**What it prevents:** Shell injection, arbitrary command execution.

**What it requires:** OP_SYS_EXEC follows same validation as OP_PROC_SPAWN: no shell metacharacters, path within workspace or whitelisted, array-style execution preferred.

**Source of this rule:**
- Standard security practice.
- PINV-01 (command validated before spawn).

**Consequence of violation:** Exec REJECTED with `ERR_PERMISSION_DENY`.

---

## SINV-02: Environment Variables Isolated

**What it prevents:** Credential leakage, unexpected behavior from inherited env.

**What it requires:** OP_SYS_ENV_SET: name validated (no special chars), value scanned for secrets (no API keys, tokens). OP_SYS_ENV_GET: returns value from isolated session env, not system env.

**Source of this rule:**
- caveman: sensitive path protection extends to env.
- PINV-10 (env sanitized).

**Consequence of violation:** Set REJECTED if value contains credential pattern. Get returns redacted value for sensitive names.

---

## SINV-03: Path Within Workspace

**What it prevents:** Directory traversal, access to system files.

**What it requires:** ALL OP_SYS_* paths validated against workspace root. `..` resolved and checked. Symlinks followed and must resolve within workspace. Absolute paths outside workspace = REJECTED.

**Source of this rule:**
- IINV-04 (workspace boundary).
- rustnet: sandbox restricts filesystem.

**Consequence of violation:** Operation REJECTED with "path outside workspace".

---

## SINV-04: Destructive Operations Confirmed

**What it prevents:** Accidental deletion, data loss.

**What it requires:** OP_SYS_FILE_DELETE, OP_SYS_DIR_REMOVE (recursive), OP_SYS_FILE_MOVE (overwrite) require explicit model confirmation if target exists. Non-recursive dir removal of empty dir does not require confirmation.

**Source of this rule:**
- IINV-08 (delete confirmation for tracked files).
- gastown: rollback on destructive ops.

**Consequence of violation:** Operation REJECTED with "confirmation required for destructive operation".

---

## SINV-05: File Copy Integrity Verified

**What it prevents:** Corrupted copies, incomplete transfers.

**What it requires:** After OP_SYS_FILE_COPY, SHA-256 hash of source compared to destination. Mismatch → retry or failure.

**Source of this rule:**
- Standard data integrity practice.
- IINV-05 (read-back verification for writes).

**Consequence of violation:** Copy returns failure with "hash mismatch". Source and destination paths reported.

---

## SINV-06: Directory Creation Idempotent

**What it prevents:** Race conditions, duplicate creation errors.

**What it requires:** OP_SYS_DIR_CREATE succeeds if directory already exists (idempotent). Returns success with "already_exists" flag. Only fails on permission denied or non-directory path conflict.

**Source of this rule:**
- Standard mkdir -p behavior.
- gastown: atomic allocation with collision check.

**Consequence of violation:** No failure for existing directory. Model receives "created" or "already_exists".

---

## SINV-07: CHMOD Restricted

**What it prevents:** World-writable files, permission escalation.

**What it requires:** OP_SYS_CHMOD: world-writable (o+w) blocked unless explicitly allowed. Setuid/setgid blocked ALWAYS. Executable bits on non-executable files flagged.

**Source of this rule:**
- Standard security hardening.
- caveman: sensitive path protection.

**Consequence of violation:** CHMOD REJECTED with "dangerous permission mode".

---

## SINV-08: File Exists Check Cached Briefly

**What it prevents:** TOCTOU race between exists check and operation.

**What it requires:** OP_SYS_FILE_EXISTS result valid for maximum 1 second within chain. After 1s, re-check required. Operations that follow exists check should use atomic syscalls (open with O_CREAT|O_EXCL) where possible.

**Source of this rule:**
- Standard TOCTOU mitigation.
- gastown ZFC: observable reality over cache.

**Consequence of violation:** Warning: "exists check stale, re-validating". Operation proceeds with fresh check.
