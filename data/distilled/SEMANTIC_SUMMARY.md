# Distillation Semantic Summary

**Date:** 2026-05-18T19:28:16.154316 UTC

## What This Means

The distillation pipeline analyzed production code repositories and extracted proven patterns that survived long enough to be considered reliable.

**Total Patterns Distilled:** 13,611

### Quality Breakdown

- **Average Precision:** 85.00% — fraction of pattern lines still present at HEAD (via git blame)
- **Average Survival Index:** 85.00% — how much of the original code survived changes over time
- **Average Z-Density:** 0.3342 — knowledge density per slot (higher = more proven invariants per unit)

### Precision Distribution

| Range | Count | Meaning |
|-------|-------|---------|
| 0.80-0.85 | 0 (0.0%) | Minimum viable — passes quality gate
| 0.85-0.90 | 13,611 (100.0%) | Good — solid provenance
| 0.90-0.95 | 0 (0.0%) | Excellent — high survival confidence
| 0.95-1.00 | 0 (0.0%) | Exceptional — near-perfect provenance

### Domains Covered

- **distributed**: 11,551 patterns
- **general**: 1,952 patterns
- **llm**: 108 patterns

### Languages Covered

- **unknown**: 13,611 patterns

### Why This Matters

Every pattern in this mesh has been **battle-tested** in production. The survival index tells us: *this code change survived because it was correct, not because it was lucky.*

When Mimic suggests a pattern to an AI agent, it's not guessing — it's recommending a solution that has already proven itself in the real world.

---
*Total errors during validation: 0*
