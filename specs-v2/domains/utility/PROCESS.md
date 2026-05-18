# Utility Domain — Hash, Compress, Encrypt

How Mimic provides cryptographic and compression utilities.

---

## What This Domain Does

Utility operations provide deterministic, stateless functions: hashing, compression, encryption/decryption. These are building blocks used by other domains (mesh, session, distillation) for integrity verification and data protection.

---

## Processes

### hash_sha256 / hash_md5

**When to use:**  
Verify data integrity. Used on every slot write/read, session persist/restore.

**Goal:**  
Compute cryptographic hash of data.

**Chain (semantically):**

1. Input data validated (not NULL, length > 0).
2. Compute hash: `OP_HASH_SHA256(data, len)` → 32-byte hash.
3. Return hash.

**Invariants:**
- Same input → same hash. Always.
- Hash verified before any decompression/decryption.

---

### compress_gzip / decompress_gzip

**When to use:**  
Compress data at rest (mesh slots, session snapshots, matrices).

**Goal:**  
Reduce storage size with integrity verification.

**Chain (semantically):**

1. Input data validated.
2. Compress: `OP_COMPRESS_GZIP(data)`.
3. Compute compression_ratio = original_size / compressed_size.
4. Track ratio per resource type.
5. Alert if ratio < 1.5 (possible corruption or already compressed data).

**Invariants:**
- Every slot write is compressed.
- Every read verifies hash before decompress.
- Compression ratio tracked.

**Result:**
```
status: "success"
compressed_size: 2048
original_size: 4096
compression_ratio: 2.0
```

---

### encrypt_aes / decrypt_aes

**When to use:**  
Encrypt sensitive artifacts (session state with credentials, DIFC-classified data).

**Goal:**  
Encrypt data with AES, decrypt with key.

**Chain (semantically):**

1. Input data and key validated.
2. Encrypt: `OP_ENCRYPT_AES(data, key, iv)`.
3. Store encrypted data + IV.
4. Decrypt: `OP_DECRYPT_AES(encrypted, key, iv)`.

**Hard constraints:**
- Key never stored in bmap. Key managed by credential pool.
- IV unique per encryption operation.
- Never encrypt with hardcoded key.

**Invariants:**
- Decrypt(Encrypt(data, key)) == data. Always.
- Key management separate from data storage.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "utility" |
| layer | "infrastructure" |
| modality | "code" |
| invariants | ["deterministic_hash", "compressed_and_verified", "key_separate_from_data"] |

---

## Cross-Domain Conflicts

Utility domain is stateless. No conflicts. Used by all other domains.
