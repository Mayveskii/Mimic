# RAG Domain — Sources

Where the RAG domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- Hybrid RAG: 5-signal ranking (vector+keyword+domain+survival+Z-density) + qdrant.
- Mesh graphs: domain/layer/state_hash indexed slots.

**What Mimic does with them:**
Retrieval uses 5-signal ranking. Slots indexed by domain, layer, modality, invariants. Qdrant for semantic search.

**What Mimic does NOT copy:**
- Embryo's specific qdrant schema.
- Embryo's vector dimensionality (Mimic uses configurable, default 384).

---

## graphify

**Principles taken:**
- IDF-weighted search: exact > prefix > substring with gap-ratio cutoff.
- Hub-throttled traversal: skip high-degree hubs as transit.
- Two-pass extraction: structural + call-graph with confidence labels.

**What Mimic does with them:**
Keyword search uses exact match first, then prefix, then substring. High-frequency patterns (hubs) not over-represented. Confidence labels filter low-quality results.

**What Mimic does NOT copy:**
- Graphify's AST parser.
- Graphify's graph database backend.

---

## hermes-agent

**Principles taken:**
- Context compression pipeline: preflight detection + multi-pass compression.
- Stable prefix: system prompt preserved during compression.

**What Mimic does with them:**
Retrieved patterns compressed for context inclusion. Stable prefix (invariant list) preserved. Preflight detects context pressure.

---

## exa-mcp-server

**Principles taken:**
- Web search tool: query → structured results.
- API key rotation on rate limit.

**What Mimic does with them:**
Retrieval results structured with metadata. Rate limit management for external search APIs.

---

## caveman

**Principles taken:**
- Sensitive path protection: no credential leakage in retrieved data.
- File type detection: appropriate handling of binary/text results.

**What Mimic does with them:**
Retrieved slots scanned for credential patterns. Binary content flagged.

---

## Standard IR Practice

**Principles taken:**
- TF-IDF / BM25 for keyword ranking.
- Vector similarity (cosine) for semantic search.
- Re-ranking: combine multiple signals.
- Pagination for large result sets.

**What Mimic does with them:**
Standard information retrieval stack adapted for slot retrieval.
