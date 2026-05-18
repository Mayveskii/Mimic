# C-Core RPC Format Specification

Exact wire formats for MCP JSON-RPC and binary mesh exchange. All formats are convertible: C struct ↔ JSON ↔ binary wire.

---

## MCP JSON-RPC Protocol

Mimic exposes an MCP (Model Context Protocol) server via JSON-RPC 2.0.

### Transport

| Transport | Framing | Default | Config Var |
|---|---|---|---|
| stdio | Newline-delimited JSON | `MIMIC_MCP_TRANSPORT=stdio` | `MIMIC_MCP_TRANSPORT` |
| TCP | Length-prefixed JSON | `MIMIC_MCP_TRANSPORT=tcp` | `MIMIC_MCP_TRANSPORT`, `MIMIC_MCP_PORT` |
| Unix socket | Length-prefixed JSON | `MIMIC_MCP_TRANSPORT=unix` | `MIMIC_MCP_TRANSPORT`, `MIMIC_MCP_UNIX_SOCKET` |

**stdio framing:**
```
{"jsonrpc":"2.0","method":"initialize","params":{...}}\n
```

**length-prefixed framing:**
```
[4 bytes big-endian length][JSON bytes]
```

### Messages

#### Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "mimic_execute",
    "arguments": {
      "intent": "commit these files safely",
      "context": {
        "session_id": "sess-uuid",
        "workspace": "/workspace/mimic"
      }
    }
  }
}
```

#### Response (Success)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "status": "success",
    "artifacts": {
      "commit_hash": "a1b2c3d",
      "files": ["main.go"]
    },
    "metrics": {
      "total_latency_ms": 145,
      "tokens_consumed": 8
    },
    "validation_report": {
      "conflict_check": "passed",
      "budget_check": "passed",
      "permission_check": "passed"
    },
    "pattern_references": [
      {"domain": "git", "pattern": "atomic_commit", "survival_index": 0.92}
    ]
  }
}
```

#### Response (Error)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "validation_failed",
    "data": {
      "reason": "conflict_pair: OP_GIT_COMMIT x OP_GIT_CHECKOUT",
      "validation_step": 6,
      "error_code": 8,
      "invalid_op_indices": [2, 3],
      "total_energy": 45.0,
      "estimated_latency_us": 15000.0
    }
  }
}
```

### Error Codes

| Code | Name | Description |
|---|---|---|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid request | JSON-RPC format error |
| -32601 | Method not found | Unknown tool |
| -32602 | Invalid params | Missing or wrong arguments |
| -32603 | Internal error | C-core panic |
| -32000 | Validation failed | Chain validation rejected |
| -32001 | Budget exceeded | Session budget exhausted |
| -32002 | Permission denied | Never-rule or dangerous op blocked |
| -32003 | Conflict detected | Pairwise conflict in chain |
| -32004 | Timeout | Operation exceeded timeout |
| -32005 | Circuit broken | Session locked, manual reset required |
| -32006 | Mesh error | Slot retrieval/storage failed |
| -32007 | Rollback failed | Atomic failure, rollback also failed |

### Methods

#### `initialize`

Client → Server. Handshake. Server responds with capabilities.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {"name": "claude-code", "version": "1.0.0"}
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {"listChanged": false},
      "resources": {"subscribe": false, "listChanged": false},
      "prompts": {"listChanged": false}
    },
    "serverInfo": {"name": "mimic", "version": "2.0.0"}
  }
}
```

#### `tools/list`

