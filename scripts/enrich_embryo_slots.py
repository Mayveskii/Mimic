#!/usr/bin/env python3
"""
enrich_embryo_slots.py

Reads embryo mesh_slots.json, classifies by repo + invariants,
generates TextSlot markdown files organized by domain.
Output: .mimic/slots/<domain>/<slot>.md
"""

import json
import os
import re
import hashlib
from pathlib import Path

SLOTS_JSON = "data/mesh/seeds/mesh_slots.json"
OUTPUT_DIR = ".mimic/slots"

# 1. Repo-based mapping (highest priority)
REPO_DOMAINS = {
    "Mayveskii/etcd": "distributed_systems",
    "Mayveskii/rtk": "rust_patterns",
    "Mayveskii/vllm": "llm_inference",
}

# 2. Invariant-based mapping (lower priority)
INVARIANT_RULES = [
    ("consensus|quorum|raft|distributed state mutation", "distributed_systems"),
    ("network partition|election|leader|heartbeat", "distributed_systems"),
    ("race condition|mutex|sync\.Mutex|atomic|concurrent access", "concurrency"),
    ("http handler|router|middleware|grpc|tcp server", "networking"),
    ("bls|ecdsa|signature|cipher|aes|rsa", "cryptography"),
    ("kubernetes|docker|deploy|pod|helm|infra", "infrastructure"),
    ("panic|unwrap|result<T|E>|error handling|optional", "error_handling"),
    ("unit test|mock|fixture|assert|benchmark|test coverage", "testing"),
    ("attention mask|batch dimension|forward pass|vllm|transformer", "llm_inference"),
    ("cache|pool|buffer|allocator|memory leak|perf", "performance"),
    ("auth|jwt|permission|security|sanitize|injection", "security"),
    ("sql|database|migration|transaction|connection pool", "database"),
]

# 3. Synthetic seed slots for domains with low/no coverage
SEED_SLOTS = {
    "api_design": {
        "invariants": [
            "api version backward compatible",
            "request validation before handler",
            "response schema stable across releases",
        ],
        "context": "REST/gRPC API design patterns from production services",
    },
    "build_systems": {
        "invariants": [
            "build reproducible across hosts",
            "dependency lockfile committed",
            "ci cache invalidates on toolchain change",
        ],
        "context": "Cargo, Bazel, Make build pipeline patterns",
    },
    "data_structures": {
        "invariants": [
            "iterator invalidation handled",
            "zero-copy where possible",
            "complexity documented in comments",
        ],
        "context": "Collections, trees, graphs used in high-performance code",
    },
    "concurrency": {
        "invariants": [
            "lock ordering documented to prevent deadlock",
            "channel buffers sized to prevent goroutine leak",
            "atomic operations aligned for cache coherency",
        ],
        "context": "Mutex, channel, atomic patterns in Go and Rust",
    },
    "networking": {
        "invariants": [
            "connection timeout set to prevent hang",
            "request body closed after read",
            "retry with exponential backoff on 5xx",
        ],
        "context": "HTTP, gRPC, TCP server/client patterns",
    },
    "error_handling": {
        "invariants": [
            "error wrapped with context at each layer",
            "sentinel errors defined as package-level vars",
            "panic only in init and main",
        ],
        "context": "Rust Result/Option and Go error wrapping patterns",
    },
    "testing": {
        "invariants": [
            "table-driven tests for stateless functions",
            "golden files versioned for output snapshots",
            "race detector enabled in ci",
        ],
        "context": "Unit, integration, and property-based testing",
    },
    "performance": {
        "invariants": [
            "benchmark comparison before and after change",
            "allocation profile checked for hot path",
            "bounded concurrency to prevent resource exhaustion",
        ],
        "context": "Profiling, caching, and optimization patterns",
    },
    "security": {
        "invariants": [
            "input validated before processing",
            "secrets never logged or returned in errors",
            "cryptographic random used for tokens",
        ],
        "context": "Auth, authorization, and input sanitization",
    },
    "database": {
        "invariants": [
            "connection pool size bounded by cpu count",
            "transactions short to prevent lock contention",
            "migrations idempotent and reversible",
        ],
        "context": "SQL, connection pooling, and migration patterns",
    },
}


