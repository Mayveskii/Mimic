# Security Domain — Sources

Where the security domain behavior comes from.

---

## bun (PR #30412)

**Principles taken:**
- Never-rules: hardcoded deny set.
- Permission pipeline: classify → budget → allow.
- 2-vote verification for critical ops.

**What Mimic does with them:**
Never-rules absolute. Permission pipeline gates all operations. 2-vote on dangerous ops.

---

## caveman

**Principles taken:**
- Sensitive path protection: .env, credentials.
- File type detection: magic bytes prevent type confusion.

**What Mimic does with them:**
Paths scanned for sensitivity. File types validated. No credential leakage.

---

## rustnet

**Principles taken:**
- Sandbox: Landlock/Seatbelt/Job Objects.
- Process isolation: no escape from workspace.

**What Mimic does with them:**
All spawned processes sandboxed. Filesystem, network, resource restrictions applied.

---

## hermes-agent

**Principles taken:**
- Credential pool: multi-key rotation.
- Error classification: distinguish retryable vs permanent vs auth.

**What Mimic does with them:**
Credentials managed via pool. Access logged. Keys rotated. Auth errors not retried blindly.

---

## graphify

**Principles taken:**
- SSRF protection: private-IP block, metadata block, DNS rebinding guard.

**What Mimic does with them:**
Network requests validated. Private IPs blocked. DNS resolved before connect.

---

## Standard Security

**Principles taken:**
- Input validation.
- Path sanitization.
- Audit logging.
- Cryptographic best practices.
- DIFC / information flow control.

**What Mimic does with them:**
Standard security stack applied throughout.
