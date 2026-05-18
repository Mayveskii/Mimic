# C-Core Mesh Exchange Specification

How mesh slots move between nodes, processes, and persistent storage. Cross-platform, versioned, endianness-safe.

---

## Design Principles

1. **Linear write, linear read**: No random access during exchange. Sequential stream.
2. **Self-describing**: Every chunk has magic + version + length. No external schema needed.
3. **Incremental**: Delta sync supported. Append-only for live mesh.
4. **Integrity**: CRC32 per chunk, SHA-256 per slot, xxHash64 for index.

---

## Exchange Protocol

### Transport Options

| Transport | Use Case | Framing |
|---|---|---|
| File mmap | Single-node persistence | Direct memory mapping |
| Unix socket | Local inter-process | Length-prefixed chunks |
| TCP | Remote node replication | Length-prefixed + TLS |
| HTTP/REST | Client-server query | JSON (for human debugging only) |

### Chunk Format (all transports)

```
[4 bytes]   chunk_magic     // 'MSLT' = mesh slot, 'MIDX' = index, 'MSNP' = snapshot
[2 bytes]   chunk_version   // 2
[2 bytes]   chunk_flags     // Bit 0: compressed, Bit 1: encrypted
[8 bytes]   chunk_length    // Total chunk size including header (little-endian)
[4 bytes]   chunk_crc32     // CRC32 of chunk_data only
[N bytes]   chunk_data      // Actual payload
```

Header size: 20 bytes. Chunk data starts at offset 20.

CRC32 computed over chunk_data only (not including header). Verified before parsing.

### Slot Chunk (MSLT)

```
[chunk_header: 20 bytes]
[slot_header: 512 bytes, padded]
[text_content: text_len bytes, padded to 8-byte boundary]
```

Slot header is the binary format from SLOT_SCHEMA.md.
Text content follows immediately after header padding.

### Index Chunk (MIDX)

```
[chunk_header: 20 bytes]
[index_header: 32 bytes]
[B-tree nodes: node_count × node_size]
```

Index header:
```
[8 bytes]   node_count
[8 bytes]   node_size    // Typically 4096 bytes (page aligned)
[8 bytes]   root_offset  // Offset to root node within this chunk
[8 bytes]   index_hash   // xxHash64 of all node data
```

### Snapshot Chunk (MSNP)

```
[chunk_header: 20 bytes]
[snapshot_data: session snapshot wire format]
```

---

## Synchronization Protocol

### Full Sync

```
Node A (sender)                          Node B (receiver)
    |                                         |
    |--- [MSLT × slot_count] --------------> |
    |--- [MIDX] ----------------------------> |
    |--- [MSNP, optional] ------------------> |
    |                                         |
    |<-- [ACK: next_slot_id] --------------- |
```

1. Sender opens mesh file, reads header.
2. Sender streams all slots sequentially (no seeking).
3. Sender streams index.
4. Receiver validates each chunk CRC32.
5. Receiver appends slots to local mesh.
6. Receiver rebuilds index from received data.
7. Receiver responds with next expected slot_id.

### Delta Sync

```
Node A                                    Node B
    |                                         |
    |<-- [REQ: last_slot_id = N] ----------- |
    |                                         |
    |--- [MSLT × (N+1 .. M)] ------------> |
    |--- [MIDX delta] ---------------------> |
    |                                         |
    |<-- [ACK: next_slot_id = M+1] --------- |
```

1. Receiver sends last known slot_id.
2. Sender seeks to slot offset (O(1) via index) — **only seek in delta sync**.
3. Sender streams only new slots.
4. Receiver appends.

### Live Replication

For hot standby nodes:
1. Primary appends new slot to mesh file.
2. Primary sends delta notification to replicas (slot_id, offset, size).
3. Replicas mmap the shared region or receive via socket.
4. Replicas validate and append.

---

## Endianness

All multi-byte integers are **little-endian**.

