# ADR-005: Text-Native Mesh Slots

## Status
- **Proposed**: 2026-05-20
- **Author**: Agent
- **Decision**: Implement — replaces binary gob with markdown-native slots

## Context

Current mesh slots are binary gob files (`data/mesh/graphs/*.gob`). This has several problems:

1. **Memory**: 1.9GB RSS for 149K slots (63% is gob slot data)
2. **Opacity**: LLM cannot read gob — must decode through `mesh.FormatActionBytes()`
3. **Embedding quality**: Binary gob contains byte offsets and pointers — text embeddings from `invariant` field are better
4. **Cross-domain edges**: No way to represent relationships between slots in gob

The GiT paper shows that universal text representation enables emergent capabilities across tasks. We apply the same principle to mesh slots.

## Decision

Convert all mesh slots from binary gob to **markdown text**. A slot is a self-contained markdown document with structured sections.

## Text-Native Slot Format

```markdown
# Slot: <hex-id>
## Domain: <domain-name>
## Invariant
<natural language invariant — what this pattern guarantees>

## Context
<where this pattern comes from, repo, commit, file>

## Actions
<structured action list — one per line>
- SYS_FILE_READ path=/path/to/file
- FILE_EDIT path=/path old="..." new="..."
- BUILD_COMPILE target=all

## Cross-Domain Links
- [slot-id-1] used_in domain=distributed
- [slot-id-2] similar_to similarity=0.85
- [slot-id-3] composes_with step=2

## Embedding
<base64-encoded int8[384] vector>

## Metadata
- created: 2026-05-20T12:00:00Z
- source_repo: github.com/etcd-io/etcd
- commit: a1b2c3d
- survival_index: 0.87
- usage_count: 42
- last_used: 2026-05-19T10:00:00Z
```

## Format Benefits

| Aspect | Gob (Current) | Text-Native (New) |
|--------|---------------|-------------------|
| Size | 1.2GB (149K slots) | ~200MB raw, ~60MB gzipped |
| Memory | 1.9GB RSS | ~200MB mmap'd |
| LLM Readable | No (binary) | Yes (markdown) |
| Embedding Source | Decoded invariant text | Inline invariant text |
| Cross-domain Edges | Not supported | `Cross-Domain Links` section |
| Version Control | Binary diffs | Text diffs |
| Searchable | Requires gob decoder | grep, ripgrep, FTS5 |

## Implementation Plan

### Phase 1: Text Slot Loader (internal/mesh/text_slot.go)

```go
type TextSlot struct {
    ID       string
    Domain   string
    Invariant string
    Context  string
    Actions  []string
    Links    []SlotLink
    Embed    [384]int8
    Metadata map[string]string
}

type SlotLink struct {
    TargetID string
    Relation string // used_in | similar_to | composes_with
    Weight   float64
}

func LoadTextSlot(path string) (*TextSlot, error) // Parse markdown
func (s *TextSlot) Save(path string) error        // Serialize to markdown
func (s *TextSlot) EmbedInt8() [384]int8          // Return embedding
```

### Phase 2: Migration Script (scripts/migrate_gob_to_text.go)

1. Read all gob files from `data/mesh/graphs/`
2. For each slot:
   - Extract invariant, actions, metadata
   - Compute int8[384] embedding via embed service
   - Write as `data/mesh/text/<domain>/<id>.md`
3. Validate: load text slot → verify embedding matches
4. Backup: move gob to `data/mesh/backup/`

### Phase 3: Registry Update

Update `MeshRegistry`:
- `LoadAllGraphs()` → `LoadAllTextSlots()`
- Keep backward compatibility: if `data/mesh/text/` exists → use text, else → gob
- `Query()` searches text slots (same int8 cosine)
- `GetLinks(slotID)` returns cross-domain edges

### Phase 4: Edge Population

During migration, infer edges from existing slots:
- Same domain → `similar_to`
- Referenced in same commit → `used_in`
- Sequential in same file → `composes_with`

Later, during distillation, explicitly extract edges from commit messages.

## Consequences

### Positive
- 6-20x smaller memory footprint
- LLM can read mesh directly → proactive suggestions possible
- Text diffs in version control
- grep/ripgrep over mesh
- Cross-domain edges enable graph traversal

### Negative
- Slightly slower loading (parsing markdown vs gob decode)
- Embedding must be stored inline (384 bytes per slot = 57MB for 149K)
- Need migration script (one-time cost)

### Risks
- Slot text could be too large if invariant is verbose → set max 4KB per slot
- Base64 embedding adds 512 bytes per slot → acceptable vs gob overhead

## Validation

Target metrics:
- Memory: < 300MB RSS for 149K slots (was 1.9GB)
- Load time: < 5s cold start (was 30s)
- Slot size: median < 2KB
- Query time: no regression (same int8 cosine)

## Relation to Other Phases

- **Phase B (qdrant-primary)**: Text-native slots make qdrant sync easier — just index `## Invariant` text
- **Phase C (generative chains)**: LLM reads slot text → generates plan → validates with C-core → saves new slot as markdown
