```yaml
repo: Mayveskii/graphify
url: https://github.com/Mayveskii/graphify
language: Python
status: partial
last_sync: "2025-05-17"

description: |
  Fork of safishamsi/graphify (48.7K stars). Code-to-knowledge-graph pipeline using tree-sitter
  AST extraction for 30+ languages, Leiden community detection, IDF-weighted search, BFS/DFS
  traversal with hub throttling, full MCP stdio server with 10 tools + 6 resources, SSRF security
  hardening, three-pass MinHash deduplication, and incremental graph merge.

advantages:
  - id: gf_ast_extraction
    what: 30+ language tree-sitter extractors — two-pass (structural nodes + call-graph edges) with EXTRACTED/INFERRED/AMBIGUOUS confidence labels
    evidence: "extract.py — _DISPATCH dict, extract_python(), extract_rust(), etc.; add_node(), add_edge() with seen_ids dedup"

  - id: gf_idf_search
    what: IDF-weighted keyword search + three-tier precedence (exact 1000x > prefix 100x > substring 1x) + source-file bonus + gap-ratio seed cutoff + hub-throttled expansion
    evidence: "serve.py — _compute_idf(), _score_nodes(), _pick_seeds() with p99 hub threshold"

  - id: gf_hub_throttled_traversal
    what: BFS/DFS bounded traversal with p99 degree hub threshold — hubs skipped as transit unless they are seeds
    evidence: "serve.py — _bfs(), _dfs() with hub_threshold from degree distribution"

  - id: gf_subgraph_render
    what: Render subgraph as structured text within token_budget — seeds first, sort by degree, truncate at budget*3 chars with omitted count
    evidence: "serve.py — _subgraph_to_text()"

  - id: gf_three_pass_dedup
    what: Exact normalization → MinHash/LSH (threshold 0.7) + Jaro-Winkler (92% merge, community boost +5) → LLM tiebreaker — Union-Find transitive merge
    evidence: "dedup.py — _norm(), _make_minhash(), _UF with path compression"

  - id: gf_leiden_community
    what: Leiden algorithm (graspologic) with Louvain fallback, deterministic seed=42, oversized split (>25% of graph), low-cohesion re-split (<0.05)
    evidence: "cluster.py — cluster(), _cohesion_score()"

  - id: gf_mcp_server
    what: Full MCP stdio server — 10 tools (query_graph, get_node, get_neighbors, get_community, god_nodes, graph_stats, shortest_path, list_prs, get_pr_impact, triage_prs) + 6 resources + mtime hot-reload
    evidence: "serve.py — serve(), @server.list_tools, @server.call_tool, @server.list_resources, @server.read_resource"

  - id: gf_ssrf_protection
    what: Private-IP blocking (loopback, RFC 1918, CGN, link-local), cloud metadata endpoint blocking, DNS rebinding TOCTOU guard, streaming fetch with 50MB/10MB caps
    evidence: "security.py — validate_url(), _ssrf_guarded_socket(), _NoFileRedirectHandler"

  - id: gf_graph_diff
    what: Structural diff between two graph snapshots — added/removed nodes and edges for tracking project evolution
    evidence: "analyze.py — graph_diff()"

  - id: gf_incremental_merge
    what: prefix_graph_for_global (repo_tag:: prefix for cross-project graphs) + build_merge (append-only, never shrinks unless prune_sources)
    evidence: "build.py — prefix_graph_for_global(), build_merge()"

applications:
  - advantage_id: gf_ast_extraction
    implemented_in: internal/graphify/extract.go
    mechanism: "Call tree-sitter via Python bridge → JSON nodes/edges → load into Go graph struct"
    invariant: "EXTRACTED edges always trusted over INFERRED. AMBIGUOUS edges flagged for review."
    status: planned

  - advantage_id: gf_idf_search
    implemented_in: core/idf_search.c
    mechanism: "Compute IDF per term → score nodes → pick seeds with gap-ratio cutoff → expand via BFS/DFS"
    invariant: "Exact match always ranks above prefix which always ranks above substring. Gap-ratio < 20% stops seed selection."
    status: planned

  - advantage_id: gf_hub_throttled_traversal
    implemented_in: core/traverse.c
    mechanism: "BFS/DFS with hub_threshold = max(p99_degree, 50) — skip hub nodes as transit unless seed"
    invariant: "Hub nodes never expanded as transit. Transit nodes have degree < hub_threshold."
    status: planned

  - advantage_id: gf_subgraph_render
    implemented_in: core/render.c
    mechanism: "Sort: seeds first, then degree descending → render nodes + edges → truncate at budget*3 chars"
    invariant: "Seed nodes always appear at top of output. Truncation reports exact omitted count."
    status: planned

  - advantage_id: gf_three_pass_dedup
    implemented_in: core/dedup.c
    mechanism: "Pass 1: exact norm → Pass 2: MinHash + JW (92% threshold) → Pass 3: LLM tiebreaker → Union-Find merge"
    invariant: "Short labels (<12 chars) never fuzzy-merged unless same-length single-char substitution. Variant patterns (M1 vs M1 Pro) blocked."
    status: planned

  - advantage_id: gf_leiden_community
    implemented_in: internal/graphify/cluster.go
    mechanism: "Call graspologic (Python) or Go Leiden lib → split oversized communities → re-split low-cohesion"
    invariant: "Deterministic with seed=42. Communities never exceed 25% of graph. Cohesion < 0.05 triggers re-split."
    status: planned

  - advantage_id: gf_mcp_server
    implemented_in: internal/mcp/server.go
    mechanism: "Go MCP server (mark3labs/mcp-go) — register 10 tools + 6 resources → stdio transport → dispatch"
    invariant: "JSON-RPC 2.0 compliant. Every tool call creates an OTel span. Hot-reload via fsnotify."
    status: planned

  - advantage_id: gf_ssrf_protection
    implemented_in: internal/graphify/security.go
    mechanism: "validate_url() → scheme whitelist → DNS resolve → private IP block → metadata endpoint block → streaming fetch with caps"
    invariant: "No request to private IP ranges. No request to 169.254.169.254. Binary cap 50MB, text cap 10MB."
    status: planned

  - advantage_id: gf_graph_diff
    implemented_in: internal/graphify/diff.go
    mechanism: "Load old + new graph → compute set difference on node IDs and edge pairs → report added/removed"
    invariant: "Diff is symmetric: diff(A,B) shows what A has that B doesn't and vice versa."
    status: planned

  - advantage_id: gf_incremental_merge
    implemented_in: internal/graphify/merge.go
    mechanism: "Prefix all node IDs with repo_tag:: → union nodes + edges from new graph into existing → never remove unless prune"
    invariant: "Merge is append-only. Prefix ensures no cross-project ID collisions."
    status: planned

control:
  - advantage_id: gf_ast_extraction
    verification: "Integration test: extract Python file → verify all classes, functions, imports appear as nodes"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_idf_search
    verification: "Unit test: index 100 nodes → search for rare term → verify ranks above common term"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_hub_throttled_traversal
    verification: "Unit test: graph with 1 hub (100 edges) → verify hub not expanded as transit"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_subgraph_render
    verification: "Unit test: 20-node subgraph with budget=200 tokens → verify truncation + omitted count"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_three_pass_dedup
    verification: "Unit test: 'UserService' + 'user_service' → exact merge; 'AuthService' + 'auth_svc' → no merge"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_leiden_community
    verification: "Integration test: 500-node graph → verify community count and no community > 25%"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_mcp_server
    verification: "Integration test: tools/list → verify 10 tools; tools/call query_graph → verify JSON response"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_ssrf_protection
    verification: "Unit test: http://127.0.0.1 → blocked; http://169.254.169.254 → blocked; https://example.com → allowed"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_graph_diff
    verification: "Unit test: add 3 nodes to graph → diff → verify 3 added, 0 removed"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never

  - advantage_id: gf_incremental_merge
    verification: "Unit test: merge graph A (tag=repo1) + graph B (tag=repo2) → verify no ID collision, all nodes present"
    update_trigger: "Re-analyze when graphify releases new version"
    last_verified: never
```
