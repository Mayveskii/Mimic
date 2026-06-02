# DISTILLATION-ARTIFACTS.md — Mimic

How distilled data becomes maximally precise, reproducible, linearly searchable atomic artifacts. How feedback becomes stored proof. How domain coverage is measured and closed.

---

## 1. Precision Principle

```
precision(extraction) = success(task)
```

The more precisely semantics are extracted from source repositories, the more successfully any model can:
- Find the exact pattern via linear search
- Apply it without error
- Store the result as a reusable atomic artifact

**Invariant**: Every artifact produced by distillation or feedback must be reproducible from the same source commit by any model, yielding bitwise-identical output.

**Source**: BEHAVIOR.md #1 (Principle of Least Action) — minimum action requires maximum precision in available knowledge.

---

## 2. Atomic Artifact Schema

Every distilled item becomes one atomic artifact stored in a bmap cell. An artifact is the smallest unit of knowledge that is:
- Self-contained (no external references needed to evaluate)
- Reproducible (same input → same output)
- Linearly searchable (flat key-value structure, no nested graphs)
- Verifiable (carries its own proof: hash, invariant, survival index)

### Artifact Structure

```protobuf
message Artifact {
  string id = 1;                    // sha256(source_commit:path:line_range:modality)
  string source_repo = 2;           // e.g. "etcd-io/etcd"
  string source_commit = 3;         // SHA of the commit this was extracted from
  string domain = 4;                // e.g. "distributed", "git", "build"
  string layer = 5;                 // e.g. "code", "decision", "diagram", "review"
  Modality modality = 6;            // TEXT, CODE, IMAGE, DIAGRAM, TABLE, METRIC

  string pattern_name = 7;          // snake_case identifier
  string pattern_description = 8;   // one-line factual description
  string pattern_code = 9;          // actual code/pattern (if CODE modality)
  bytes pattern_bytes = 10;         // raw bytes (if IMAGE/DIAGRAM modality)

  float survival_index = 11;       // surviving_lines / total_lines_added
  float z_density = 12;             // Z(slot) from BEHAVIOR.md #5
  float decision_survival = 13;     // for DECISION layer: has this decision been reverted?

  repeated Invariant invariants = 14;  // preconditions from BEHAVIOR.md #10
  string invariant_hash = 15;       // sha256 of serialized invariants

  string extracted_by = 16;         // "distill.sh", "distill_decisions.py", "extract_multimodal.py"
  string extracted_at = 17;         // ISO 8601 timestamp
  string extraction_hash = 18;      // sha256 of extraction tool version + parameters

  int32 token_count = 19;           // estimated tokens for this artifact
  int32 latency_us = 20;            // measured execution latency (for CODE artifacts)
  int32 memory_bytes = 21;          // measured memory usage (for CODE artifacts)

  // Polarity: POSITIVE = this works, NEGATIVE = this failed/was reverted, COUNTER = replaces a NEGATIVE
  Polarity polarity = 22;           // POSITIVE (default), NEGATIVE, COUNTER
  string counter_pattern_id = 23;   // for NEGATIVE: id of POSITIVE artifact that replaces this
  string anti_pattern_id = 24;      // for COUNTER: id of NEGATIVE artifact this replaces
  string failure_evidence = 25;     // for NEGATIVE: source_commit + description of failure
  string qac_violated = 26;         // for NEGATIVE: which QAC(s) this failure violates
}

enum Polarity {
  POSITIVE = 0;   // this works, survived git blame, high decision_survival
  NEGATIVE = 1;   // this failed, was reverted, or decision_survival = 0.0
  COUNTER = 2;    // this replaces a NEGATIVE pattern (positive alternative)
}

enum Modality {
  TEXT = 0;       // prose: comments, documentation, PR descriptions
  CODE = 1;       // source code: functions, types, patterns
  IMAGE = 2;      // raster: screenshots, architecture diagrams
  DIAGRAM = 3;    // vector: mermaid, plantuml, d2, excalidraw JSON
  TABLE = 4;      // structured: benchmark results, comparison tables
  METRIC = 5;     // numeric: performance counters, p99 latencies
}

message Invariant {
  string condition = 1;             // e.g. "survival_index >= 0.7"
  string source = 2;                // e.g. "BEHAVIOR.md #4"
  string verification = 3;          // e.g. "2-vote", "CI", "git blame"
}
```

