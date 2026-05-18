# Process Domain — Spawn, Wait, Signal, Kill

How a model spawns and manages subprocesses through Mimic.

---

## What This Domain Does

Process operations spawn child processes, wait for completion, send signals, and terminate processes. Used for build pipelines, test runners, and long-running tasks. Every spawned process is tracked, sandboxed, and subject to timeout.

---

## Processes

### proc_spawn

**When to use:**  
Model needs to run a command as separate process (build, test, server).

**Goal:**  
Spawn process with sandbox, timeout, resource limits.

**Chain (semantically):**

1. Validate command (same rules as OP_SYS_EXEC).
2. Check resource limits (memory, CPU, file descriptors).
3. Apply sandbox (Landlock/Seatbelt/Job Objects from rustnet).
4. Spawn: `OP_PROC_SPAWN(command, args, env, cwd, timeout)`.
5. Track PID in ExecContext.

**Hard constraints:**
- Command validated before spawn.
- Resource limits enforced.
- Sandbox active before execution.
- Timeout always set.

**Invariants:**
- PID tracked for lifecycle management.
- Resources monitored during execution.
- Sandbox prevents filesystem/network escape.

---

### proc_wait

**When to use:**  
Wait for spawned process to complete.

**Goal:**  
Collect exit code and output.

**Chain (semantically):**

1. Wait for PID with timeout.
2. Collect stdout, stderr, exit code.
3. Clean up process record.

**Invariants:**
- Never block forever (timeout enforced).
- Process record cleaned up after wait.

---

### proc_signal / proc_kill

**When to use:**  
Terminate or signal a running process.

**Goal:**  
Safe process termination.

**Chain (semantically):**

1. Validate PID belongs to current session (cannot kill other sessions' processes).
2. Send signal (SIGTERM → graceful, SIGKILL → force after timeout).
3. Verify process terminated.

**Hard constraints:**
- Can only signal own session's processes.
- SIGTERM first, SIGKILL only if SIGTERM fails after timeout.
- Never kill system processes.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "process" |
| layer | "process" |
| modality | "code" |
| invariants | ["command_validated", "sandboxed", "timeout_enforced", "own_session_only"] |

---

## Cross-Domain Conflicts

Process domain conflicts with:
- **system domain**: concurrent exec and spawn on same command = race.
- **build domain**: concurrent build processes must be isolated.
