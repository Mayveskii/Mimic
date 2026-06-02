# specs/12-EXA-RESEARCH.md — How Models Use Exa Tools in Mimic

## Purpose

This spec defines when and how AI models (Claude, GPT, Kimi, etc.) should call Exa tools through the Mimic MCP server. It is written for both the model (instruction) and the developer (reference).

## The Three Tools

### EXA_SEARCH — Discovery
**Call when**: You don't know which URL has the answer.
**Examples**:
- "Find the best Go HTTP router library"
- "How does etcd handle leader election?"
- "Compare Prometheus vs Grafana for metrics"
**Returns**: 5-10 results with title, URL, highlights (1-sentence extract)
**Model strategy**: Use for initial discovery, then EXA_FETCH the most relevant URL(s).

### EXA_FETCH — Extraction  
**Call when**: You already have URL(s) and need the content.
**Examples**:
- "Get me the code example from https://github.com/etcd-io/raft"
- "Extract the docs from https://pkg.go.dev/..."
**Returns**: Full text (respecting `max_characters` limit)
**Model strategy**: Always specify `max_characters` to avoid huge context dumps. Start with 500 chars, expand if needed.

### MIMIC_RESEARCH — Synthesis
**Call when**: You need a comprehensive answer, not just one source.
**Examples**:
- "Explain Go graceful shutdown patterns"
- "How to implement circuit breaker in Go?"
**`depth=shallow`**: Returns titles + URLs (fast, ~0.5s). Use when you just need pointers.
**`depth=deep`**: Returns search + fetch + RTK compress of top result (~3-8s). Use when you need actual content.
**Model strategy**: Shallow for orientation, deep for implementation details.

## Decision Tree for Models

```
Do I need to research something I don't know?
├── Yes → Do I have specific URLs?
│   ├── Yes → EXA_FETCH (extract content)
│   └── No → EXA_SEARCH (find URLs)
│       └── After search: EXA_FETCH (best URL)
└── No → Do I need a comprehensive overview?
    ├── Yes → MIMIC_RESEARCH(deep)
    └── No → Use mesh/project tools (local knowledge)
```

## Anti-Patterns (Don't Do This)

1. **Don't** use MIMIC_RESEARCH(deep) when you just need one fact. Use EXA_SEARCH with num_results=1.
2. **Don't** fetch huge pages without max_characters. You will waste context window.
3. **Don't** use deep search type (`type=deep`) for simple queries. It takes 4-15s. Use `auto` (1s).
4. **Don't** retry failed requests manually. Mimic handles retries. Just handle the error.

## Context Window Budget

| Tool | Typical response size | Token cost |
|------|----------------------|------------|
| EXA_SEARCH (10 results) | ~3KB | ~750 tokens |
| EXA_FETCH (500 chars) | ~500B | ~125 tokens |
| EXA_FETCH (5000 chars) | ~5KB | ~1250 tokens |
| MIMIC_RESEARCH(deep) | ~2-4KB | ~500-1000 tokens |

**Rule**: If your remaining context budget < 2000 tokens, use EXA_SEARCH with `num_results=3` or EXA_FETCH with `max_characters=500`.

## Example Workflow: "Implement rate limiter in Go"

1. **MIMIC_RESEARCH(shallow)** on "Go rate limiter patterns"
   → Gets: 5 URLs (token bucket, sliding window, etc.)

2. **EXA_FETCH** top URL with max_characters=2000
   → Gets: Code example + explanation

3. **Write code** using SYS_FILE_WRITE
   → Implements token bucket

4. **EXA_FETCH** second URL (if first wasn't enough)
   → Alternative approach

Total research tokens: ~500 + ~500 = ~1000 (very efficient)

## Integration with Mesh

Exa is **complementary** to mesh (local distilled knowledge):
- **Mesh**: Fast, offline, proven patterns from production repos
- **Exa**: Online, current, documentation and recent articles

**Best practice**: Query mesh first (MESH_QUERY), then Exa if mesh has no answer.