### Linear Searchability

Every artifact is indexed by flat composite key:

```
key = domain:layer:modality:pattern_name
```

Query examples:
- `si_query_domain_layer("distributed", "code")` → all CODE artifacts in distributed domain
- `si_query_domain_layer("git", "decision")` → all DECISION artifacts in git domain
- `si_query_state_hash(artifact.invariant_hash)` → all artifacts sharing same invariants

No graph traversal required. No vector search required for first-level retrieval. Vector search (binary_rag) is second-level ranking only.

**Invariant**: first-level retrieval = O(1) slot index lookup. No model needs to "understand" embeddings to find an artifact — it needs domain + layer + modality + name.

---

## 3. Domain Coverage for Research Level

### 3.1 OpCode Domain Coverage

| Domain | OpCodes | Implemented | Coverage | Research-Ready |
|--------|---------|-------------|----------|----------------|
| Memory | 5 | 5 | 100% | Yes |
| Git | 11 | 11 | 100% | Yes |
| I/O | 5 | 0 | 0% | No |
| Build | 5 | 0 | 0% | No |
| Network | 6 | 0 | 0% | No |
| Process | 4 | 0 | 0% | No |
| Utility | 6 | 0 | 0% | No |
| System | 9 | 2 | 22% | No |
| **Total** | **46** | **18** | **39.1%** | **No** |

**Research-level threshold**: 100% OpCode implementation (46/46). Without Build, Network, I/O — no model can execute research tasks that require compilation, network access, or file I/O through Mimic.

### 3.2 Knowledge Domain Coverage

12 knowledge domains mapped from repos-manifest.yaml categories:

| Knowledge Domain | Repos in Manifest | Distilled | Distillation Coverage | Behavior Source | Gap |
|-----------------|-------------------|-----------|----------------------|----------------|-----|
| Distributed Systems | 6 | 0 | 0% | embryo (partial) | Need etcd, k8s, cockroach distilled |
| Database/Storage | 5 | 0 | 0% | None | Need postgres, redis, kafka distilled |
| Network/Proxy | 3 | 0 | 0% | rustnet, gh-aw-mcpg | Need nginx, envoy distilled |
| Security/Identity | 4 | 0 | 0% | gh-aw-mcpg (DIFC) | Need vault, openssl distilled |
| Observability | 4 | 0 | 0% | None | Need prometheus, jaeger distilled |
| Build/Packaging | 3 | 0 | 0% | bun, netboot.xyz | Need turbo distilled |
| Runtime/Interpreter | 3 | 0 | 0% | None | Need cpython, go distilled |
| LLM/Inference | 5 | 0 | 0% | None | Need vllm, transformers distilled |
| Agent/AI | 7 | 0 | 0% | hermes, code-mode, agency-agents | Need crewai, swarm distilled |
| Git/VCS | 6 | 0 | 0% | git (partial) | Need git/git, libgit2 distilled |
| OS/System | 4 | 0 | 0% | None | Need linux, llvm distilled |
| Data/Compute | 3 | 0 | 0% | None | Need pandas, scipy distilled |
| **Total** | **53** | **0** | **0%** | — | **All pending** |

**Research-level threshold**: ≥3 repos distilled per knowledge domain, with mean Z-density ≥ 0.5 per domain. Total: ≥36 repos distilled.

### 3.3 Behavior Source Coverage

