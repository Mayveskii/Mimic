# C-Core Memory Layout — Linear Access Patterns

The default OpPacket struct (4992 bytes) is correct for simple implementations but cache-inefficient for chains with few arguments. This spec defines an arena-based alternative for high-performance, cache-friendly linear access.

Both formats are valid. The implementation chooses based on `MIMIC_ARENA_ALLOC` env/config.

---

## Problem with Fixed-Size Arrays

**OpArg.value[256]:**
- 80% of values are < 64 bytes (short paths, small numbers).
- Waste: 192 bytes per arg × 16 args = 3KB unused per packet.
- Cache line pollution: only 64 bytes useful, 240 bytes loaded into cache.

**OpPacket.args[16]:**
- 90% of packets have < 4 arguments.
- Waste: 12 args × 304 bytes = 3.6KB unused per packet.
- Sequential packet access: args[0] loaded, args[1] in same cache line (good), but args[2..15] fill 4 cache lines with mostly zeros.

---

## Arena-Based Linear Layout

### Arena Structure

```c
typedef struct {
    uint8_t* base;         // Arena base pointer
    size_t size;           // Total arena size
    size_t used;           // Bytes committed
    size_t committed;      // Bytes reserved (for growth)
    uint32_t alignment;    // Default alignment (8 bytes)
} MimicArena;
```

### Packet Header (Fixed Size)

```c
typedef struct {
    // Identity: 16 bytes
    uint32_t id;
    uint32_t opcode;
    uint32_t flags;
    uint32_t chain_id;
    
    // Metadata: 16 bytes
    uint64_t timestamp_ns;
    uint32_t timeout_ms;
    uint32_t retry_count;
    
    // Result: 16 bytes
    int32_t result_code;
    uint32_t bytes_processed_lo;  // Split for 32-bit alignment
    uint32_t bytes_processed_hi;
    uint64_t latency_ns;
    
    // Links: 12 bytes
    uint32_t prev_op_id;
    uint32_t next_op_id;
    uint32_t arg_count;
    
    // Resources: 8 bytes
    int32_t fd_in;
    int32_t fd_out;
    
    // Buffer: 16 bytes
    uint64_t buffer_offset;   // Offset into arena (or 0 if none)
    uint64_t buffer_size;
    
    // Total: 84 bytes → padded to 96 bytes (12×8)
    uint8_t padding[12];
} OpPacketHeader;
```

`sizeof(OpPacketHeader)` = 96 bytes. Exactly 1.5 cache lines (64-byte lines).

### Argument Storage (Variable, Arena-Allocated)

Arguments stored contiguously after all headers in the arena:

```
[packet_0_header: 96 bytes]
[packet_1_header: 96 bytes]
...
[packet_N-1_header: 96 bytes]
[arg_data_block]        // All args for all packets, contiguous
[buffer_data_block]     // All buffers, contiguous
```

**Per-packet arg layout:**
```
[arg_count × 8 bytes]   // Arg descriptor array (offset + length + type)
[key_data]              // Concatenated null-terminated keys
[value_data]            // Concatenated values (not null-terminated, length known)
```

**Arg descriptor (8 bytes):**
```
[3 bytes] key_offset    // Offset from arg block start (max 16MB)
[3 bytes] value_offset  // Offset from arg block start
[1 byte]  value_len     // Max 255 bytes (sufficient for 95% of values)
[1 byte]  type            // 0-4
```

If value_len > 255: type = 4 (blob), value_offset points to buffer_data_block.

### Example: 3-Packet Chain in Arena

```c
// Chain: status → diff → commit
// Packet 0: 0 args
// Packet 1: 2 args (path, cached)
// Packet 2: 2 args (message, author)

// Arena layout:
[0:96]    packet_0_header (id=1, opcode=GIT_STATUS, arg_count=0)
[96:192]  packet_1_header (id=2, opcode=GIT_DIFF, arg_count=2, arg_offset=192)
[192:288] packet_2_header (id=3, opcode=GIT_COMMIT, arg_count=2, arg_offset=216)

// Arg descriptors start at offset 192 (packet_0 has 0 args, so no descriptor)
[192:200] packet_1_arg0: key_offset=0 ("path"), value_offset=5 ("/workspace"), len=10, type=3
[200:208] packet_1_arg1: key_offset=16 ("cached"), value_offset=23 ("false"), len=5, type=3

// Keys:
[208:213] "path\0"
[213:220] "cached\0"

// Values:
[220:230] "/workspace"
[230:235] "false"

// Total arena for 3 packets: ~288 + 43 = 331 bytes
// vs fixed-size: 3 × 4992 = 14976 bytes
// Savings: 97.8%
```

