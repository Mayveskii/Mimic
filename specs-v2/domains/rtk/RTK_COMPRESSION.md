# RTK Compression Domain

## Purpose
Token-efficient output handling for large tool results. Prevents context window exhaustion when models interact with large files, logs, command output.

## Behaviors Extracted
From Mayveskii/rtk (CLI proxy for token reduction, 49K stars, Rust).

## Implemented Behaviors

### stream_filter_trait → Compress()
Batch-mode output compression per tool call. Content-type aware.

### language_aware_code_filter
- Strip comments (//, /* */)
- Strip function bodies, keep signatures
- Supported: Go, Rust, Python, C/C++

### toml_filter_pipeline (partial)
- strip_ansi: remove color escape codes
- collapse_blank_lines: 3+ newlines → 1
- max_lines: configurable limit
- head_tail: keep first N + last M lines
- smart_truncate: "... X lines omitted ..." indicator

## API

```go
// Detect content type from sample
rtk.DetectContentType(sample string) ContentType

// Compress output
rtk.Compress(output string, contentType ContentType, cfg Config) string

// Estimate token count (~4 chars/token)
rtk.EstimateTokens(s string) int
```

## Configurations

| Mode | MaxLines | MaxChars | StripComments | StripBodies | Use Case |
|------|----------|----------|---------------|-------------|----------|
| Default | 200 | 10000 | false | false | General tools |
| Aggressive | 50 | 5000 | true | true | Large code files |
| HeadTail | 100 | - | - | - | Logs (first 50 + last 50) |

## Performance
- Simple text: ~500 ns/op
- Code with body stripping: ~2-5 μs/op
- Large file (1200 lines → 50 lines): ~10 μs/op

## Integration Points
- MCP server: apply to ALL tool outputs before returning to model
- Orchestrator: budget tracking uses EstimateTokens()
- ContextCompressor: metadata compression (project structure)

## Example

```go
// Large git log output
output := "a1b2c3d feat: something\n... 5000 more lines ..."
compressed := rtk.Compress(output, rtk.ContentLog, rtk.AggressiveConfig())
// Result: first 25 lines + "... 4975 lines omitted ..." + last 15 lines
// Token reduction: 95%+ for large outputs
```

## Invariants
- Output is ALWAYS returned to model (never dropped silently)
- Omission indicators show how much was truncated
- Content type is auto-detected if not specified
- Compression preserves semantic meaning (signatures, errors, timestamps)