| Source | Behaviors | Analyzed | Implemented | Gap |
|--------|-----------|----------|-------------|-----|
| embryo | 10 | 2 | 1 (c-core) | 9 planned |
| hermes-agent | 15 | 0 | 0 | 15 planned |
| gastown | 12 | 0 | 0 | 12 planned |
| bun | 7 | 0 | 0 | 7 planned |
| rustnet | 10 | 0 | 0 | 10 planned |
| netboot.xyz | 4 | 0 | 0 | 4 planned |
| exa-mcp-server | 5 | 0 | 0 | 5 planned |
| git | 6 | 0 | 0 | 6 planned |
| gitingest | 6 | 0 | 0 | 6 planned |
| awesome-mcp-servers | 2 | 0 | 0 | 2 reference |
| agency-agents | 6 | 0 | 0 | 6 planned |
| gh-aw-mcpg | 6 | 0 | 0 | 6 planned |
| openmythos | 6 | 0 | 0 | 6 planned |
| caveman | 6 | 0 | 0 | 6 planned |
| opencode-anomalyco- | 6 | 0 | 0 | 6 planned |
| code-mode | 6 | 0 | 0 | 6 planned |
| minbpe | 4 | 0 | 0 | 4 planned |
| rtk | 10 | 0 | 0 | 10 planned |
| graphify | 10 | 0 | 0 | 10 planned |
| go-service-template-rest | 10 | 0 | 0 | 10 planned |
| **vllm** | 4 | 0 | 0 | **4 planned (NEW)** |
| **Total** | **141** | **2** | **1** | **138 planned** |

**Research-level threshold**: all behaviors analyzed (status ≥ partial), ≥50% implemented (status ≥ partial in Mimic code).

### 3.4 Composite Coverage Score

```
coverage = (opcodes_implemented / 46) × 0.4
         + (repos_distilled / 36) × 0.3
         + (behaviors_implemented / 141) × 0.3
```

Current: (18/46)×0.4 + (0/36)×0.3 + (1/141)×0.3 = 0.157 + 0 + 0.002 = **15.9%**

Research-level: ≥ 0.80

---

## 4. Multimodal Distillation

### 4.1 Modalities per Knowledge Domain

| Modality | Sources | Extraction Method | Storage |
|----------|---------|-------------------|---------|
| TEXT | PR descriptions, issues, reviews, docs, comments | distill_decisions.py | artifact.pattern_code (UTF-8) |
| CODE | Source files, functions, types, patterns | distill.sh | artifact.pattern_code (UTF-8) |
| IMAGE | Architecture diagrams, screenshots, flowcharts | extract_multimodal.py | artifact.pattern_bytes (PNG/WEBP) |
| DIAGRAM | mermaid, plantuml, d2, excalidraw JSON | extract_multimodal.py | artifact.pattern_code (source text) |
| TABLE | Benchmark results, comparison tables, CI reports | extract_multimodal.py | artifact.pattern_code (CSV/markdown) |
| METRIC | p99 latency, throughput, memory usage | extract_multimodal.py | artifact fields: latency_us, memory_bytes, token_count |

### 4.2 Multimodal Extraction Pipeline

```
source_repo (git clone)
    ↓
walk files:
    ├── *.c, *.go, *.py, *.rs, *.zig → CODE modality
    ├── *.md, *.txt, *.rst            → TEXT modality
    ├── *.png, *.jpg, *.svg           → IMAGE modality
    ├── *.mermaid, *.puml, *.d2      → DIAGRAM modality
    ├── *.csv, benchmark outputs      → TABLE modality
    └── perf reports, CI logs         → METRIC modality
    ↓
for CODE: git blame → survival_index → extract functions with SI ≥ 0.7
for TEXT: parse PR/issue comments → extract decisions → decision_survival
for IMAGE: OCR + vision model → extract architecture concepts → tag with domain
for DIAGRAM: parse source → extract graph structure → tag with domain
for TABLE: parse headers + values → extract comparison patterns → tag with domain
for METRIC: parse numbers → extract performance patterns → tag with domain
    ↓
encode → artifact (protobuf) → compress (OP_COMPRESS_GZIP) → bmap_write_cell
    ↓
index: si_insert(domain, layer, modality, pattern_name)
```

### 4.3 Model Accessibility

Every artifact is accessible to any model via three paths:

1. **Linear path** (no intelligence needed): domain + layer + modality + pattern_name → exact artifact
2. **Keyword path** (minimal intelligence): pattern_name keywords → si_query_state_hash → filtered set
3. **Semantic path** (full intelligence): natural language query → int8_quantize → batch_cosine_int8 → top-k → re-rank by survival_index + z_density

