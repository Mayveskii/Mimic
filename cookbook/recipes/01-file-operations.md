# Recipe 1: File Operations — Read, Check, Hash

> **Tier:** 1 (Single tool)  
> **Complexity:** Low  
> **Status:** ✅ Tested and passing  
> **Token savings:** Low (small outputs)

## Problem

Your agent needs to read files, check existence, or compute hashes without guessing paths or modes.

## Solution

### Read a file by path

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "SYS_FILE_READ",
    "arguments": {
      "path": "README.md",
      "limit": 4096
    }
  }
}
```

**Response (real output, truncated):**
```text
# Mimic – Deterministic AI-Agent Tool Orchestration

> **Mimic is not the agent. The agent is autonomous.**
```

**Measured:** 3,059 chars returned instantly.

### Check if file exists

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "SYS_FILE_EXISTS",
    "arguments": {
      "path": "bin/mimic"
    }
  }
}
```

**Response:**
```text
exists: true
```

### Compute SHA256 hash

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "HASH_SHA256",
    "arguments": {
      "data": "hello world"
    }
  }
}
```

**Response:**
```text
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

## Before vs After

| Metric | Before (raw bash) | After (Mimic) |
|--------|------------------|---------------|
| Path errors | ~10% (wrong cwd) | 0% (resolved against project root) |
| Mode errors | ~30% (IO_READ needs fd) | 0% (SYS_FILE_READ by path) |
| Output capture | Manual | Automatic in `result` |

## Data Source

Real outputs from `2026-05-18` run against commit `47eb8fa`.
