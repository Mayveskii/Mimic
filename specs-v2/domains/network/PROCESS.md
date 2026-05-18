# Network Domain — HTTP, TCP

How a model makes network requests through Mimic.

---

## What This Domain Does

Network operations enable external API calls, web search, content fetching, remote distillation. Every request is validated for safety (SSRF protection), bounded by timeout, rate-limited, and logged. The model cannot make arbitrary outbound connections — it uses tokenized network operations with explicit guards.

---

## Processes

### http_get / http_post

**When to use:**  
Fetch data from external API. Web search. Content retrieval.

**Goal:**  
Safe HTTP request with SSRF protection, timeout, size limits.

**Chain (semantically):**

1. **Validate URL.**
   - Scheme whitelist: https preferred, http allowed only for specific domains.
   - Private IP block: loopback, RFC 1918, CGN, link-local → reject.
   - Cloud metadata block: 169.254.169.254 → reject.
   - DNS rebinding guard: resolve IP before request, block if resolved to private.

2. **Check rate limits.**
   - Per-domain rate limit tracking.
   - Exceeded → queue or reject with "rate_limited, retry_after Ns".

3. **Execute request.**
   - `OP_NET_HTTP_GET(url, headers, timeout)` or `OP_NET_HTTP_POST(url, body, headers, timeout)`.
   - Default timeout: 30s.
   - Max response size: 50MB binary, 10MB text.

4. **Process response.**
   - Status code, headers, body.
   - Body truncated at 10MB, overflow to temp file.
   - Return structured result.

**Hard constraints:**
- Never request private IP ranges.
- Never request cloud metadata endpoints.
- Never exceed rate limits.
- Timeout always enforced.

**Invariants:**
- URL validated before request.
- Response size bounded.
- Every request logged with timestamp, duration, result.

**Result:**
```
status: "success"
url: "https://api.example.com/v1/data"
status_code: 200
headers: {...}
body: "..."
duration_ms: 450
size: 2048
```

**Result when blocked:**
```
status: "blocked"
reason: "private_ip" | "cloud_metadata" | "rate_limited"
retry_after: 60
```

---

### tcp_connect / tcp_send / tcp_recv / tcp_close

**When to use:**  
Low-level TCP connections (rare, for specific protocols not covered by HTTP).

**Goal:**  
Managed TCP connection with timeout and size limits.

**Chain (semantically):**

1. Validate destination (same SSRF rules as HTTP).
2. Connect with timeout.
3. Send data (size bounded).
4. Receive data (size bounded, timeout enforced).
5. Close connection.

**Hard constraints:**
- Same SSRF protection as HTTP.
- Connection timeout enforced.
- Data size bounded per send/recv.

---

## Principles From Sources

### hermes-agent

**Principles taken:**
- Error classifier: network errors classified as retryable (rate limit, timeout) vs permanent (auth, not_found).
- Streaming health: 90s stale-stream detection, 60s read timeout.
- Credential pool: multi-key rotation, env fallback.

### graphify

**Principles taken:**
- SSRF protection: private-IP block, metadata block, DNS rebinding guard.
- Streaming fetch with size caps.

### exa-mcp-server

**Principles taken:**
- Web search tool: query → structured results with metadata.
- Rate limit management: API key rotation.

---

## Artifact Storage

| Field | Value |
|-------|-------|
| domain | "network" |
| layer | "process" |
| modality | "code" |
| invariants | ["url_validated", "ssrf_protected", "rate_limited", "timeout_enforced", "size_bounded"] |

---

## Cross-Domain Conflicts

Network domain conflicts with:
- **build domain**: network fetch during build = possible non-determinism.
- **distillation domain**: concurrent git clone and HTTP request = bandwidth contention.