**Conversion macros (for big-endian hosts):**
```c
#define MIMIC_LE16(x) (x)  // No-op on little-endian
#define MIMIC_LE32(x) (x)
#define MIMIC_LE64(x) (x)

// On big-endian:
#define MIMIC_LE16(x) __builtin_bswap16(x)
#define MIMIC_LE32(x) __builtin_bswap32(x)
#define MIMIC_LE64(x) __builtin_bswap64(x)
```

**Floats:** IEEE 754 binary32. Byte order follows integer endianness convention (little-endian bytes).

**Text:** UTF-8. No byte order mark. Platform-independent.

---

## Versioning

**Chunk version = 2.**

- Version 1: Deprecated. No longer supported.
- Version 2: Current. All fields as documented.
- Version 3+: Reserved. Receiver MUST reject unknown versions.

**Forward compatibility:** Unknown chunk types skipped (with warning). Unknown flags handled as fatal error (safety).

---

## Integrity

**Per-chunk:** CRC32 (zlib algorithm). Detects transmission corruption.
**Per-slot:** SHA-256 of text content. Detects content tampering.
**Per-index:** xxHash64 of all nodes. Detects index corruption.
**Per-file:** HMAC-SHA256 of entire mesh file (optional, keyed by credential pool).

**Recovery:**
- CRC32 fail → re-request chunk.
- SHA-256 fail → quarantine slot, request re-send.
- xxHash64 fail → rebuild index from slots (slow but safe).

---

## Compression

Chunk flag bit 0: chunk_data is LZ4 compressed.

**Decision rule:**
- text_content > 4096 bytes → compress.
- text_content ≤ 4096 bytes → uncompressed (overhead not worth it).
- index chunk → always uncompressed (already compact).
- snapshot chunk → compressed if > 4096 bytes.

**Format:**
```
[8 bytes]   uncompressed_size
[N bytes]   LZ4 compressed data
```

---

## Encryption

Chunk flag bit 1: chunk_data is AES-256-GCM encrypted.

**Header:**
```
[12 bytes]  nonce (random per chunk)
[N bytes]   ciphertext
[16 bytes]  authentication tag
```

Key from credential pool (`MIMIC_CREDENTIAL_POOL_PATH`). Key ID prepended to chunk.

**Use case:** Remote replication over untrusted network. Local file NOT encrypted (sandbox + permissions are sufficient).

---

## Go ↔ C Interop

Go bridge reads mesh via shared mmap or socket.

**Go struct (cgogen from C header):**
```go
type SlotHeader struct {
    Magic        uint32
    Version      uint16
    Flags        uint16
    SlotID       uint64
    TimestampNs  uint64
    Name         [64]byte
    Domain       uint8
    Layer        uint8
    Modality     uint8
    Reserved     uint8
    SurvivalIndex float32
    ZDensity      float32
    ArtifactPrecision float32
    UsageFrequency float32
    Polarity     uint8
    CounterSlotID uint64
    AntiPatternID uint64
    // ... remaining fields
}
```

**Alignment:** Go's struct layout matches C with `#pragma pack(1)` or explicit padding. cgo generates compatible layout on little-endian platforms.

**Byte order:** Go `binary.LittleEndian` for all reads/writes.

**String conversion:** C null-padded `[N]byte` → Go `string(bytes[:clz(bytes)])`.

---

## Wire Format Summary

| Layer | Format | Size |
|---|---|---|
| Transport | Length-prefixed or mmap | Variable |
| Chunk | Magic(4) + Version(2) + Flags(2) + Length(8) + CRC32(4) + Data(N) | 20 + N |
| Slot | Header(512) + Padding + Text(text_len) + Padding | 512 + pad + text_len + pad |
| Index | Header(32) + Nodes(node_count × node_size) | 32 + node_count × 4096 |
| Snapshot | Session wire format | Variable |

All sizes are deterministic. No heap allocation during parsing (stack + arena only).
