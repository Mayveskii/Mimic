# Process Domain — Sources

Where the process domain behavior comes from.

---

## rustnet

**Principles taken:**
- Sandbox: Landlock on Linux, Seatbelt on macOS, Job Objects on Windows.
- Process isolation: spawned processes cannot escape workspace.

**What Mimic does with them:**
Every OP_PROC_SPAWN applies sandbox before exec. Filesystem restrictions: workspace only. Network blocked unless test explicitly requests.

**What Mimic does NOT copy:**
- Rustnet's eBPF integration.
- Rustnet's sandbox policy DSL.
- Rustnet's specific Windows ACL configuration.

---

## hermes-agent

**Principles taken:**
- Error classification: process errors classified as retryable vs permanent.
- Streaming health: process output monitored for staleness.

**What Mimic does with them:**
Process exit codes captured and classified. Output monitored for hangs.

---

## bun (PR #30412)

**Principles taken:**
- Agent resource isolation: memory, CPU limits per process.
- Edit scope isolation: process operations scoped to workspace.

**What Mimic does with them:**
Resource limits applied via cgroups/rlimit. Process scope restricted.

---

## gastown

**Principles taken:**
- Rollback: if process fails, state restored.
- Graceful termination: SIGTERM before SIGKILL.

**What Mimic does with them:**
Process termination follows graceful-then-force sequence. Failed processes trigger rollback if atomic.

---

## embryo

**Principles taken:**
- BinaryRuntime: subprocesses managed by OpPacket chains.
- Tool loop: session tracks process state.

**What Mimic does with them:**
Process operations tokenized. Session tracks PIDs for cleanup.

---

## Standard Unix Practice

**Principles taken:**
- execve over system(): array arguments prevent injection.
- Exit code checking: always check $?, never ignore.
- Resource limits: ulimit before execution.
- Process groups: kill entire group, not single PID.

**What Mimic does with them:**
Array-style execution preferred. Exit codes captured. rlimits applied. Process groups managed.
