# Mesh Performance & Scaling — Mimic v0.3

## Current Bottlenecks

### 1. Brute-Force int8 Cosine (Local Mesh)

**Current:** `MeshRegistry.Query()` iterates ALL 149K slots, computes int8 cosine for each.

```
Time complexity: O(N × D) where N=149642, D=384
Actual: ~150ms per query (warm), ~400ms (cold)
Memory: 1.9GB for all slots (go-loaded + centroid routing table)
```

**Problem:** Linear scaling. At 1M slots → ~1s per query, 13GB RAM.

### 2. Centroid Routing — Coarse-Grained

**Current:** 200 centroids, exact search within nearest centroid only.

```
Recall: ~0.85 (15% of relevant slots missed)
Speedup: ~200x vs brute force
Trade-off: speed vs accuracy
```

**Problem:** Single centroid = missing cross-domain patterns. Error handling in Go might be in `go` centroid, but distributed error handling — in `distributed`. Fixed by qdrant fallback.

### 3. Qdrant Fallback — Network Latency

**Current:** Local mesh first → qdrant when < 3 results or sim < 0.35.

```
Local mesh: 50-150ms
Qdrant HNSW: 10-50ms (but + network RTT ~5ms)
Total hybrid: 60-200ms
```

**Problem:** Dual path complexity. Two different embedding formats (int8 local, float32 qdrant).

## Scaling Plan

### Phase 1: Qdrant as Primary (Immediate)

**Action:** Swap priority — qdrant first, local mesh as offline cache.

```go
// New Query() logic:
1. Embed query → float32[384]
2. Qdrant search (HNSW, ef=128) → top 20 in 10-30ms
3. Filter by threshold, format results
4. IF qdrant unavailable: fall back to local mesh
```

**Benefits:**
- Sub-50ms search at 1M+ vectors
- No local RAM growth (qdrant handles it)
- Consistent float32 similarity (no int8 quantization loss)

**Cost:** Qdrant RAM ~2GB for 1M vectors.

### Phase 2: Quantized HNSW in Mimic (Month 2-3)

**Action:** Replace brute-force with go-hnsw or similar.

```
Library: github.com/byron-janrain/go-hnsw (pure Go)
Parameters: M=16, efConstruction=200, efSearch=128
Index size: ~200MB for 1M vectors (vs 1.9GB now)
Build time: ~5 minutes for 1M vectors
Query time: < 5ms
```

**Benefits:**
- No network dependency for search
- Sub-5ms latency
- Fits on edge devices

**Implementation:**
```go
// internal/mesh/hnsw_index.go
type HNSWIndex struct {
    index *hnsw.HnswIndex
    ids   []string // slot IDs parallel to vectors
}

func (idx *HNSWIndex) Search(query []float32, k int) []SearchResult {
    neighbors := idx.index.Search(query, k)
    // map indices to slots
}
```

### Phase 3: Tiered Storage (Month 3-4)

**Action:** Hot/warm/cold tiering based on access frequency.

```
Hot (in-memory HNSW): top 10K most-used slots
  → < 1ms search, 200MB RAM

Warm (local int8 brute): next 100K slots
  → ~50ms search, 1GB RAM

Cold (qdrant): remaining 890K slots
  → ~30ms search, qdrant handles storage
```

**Promotion:** LRU cache + explicit agent bookmarks (`MESH_BOOKMARK` tool).

## Memory Optimization

### Current Memory Breakdown (server, 149K slots)

| Component | Size | % |
|-----------|------|---|
| gob slot data (invariant text, actions) | 1.2GB | 63% |
| int8[384] embeddings | 57MB | 3% |
| Centroid routing table | 77MB | 4% |
| Go runtime overhead | 550MB | 29% |
| **Total** | **1.9GB** | **100%** |

### Text-Native Slots (ADR-005)

If slots become markdown-text instead of gob:

```
Current gob: struct with maps, strings, byte slices → 1.2GB
Text markdown: plain text + int8[384] → ~200MB
Compression gzip: ~60MB

Savings: 6-20x smaller
Bonus: LLM can read mesh directly
```

### mmap для графов

```go
// internal/mesh/mmap_loader.go
func LoadGraphMmap(path string) (*Graph, error) {
    f, _ := os.Open(path)
    data, _ := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_PRIVATE)
    // Parse slots lazily from mmap'd region
}
```

OS pages in data on demand. RSS grows gradually as slots are accessed.

## Benchmark Targets

| Metric | Current | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|---------|
| Query latency | 150ms | 30ms | 5ms | <1ms (hot) |
| Max slots | 149K | 1M | 1M | 10M |
| RAM | 1.9GB | 2GB (+qdrant) | 200MB | 200MB |
| Recall@10 | 0.85 | 0.95 | 0.95 | 0.95 |
| Cold start | 30s | 30s | 5s | 5s |

## Migration Path

```
v0.3 (now):     Hybrid (local brute + qdrant fallback)
v0.4 (week 4):  Qdrant primary, local cache
v0.5 (week 8):  HNSW in-process, qdrant optional
v0.6 (week 12): Tiered storage, 10M+ slots
```
