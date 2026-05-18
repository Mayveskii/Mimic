# IO Domain — File Operations

How a model reads, writes, and manipulates files through Mimic.

---

## What This Domain Does

File operations are the most common need for any model working with code. The IO domain provides tokenized processes for file read, write, open, close, seek. Every operation is validated before execution: existence checks, path validation, conflict detection with other operations.

The model does not use shell commands (`cat`, `echo`, `mv`). It uses tokenized file operations with explicit validation.

---

## Processes

### read_file

**When to use:**  
Model wants to read contents of a file. "Show me main.go" or "Read the config file."

**Goal:**  
Return file contents with integrity verification and metadata.

**Chain (semantically):**

1. Verify file exists: `OP_SYS_FILE_EXISTS(path)`.
2. If exists → read: `OP_IO_READ(path, offset=0, limit=<max>)`.
3. Compute hash: `OP_HASH_SHA256(content)`.
4. Return content + hash + metadata.

**Hard constraints:**
- Never read file that does not exist. Return "file_not_found" with suggestions.
- Never read beyond file bounds.
- Sensitive paths (.env, credentials, keys) → return "access_denied" + log.

**Invariants:**
- File exists before read.
- Content hash returned for integrity verification.
- Read-only operation, no side effects.

**Result when successful:**
```
status: "success"
path: "/path/to/file.go"
content: "package main\n..."
hash: "sha256..."
size: 4096
encoding: "utf-8"
```

**Result when failed:**
```
status: "failure"
reason: "file_not_found" | "permission_denied" | "path_outside_workspace"
suggestion: "Did you mean /path/to/other_file.go?"
```

**How a model uses this:**  
Model says "read main.go" → Mimic checks exists → reads → returns content + hash. If file not found → Model gets suggestions for similar filenames. No silent failures.

---

### write_file

**When to use:**  
Model wants to create or overwrite a file. "Write this content to file.go" or "Update the config."

**Goal:**  
Write content to file with atomic operation, backup of previous version, and integrity verification.

**Chain (semantically):**

1. Verify parent directory exists: `OP_SYS_FILE_EXISTS(parent_dir)`.
2. If parent missing → create: `OP_SYS_DIR_CREATE(parent_dir)`.
3. If file exists → create backup: compress previous version, store with hash.
4. Write content: `OP_IO_WRITE(path, content, create=true)`.
5. Verify: read back, hash matches.
6. Trigger workspace re-index (tree + symbols updated).

**Hard constraints:**
- Never overwrite without backup of previous version.
- Never write to sensitive path (.env, credentials, keys, tokens).
- Parent directory must exist or be created.
- Write within workspace boundaries only.

**Invariants:**
- File content matches what was written (verified by read-back hash).
- Backup of previous version exists and is recoverable.
- Index is stale after write → re-index triggered.

**Result when successful:**
```
status: "success"
path: "/path/to/file.go"
bytes_written: 2048
previous_version_backed_up: true
backup_path: ".mimic/backups/file.go.2025-05-17.sha256.gz"
```

**Result when failed:**
```
status: "failure"
reason: "permission_denied" | "disk_full" | "path_outside_workspace"
rollback: "previous_version_restored"
```

**How a model uses this:**  
Model says "write this code to handler.go" → Mimic checks parent dir → creates if needed → writes → verifies → triggers re-index. If file existed → Model gets "previous version backed up at...". Model never loses work.

---

### delete_file

**When to use:**  
Model wants to delete a file. "Delete old_test.go" or "Remove temporary files."

**Goal:**  
Remove file with confirmation for non-temporary files, backup if specified.

**Chain (semantically):**

1. Verify file exists.
2. Check if file is tracked by git.
3. If tracked and not explicitly marked as temporary → require model confirmation.
4. If confirmation given → delete: `OP_SYS_FILE_DELETE(path)`.
5. Log deletion with reason.

**Hard constraints:**
- Never delete tracked file without explicit confirmation.
- Never delete outside workspace.
- Deletion logged with model-provided reason.

**Invariants:**
- File no longer exists after deletion.
- Git status updated (file marked as deleted).
- Index updated.

---

## Principles From Sources

### embryo (pkg/projectmap/)

**Principles taken:**
- SQLite-based project navigation: index files, symbols, imports.
- FTS5 full-text search for symbol lookup.
- Updated after every WRITE opcode.

**What Mimic does with them:**
Every file write triggers index update. Symbol lookup via FTS5. Project map stays current.

### caveman

**Principles taken:**
- Sensitive path protection: .env, credentials, keys → auto-exclude.
- File type detection: magic bytes + content, not just extension.

**What Mimic does with them:**
Sensitive paths blocked from read/write. File types detected for appropriate handling.

### anti-patterns (AP-04, AP-05)

**Principles taken:**
- Unvalidated input before I/O: never write without validating path and content.
- Context injection without structure stability: file writes must not corrupt shared structures.

**What Mimic does with them:**
Every write validated: path within workspace, not sensitive, parent exists. Content validated for syntax if known file type.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "io" |
| layer | "process" |
| modality | "code" |
| pattern_name | "read_file" / "write_file" / "delete_file" |
| invariants | ["file_exists_before_read", "backup_before_overwrite", "sensitive_path_blocked"] |

---

## Cross-Domain Conflicts

IO domain conflicts with:
- **git domain**: cannot read/write files during git operations on same path.
- **build domain**: cannot write source files during compilation.
- **io domain self**: WRITE × READ without SYNC = conflict (stale read).
