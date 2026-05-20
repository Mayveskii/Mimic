# Mimic Cookbook

> Inspired by [OpenAI Cookbook](https://github.com/openai/openai-cookbook) — recipes for building with LLMs.
>
> **Mimic Cookbook** — recipes for building deterministic AI agents with MCP tools.

## Philosophy

Every recipe contains:
- **Problem:** What you need to do
- **Solution:** Exact Mimic tool calls (copy-paste ready)
- **Before/After:** Metrics proving effectiveness
- **Data:** Real outputs from actual runs (no mocks)

---

## Recipe Index

| # | Recipe | Tier | Status | Tokens Saved |
|---|--------|------|--------|-------------|
| 1 | [File Operations](recipes/01-file-operations.md) | 1 | ✅ Tested | 0% |
| 2 | [Git Workflow](recipes/02-git-workflow.md) | 2 | ✅ Tested | 15% |
| 3 | [Build & Test](recipes/03-build-test.md) | 2 | ✅ Tested | 30% |
| 4 | [Code Analysis](recipes/04-code-analysis.md) | 3 | ✅ Tested | 85% |
| 5 | [Batch Processing](recipes/05-batch-processing.md) | 3 | 🔄 Planned | 95% |
| 6 | [K8s Operations](recipes/06-k8s-operations.md) | 3 | ⏳ Future | - |

---

## Benchmark Results

See [benchmarks/results.md](benchmarks/results.md) for full metrics.

**Quick summary:**
- **Tier 1 (Single tools):** 4/4 PASS, 0 retries, $0.001 per call
- **Tier 2 (Tool chains):** 2/2 PASS with `SYS_FILE_READ`, 1 retry avg
- **Tier 3 (Complex analysis):** 2/2 PASS with compression, 95% token reduction

---

## How to Use This Cookbook

```bash
# Start Mimic MCP server
./bin/mimic serve

# Run any recipe example (copy from recipe page)
mimic_SYS_FILE_READ path="README.md"
```

**For agents:** Each recipe includes the exact JSON-RPC request the agent should send.

---

## Contributing Recipes

1. Test the recipe on real data
2. Measure before/after metrics
3. Submit PR with recipe + benchmark results
4. All recipes must pass 13 QAC checks

---

*Built with determination. Every recipe has a source. Every metric is measured.*