List available tools (domains + patterns).

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "mimic_execute",
        "description": "Execute a validated operation chain",
        "inputSchema": {
          "type": "object",
          "properties": {
            "intent": {"type": "string", "description": "Natural language intent"},
            "domain": {"type": "string", "enum": ["git", "build", "io", "network", "system"]},
            "scenario": {"type": "string"},
            "args": {"type": "object"},
            "context": {
              "type": "object",
              "properties": {
                "session_id": {"type": "string"},
                "workspace": {"type": "string"}
              }
            }
          },
          "required": ["intent"]
        }
      },
      {
        "name": "mimic_query",
        "description": "Query mesh for proven patterns",
        "inputSchema": {
          "type": "object",
          "properties": {
            "query": {"type": "string"},
            "domain": {"type": "string"},
            "retrieval_path": {"type": "string", "enum": ["linear", "keyword", "semantic"]}
          },
          "required": ["query"]
        }
      },
      {
        "name": "mimic_validate",
        "description": "Validate a proposed OpPacket chain without executing",
        "inputSchema": {
          "type": "object",
          "properties": {
            "chain": {
              "type": "array",
              "items": {"type": "object"}
            }
          },
          "required": ["chain"]
        }
      }
    ]
  }
}
```

#### `tools/call`

Execute a tool.

**mimic_execute:**
- Input: intent (string, required), domain (string, optional), scenario (string, optional), args (object, optional), context (object, optional).
- Output: result with status, artifacts, metrics, validation_report, pattern_references.

**mimic_query:**
- Input: query (string, required), domain (string, optional), retrieval_path (string, optional).
- Output: results array with pattern_name, survival_index, z_density, polarity, source, invariants.

**mimic_validate:**
- Input: chain (array of OpPacket JSON, required).
- Output: valid (bool), error_code, error_msg, conflicts, energy, latency.

#### `resources/list`

List resources (session state, metrics, config).

**Response:**
```json
{
  "resources": [
    {"uri": "session://budget", "name": "Session Budget", "mimeType": "application/json"},
    {"uri": "session://context", "name": "Session Context", "mimeType": "application/json"},
    {"uri": "mesh://domains", "name": "Mesh Domains", "mimeType": "application/json"},
    {"uri": "metrics://latency", "name": "Operation Latencies", "mimeType": "application/json"}
  ]
}
```

#### `resources/read`

Read resource by URI.

#### `prompts/list`

List prompts (system prompts for model interaction).

#### `prompts/get`

Get prompt by name.

### C ↔ JSON Conversion

**OpPacket → JSON:**
- `id`: uint32 → number
- `opcode`: uint32 → number (enum value)
- `opcode_name`: string (lookup table)
- `flags`: uint32 → number
- `flags_names`: array of strings (bit decode)
- `args`: array of {"key": string, "value": string, "type": number, "value_len": number}
- `arg_count`: uint8 → number
- `fd_in`, `fd_out`: int32 → number (-1 omitted in JSON)
- `buffer_size`: size_t → number
- `buffer`: base64 string (only if buffer_size > 0)
- `timestamp_ns`, `latency_ns`: uint64 → string (JSON lacks uint64)
- `timeout_ms`, `retry_count`: uint32 → number
- `result_code`: int32 → number
- `bytes_processed`: size_t → number
- `prev_op_id`, `next_op_id`, `chain_id`: uint32 → number

**JSON → OpPacket:**
- Validate all numeric ranges.
- Decode base64 buffer.
- Validate opcode against registered opcodes.
- Verify arg_count matches args array length.
- Ensure fd_in/fd_out ≥ -1.
- Set timestamp_ns from current time if missing.

---

## Binary Mesh Wire Format

For inter-node slot exchange, backup, and cold storage.

### Byte Order

All multi-byte integers are **little-endian** (x86_64 default).
Nodes on big-endian platforms MUST convert on send/receive.

### Slot Wire Format

```
[4 bytes]   magic = 0x534C4F54  // 'SLOT'
[2 bytes]   version = 2
[2 bytes]   flags
[8 bytes]   slot_id
[8 bytes]   timestamp_ns
[64 bytes]  name (null-padded)
[1 byte]    domain
[1 byte]    layer
[1 byte]    modality
[1 byte]    reserved
[4 bytes]   survival_index (IEEE 754 float32)
[4 bytes]   z_density (float32)
[4 bytes]   artifact_precision (float32)
[4 bytes]   usage_frequency (float32)
[1 byte]    polarity
[8 bytes]   counter_slot_id
[8 bytes]   anti_pattern_id
[1 byte]    invariant_count
[N × 64]    invariants (null-padded strings)
[1 byte]    tag_count
[M × 64]    tags (null-padded strings)
[1 byte]    source_count
[source_count × (128+40+256+4+4+8)] sources
[4 bytes]   text_len
[4 bytes]   text_offset
[32 bytes]  text_hash (SHA-256 binary)
[32 bytes]  extractor (null-padded)
[32 bytes]  extraction_hash (SHA-256 binary)
[8 bytes]   retrieval_count
[8 bytes]   success_count
[8 bytes]   failure_count
[8 bytes]   originating_session_id
[P bytes]   padding to 512-byte boundary
[text_len bytes] text content (follows header)
```

### Conversion Rules

**C struct → wire:**
- All strings null-padded to fixed length.
- Floats written as IEEE 754 binary32.
- slot_id written as little-endian uint64.
- text_offset = header_size_including_padding.

**Wire → C struct:**
- Validate magic (0x534C4F54).
- Validate version (2).
- Bounds-check invariant_count ≤ 16, tag_count ≤ 32, source_count ≤ 4.
- Bounds-check text_len ≤ MAX_SLOT_TEXT_SIZE (env config, default 65536).
- Verify text_offset within packet bounds.
- Verify SHA-256 of text content.

**Wire → JSON:**
- Binary fields hex-encoded.
- Floats converted to decimal with 6 digits precision.
- Strings trimmed of null padding.
- text content UTF-8 decoded.

**JSON → wire:**
- Strings null-padded to fixed length.
- Hex fields decoded to binary.
- Floats parsed from decimal.
- text content UTF-8 encoded.

### Mesh File Exchange

```
[8 bytes]   file_magic = 0x4D455348424D4150  // 'MESHBMAP'
[4 bytes]   file_version = 2
[4 bytes]   reserved
[8 bytes]   slot_count (little-endian uint64)
[8 bytes]   total_bytes_used
[8 bytes]   total_bytes_capacity
[8 bytes]   next_slot_id
[8 bytes]   index_offset
[8 bytes]   index_size
[slot_count × slot_size] slots (variable, aligned to page_size)
[index_size bytes] index (B-tree nodes)
```

**Exchange protocol:**
1. Sender reads file header.
2. Sender sends header (8 + 4 + 4 + 8*6 = 64 bytes).
3. Sender sends slots sequentially (no random access).
4. Sender sends index.
5. Receiver validates header, counts slots, verifies index checksum.
6. Receiver writes to local mesh file.

**Incremental sync:**
- Only slots with slot_id > receiver's next_slot_id - 1.
- Sender sends: header with incremental flag, delta slot count, slots.
- Receiver appends to local file, updates index.

---

## Session Snapshot Wire Format

```
[4 bytes]   magic = 0x52424C4B  // 'RBLK'
[4 bytes]   version = 1
[8 bytes]   timestamp_ns
[8 bytes]   cwd_hash (xxHash64)
[4 bytes]   env_count
[env_count × variable]:
  [2 bytes] name_len
  [name_len bytes] name
  [2 bytes] value_len
  [value_len bytes] value
[4 bytes]   git_repo_count
[repo_count × repo]:
  [2 bytes] path_len
  [path_len bytes] path
  [20 bytes] HEAD hash (binary)
  [4 bytes] branch_name_len
  [branch_name_len bytes] branch_name
[4 bytes]   fd_count
[fd_count × fd]:
  [4 bytes] fd number
  [2 bytes] path_len
  [path_len bytes] path
[4 bytes]   mmap_count
[mmap_count × mmap]:
  [8 bytes] pointer (as uint64)
  [8 bytes] size
[8 bytes]   state_hash (xxHash64)
```

---

## Convertibility Matrix

| Format | To C struct | To JSON | To binary wire |
|---|---|---|---|
| C struct | — | Yes, via serialization | Yes, memcpy + endian swap |
| JSON | Yes, via deserialization | — | Yes, encode fields |
| Binary wire | Yes, parse + validate | Yes, decode + trim | — |

All conversions are lossless for valid data. Invalid data → conversion error with specific field name.