Path 1 is always available. Path 2 requires keyword extraction. Path 3 requires embedding model. All three paths yield the same artifact for the same query if the index is consistent.

**Invariant**: linear path must always work. Semantic path is optimization, not requirement.

---

## 5. Feedback as Atomic Artifact

### 5.1 The Precision Formula

```
artifact_precision = survival_index × invariant_coverage × extraction_reproducibility
```

Where:
- `survival_index` = surviving_lines / total_lines_added (BEHAVIOR.md #4)
- `invariant_coverage` = verified_invariants / required_invariants (BEHAVIOR.md #10)
- `extraction_reproducibility` = 1.0 if same commit → same artifact (verified by extraction_hash)

**Invariant**: artifact_precision = 1.0 means this artifact is proven, fully verified, and exactly reproducible. Any model can use it without hesitation.

### 5.2 Feedback Artifact

When a model executes a task through Mimic and produces a result, the result becomes a feedback artifact:

```protobuf
message FeedbackArtifact {
  string id = 1;                      // sha256(task_id:model_id:result_hash)
  string task_id = 2;                 // links to the OpPacket chain that produced this
  string model_id = 3;               // which model produced this result
  string source_artifact_id = 4;      // which distilled artifact was used (if any)

  bool success = 5;                   // did the task succeed?
  string result_hash = 6;            // sha256 of result output
  int64 execution_time_ns = 7;       // measured execution time
  int32 tokens_used = 8;             // tokens consumed

  repeated string invariant_checks = 9;  // which invariants were verified
  float artifact_precision = 10;         // from formula above

  string consilium_votes = 11;       // if consilium was used: model_a:yes,model_b:no,model_c:yes
  string consilium_decision = 12;    // final consilium decision

  int64 timestamp = 13;              // when this feedback was created
}
```

### 5.3 Feedback → Deep Cache → mimic-server

```
Model executes task via Mimic
    ↓
Result + metrics → FeedbackArtifact
    ↓
Compress (OP_COMPRESS_GZIP) → bmap_write_cell (local mesh)
    ↓
If artifact_precision ≥ 0.8 AND survival confirmed (6mo git blame):
    → Push to mimic-server deep cache
    → mimic-server stores in shared bmap
    → Other models pull this artifact on next si_query
    ↓
If artifact_precision < 0.8:
    → Local mesh only, not shared
    → Mark for re-extraction when more data available
```

**Invariant**: only artifacts with precision ≥ 0.8 enter the shared deep cache. Lower precision stays local.

---

## 6. Decision Pattern Distillation

### 6.1 Decision Survival Index

For DECISION layer artifacts (from PR comments, reviews):

```
decision_survival = has_not_been_reverted_after / months_since_decision
```

If a PR comment says "rejected CUDA Graphs because of aliasing errors" and 6 months later no PR has re-enabled CUDA Graphs — decision_survival = 1.0 / 0.5 = 2.0 (high confidence).

If a PR was reverted after 2 weeks — decision_survival = 0.0 / 0.03 = 0.0 (zero confidence).

Threshold: decision_survival ≥ 1.0 → candidate for deep cache.

### 6.2 Decision Artifact Fields

Decision artifacts use the same Artifact schema with:
- `layer = "decision"`
- `modality = TEXT`
- `pattern_code = "REJECTED: <alternative> BECAUSE: <reason> MEASURED: <metric>"`
- `decision_survival = <value>`
- `invariants = [precondition for this decision]`

Example (from gonka-ai/vllm#36):
```
pattern_name: "reject_cuda_graphs_for_multi_instance"
pattern_code: "REJECTED: mode=reduce-overhead (CUDA Graphs) BECAUSE: output-tensor aliasing errors in multi-instance setup MEASURED: no throughput gain, aliasing failure"
decision_survival: 1.5
domain: "llm"
layer: "decision"
modality: TEXT
survival_index: 1.0
invariants: [{condition: "multi_instance == true", source: "PR#36", verification: "git blame 6mo"}]
```

---

## 7. Domain Gap Analysis and Priority

### 7.1 OpCode Implementation Priority (by research task dependency)

| Priority | Domain | Reason | Blocks These Research Tasks |
|----------|--------|--------|---------------------------|
| P0 | Build | No compilation = no validation of extracted code patterns | Pattern application, safe_deploy, parallel_build |
| P0 | I/O | No file I/O = no workspace indexing, no config loading | Workspace self-build, any task reading files |
| P0 | System (remaining 7 ops) | No file manipulation = no project scaffolding | Feature creation, refactoring, hotfix |
| P1 | Network | No HTTP = no external API calls, no fetch | Web search, remote distillation, deploy |
| P1 | Process | No subprocess = no build/test orchestration | Build pipeline, CI integration |
| P2 | Utility | Hash/compress/encrypt needed for integrity | Compression integrity, security checks |

### 7.2 Distillation Priority (by knowledge domain criticality)

| Priority | Domain | First 3 Repos to Distill | Expected Z-density | Yield (est. slots) |
|----------|--------|--------------------------|--------------------|--------------------|
| P0 | Git/VCS | git/git, libgit2/libgit2, go-git/go-git | 0.7+ | 200+ |
| P0 | LLM/Inference | Mayveskii/vllm, vllm-project/vllm, ollama/ollama | 0.6+ | 150+ |
| P0 | Distributed | etcd-io/etcd, kubernetes/kubernetes, cockroachdb/cockroach | 0.8+ | 300+ |
| P1 | Database | postgres/postgres, redis/redis, confluentinc/kafka | 0.7+ | 250+ |
| P1 | Security | hashicorp/vault, openssl/openssl, google/tink | 0.6+ | 100+ |
| P1 | Observability | prometheus/prometheus, jaegertracing/jaeger, grafana/loki | 0.5+ | 120+ |
| P2 | Build/Packaging | vercel/turbo, nodejs/node, npm/cli | 0.4+ | 80+ |
| P2 | Runtime | golang/go, python/cpython, tokio-rs/tokio | 0.6+ | 180+ |
| P2 | OS/System | torvalds/linux, llvm/llvm-project, docker/docker | 0.5+ | 150+ |
| P0 | Resource Cleanup Under Lock | gonka (managed_storage), etcd-io/etcd (compaction), cockroachdb/cockroach (GC) | 0.7+ | 20+ |
| P0 | Context Structure Stability | binary-mesh (buildMessages), gonka (buildMessages), envoy (stream state) | 0.6+ | 15+ |
| P0 | Input Validation Before I/O | gonka (participant), bitcoin (coin), envoy (bounds) | 0.7+ | 30+ |
| P1 | Idempotent Close/Cleanup | gonka (UnboundedQueue), etcd (bolt Close), redis (aof rewrite) | 0.6+ | 10+ |
| P1 | Economic Invariant Enforcement | gonka (rewards), bitcoin (CAmount), cosmos-sdk (coins) | 0.8+ | 25+ |

**Estimated total slots for research level**: 1,530+ (from 36 repos)

---

## 8. Reproducibility Protocol

Every distillation run must produce identical output given identical input:

1. Pin source commit SHA in repos-manifest.yaml
2. Pin extraction tool version in extraction_hash
3. Pin all parameters (SI threshold, Z-density threshold, max tokens)
4. Verify: `sha256(artifact_bytes) == artifact.id`

```
make distill-code REPO=etcd-io/etcd COMMIT=abc123
    → output: data/seeds/etcd-io-etcd-abc123.bmap
    → verify: sha256 of bmap matches recorded hash

make distill-decisions REPO=gonka-ai/vllm PR=36
    → output: data/seeds/gonka-ai-vllm-pr36-decisions.bmap
    → verify: sha256 of bmap matches recorded hash

make distill-multimodal REPO=Mayveskii/vllm COMMIT=def456
    → output: data/seeds/Mayveskii-vllm-def456.bmap (includes IMAGE/DIAGRAM/TABLE/METRIC)
    → verify: sha256 of bmap matches recorded hash
```

**Invariant**: same commit + same tool version + same parameters → same bmap. If not, distillation is broken.
