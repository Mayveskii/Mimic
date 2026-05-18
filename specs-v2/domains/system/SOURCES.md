# System Domain — Sources

Where the system domain behavior comes from.

---

## rustnet

**Principles taken:**
- Sandbox: Landlock/Seatbelt restricts filesystem access.
- Path validation: all paths within workspace.

**What Mimic does with them:**
System operations respect workspace boundary. Sandbox rules apply.

---

## caveman

**Principles taken:**
- Sensitive path protection: .env, credentials.
- File type detection: magic bytes.

**What Mimic does with them:**
System paths scanned for sensitivity. File operations typed appropriately.

---

## embryo

**Principles taken:**
- BinaryRuntime: system operations tokenized.
- Project map: workspace structure tracked.

**What Mimic does with them:**
System ops part of OpPacket chains. Workspace structure indexed.

---

## gastown

**Principles taken:**
- Rollback: system operations reversed on failure.
- Atomic operations: collision check, idempotent creation.

**What Mimic does with them:**
Destructive operations confirmed. Creation idempotent. Rollback restores moved/deleted files.

---

## Standard Unix Practice

**Principles taken:**
- mkdir -p: idempotent directory creation.
- cp --preserve: copy with integrity.
- rm -i: confirmation for destructive ops.
- chmod: restricted dangerous modes.

**What Mimic does with them:**
Standard Unix semantics adapted for safety.
