```yaml
repo: Mayveskii/embryo
url: https://github.com/Mayveskii/embryo
language: Go
status: partial
last_sync: "2025-05-17"

description: |
  The repository where Mimic was born. Go implementations of binary runtime,
  mesh graph, RAG, hunt, survival, orchestrator, and 45+ MCP tools.
  The C-core in Mimic is derived from embryo's Go pkg/do/BinaryRuntime.
  Contains the original implementations that inform all Mimic architecture.

advantages:
  - id: emb_binary_runtime
    what: BinaryRuntime with OpCodes 0x10-0x41, tool loop, session management, inference pool, and hook system; execute chain of opcodes against tool registry
    evidence: "pkg/do/binary_runtime.go — BinaryRuntime struct with Execute(), opcode switch; pkg/do/opcodes.go — OpCode constants 0x10-0x41; pkg/do/session.go — session tracking"

  - id: emb_orchestrator_pipeline
    what: 7-stage pipeline: state→mesh→DIRECT→classify→exec→flywheel→respond; each stage is a Go interface with Process() method; chain of responsibility pattern
    evidence: "pkg/orchestrator/pipeline.go — Pipeline struct with stages; pkg/orchestrator/classify.go, exec.go, flywheel.go, respond.go — stage implementations"

  - id: emb_mesh_agent_cascade
    what: MeshAgent + Cascade: 3-tier inference (local→medium→top) with cost-aware escalation; local tier uses cheap model, top tier uses expensive model
    evidence: "pkg/agent/mesh_agent.go — MeshAgent struct; pkg/agent/cascade.go — Cascade struct with tier escalation logic, cost tracking per tier"

  - id: emb_hunt_system
    what: Hunt system with 17 modules: assess(complexity scoring), compress(context), github(search), meshscan(network discovery), and more; modular hunt pipeline
    evidence: "pkg/hunt/ — 17 files: assess.go, compress.go, github.go, meshscan.go, patterns.go, z_density.go, etc."

  - id: emb_mesh_graph_slots
    what: Mesh graphs with slots (named capabilities) and domains (capability groups); slot_index for fast lookup by domain/layer/state_hash
    evidence: "pkg/mesh/mesh.go — Mesh struct with slots; pkg/mesh/slot.go — Slot struct with domain, layer, state_hash; pkg/mesh/slot_index.go — indexing by composite key"

  - id: emb_hybrid_rag
    what: 5-signal hybrid RAG: vector_similarity + keyword_match + domain_filter + survival_index + Z_density; qdrant client for vector search; results ranked by weighted combination
    evidence: "pkg/rag/rag.go — HybridRAG struct with 5 signals; pkg/rag/qdrant_client.go — vector search; pkg/rag/ranking.go — weighted score combination"

  - id: emb_projectmap_sqlite
    what: SQLite-based project navigation: index files, symbols, imports, definitions; FTS5 full-text search for symbol lookup; updated after every WRITE opcode
    evidence: "pkg/projectmap/projectmap.go — ProjectMap struct with SQLite backend; pkg/projectmap/schema.go — FTS5 schema; pkg/projectmap/indexer.go — file/symbol indexing"

  - id: emb_survival_index
    what: git blame → Survival Index: surviving_lines / total_lines_per_commit → score 0-1; commits with survival ≥ 0.7 = slot candidates; < 0.1 = discard
    evidence: "pkg/survival/survival.go — ComputeSurvivalIndex() from git blame output; pkg/survival/blame.go — parse blame output; threshold constants"

  - id: emb_mapstore_slots_invariants
    what: Slot storage with invariant registry: each slot must have ≥1 invariant; inv_create/invariant_add/find_similar for invariant matching; si_create/insert/query for slot indexing
    evidence: "pkg/mapstore/store.go — Store struct; pkg/mapstore/invariant_registry.go — inv_create, invariant_add, find_similar; pkg/mapstore/slot_index.go — si_create, insert, query"

  - id: emb_mcp_tools
    what: 45+ MCP tools registered as JSON-RPC methods; HTTP connection pooling with circuit breaker per backend; tool schema validation
    evidence: "cmd/mcp/server.go — JSON-RPC server; cmd/mcp/tools/ — 45+ tool definitions; cmd/mcp/pool.go — connection pooling; cmd/mcp/circuit.go — circuit breaker"

applications:
  - advantage_id: emb_binary_runtime
    implemented_in: core/ops.c, core/ops.h
    mechanism: "Go BinaryRuntime.Execute(opcode_chain) → C ops_execute_chain(): OpPacket{opcode, args} → switch(opcode) → dispatch to C handler → return OpResult; session tracked in C session struct"
    invariant: "C-core behavior matches embryo Go BinaryRuntime for same opcode inputs. Unknown opcode → OP_UNKNOWN error. Session state consistent across Go↔C boundary."
    status: partial (c-core derived from this, Go→C translation done)

  - advantage_id: emb_orchestrator_pipeline
    implemented_in: internal/orchestrator/pipeline.go
    mechanism: "Go Pipeline struct: stages []Stage → Process(input) → chain stages: state.Process→mesh.Process→direct.Process→classify.Process→exec.Process→flywheel.Process→respond.Process; each stage returns PipelineResult"
    invariant: "All 7 stages present in pipeline. No stage skipped. Failed stage → PipelineResult{Error} → downstream stages receive error context."
    status: planned

  - advantage_id: emb_mesh_agent_cascade
    implemented_in: internal/orchestrator/cascade.go
    mechanism: "Go Cascade: try local_tier.Infer(prompt) → if confidence < threshold → medium_tier.Infer → if still insufficient → top_tier.Infer; cost accumulated per tier"
    invariant: "Local tier attempted first. Escalation only if local confidence < threshold. Total cost = Σ(tier_cost). Top tier always succeeds or returns error."
    status: planned

  - advantage_id: emb_hunt_system
    implemented_in: internal/tool/tools_hunt.go
    mechanism: "Go hunt pipeline: hunt.Assess(complexity) → hunt.Compress(context) → hunt.Search(mesh, patterns) → rank by Z-density → return HuntResult{patterns, z_density_scores}"
    invariant: "Hunt always returns ranked results with Z-density scores. Empty results = valid (no patterns found). Complexity score determines search depth."
    status: planned

  - advantage_id: emb_mesh_graph_slots
    implemented_in: core/bmap/ (libbmap.a)
    mechanism: "Go mesh.go → C bmap: si_create(domain, layer) → si_insert(slot_index, slot) → si_query_domain(domain) → returns matching slots; slot.state_hash for dedup"
    invariant: "libbmap.a functions match embryo mesh.go behavior for slot operations. Same domain+layer query returns same results."
    status: partial (libbmap.a exists but no .c sources)

  - advantage_id: emb_hybrid_rag
    implemented_in: internal/orchestrator/rag.go
    mechanism: "Go HybridRAG: vector_sim(query) → keyword_match(query) → domain_filter() → survival_weight() → z_density_weight() → combine with weights [0.3, 0.2, 0.2, 0.15, 0.15] → rank → return top-K"
    invariant: "RAG without survival signal = unverified. Every result carries survival index. All 5 signals contribute to final ranking score."
    status: planned

  - advantage_id: emb_projectmap_sqlite
    implemented_in: internal/orchestrator/projectmap.go
    mechanism: "Go ProjectMap: SQLite with FTS5 → IndexFile(path, symbols, imports) → QuerySymbol(name) via FTS5 MATCH → UpdateIndex after WRITE opcode execution"
    invariant: "Index updated after every WRITE operation to workspace. Query always reflects latest file state. FTS5 ensures sub-millisecond symbol lookup."
    status: planned

  - advantage_id: emb_survival_index
    implemented_in: data/extraction/ (distillation scripts)
    mechanism: "Go ComputeSurvivalIndex: git blame --porcelain → parse surviving/total lines per commit → survival = surviving/total → score 0-1; batch process for all commits"
    invariant: "survival ≥ 0.7 → slot candidate; < 0.1 → discard. Score is deterministic for same commit at same point in history."
    status: planned

  - advantage_id: emb_mapstore_slots_invariants
    implemented_in: core/bmap/ (libbmap.a inv_*, si_*)
    mechanism: "Go mapstore → C bmap: inv_create() → invariant_add(invariant) → find_similar(invariant, threshold); si_create() → insert(slot) → query(domain, layer); each slot validated for ≥1 invariant"
    invariant: "Every slot has at least one invariant. No slot without invariant. invariant_add rejects duplicates. find_similar uses cosine similarity."
    status: planned

  - advantage_id: emb_mcp_tools
    implemented_in: internal/mcp/, internal/tool/
    mechanism: "Go MCP server: JSON-RPC handler → route method to tool → validate params → execute → return result; connection pool with per-backend circuit breaker; 45+ tools in cmd/mcp/tools/"
    invariant: "All 45+ embryo MCP tools available as Mimic MCP tools. Circuit breaker prevents cascade failure. Tool schema validated before execution."
    status: planned

control:
  - advantage_id: emb_binary_runtime
    verification: "Comparison test: same opcode input → same output in C-core vs embryo Go; unknown opcode → OP_UNKNOWN in both"
    update_trigger: "Re-analyze when embryo updates pkg/do/"
    last_verified: never

  - advantage_id: emb_orchestrator_pipeline
    verification: "Integration test: submit task → verify all 7 stages execute; inject stage failure → verify downstream receives error context"
    update_trigger: "Re-analyze when embryo updates pkg/orchestrator/"
    last_verified: never

  - advantage_id: emb_mesh_agent_cascade
    verification: "Unit test: local tier resolves with high confidence → verify no escalation; low confidence → verify medium tier; medium fails → verify top tier"
    update_trigger: "Re-analyze when embryo updates pkg/agent/"
    last_verified: never

  - advantage_id: emb_hunt_system
    verification: "Integration test: hunt for pattern in indexed mesh → verify ranked results with Z-density; empty mesh → verify empty results (not error)"
    update_trigger: "Re-analyze when embryo updates pkg/hunt/"
    last_verified: never

  - advantage_id: emb_mesh_graph_slots
    verification: "Unit test: si_insert(slot_A) → si_query_domain(same_domain) → verify slot_A returned; different domain → verify empty"
    update_trigger: "Re-analyze when embryo updates pkg/mesh/"
    last_verified: never

  - advantage_id: emb_hybrid_rag
    verification: "Integration test: query RAG → verify all 5 signals contribute to ranking; disable survival signal → verify 'unverified' tag on results"
    update_trigger: "Re-analyze when embryo updates pkg/rag/"
    last_verified: never

  - advantage_id: emb_projectmap_sqlite
    verification: "Integration test: index Go project → query symbol 'BinaryRuntime' → verify found; after WRITE → verify index updated"
    update_trigger: "Re-analyze when embryo updates pkg/projectmap/"
    last_verified: never

  - advantage_id: emb_survival_index
    verification: "Integration test: git blame on known repo → verify survival calculation matches manual count; new commit → verify survival > 0.7"
    update_trigger: "Re-analyze when embryo updates pkg/survival/"
    last_verified: never

  - advantage_id: emb_mapstore_slots_invariants
    verification: "Unit test: create slot without invariant → verify rejected; add 2 invariants → find_similar returns both ranked by similarity"
    update_trigger: "Re-analyze when embryo updates pkg/mapstore/"
    last_verified: never

  - advantage_id: emb_mcp_tools
    verification: "Integration test: call each MCP tool via JSON-RPC → verify response format matches embryo; circuit breaker open → verify fail fast"
    update_trigger: "Re-analyze when embryo updates cmd/mcp/"
    last_verified: never
```