def classify_domain(slot):
    # 1. Repo mapping
    repo = slot.get("provenance", {}).get("repo", "")
    for key, domain in REPO_DOMAINS.items():
        if key in repo:
            return domain

    # 2. Invariant mapping
    text = " ".join(slot.get("pattern", {}).get("invariants", [])).lower()
    for pattern, domain in INVARIANT_RULES:
        if re.search(pattern, text):
            return domain

    # 3. Fallback
    return "general"


def slot_to_markdown(slot, domain):
    sid = slot["slot_id"]
    prov = slot.get("provenance", {})
    metrics = slot.get("metrics", {})
    pattern = slot.get("pattern", {})
    meta = slot.get("metadata", {})

    lines = [
        "---",
        f'id: {sid}',
        f'domain: {domain}',
        f'subdomain: {slot.get("subdomain","") or domain}',
        f'state_hash: {hashlib.sha256(sid.encode()).hexdigest()[:16]}',
        f'source_repo: {prov.get("repo","")}',
        f'source_commit: {prov.get("commit","")}',
        f'survival_index: {metrics.get("survival_index",0)}',
        f'z_density: {metrics.get("z_density",0)}',
        f'artifact_precision: {metrics.get("artifact_precision",0)}',
        f'invariant_coverage: {metrics.get("invariant_coverage",0)}',
        f'polarity: {meta.get("polarity","POSITIVE")}',
        f'distilled_at: {meta.get("distilled_at","")}',
        '---',
        '',
        f'# Slot {sid[:8]} — {domain}',
        '',
        f'**Source:** `{prov.get("repo","")}` @ `{prov.get("commit","")[:8]}`  ',
        f'**Metrics:** SI={metrics.get("survival_index",0)} | Z={metrics.get("z_density",0):.4f} | Precision={metrics.get("artifact_precision",0)}',
        '',
        '## Invariants',
    ]
    for inv in pattern.get("invariants", []):
        lines.append(f'- {inv}')
    lines.append('')
    return '\n'.join(lines)


def write_seed_slots():
    for domain, data in SEED_SLOTS.items():
        sid = hashlib.sha256(domain.encode()).hexdigest()[:32]
        path = Path(OUTPUT_DIR) / domain / f"seed_{domain}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        lines = [
            "---",
            f'id: {sid}',
            f'domain: {domain}',
            f'subdomain: {domain}',
            f'state_hash: {hashlib.sha256(domain.encode()).hexdigest()[:16]}',
            f'source_repo: synthetic/seed',
            f'survival_index: 0.75',
            f'z_density: 0.5',
            f'polarity: POSITIVE',
            '---',
            '',
            f'# Seed Slot: {domain}',
            '',
            f'**Domain:** {domain}  ',
            f'**Type:** Synthetic seed for v16Q playground',
            '',
            '## Invariants',
        ]
        for inv in data["invariants"]:
            lines.append(f'- {inv}')
        lines.append('')
        lines.append('## Context')
        lines.append(data['context'])
        lines.append('')
        path.write_text('\n'.join(lines))


def main():
    print("[enrich] Loading embryo mesh_slots.json...")
    with open(SLOTS_JSON) as f:
        slots = json.load(f)

    print(f"[enrich] Total slots: {len(slots)}")

    domain_counts = {}
    for slot in slots:
        domain = classify_domain(slot)
        domain_counts[domain] = domain_counts.get(domain, 0) + 1
        out_dir = Path(OUTPUT_DIR) / domain
        out_dir.mkdir(parents=True, exist_ok=True)
        md = slot_to_markdown(slot, domain)
        sid = slot["slot_id"]
        (out_dir / f"{sid}.md").write_text(md)

    write_seed_slots()

    print("\n[enrich] Domain distribution:")
    for d, c in sorted(domain_counts.items(), key=lambda x: -x[1]):
        print(f"  {d}: {c}")
    print(f"\n[enrich] Synthetic seeds: {len(SEED_SLOTS)}")
    print(f"[enrich] Total domains: {len(domain_counts) + len(SEED_SLOTS)}")
    print(f"[enrich] Output: {OUTPUT_DIR}/")


if __name__ == "__main__":
    main()
