# Security Domain — Invariants

Rules that MUST hold for every process in the security domain.

---

## SECINV-01: Never-Rules Absolute

**What it prevents:** Dangerous operations executed despite prohibition.

**What it requires:** Never-rules set is hardcoded and CANNOT be overridden by model, auto-classifier, or any permission level. Contains: git reset, git checkout for destructive purpose, git restore, git stash, git rebase, git push --force, rm -rf /, eval(), system() with unsanitized input.

**Source of this rule:**
- bun PR #30412: never-rules hardcoded deny set.
- META_INVARIANT.md: no side effect without prior validation.

**Consequence of violation:** Operation REJECTED with `ERR_PERMISSION_DENY`. Denial logged. No appeal.

---

## SECINV-02: DIFC Labels Enforced

**What it prevents:** Information leakage between sensitivity levels.

**What it requires:** DIFC (Dynamic Information Flow Control) labels assigned to all data:
- PUBLIC: mesh slots, public API responses.
- INTERNAL: session context, operation logs.
- CONFIDENTIAL: credential pool, API keys.
- RESTRICTED: user PII, sensitive workspace data.
Data flow: low → high is allowed. high → low is DENIED (unless explicitly declassified with audit trail).

**Source of this rule:**
- Standard information flow control.
- caveman: sensitive path protection.

**Consequence of violation:** Data flow REJECTED with `ERR_PERMISSION_DENY`. Attempt logged.

---

## SECINV-03: Credential Pool Isolation

**What it prevents:** Key leakage, unauthorized usage.

**What it requires:** Credential pool:
1. Encrypted at rest (AES-256-GCM).
2. Keys referenced by ID, never inline.
3. Access logged: who (session), when, what operation.
4. Rotation on suspected compromise or schedule.
5. Memory scrub after use (explicit_bzero).

**Source of this rule:**
- hermes-agent: credential pool with rotation.
- AP-07 (hardcoded secrets).

**Consequence of violation:** Inline key → REJECTED. Unauthorized access → REJECTED + alert.

---

## SECINV-04: Path Sanitization

**What it prevents:** Directory traversal, escape from workspace.

**What it requires:** ALL paths:
1. Resolved to absolute (realpath).
2. Checked against workspace root prefix.
3. Symlinks followed and validated.
4. `..` components resolved and checked.
5. Null bytes rejected.

**Source of this rule:**
- Standard filesystem security.
- IINV-04 (workspace boundary).

**Consequence of violation:** Path REJECTED with `ERR_PERMISSION_DENY`.

---

## SECINV-05: Input Validation Before Processing

**What it prevents:** Injection attacks, state corruption.

**What it requires:** ALL external input validated:
1. Size bounded.
2. Type checked (no type confusion).
3. Format validated (JSON schema, regex).
4. Encoding normalized (UTF-8).
5. No control characters (except newline, tab).

**Source of this rule:**
- AP-04 (unvalidated input before I/O).
- Standard input validation.

**Consequence of violation:** Input REJECTED with `ERR_INVALID_ARG`.

---

## SECINV-06: Audit Log Immutable

**What it prevents:** Log tampering, audit evasion.

**What it requires:** Security audit log:
1. Append-only.
2. Cryptographically signed (hash chain: each entry contains hash of previous).
3. No deletion by any operation.
4. Retention: 90 days hot, 1 year cold.

**Source of this rule:**
- Standard audit requirements.
- bun PR #30412: session logging.

**Consequence of violation:** Log tampering detected by hash chain break. Alert raised. Session flagged.

---

## SECINV-07: Sandboxed Execution

**What it prevents:** Escape from workspace, system compromise.

**What it requires:** ALL spawned processes (build, test, arbitrary) run in sandbox. Network blocked unless explicitly enabled. Filesystem restricted to workspace. Resource limits applied.

**Source of this rule:**
- rustnet: Landlock/Seatbelt sandbox.
- PINV-02 (sandbox active before execution).

**Consequence of violation:** Sandbox failure → process MUST NOT run. `ERR_PERMISSION_DENY`.

---

## SECINV-08: Side-Channel Mitigation

**What it prevents:** Timing attacks, cache-based leakage.

**What it requires:** Cryptographic operations (hash, encrypt) use constant-time implementations where available. No branching on secret data.

**Source of this rule:**
- Cryptographic best practice.
- Standard side-channel defense.

**Consequence of violation:** Non-constant-time crypto → warning. Operation proceeds but flagged for review.
