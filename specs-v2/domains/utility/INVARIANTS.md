# Utility Domain — Invariants

Rules that MUST hold for every process in the utility domain.

---

## UINV-01: Hash Deterministic

**What it prevents:** Non-deterministic outputs from hash functions.

**What it requires:** OP_HASH_SHA256 and OP_HASH_MD5 produce identical output for identical input, always. No seed variation, no algorithm changes.

**Source of this rule:**
- Cryptographic standard requirements.
- Determinism principle.

**Consequence of violation:** Hash mismatch detected in verification. Operation returns failure.

---

## UINV-02: Compression Reversible

**What it prevents:** Data loss from compression.

**What it requires:** OP_COMPRESS_GZIP output MUST be decompressible by OP_DECOMPRESS_GZIP with identical result. Verified with test round-trip on random sample.

**Source of this rule:**
- Data integrity principle.
- Standard compression library behavior.

**Consequence of violation:** Decompression failure detected. Compression operation marked as PARTIAL.

---

## UINV-03: Encryption Authenticated

**What it prevents:** Tampering, padding oracle attacks, unauthorized decryption.

**What it requires:** OP_ENCRYPT_AES uses AES-256-GCM with authentication tag. OP_DECRYPT_AES verifies tag before returning plaintext. Key referenced by ID from credential pool, never inline.

**Source of this rule:**
- Modern cryptography best practice.
- AP-07 (hardcoded secrets).

**Consequence of violation:** Decryption REJECTED with "authentication failed" if tag mismatch. Key inline → REJECTED at validation.

---

## UINV-04: Key Management Isolated

**What it prevents:** Key leakage, unauthorized key usage.

**What it requires:** Encryption/decryption keys stored in credential pool. Key IDs used in operations, not key material. Key pool encrypted at rest. Key rotation supported.

**Source of this rule:**
- Standard key management practice.
- caveman: sensitive path protection.

**Consequence of violation:** Inline key detected → REJECTED with `ERR_PERMISSION_DENY`.

---

## UINV-05: Input Size Bounded

**What it prevents:** DoS via huge input to utility functions.

**What it requires:** Max input size: 1GB for hash, 1GB for compress, 100MB for encrypt. Exceeded → stream processing or REJECTED.

**Source of this rule:**
- Resource management.
- Standard DoS protection.

**Consequence of violation:** Operation REJECTED with `ERR_OOM`.

---

## UINV-06: Side-Effect Free

**What it prevents:** Utility functions modifying state unexpectedly.

**What it requires:** ALL utility ops (hash, compress, encrypt) are PURE FUNCTIONS. They read input, produce output, modify nothing else. No filesystem, no network, no env changes.

**Source of this rule:**
- Functional programming principle.
- Determinism requirement.

**Consequence of violation:** If utility op attempts file/network access → REJECTED with `ERR_PERMISSION_DENY`.
