# Network Domain — Sources

Where the network domain behavior comes from.

---

## graphify

**Principles taken:**
- SSRF protection: private-IP block, metadata block, DNS rebinding guard.
- Streaming fetch with size caps.

**What Mimic does with them:**
URL validated before every request. Private IPs blocked. DNS resolved before connect. Response size bounded.

**What Mimic does NOT copy:**
- Graphify's specific proxy configuration.
- Graphify's CDN routing logic.

---

## hermes-agent

**Principles taken:**
- Error classifier: network errors classified as retryable (rate limit, timeout) vs permanent (auth, not_found).
- Streaming health: 90s stale-stream detection, 60s read timeout.
- Credential pool: multi-key rotation, env fallback.

**What Mimic does with them:**
Network errors classified before retry decision. Streaming monitored for staleness. Credentials managed via pool, never hardcoded.

**What Mimic does NOT copy:**
- Hermes-agent's specific API retry formula.
- Hermes-agent's backoff curve (Mimic uses exponential backoff).

---

## exa-mcp-server

**Principles taken:**
- Web search tool: query → structured results with metadata.
- Rate limit management: API key rotation.

**What Mimic does with them:**
HTTP GET/POST structured responses. Rate limit tracking per domain. Key rotation on 429.

**What Mimic does NOT copy:**
- exa-mcp-server's specific MCP protocol framing.
- exa-mcp-server's vector search API.

---

## bun (PR #30412)

**Principles taken:**
- Session enrichment: every network request logged with session context.
- Permission pipeline: network requests classified by safety level.

**What Mimic does with them:**
Every request logged. Network ops classified as safety level 2 (network = external, not necessarily dangerous). Dangerous network ops (deploy, push) classified as level 0.

---

## caveman

**Principles taken:**
- Sensitive path protection: applies to network — credential leakage in URLs.
- File type detection: response Content-Type validated against body.

**What Mimic does with them:**
URLs scanned for credential patterns. Response types validated.

---

## rustnet

**Principles taken:**
- Sandbox: network access can be blocked per process.
- Outbound filtering: Landlock can restrict network connections.

**What Mimic does with them:**
Network operations respect sandbox rules. Build/test processes in sandbox may have network blocked unless explicitly enabled.

**What Mimic does NOT copy:**
- Rustnet's eBPF-based network filtering.
- Rustnet's specific sandbox policy DSL.

---

## Standard Security Practice

**Principles taken:**
- SSRF prevention: private IP block, metadata endpoint block.
- Rate limiting: token bucket algorithm.
- Timeout enforcement: connection + read + write timeouts.
- Credential rotation: never hardcode, use pool.

**What Mimic does with them:**
Standard security stack applied to all network operations.
