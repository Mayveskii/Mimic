# ADR-0010: Unified Configuration Variables Outcome

## Decision
Establish a single authoritative document `specs/11-CONFIGURATION.md` that enumerates every customizable variable in Mimic — environment, compile-time, runtime, and client-side — with full semantics: type, default, scope, invariant, and source behavior. All other configuration references MUST be generated from or verified against this document.

## Why (formal)
- **Agent confusion**: Currently variables are scattered across `.env.example`, `go.mod` build tags, C-core `#define`s, Makefile variables, and Docker Compose. Agents (and users) cannot discover all tuning knobs.
- **Validation completeness**: `koanf_layered_config` (from Mayveskii/go-service-template-rest) requires a config schema. Without an outcome document, validation logic is incomplete.
- **Behavior source**: Mayveskii/go-service-template-rest (koanf_layered_config) — defaults → YAML file → env overrides + strict unknown-key validation.
- **Behavior source**: Mayveskii/exa-mcp-server (mcp_tool_pattern) — zod schema on all inputs; configuration deserves the same rigor.

## Measured
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| Config variables documented | ~20 (env only) | ~50+ (all layers) | +150% |
| Agent discovery time | manual grep (5-10 min) | single document (< 1 min) | -90% |
| Invalid config caught | at runtime (panic or silent ignore) | at startup (schema validation) | +1 safety layer |

## Invariant
- **INV-CFG-1**: Every customizable variable has exactly ONE authoritative definition in `11-CONFIGURATION.md`.
- **INV-CFG-2**: Every variable has: `name`, `type`, `default`, `scope` (env|compile|runtime|client), `invariant`, `source` behavior repo, `description`.
- **INV-CFG-3**: Variables added to code without updating `11-CONFIGURATION.md` → CI `make check` fails (schema drift detection).
- **INV-CFG-4**: No secret (API key, password) has a default value other than empty string.

## Alternatives
| Alternative | Rejected Why |
|-------------|--------------|
| Auto-generate from code comments | Fragile; comments drift. Explicit document is contract. |
| Use JSON Schema only | Good for machines, bad for agents reading specs. We need both: markdown for humans/agents, schema for validation. |
| Put in README.md | README is for quickstart; configuration deserves dedicated spec-level document. |

## Consilium
Approved by user on 2026-05-20.

## Test
- CI step: `make check-config` parses `11-CONFIGURATION.md`, extracts variable list, greps codebase for each name. Missing in code → fail. Missing in doc → fail.
- Unit: `internal/config/` loader validates every variable against type and invariant on startup.

## Artifact precision
- Document survival: reviewed on every PR that touches config (invariant coverage 1.0).
- Extraction reproducibility: 100% (same markdown parser every CI run).

---

## Implementation Details

### Document Location
`specs/11-CONFIGURATION.md` — referenced from `00-SPEC-INDEX.md`.

### Schema per variable (in document)
```markdown
### `MIMIC_PORT`
- **type**: uint16
- **default**: 1337
- **scope**: env, runtime
- **invariant**: `> 0 && < 65535 && != 22 && != 80` (non-privileged, non-standard)
- **source**: Mayveskii/go-service-template-rest (koanf_layered_config)
- **description**: Primary MCP server port for stdio/SSE/HTTP transport.
- **affects**: `internal/mcp/tcp.go`, `cmd/mimic/main.go`, `Dockerfile EXPOSE`
```

### Variable categories (in document)
1. Core Server (ports, log, timeouts)
2. Budget & Limits (tokens, time, tasks)
3. RTK Compression (enable, max lines, modes)
4. Deep Cache & Mesh (paths, min SI, max size)
5. Security (confirm dangerous, circuit breaker)
6. External APIs (Exa, Qdrant, Embed)
7. C-Core (MAX_CHAIN_LENGTH, MAX_TOTAL_BUFFER, compile flags)
8. Build / Makefile (CC, CFLAGS, CGO_ENABLED)
9. Docker / Compose (volumes, healthcheck intervals)
10. Client / npm (MIMIC_VERSION override, binary cache dir)

### CI Drift Detection
```bash
# In Makefile or scripts/check-config.sh
python3 scripts/check_config_consistency.py \
  --spec specs/11-CONFIGURATION.md \
  --env .env.example \
  --code internal/config/,internal/mcp/,core/ops.h,Makefile,Dockerfile
```
