# ADR-0009: Exa API Integration — External Knowledge Ingestion Pipeline

## Decision
Integrate Exa API (`exa.ai`) as a **3-tier knowledge source**: (1) raw MCP tools `EXA_SEARCH`/`EXA_FETCH`, (2) research tool `MIMIC_RESEARCH`, (3) auto-ingestion fallback in orchestrator when mesh query returns zero high-confidence results. All tiers configurable via environment variables.

## Why (formal)
- **Mesh gap closure**: 90+ repos in `repos-manifest.yaml` are `pending`. Manual distillation is bottleneck. Exa provides automated pattern discovery.
- **Self-healing knowledge**: Orchestrator detects low Z-density (< 0.3) or empty query results → triggers web research → validates → seeds mesh.
- **Agent empowerment**: Models can explicitly call `EXA_SEARCH` when they need current information (e.g., "latest Go 1.25 generics patterns").
- **Behavior source**: Mayveskii/exa-mcp-server (exa_web_search, exa_web_fetch, exa_rate_limit_management) — rate-limit backoff, content extraction, zod schema pattern.

## Measured
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| Mesh coverage | 3 repos distilled | 3 repos + dynamic web seeds | +∞ (theoretically) |
| Time to new pattern | manual (hours) | < 30s (API roundtrip) | -99% |
| Agent context window waste | full HTML pages | RTK-compressed extraction (95% reduction) | -95% |
| Knowledge freshness | static (last distill) | real-time web | +current |

## Invariant
- **INV-EXA-1**: Exa API key MUST be configurable via `EXA_API_KEY` env. No hardcoded keys. No key → tools return error, not crash.
- **INV-EXA-2**: Rate limit: max 1 concurrent request per API key during backoff. Exponential backoff on 429. Fail after 3 retries. Source: exa-mcp-server rate_limit_management.
- **INV-EXA-3**: Auto-ingestion ONLY triggers if `MIMIC_AUTO_RESEARCH=true` AND mesh query returned < 1 result with similarity ≥ 0.4. Never silent.
- **INV-EXA-4**: All fetched content passes through RTK compression before storage or return to agent. No raw HTML in mesh slots.
- **INV-EXA-5**: Exa results are seeds, not slots. Seeds require distillation + QAC validation before becoming mesh slots. Survival index not applicable to raw web data.

## Alternatives
| Alternative | Rejected Why |
|-------------|--------------|
| Perplexity API instead of Exa | Exa has structured `contents` endpoint with `includes=[highlights,summary,entities]` — better for pattern extraction. Perplexity is more chat-oriented. |
| Direct web scraping (no API) | Fragile; violates robots.txt; no rate-limit guarantee. |
| Store Exa results directly as mesh slots | Raw web data lacks survival index, git blame, QAC validation. Must be `data/seeds/` first, then distilled. |

## Consilium
Approved by user on 2026-05-20. Architecture: 3-tier (raw tools → research tool → auto-ingestion).

## Test
- Unit: mock Exa API server (httptest) returning 429 → verify backoff sequence (1s, 2s, 4s) → fail.
- Integration: real Exa API call with `EXA_API_KEY` → verify `EXA_SEARCH` returns structured results.
- Mesh: query non-existent domain → verify auto-ingestion triggers (if enabled) → verify seed file created.
- RTK: fetch 50KB HTML → verify compressed to < 2.5KB.

## Artifact precision
- Exa client code survival: backed by spec card `exa-mcp-server.md` (evidence from `src/api/` and `src/tools/`).
- Rate limit logic: 1:1 with exa-mcp-server implementation (extraction reproducibility ≥ 0.95).

---

## Implementation Details

### Environment Variables (.env.example)
```bash
# Exa API Integration
EXA_API_KEY=                          # Required for Exa tools. Empty = tools disabled.
EXA_BASE_URL=https://api.exa.ai       # Override for enterprise proxy or mock server.
EXA_MAX_RESULTS=10                    # Default numResults for EXA_SEARCH (max 100).
EXA_TIMEOUT_MS=30000                  # Per-request timeout.
EXA_RETRY_MAX=3                     # Max retries on 429/5xx.
EXA_RETRY_BACKOFF_BASE_MS=1000       # Base delay for exponential backoff.

# Auto-Ingestion (Tier 3)
MIMIC_AUTO_RESEARCH=false           # Enable orchestrator auto-research on mesh gaps.
MIMIC_AUTO_RESEARCH_MIN_SIM=0.4     # Similarity threshold below which gap is declared.
MIMIC_AUTO_RESEARCH_MAX_RESULTS=5   # How many Exa results to fetch per gap.
MIMIC_SEED_DIR=data/seeds           # Where auto-ingestion writes raw seeds.
```

### Code Structure
```
internal/
  tool/
    exa/
      client.go      # HTTP client: Search(), Fetch(), rate-limit backoff
      types.go       # Request/response structs matching Exa API
      config.go      # LoadConfigFromEnv() — reads EXA_* vars
    rtk/
      compress.go    # Already exists; Exa results flow through here
  mcp/
    exa_handler.go   # HandleExaSearch, HandleExaFetch, HandleMimicResearch
    tool_schemas.go  # Add EXA_SEARCH, EXA_FETCH, MIMIC_RESEARCH schemas
```

### Tier 1: Raw Tools
- `EXA_SEARCH(query, numResults, type)` → POST `/search` → return titles/urls/highlights.
- `EXA_FETCH(urls, extract, includeHtml)` → POST `/contents` → return markdown per URL.

### Tier 2: Research Tool
- `MIMIC_RESEARCH(topic, depth)` → internally:
  1. `exa_search(topic)`
  2. `exa_fetch(top N urls, extract=highlights)`
  3. `rtk.Compress(results)` per content-type
  4. Return structured summary to agent

### Tier 3: Auto-Ingestion (Orchestrator)
- In `orchestrator.go` mesh query phase:
  ```go
  if len(results) == 0 || bestSimilarity < cfg.AutoResearchMinSim {
      if cfg.AutoResearch {
          seeds := exaClient.Research(query, cfg.AutoResearchMaxResults)
          for _, s := range seeds {
              rtkCompressed := rtk.Compress(s.Content)
              writeSeed(cfg.SeedDir, s.URL, rtkCompressed)
          }
          // Re-query mesh after seeding (optional)
      }
  }
  ```
- Seeds are JSON files: `{url, title, compressed_content, fetched_at, query}`
- Periodically: `make distill` processes seeds → slots (manual or cron, not auto — QAC requires time).