### Cache Efficiency

**Sequential packet access:**
- Read packet_0 header: 96 bytes = 1.5 cache lines.
- Read packet_1 header: next 96 bytes = already in cache (2 lines fetched).
- No wasted cache on unused args.

**Random arg access:**
- Arg descriptors: 8 bytes each, 8 per cache line.
- Keys and values: densely packed, sequential access during parsing.

---

## Macro API (Implementation Detail Hidden)

User code does NOT manipulate arena directly. Helper macros provide same API as fixed-size:

```c
OpPacketArena pkt;
ops_packet_arena_init(&arena, &pkt, OP_GIT_STATUS);
ops_packet_arena_set_string(&arena, &pkt, "path", "/workspace/repo");
ops_packet_arena_set_string(&arena, &pkt, "cached", "false");

// Access:
const char* path = ops_packet_arena_get_string(&pkt, "path");
```

Under the hood, macros compute offsets into the arena.

---

## Conversion: Fixed ↔ Arena

**Fixed → Arena:**
1. Count total arg data size.
2. Allocate arena: `N × sizeof(OpPacketHeader) + arg_data_size + buffer_size`.
3. Copy headers (fixed fields only).
4. Pack args sequentially.
5. Update arg_count, arg_offset in each header.

**Arena → Fixed:**
1. Allocate `N × sizeof(OpPacket)`.
2. Copy headers.
3. Unpack args into `args[MAX_ARGS]`.
4. Copy values into `value[256]`, truncate if > 255 (rare).
5. Set remaining args to zero.

**Use case:** Arena for internal execution (cache-friendly). Fixed for JSON serialization (simple schema).

---

## Buffer Handling

Buffers (> 1KB) stored in separate arena region to avoid fragmentation:

```
[headers_region: N × 96 bytes]
[args_region: variable]
[buffers_region: variable, 4096-byte aligned]
```

Buffer allocation: bump pointer in buffers_region, aligned to 8 bytes.

Large file I/O: buffer NOT stored in arena. FD-based I/O used instead.

---

## Validation on Arena Format

Same 11 validation steps as fixed-size:
1. arg_count ≤ MAX_ARGS (16).
2. Total arg data size ≤ 64KB per packet (sanity limit).
3. Key non-empty, null-terminated.
4. No duplicate keys (linear scan, O(arg_count²), arg_count ≤ 16 → negligible).
5. Value length matches descriptor.
6. Type in valid range.

All checks O(arg_count) per packet, not O(MAX_ARGS).

---

## RPC / Mesh Exchange

**Over-the-wire:** Always fixed-size format (simpler, no arena pointer issues).
**Internal execution:** Arena format (cache-friendly, compact).
**Mesh storage:** Fixed-size (random access, predictable offsets).

Conversion happens at boundaries:
- JSON-RPC request arrives → parse to fixed → convert to arena before EXEC.
- Mesh slot retrieval → fixed format from disk → convert to arena.
- Result assembly → arena → convert to fixed → serialize to JSON.

Conversion cost: O(N × arg_count). For typical chains (N < 100, arg_count < 4): < 1ms.

---

## Size Comparison

| Chain | Fixed Size | Arena Size | Savings |
|---|---|---|---|
| 10 packets, avg 2 args | 49,920 bytes | ~1,500 bytes | 97% |
| 100 packets, avg 3 args | 499,200 bytes | ~18,000 bytes | 96% |
| 1024 packets, avg 4 args | 4.9 MB | ~220 KB | 95% |
| 1 packet, 16 args, long values | 4,992 bytes | ~4,200 bytes | 16% |

Arena format shines for typical chains. Fixed-size acceptable for edge cases.

---

## Configuration

```c
// Env/config selects format:
// MIMIC_ARENA_ALLOC=1 (default): use arena for execution
// MIMIC_ARENA_ALLOC=0: use fixed-size for everything

#if MIMIC_ARENA_ALLOC
typedef OpPacketHeader OpPacketExec;
#else
typedef OpPacket OpPacketExec;
#endif
```

Both formats compile. Tests run with both. CI validates bit-exact results.
