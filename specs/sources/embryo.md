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

advantages:
  - id: emb_binary_runtime
    what: BinaryRuntime with OpCodes 0x10-0x41, tool loop, session, inference pool, hooks
    evidence: "pkg/do/ — BinaryRuntime implementation, opcode definitions, session management"

  - id: emb_orchestrator_pipeline
    what: Pipeline: state→mesh→DIRECT→classify→exec→flywheel→respond
    evidence: "pkg/orchestrator/ — pipeline implementation with all stages"

  - id: emb_mesh_agent_cascade
    what: MeshAgent + Cascade (3-tier: local→medium→top)
    evidence: "pkg/agent/ — MeshAgent and Cascade implementations"

  - id: emb_hunt_system
    what: Hunt system with 17 files: assess, compress, github, meshscan and more
    evidence: "pkg/hunt/ — 17 hunt module files"

  - id: emb_mesh_graph_slots
    what: Mesh graphs with slots and domains
    evidence: "pkg/mesh/ — mesh.go with graph, slot, domain structures"

  - id: emb_hybrid_rag
    what: 5-signal hybrid RAG + qdrant integration
    evidence: "pkg/rag/ — 5-signal RAG implementation with qdrant client"

  - id: emb_projectmap_sqlite
    what: SQLite-based project navigation
    evidence: "pkg/projectmap/ — SQLite navigation implementation"

  - id: emb_survival_index
    what: git blame → Survival Index calculation
    evidence: "pkg/survival/ — survival index from git blame data"

  - id: emb_mapstore_slots_invariants
    what: Slot storage, store, invariant registry
    evidence: "pkg/mapstore/ — slot, store, invariant_registry implementations"

  - id: emb_mcp_tools
    what: 45+ MCP tools, JSON-RPC, HTTP connection pooling, circuit breaker
    evidence: "cmd/mcp/ — 45+ tool definitions, JSON-RPC handler, connection pooling"

applications:
  - advantage_id: emb_binary_runtime
    implemented_in: core/ops.c, core/ops.h
    mechanism: "Go BinaryRuntime → C-core: OpPacket + OpCodeDef + ops_execute_chain"
    invariant: "C-core behavior matches embryo Go BinaryRuntime for same opcode inputs."
    status: partial (c-core derived from this, Go→C translation done)

  - advantage_id: emb_orchestrator_pipeline
    implemented_in: internal/orchestrator/
    mechanism: "Go pipeline stages → Go orchestrator modules (classify, plan, validate, exec, verify, respond)"
    invariant: "All 6 pipeline stages present. No stage skipped."
    status: planned

  - advantage_id: emb_mesh_agent_cascade
    implemented_in: internal/orchestrator/cascade.go
    mechanism: "3-tier cascade: local (cheap/fast) → medium → top (expensive/thorough)"
    invariant: "Local tier attempted first. Escalation only if local fails."
    status: planned

  - advantage_id: emb_hunt_system
    implemented_in: internal/tool/tools_hunt.go
    mechanism: "Hunt tool: assess(complexity) → compress(context) → search(mesh) → return patterns"
    invariant: "Hunt always returns ranked results with Z-density scores."
    status: planned

  - advantage_id: emb_mesh_graph_slots
    implemented_in: core/bmap/ (libbmap.a)
    mechanism: "Go mesh.go → C bmap: slot_index with domain/layer/state_hash indexing"
    invariant: "libbmap.a functions match embryo mesh.go behavior for slot operations."
    status: partial (libbmap.a exists but no .c sources)

  - advantage_id: emb_hybrid_rag
    implemented_in: internal/orchestrator/rag.go
    mechanism: "5-signal RAG: vector similarity + keyword + domain filter + survival + Z-density"
    invariant: "RAG without survival signal = unverified. Every result carries survival index."
    status: planned

  - advantage_id: emb_projectmap_sqlite
    implemented_in: internal/orchestrator/projectmap.go
    mechanism: "SQLite index of project files/symbols for fast navigation queries"
    invariant: "Index updated after every WRITE operation to workspace."
    status: planned

  - advantage_id: emb_survival_index
    implemented_in: data/extraction/ (distillation scripts)
    mechanism: "git blame → count surviving lines per commit → survival = surviving/total"
    invariant: "survival ≥ 0.7 → slot candidate; < 0.1 → discard"
    status: planned

  - advantage_id: emb_mapstore_slots_invariants
    implemented_in: core/bmap/ (libbmap.a inv_*, si_*)
    mechanism: "Go mapstore → C bmap: inv_create/add/find_similar, si_create/insert/query"
    invariant: "Every slot has at least one invariant. No slot without invariant."
    status: planned

  - advantage_id: emb_mcp_tools
    implemented_in: internal/mcp/, internal/tool/
    mechanism: "Go MCP tools → Go tool registry + JSON-RPC server"
    invariant: "All 45+ embryo MCP tools available as Mimic MCP tools."
    status: planned

control:
  - advantage_id: emb_binary_runtime
    verification: "Comparison test: same opcode input → same output in C-core vs embryo Go"
    update_trigger: "Re-analyze when embryo updates pkg/do/"
    last_verified: never

  - advantage_id: emb_orchestrator_pipeline
    verification: "Integration test: submit task → verify all 6 stages execute"
    update_trigger: "Re-analyze when embryo updates pkg/orchestrator/"
    last_verified: never

  - advantage_id: emb_mesh_agent_cascade
    verification: "Unit test: local tier resolves → verify no escalation; local fails → verify escalation"
    update_trigger: "Re-analyze when embryo updates pkg/agent/"
    last_verified: never

  - advantage_id: emb_hunt_system
    verification: "Integration test: hunt for pattern → verify ranked results with Z-density"
    update_trigger: "Re-analyze when embryo updates pkg/hunt/"
    last_verified: never

  - advantage_id: emb_mesh_graph_slots
    verification: "Unit test: si_insert → si_query_domain → verify correct results"
    update_trigger: "Re-analyze when embryo updates pkg/mesh/"
    last_verified: never

  - advantage_id: emb_hybrid_rag
    verification: "Integration test: query RAG → verify all 5 signals contribute to ranking"
    update_trigger: "Re-analyze when embryo updates pkg/rag/"
    last_verified: never

  - advantage_id: emb_projectmap_sqlite
    verification: "Integration test: index project → query symbol → verify found"
    update_trigger: "Re-analyze when embryo updates pkg/projectmap/"
    last_verified: never

  - advantage_id: emb_survival_index
    verification: "Integration test: git blame on known repo → verify survival calculation matches manual count"
    update_trigger: "Re-analyze when embryo updates pkg/survival/"
    last_verified: never

  - advantage_id: emb_mapstore_slots_invariants
    verification: "Unit test: create slot without invariant → verify rejected"
    update_trigger: "Re-analyze when embryo updates pkg/mapstore/"
    last_verified: never

  - advantage_id: emb_mcp_tools
    verification: "Integration test: call each MCP tool → verify response format matches embryo"
    update_trigger: "Re-analyze when embryo updates cmd/mcp/"
    last_verified: never
```
