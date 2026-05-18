# Slot Schema — Exact Structure for Compilation

A slot is the fundamental storage unit in the Mimic mesh. It holds one proven pattern, extracted from one source, with full provenance and measured quality signals.

Every slot is immutable after creation. Updates append new slots; old slots are archived.

---

## Binary Layout (mmap-backed)

```c
#define SLOT_MAGIC          0x534C4F54  // 'SLOT'
#define SLOT_VERSION        2
#define SLOT_MAX_INVARIANTS 16
#define SLOT_MAX_TAGS       32
#define SLOT_MAX_SOURCES    4
#define SLOT_NAME_LEN       64
#define SLOT_HASH_LEN       32          // SHA-256 binary
#define SLOT_TEXT_MAX       65536       // 64KB pattern text

typedef struct {
    // Header
    uint32_t magic;              // SLOT_MAGIC
    uint16_t version;            // SLOT_VERSION
    uint16_t flags;              // Bit 0: has_counter, Bit 1: is_negative
    
    // Identity
    uint64_t slot_id;            // Global unique ID (monotonic)
    uint64_t timestamp_ns;       // Creation time
    char name[SLOT_NAME_LEN];    // Pattern name (null-terminated)
    
    // Domain classification
    uint8_t domain;              // 0-15 enum (see DomainEnum below)
    uint8_t layer;               // 0=primitive, 1=composite, 2=process
    uint8_t modality;            // 0=code, 1=text, 2=binary, 3=hybrid
    uint8_t reserved;            // Padding
    
    // Quality signals
    float survival_index;        // 0.0-1.0, git blame derived
    float z_density;             // 0.0-1.0, compression ratio derived
    float artifact_precision;    // 0.0-1.0, extraction quality
    float usage_frequency;       // 0.0+, times retrieved / total queries
    
    // Polarity
    uint8_t polarity;            // 0=POSITIVE, 1=NEGATIVE, 2=COUNTER
    uint64_t counter_slot_id;    // If NEGATIVE, links to COUNTER slot
    uint64_t anti_pattern_id;  // If NEGATIVE, links to anti-pattern record
    
    // Invariants (up to 16)
    uint8_t invariant_count;
    char invariants[SLOT_MAX_INVARIANTS][SLOT_NAME_LEN];
    
    // Tags (up to 32)
    uint8_t tag_count;
    char tags[SLOT_MAX_TAGS][SLOT_NAME_LEN];
    
    // Source provenance (up to 4 stacked sources)
    uint8_t source_count;
    struct {
        char repo[128];          // github.com/owner/repo
        char commit[40];         // Hex SHA-1
        char path[256];          // File path in repo
        uint32_t line_start;
        uint32_t line_end;
        uint64_t blame_timestamp;
    } sources[SLOT_MAX_SOURCES];
    
    // Content
    uint32_t text_len;           // Length of pattern text
    uint32_t text_offset;        // Offset from slot start to text
    uint8_t text_hash[SLOT_HASH_LEN]; // SHA-256 of text
    
    // Extraction metadata
    char extractor[32];          // Tool name + version
    uint8_t extraction_hash[SLOT_HASH_LEN]; // Hash of extractor + parameters
    
    // Statistics
    uint64_t retrieval_count;    // How many times this slot was retrieved
    uint64_t success_count;    // How many times applied successfully
    uint64_t failure_count;    // How many times applied and failed
    
    // Session link (for feedback)
    uint64_t originating_session_id;
    
    // Padding to 512-byte boundary
    uint8_t padding[64];
} SlotHeader;
```

### Size

Header only: ~4 + 2 + 2 + 8 + 8 + 64 + 1 + 1 + 1 + 1 + 4 + 4 + 4 + 4 + 1 + 8 + 8 + 1 + 16*64 + 1 + 32*64 + 1 + 4* (128+40+256+4+4+8) + 4 + 4 + 32 + 32 + 32 + 8 + 8 + 8 + 8 + 64
= 4+2+2+8+8+64+4+16+1+8+8+1024+1+2048+1+1744+4+4+32+32+32+8+8+8+8+64
= 5120 bytes + padding to 5120 (already aligned if calculated correctly).

Actually: 4+2+2=8, +8+8=24, +64=88, +4=92, +16=108, +1+8+8=125, +1024=1149, +1+2048=3198, +1+1744=4943, +4+4+32+32+32=5047, +8+8+8+8=5079, +64=5143.
So padding = 57 bytes to reach 5200 (multiple of 8).

