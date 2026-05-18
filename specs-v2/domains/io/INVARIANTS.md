# IO Domain — Invariants

Rules that MUST hold for every process in the IO domain.

---

## IINV-01: File Exists Before Read

**What it prevents:** Reading non-existent files, misleading errors.

**What it requires:** Before OP_IO_READ or OP_SYS_FILE_EXISTS returning true, the path must exist and be within workspace.

**Source of this rule:**
- AP-04 (unvalidated input before I/O): invalid data enters system.
- gastown ZFC: verify existence from observable reality.

**Consequence of violation:** Read returns `ERR_INVALID_ARG` with "file not found".

---

## IINV-02: Backup Before Overwrite

**What it prevents:** Data loss from file overwrites.

**What it requires:** Before OP_IO_WRITE on an existing file, a backup MUST be created. Backup stored at `.mimic/backups/<file>.<timestamp>.<hash>.gz`.

**Source of this rule:**
- AP-15 (partial rollback): incomplete cleanup leaves orphans.
- gastown: rollback capability requires backups.

**Consequence of violation:** Write returns `ERR_PERMISSION_DENY` with "backup required for overwrite".

---

## IINV-03: Sensitive Path Protection

**What it prevents:** Reading/writing credentials, tokens, keys.

**What it requires:** Paths matching sensitive patterns (.env, .ssh/, credentials*, *token*, *key*) are BLOCKED. Read returns "access_denied". Write returns "sensitive_path_blocked".

**Source of this rule:**
- AP-07 (hardcoded secrets): committed to source.
- caveman: sensitive path auto-exclude.

**Consequence of violation:** Operation REJECTED with specific sensitive path identified.

---

## IINV-04: Workspace Boundary Enforcement

**What it prevents:** Escape from designated workspace, filesystem traversal attacks.

**What it requires:** ALL paths must resolve within workspace root. `..` components resolved and checked against root. Symlinks followed and verified.

**Source of this rule:**
- Standard security practice.
- rustnet sandbox: Landlock prevents filesystem escape.

**Consequence of violation:** Operation REJECTED with "path outside workspace".

---

## IINV-05: Read-Back Verification for Writes

**What it prevents:** Silent write failures, disk corruption, partial writes.

**What it requires:** After OP_IO_WRITE, read back the file and verify SHA-256 hash matches expected.

**Source of this rule:**
- AP-29 (swallowed error): error silently ignored.
- embryo projectmap: index integrity requires verified writes.

**Consequence of violation:** Write returns failure with "write verification failed".

---

## IINV-06: Index Update After Write

**What it prevents:** Stale project map, incorrect symbol lookups.

**What it requires:** After OP_IO_WRITE on a source file, workspace index (SQLite FTS5) MUST be updated. Update is asynchronous (non-blocking) but queued.

**Source of this rule:**
- embryo projectmap: auto-index on WRITE opcode.
- AP-11 (stale cached state): decisions based on stale data.

**Consequence of violation:** Index out of sync. Symbol lookups may return stale results. Detected at next index query.

---

## IINV-07: Parent Directory Creation

**What it prevents:** "No such file or directory" errors on write.

**What it requires:** Before OP_IO_WRITE, parent directory must exist. If missing, OP_SYS_DIR_CREATE is auto-inserted (with model notification).

**Source of this rule:**
- Convenience. Reduces model friction.
- gastown: atomic allocation ensures path exists before operation.

**Consequence of violation:** Parent created automatically. Model receives "created directory X" notification.

---

## IINV-08: Delete Confirmation for Tracked Files

**What it prevents:** Accidental deletion of source code.

**What it requires:** Before OP_SYS_FILE_DELETE, if file is tracked by git, model must explicitly confirm. Temporary files (matching *.tmp, *~) require no confirmation.

**Source of this rule:**
- AP-15 (partial rollback): orphaned resources.
- Common IDE behavior.

**Consequence of violation:** Delete REJECTED with "file tracked by git, confirmation required".

---

## IINV-09: File Type Detection

**What it prevents:** Wrong handling of file content (e.g., treating binary as text).

**What it requires:** File type detected via magic bytes and content analysis, not just extension. Binary files flagged for appropriate handling.

**Source of this rule:**
- caveman: file type detection via magic bytes.
- Standard Unix `file` command.

**Consequence of violation:** Binary files processed as text may produce garbled output. Type mismatch detected and reported.

---

## IINV-10: Write Operation Atomicity

**What it prevents:** Torn writes, partial file updates.

**What it requires:** OP_IO_WRITE uses atomic replace: write to temp file, fsync, rename to target. Readers never see partial writes.

**Source of this rule:**
- Standard filesystem best practice.
- AP-26 (out-of-bound write): writing beyond buffer.

**Consequence of violation:** Non-atomic write detected (if filesystem doesn't support rename). Operation proceeds with warning.
