```yaml
repo: Mayveskii/rustnet
url: https://github.com/Mayveskii/rustnet
language: Rust
status: partial
last_sync: "2025-05-17"

description: |
  Fork of domcyrus/rustnet (3.3K stars). Per-process network monitor with TUI and deep
  packet inspection for 17 application protocols. Cross-platform (Linux eBPF, macOS, Windows,
  FreeBSD) with sandboxing (Landlock, Seatbelt, Job Objects). Lock-free concurrent pipeline.

advantages:
  - id: rn_dpi_detection_chain
    what: Ordered protocol detection: port-gated fast-path → signature-based fallback, 17 application protocols
    evidence: "src/network/dpi/mod.rs — analyze_tcp_packet(), analyze_udp_packet() with ordered protocol chain"

  - id: rn_quic_fragment_reassembly
    what: Multi-strategy data extraction: cache→contiguous→fragment→reconstruction→greedy, 5 fallback levels
    evidence: "src/network/dpi/quic.rs — try_extract_tls_from_reassembler() with 5 strategies, CryptoFrameReassembler"

  - id: rn_connection_state_machine
    what: Protocol-aware state machine (TCP RFC 793 + QUIC) with progressive staleness indicators (white→yellow→red)
    evidence: "src/network/merge.rs — update_tcp_state(); src/network/types.rs — TcpState, QuicConnectionState"

  - id: rn_immutable_after_set
    what: First-writer-wins for identity fields, latest-wins for transient — prevents attribution conflicts
    evidence: "src/network/merge.rs — process_name immutability check with IMMUTABILITY VIOLATION log"

  - id: rn_dpi_info_merge
    what: Per-protocol merge strategy: identity→first-wins, dialog→latest-wins, TLS→prefer complete over partial
    evidence: "src/network/merge.rs — merge_packet_into_connection() with per-field merge rules"

  - id: rn_structured_filter
    what: Vim/fzf-style filter: 12 keyword:value pairs (port:, src:, sni:, process:, state:, etc.) AND-combined with regex
    evidence: "src/filter.rs — ConnectionFilter::parse(), FilterCriteria enum with 12 variants"

  - id: rn_platform_abstraction
    what: ProcessLookup trait with DegradationReason enum — actionable diagnostics not just 'failed'
    evidence: "src/network/platform/mod.rs — ProcessLookup trait, DegradationReason enum; platform/*/process.rs implementations"

  - id: rn_capability_detection
    what: eBPF capability check: CapEff bitmask → CAP_BPF+CAP_PERFMON → fallback CAP_SYS_ADMIN → nosuid detection
    evidence: "src/network/platform/linux/ebpf/loader.rs — EbpfLoader::try_load(), check_capabilities_detailed()"

  - id: rn_concurrent_pipeline
    what: Producer-consumer with lock-free DashMap + crossbeam channels + snapshot isolation for readers
    evidence: "src/main.rs — thread spawning; src/app.rs — DashMap for connections, snapshot creation"

  - id: rn_sandbox_multiplatform
    what: Landlock (Linux) + Seatbelt (macOS) + Job Objects (Windows) — 3-platform sandbox
    evidence: "src/network/platform/linux/sandbox/, macos/sandbox/seatbelt.rs, windows/sandbox/"

applications:
  - advantage_id: rn_dpi_detection_chain
    implemented_in: core/dpi.c
    mechanism: "Ordered protocol list: try each by port-gate → signature match → first match wins, DPI limit per connection"
    invariant: "DPI packet limit per connection (default 10). Port-gated protocols checked first for speed."
    status: planned

  - advantage_id: rn_quic_fragment_reassembly
    implemented_in: core/reassemble.c
    mechanism: "Accumulate fragments in BTreeMap by offset → 5 strategies with increasing desperation → cache result"
    invariant: "Max reassembly buffer 64KB. Each strategy only tried if previous failed. Partial results tagged [PARTIAL]."
    status: planned

  - advantage_id: rn_connection_state_machine
    implemented_in: core/conn_state.c
    mechanism: "TCP: RFC 793 state transitions + retransmit/OOO detection; QUIC: Initial→Handshaking→Connected→Draining→Closed"
    invariant: "No state skip. Staleness indicator: <75% white, 75-90% yellow, >90% red."
    status: planned

  - advantage_id: rn_immutable_after_set
    implemented_in: core/immutable.c
    mechanism: "Identity fields: if set → reject update; Transient fields: always update to latest"
    invariant: "Process name immutable once set. Rate counters always latest. No mixed strategy."
    status: planned

  - advantage_id: rn_dpi_info_merge
    implemented_in: core/merge_strategy.c
    mechanism: "Per-field merge: identity→first-wins, state→latest-wins, TLS SNI→prefer complete, QUIC→accumulate+re-extract"
    invariant: "Merge never drops data — always prefers more complete over less complete."
    status: planned

  - advantage_id: rn_structured_filter
    implemented_in: core/filter.c
    mechanism: "Parse keyword:value pairs → AND-combine criteria → match against connection fields"
    invariant: "Empty filter matches all. Multiple criteria AND-combined. Regex supported per keyword."
    status: planned

  - advantage_id: rn_platform_abstraction
    implemented_in: internal/platform/
    mechanism: "Platform trait with get_process() + get_degradation_reason() per OS implementation"
    invariant: "Every platform returns DegradationReason if detection is impaired. No silent failures."
    status: planned

  - advantage_id: rn_capability_detection
    implemented_in: core/capability.c
    mechanism: "Read /proc/self/status CapEff → check bits 39+38 (modern) or 21 (legacy) → detect nosuid"
    invariant: "Capability denied → actionable reason returned. nosuid mount → specific detection."
    status: planned

  - advantage_id: rn_concurrent_pipeline
    implemented_in: internal/pipeline/
    mechanism: "Capture thread → crossbeam channel → worker threads → DashMap → snapshot for readers"
    invariant: "Readers never block writers. Snapshots are consistent point-in-time views."
    status: planned

  - advantage_id: rn_sandbox_multiplatform
    implemented_in: internal/sandbox/
    mechanism: "Landlock rules (Linux) / Seatbelt profile (macOS) / Job Object (Windows) — restrict filesystem+network"
    invariant: "Sandbox active after capability drop. No filesystem write outside allowed paths."
    status: future

control:
  - advantage_id: rn_dpi_detection_chain
    verification: "Unit test: HTTP on port 80 → verify detected; TLS on port 443 → verify detected; custom port with signature → verify fallback"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_quic_fragment_reassembly
    verification: "Unit test: 3 CRYPTO fragments → verify contiguous extraction; partial SNI → verify [PARTIAL] tag"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_connection_state_machine
    verification: "Unit test: SYN→SYN/ACK→ACK → verify Established; FIN+no response → verify stale yellow"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_immutable_after_set
    verification: "Unit test: set process_name='firefox' → set process_name='chrome' → verify 'firefox' retained"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_dpi_info_merge
    verification: "Unit test: partial SNI → complete SNI → verify complete wins; 2 usernames → verify first wins"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_structured_filter
    verification: "Unit test: 'port:443 sni:google' → verify AND-combined; 'process:/firefox/' → verify regex"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_platform_abstraction
    verification: "Integration test: each platform → verify ProcessLookup works or returns DegradationReason"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_capability_detection
    verification: "Unit test: mock CapEff bitmask → verify correct capability extraction; nosuid path → verify detected"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_concurrent_pipeline
    verification: "Stress test: 10K packets/sec → verify no data loss; snapshot during write → verify consistency"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never

  - advantage_id: rn_sandbox_multiplatform
    verification: "Integration test: sandbox active → verify file write blocked outside allowed path"
    update_trigger: "Re-analyze when rustnet releases new version"
    last_verified: never
```
