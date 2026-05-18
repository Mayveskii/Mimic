# Process Domain — Invariants

Rules that MUST hold for every process in the process domain.

---

## PINV-01: Command Validated Before Spawn

**What it prevents:** Arbitrary code execution, injection attacks.

**What it requires:** Before OP_PROC_SPAWN, command string validated:
1. No shell metacharacters (`;`, `|`, `&&`, `||`, `` ` ``, `$()`) unless explicitly allowed by model.
2. Command path exists and is within workspace (or system PATH for whitelisted tools).
3. Arguments are individually validated, not concatenated into shell string.
4. Prefer array-style execution (execve) over shell invocation.

**Source of this rule:**
- Standard security practice.
- AP-04 (unvalidated input before I/O).

**Consequence of violation:** Spawn REJECTED with `ERR_PERMISSION_DENY` and "command contains disallowed characters".

---

## PINV-02: Sandbox Active Before Execution

**What it prevents:** Filesystem escape, network access from untrusted code, resource exhaustion.

**What it requires:** OP_PROC_SPAWN MUST apply sandbox before exec:
- Linux: Landlock (filesystem restrictions).
- macOS: Seatbelt (sandbox profile).
- Windows: Job Objects + ACL restrictions.
Sandbox rules: workspace root = read+write, all other paths = deny (unless explicitly allowed). Network = deny (unless test explicitly requests).

**Source of this rule:**
- rustnet: Landlock/Seatbelt/Job Objects sandbox.
- Standard container security.

**Consequence of violation:** Spawn REJECTED with `ERR_PERMISSION_DENY`. Sandbox failure is FATAL — process must not run unsandboxed.

---

## PINV-03: Timeout Always Set

**What it prevents:** Runaway processes, zombie children, budget exhaustion.

**What it requires:** OP_PROC_SPAWN timeout_ms > 0. Default: 300000ms (5min). Max: 3600000ms (1hour). Zero timeout = REJECTED.

**Source of this rule:**
- AP-14 (infinite wait).
- AP-30 (missing cancellation boundary).

**Consequence of violation:** Validation REJECTS. If timeout fires → SIGKILL sent, process terminated, `ERR_TIMEOUT` returned.

---

## PINV-04: Resource Limits Enforced

**What it prevents:** Fork bombs, memory exhaustion, CPU monopolization.

**What it requires:** OP_PROC_SPAWN with limits:
- Max memory: 4GB (configurable).
- Max CPU time: 300s (configurable).
- Max open files: 1024.
- Max subprocesses: 16.
- cgroup / rlimit applied before exec.

**Source of this rule:**
- Standard container/resource management.
- bun PR #30412: agent resource isolation.

**Consequence of violation:** Process killed by OOM killer or rlimit. `ERR_OOM` or `ERR_TIMEOUT` returned.

---

## PINV-05: PID Tracked for Lifecycle

**What it prevents:** Orphaned processes, resource leaks.

**What it requires:** Every spawned PID tracked in ExecContext. OP_PROC_WAIT or OP_PROC_KILL must reference tracked PID. Auto-cleanup on chain completion: SIGKILL any untracked running children.

**Source of this rule:**
- EXEC_CONTEXT_SPEC.md: process tracking.
- AP-15 (partial rollback): orphaned resources.

**Consequence of violation:** Untracked PID referenced → `ERR_INVALID_ARG`. Orphaned processes detected at chain cleanup, logged, SIGKILL sent.

---

## PINV-06: SIGTERM Before SIGKILL

**What it prevents:** Data loss from forceful termination.

**What it requires:** OP_PROC_KILL sequence: send SIGTERM, wait up to 5s, if still alive → SIGKILL. Direct SIGKILL only allowed for processes that ignore SIGTERM or are known unresponsive.

**Source of this rule:**
- Standard Unix process management.
- gastown: rollback requires graceful cleanup.

**Consequence of violation:** Direct SIGKILL → warning logged. Data loss risk noted in session log.

---

## PINV-07: Own Session Only

**What it prevents:** Cross-session process manipulation, security boundary violation.

**What it requires:** OP_PROC_KILL and OP_PROC_SIGNAL can only target PIDs spawned in the SAME session. Cannot kill system processes (PID < 1000 on Linux) or other sessions' processes.

**Source of this rule:**
- Security isolation principle.
- bun PR #30412: edit scope isolation.

**Consequence of violation:** Kill REJECTED with `ERR_PERMISSION_DENY`. Attempt logged as potential attack.

---

## PINV-08: Exit Code Captured

**What it prevents:** Ignored failures, false confidence from missing error checks.

**What it requires:** OP_PROC_WAIT MUST capture and return exit code. Non-zero exit code = failure (unless model explicitly defines success codes). Exit code included in result.

**Source of this rule:**
- Standard shell scripting practice.
- AP-29 (swallowed error).

**Consequence of violation:** Missing exit code → result marked as PARTIAL. Model receives warning: "process exited but exit code not captured".

---

## PINV-09: Output Size Bounded

**What it prevents:** Memory exhaustion from large stdout/stderr.

**What it requires:** stdout and stderr captured with size limit: 10MB each. Exceeded → truncate with `truncated: true`. Large outputs redirected to temp files.

**Source of this rule:**
- Resource management best practice.
- graphify: streaming with size caps.

**Consequence of violation:** Output truncated. Model receives `truncated: true` with actual size.

---

## PINV-10: Environment Sanitized

**What it prevents:** Credential leakage via env, unexpected behavior from inherited env.

**What it requires:** OP_PROC_SPAWN environment:
1. PATH sanitized (only workspace + whitelisted system dirs).
2. No MIMIC_*, AGENT_*, API_KEY_* variables passed (unless explicitly requested).
3. HOME set to workspace tmp dir (not real home).
4. TMPDIR set to workspace tmp dir.

**Source of this rule:**
- Security best practice.
- caveman: sensitive path protection extends to env.

**Consequence of violation:** Environment scrubbed before spawn. Sensitive vars detected → warning logged, values redacted.
