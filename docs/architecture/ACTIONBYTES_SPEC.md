# ActionBytes Decoder Specification

## Overview

ActionBytes in embryo binary mesh follow a **binary patch format** for deterministic, replayable edits. This format encodes file modifications in a compact binary representation that can be executed without an LLM.

## Format (Embryo Binary Patch)

ActionBytes is a sequence of **patch entries**. Each entry:

```
Entry := Marker(1) + PathLength(2) + Path(N) + ContentLength(4) + Content(M)
```

| Field | Size | Description |
|-------|------|-------------|
| Marker | 1 byte | Always `0x21` (`!`) |
| PathLength | 2 bytes (big-endian) | Length of path string |
| Path | N bytes | UTF-8 file path |
| ContentLength | 4 bytes (big-endian) | Length of content to write |
| Content | M bytes | Raw content bytes |

Multiple entries concatenate without delimiter.

## Translation to MCP Tools

Each entry translates to either:
- **FILE_EDIT** — if the target file exists and the entry represents replacement
- **SYS_EXEC** — if the entry represents a command
- **FILE_INSERT** — if the entry represents insertion at offset

## Example (from etcd slot)

```
21 03 1a CHANGELOG/CHANGELOG-3.7.md (len=26)
00 00 00 8d [141 bytes content]
21 03 20  server/etcdserver/server_test.go (len=32)
00 00 88 09 [34825 bytes content]
```

This updates two files in a single deterministic operation.

## Decoder Implementation

Location: `internal/mesh/actionbytes.go`

Function: `DecodeActionBytes(data []byte) ([]PatchEntry, error)`

Returns slice of `PatchEntry{Path, Content}` for agent analysis and execution.

## Safety

Binary patches are **never auto-executed**. They are exposed as hex for agent analysis, then translated to FILE_EDIT chains with human confirmation.

## Future

Full binary format may include:
- Operation type byte (replace/insert/delete)
- Offset for insertions
- Checksums for verification
- Compression (gzip)
