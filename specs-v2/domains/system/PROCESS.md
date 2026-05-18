# System Domain — Exec, Env, Dirs, Files

How a model interacts with the operating system through Mimic.

---

## What This Domain Does

System operations provide controlled access to OS-level functions: shell execution, environment variables, directory operations, file copy/move/delete. Every operation is validated, sandboxed, and tracked. The model cannot execute arbitrary shell commands — it uses tokenized system operations with explicit safety checks.

---

## Processes

### sys_exec

**When to use:**  
Model needs to run a shell command. Used sparingly — prefer domain-specific operations (git, build) when available.

**Goal:**  
Execute command with timeout, capture output, validate safety.

**Chain (semantically):**

1. **Validate command safety.**
   - Check against never-rules (no rm -rf /, no curl | sh).
   - Check against command whitelist if configured.
   - # UNCERTAIN: exact whitelist needs calibration per deployment

2. **Check conflicts.**
   - `OP_SYS_EXEC` × `OP_SYS_EXEC` = conflict (race condition).
   - Serialize if another exec is in progress.

3. **Execute with timeout.**
   - `OP_SYS_EXEC(command, cwd, timeout, env)`.
   - Default timeout: 300s.

4. **Capture and return output.**
   - stdout, stderr, exit code.
   - Output truncated at 50KB, overflow to temp file.

**Hard constraints:**
- Never execute without safety validation.
- Never execute without timeout.
- Concurrent exec operations serialized (conflict_matrix).

**Invariants:**
- Command validated before execution.
- Output captured and returned.
- Timeout enforced.

**Result:**
```
status: "success"
command: "go version"
stdout: "go version go1.23 linux/amd64"
stderr: ""
exit_code: 0
duration_ms: 42
```

---

### env_get / env_set

**When to use:**  
Read or set environment variables.

**Goal:**  
Safe environment variable access.

**Chain (semantically):**

- **env_get:**
  1. Validate variable name (alphanumeric + underscore).
  2. Read value.
  3. Return value or "not_set".

- **env_set:**
  1. Validate variable name.
  2. Validate value (no shell metacharacters unless explicitly allowed).
  3. Set value.
  4. Log change.

**Hard constraints:**
- Never set PATH, LD_PRELOAD, or other security-sensitive variables without explicit allow.
- Never set secrets in env (use credential pool instead).

---

### dir_create / dir_remove

**When to use:**  
Create or remove directories.

**Goal:**  
Directory operations with validation.

**Chain (semantically):**

- **dir_create:**
  1. Validate parent exists.
  2. Validate name (no .., no symlinks to outside workspace).
  3. Create with permissions.

- **dir_remove:**
  1. Validate directory exists.
  2. Validate directory is empty (or force flag with explicit confirmation).
  3. Validate not sensitive path.
  4. Remove.

**Hard constraints:**
- Never create directory outside workspace.
- Never remove non-empty directory without explicit confirmation.
- Never remove .git directory.

---

### file_copy / file_move / file_delete

**When to use:**  
File manipulation operations.

**Goal:**  
Copy, move, or delete files safely.

**Chain (semantically):**

- **file_copy:**
  1. Validate source exists.
  2. Validate destination is within workspace.
  3. Validate not overwriting sensitive file.
  4. Copy with verification (hash match).

- **file_move:**
  1. Validate source exists.
  2. Validate destination parent exists.
  3. Validate not overwriting without flag.
  4. Move.

- **file_delete:**
  1. Validate file exists.
  2. Validate not sensitive path.
  3. Validate not tracked by git (or explicit confirmation).
  4. Delete.

**Hard constraints:**
- Never copy/move/delete outside workspace.
- Never overwrite without explicit flag.
- Never delete tracked file without confirmation.
- Never delete sensitive files (.env, credentials).

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "system" |
| layer | "process" |
| modality | "code" |
| invariants | ["command_validated", "timeout_enforced", "within_workspace", "sensitive_blocked"] |

---

## Cross-Domain Conflicts

System domain conflicts with:
- **build domain**: sys_exec during compilation = race.
- **git domain**: sys_exec modifying git repo = corruption risk.
- **io domain**: concurrent file operations on same path = race.
