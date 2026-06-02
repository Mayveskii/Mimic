```yaml
repo: Mayveskii/git
url: https://github.com/Mayveskii/git
language: C
status: partial
last_sync: "2025-05-17"

description: |
  Fork of git/git (55K+ stars). Distributed version control system with content-addressable
  object store, delta compression, hook middleware, and performance tracing. Core architecture
  reveals battle-tested patterns for command dispatch, object management, and data integrity.

advantages:
  - id: git_command_dispatch
    what: Static command dispatch table (cmd_struct[]) maps subcommand strings → handler functions; unknown command → fallback to git-<cmd> PATH lookup
    evidence: "git.c — cmd_struct array with {name, fn, option} entries; run_argv() for external command fallback"

  - id: git_content_addressable_store
    what: SHA-1 content-addressed object database: hash = identity, duplicate content deduplicates automatically, integrity verified on every read
    evidence: "object.c — lookup_object() by SHA-1; sha1_file.c — hash_object_file(), read_object_file() with integrity check"

  - id: git_pack_delta_compression
    what: Packfiles with delta compression: find base object → compute delta instruction stream (copy/insert) → store delta; reverse to reconstruct
    evidence: "packfile.c — unpack_entry() reconstructs from delta; diff-delta.c — create_delta(), get_delta_header_size()"

  - id: git_hook_middleware
    what: Hook pipeline: find_hook(name) → fork+exec hook script → exit code 0=proceed, non-zero=abort; hooks run in defined order per operation
    evidence: "run-command.c — find_hook() scans .git/hooks/; run_hook_ve() forks hook with specified env"

  - id: git_strbuf_dynamic_string
    what: strbuf: heap-allocated string buffer with grow-by-doubling, guaranteed NUL-termination, detach/attach for zero-copy ownership transfer
    evidence: "strbuf.c — strbuf_grow() doubles alloc; strbuf_detach() returns malloc'd buffer; strbuf_attach() takes ownership"

  - id: git_perf_tracing
    what: trace2 hierarchical performance tracing: region enter/leave, data counters, timers; JSON/normal/perf target formats; nested regions for call graph
    evidence: "trace2.c — trace2_region_enter/leave(), trace2_data_intmax(); trace2/tr2_tgt_perf.c — perf format output"

applications:
  - advantage_id: git_command_dispatch
    implemented_in: core/dispatch.c
    mechanism: "Static dispatch table: {name, handler_fn, flags}[] → binary search by name → call handler(argc, argv); unknown → PATH lookup"
    invariant: "Every registered command has unique name. Dispatch is O(log n) via sorted table. Fallback never overrides built-in."
    status: planned

  - advantage_id: git_content_addressable_store
    implemented_in: core/store.c
    mechanism: "SHA-256 hash of content → store object → lookup by hash → verify integrity on read; duplicate hash = no-op"
    invariant: "Same content always produces same hash. Read always verifies stored hash matches computed hash. No silent corruption."
    status: planned

  - advantage_id: git_pack_delta_compression
    implemented_in: core/delta.c
    mechanism: "Find similar base object → compute delta (copy offset/len + insert bytes) → store delta; reconstruct by applying delta stream to base"
    invariant: "Delta always reconstructs to exact original. Base object always stored as full object, never as delta."
    status: future

  - advantage_id: git_hook_middleware
    implemented_in: internal/mcp/middleware.go
    mechanism: "Hook chain: before operation → find_hook → execute scripts in order → 0=proceed, non-zero=abort; after operation → post-hooks"
    invariant: "Hook abort = operation never started. Post-hooks always run if operation completed. Missing hook file = skip (not error)."
    status: planned

  - advantage_id: git_strbuf_dynamic_string
    implemented_in: core/strbuf.c
    mechanism: "strbuf_init → strbuf_grow(doubling) → append operations → strbuf_detach(zero-copy out); guaranteed NUL-terminated"
    invariant: "len < alloc always. NUL at buf[len]. Grow doubles capacity. Detach transfers ownership, caller must free."
    status: planned

  - advantage_id: git_perf_tracing
    implemented_in: internal/quality/trace.go
    mechanism: "Hierarchical regions: enter(name, labels) → nested operations → leave(name, duration); data counters for metrics; output to JSON/perf targets"
    invariant: "Every enter has matching leave. Nested regions form call graph. Duration measured in nanoseconds."
    status: planned

control:
  - advantage_id: git_command_dispatch
    verification: "Unit test: dispatch 'version' → verify handler called; dispatch 'nonexistent' → verify fallback path"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never

  - advantage_id: git_content_addressable_store
    verification: "Unit test: store same content twice → verify single object; corrupt stored object → verify read detects corruption"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never

  - advantage_id: git_pack_delta_compression
    verification: "Unit test: delta-encode object → apply delta → verify byte-identical to original"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never

  - advantage_id: git_hook_middleware
    verification: "Integration test: hook returns 1 → verify operation aborted; missing hook → verify operation proceeds"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never

  - advantage_id: git_strbuf_dynamic_string
    verification: "Unit test: append 10KB to strbuf → verify single realloc; detach → verify buf valid + strbuf reset"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never

  - advantage_id: git_perf_tracing
    verification: "Integration test: nested region enter/leave → verify JSON output with correct nesting and durations"
    update_trigger: "Re-analyze when git releases new version"
    last_verified: never
```
