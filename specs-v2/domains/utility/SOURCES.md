# Utility Domain — Sources

Where the utility domain behavior comes from.

---

## Standard Cryptography

**Principles taken:**
- SHA-256 for integrity verification.
- MD5 for legacy compatibility (not for security).
- AES-256-GCM for authenticated encryption.

**What Mimic does with them:**
Hash for file integrity, slot verification, state snapshots. AES-GCM for credential pool and sensitive data.

---

## Standard Compression

**Principles taken:**
- Gzip (zlib) for compression/decompression.
- Level 6 as default (speed/size balance).

**What Mimic does with them:**
Session snapshots compressed. Backups compressed. Large responses compressed for storage.

---

## caveman

**Principles taken:**
- File type detection via magic bytes (for validation).
- Sensitive path protection (key management).

**What Mimic does with them:**
Utility ops used for file validation. Keys managed via credential pool.

---

## hermes-agent

**Principles taken:**
- Credential rotation.
- Multi-key pools.

**What Mimic does with them:**
Encryption keys rotated from pool. Key IDs used in operations.

---

## Standard Practice

**Principles taken:**
- Pure functions for utility ops.
- Deterministic outputs.
- Bounded inputs.

**What Mimic does with them:**
Utility ops are read-only, deterministic, bounded. No side effects.
