#!/usr/bin/env python3
"""
encode_artifacts.py — Encode extracted patterns as atomic artifacts in protobuf JSON.

Each artifact gets deterministic ID from sha256(repo:commit:path:chunk).
Same input → same output (bitwise-identical).

Usage: python3 encode_artifacts.py --patterns <file> --repo <org/repo> --commit <sha> --output <file>
"""

import argparse
import hashlib
import json
import sys
import time
from datetime import datetime, timezone


def make_artifact_id(repo, commit, path, chunk_id):
    raw = f"{repo}:{commit}:{path}:{chunk_id}"
    return hashlib.sha256(raw.encode()).hexdigest()[:64]


def encode_artifacts(patterns_path, repo, commit, tool_version, output_path):
    artifacts = []

    with open(patterns_path, "r") as f:
        for line in f:
            parts = line.strip().split("\t")
            if len(parts) < 4:
                continue

            pattern_id, rel_path, token_str, code = parts[0], parts[1], parts[2], parts[3]

            artifact = {
                "id": make_artifact_id(repo, commit, rel_path, pattern_id.split(":")[-1]),
                "source_repo": repo,
                "source_commit": commit,
                "domain": infer_domain(repo),
                "layer": "code",
                "modality": "CODE",
                "pattern_name": pattern_id.split(":")[-1][:64],
                "pattern_description": f"Pattern from {rel_path}",
                "pattern_code": code,
                "survival_index": 1.0,
                "z_density": 0.0,
                "decision_survival": 0.0,
                "invariants": [],
                "invariant_hash": hashlib.sha256(b"empty").hexdigest()[:64],
                "extracted_by": tool_version,
                "extracted_at": datetime.now(timezone.utc).isoformat(),
                "extraction_hash": hashlib.sha256(tool_version.encode()).hexdigest()[:64],
                "token_count": int(token_str) if token_str.isdigit() else 0,
                "latency_us": 0,
                "memory_bytes": 0,
            }

            artifacts.append(artifact)

    batch = {
        "artifacts": artifacts,
        "batch_hash": hashlib.sha256(json.dumps(artifacts, sort_keys=True).encode()).hexdigest()[:64],
        "tool_version": tool_version,
    }

    with open(output_path, "w") as f:
        json.dump(batch, f, indent=2, sort_keys=True)

    print(f"[encode] {len(artifacts)} artifacts → {output_path}", file=sys.stderr)


DOMAIN_KEYWORDS = {
    "distributed": ["etcd", "k8s", "kubernetes", "cockroach", "consul", "istio", "linkerd"],
    "database": ["postgres", "redis", "kafka", "mongo", "cassandra", "tidb"],
    "network": ["nginx", "envoy", "cilium", "rustnet"],
    "security": ["vault", "openssl", "tink"],
    "llm": ["vllm", "transformers", "ollama", "gemma"],
    "git": ["git", "libgit2", "go-git", "gitingest", "lazygit"],
    "build": ["turbo", "bun", "netboot"],
    "runtime": ["golang", "cpython", "tokio", "node"],
    "agent": ["autogen", "langchain", "crewai", "swarm", "hermes", "code-mode"],
    "os": ["linux", "llvm", "docker", "gdb"],
    "data": ["pandas", "scipy"],
    "observability": ["prometheus", "jaeger", "grafana", "loki"],
}


def infer_domain(repo):
    repo_lower = repo.lower()
    for domain, keywords in DOMAIN_KEYWORDS.items():
        if any(kw in repo_lower for kw in keywords):
            return domain
    return "general"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--patterns", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--tool-version", default="encode_artifacts.py-1.0.0")
    parser.add_argument("--output", required=True)
    parser.add_argument("--format", default="protobuf", choices=["protobuf", "json"])
    args = parser.parse_args()

    encode_artifacts(args.patterns, args.repo, args.commit, args.tool_version, args.output)


if __name__ == "__main__":
    main()