Text content follows header: up to 64KB.
Total slot size: header (5200) + text (up to 65536) = up to 70736 bytes.

### Alignment

Slots are aligned to 4096-byte boundaries for mmap page alignment.
`slot_offset = slot_id * 4096` (simplified; actual index uses B-tree).

---

## Domain Enum

```c
typedef enum {
    DOMAIN_GIT = 0,
    DOMAIN_BUILD = 1,
    DOMAIN_IO = 2,
    DOMAIN_NETWORK = 3,
    DOMAIN_PROCESS = 4,
    DOMAIN_MEMORY = 5,
    DOMAIN_SYSTEM = 6,
    DOMAIN_UTILITY = 7,
    DOMAIN_ORCHESTRATOR = 8,
    DOMAIN_SESSION = 9,
    DOMAIN_RAG = 10,
    DOMAIN_MESH = 11,
    DOMAIN_DISTILLATION = 12,
    DOMAIN_SECURITY = 13,
    DOMAIN_QUALITY = 14,
    DOMAIN_ANTI_PATTERNS = 15,
    DOMAIN_MAX = 16
} DomainEnum;
```

---

## Polarity Enum

```c
typedef enum {
    POLARITY_POSITIVE = 0,   // What TO do
    POLARITY_NEGATIVE = 1,   // What NOT to do (must link to counter)
    POLARITY_COUNTER  = 2    // The correct alternative to a NEGATIVE
} PolarityEnum;
```

---

## Flags

```c
#define SLOT_FLAG_HAS_COUNTER   0x0001  // counter_slot_id is valid
#define SLOT_FLAG_IS_NEGATIVE   0x0002  // This slot documents a failure mode
#define SLOT_FLAG_FROM_FEEDBACK 0x0004  // Created from execution feedback, not distillation
#define SLOT_FLAG_ARCHIVED      0x0008  // Superseded by newer slot
#define SLOT_FLAG_VERIFIED      0x0010  // 2-vote verified
```

---

## Index Structure

Slots are indexed by multiple keys for O(1) or O(log n) retrieval:

```c
typedef struct {
    uint64_t slot_id;
    uint64_t hash64;           // xxHash64 of name + domain + layer
} SlotIndexEntry;

// Primary index: slot_id -> file offset (B-tree)
// Secondary index: domain:layer:modality:name -> slot_id (hash table)
// Tertiary index: invariant_hash -> list of slot_ids (inverted index)
// Semantic index: int8 quantized embedding -> vector similarity (Qdrant)
```

### Retrieval Paths

1. **Linear**: `si_query_exact(domain, layer, modality, name)` → single slot_id.
2. **Keyword**: `si_query_invariant(invariant_hash)` → list of slot_ids sharing invariant.
3. **Semantic**: `int8_quantize(query_embedding)` → batch_cosine_int8 → top-k by survival_index * z_density reranking.

---

## Slot Lifecycle

```
EXTRACTION → VALIDATION → INDEXING → STORAGE → RETRIEVAL → APPLICATION → FEEDBACK
```

1. **Extraction**: Distillation pipeline creates raw slot content.
2. **Validation**: Quality gates check artifact_precision ≥ 0.8, survival_index > 0.0.
3. **Indexing**: All indexes updated atomically.
4. **Storage**: Slot appended to mmap region. Old slots never modified.
5. **Retrieval**: Query resolved via linear → keyword → semantic path.
6. **Application**: Pattern applied in execution context.
7. **Feedback**: Success/failure recorded. If failure rate > threshold, slot flagged for review.

---

## Slot File Format

```
[8 bytes]  file_magic = 0x4D455348424D4150 // 'MESHBMAP'
[4 bytes]  file_version = 2
[8 bytes]  slot_count
[8 bytes]  total_bytes_used
[8 bytes]  total_bytes_capacity
[8 bytes]  next_slot_id
[8 bytes]  index_offset
[8 bytes]  index_size
[... slots ...]
[... index ...]
```

The file is memory-mapped with `MAP_SHARED`. Writes use atomic append.
Index is rebuilt on open if `index_offset == 0` (first open) or on explicit `reindex()`.

---

## Integrity

- Every slot header and text has SHA-256 hash.
- File header has CRC32 of the index region.
- On open: verify file_magic, verify each slot's magic, verify text hashes.
- Corrupted slots are logged and skipped, not loaded.
- Backup files: `.bmap.YYYYMMDD.HHMMSS` created before any write operation that modifies > 1% of file.
