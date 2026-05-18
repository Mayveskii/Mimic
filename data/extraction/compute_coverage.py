#!/usr/bin/env python3
"""
compute_coverage.py — Domain coverage calculator with gap analysis.

Computes composite coverage score per 09-DISTILLATION-ARTIFACTS.md:
  coverage = (opcodes/46)*0.4 + (repos_distilled/36)*0.3 + (behaviors_implemented/141)*0.3

Outputs: coverage report, gap analysis, priority ordering for closing gaps.

Usage: python3 compute_coverage.py [--manifest <path>] [--behaviors <path>]
"""

import argparse
import json
import sys
from pathlib import Path


OPCODE_DOMAINS = {
    "Memory": {"total": 5, "ops": ["OP_MMAP_ALLOC", "OP_MMAP_FREE", "OP_MMAP_READ", "OP_MMAP_WRITE", "OP_MMAP_SYNC"]},
    "I/O": {"total": 5, "ops": ["OP_IO_READ", "OP_IO_WRITE", "OP_IO_OPEN", "OP_IO_CLOSE", "OP_IO_SEEK"]},
    "Git": {"total": 11, "ops": ["OP_GIT_INIT", "OP_GIT_CLONE", "OP_GIT_FETCH", "OP_GIT_COMMIT", "OP_GIT_PUSH", "OP_GIT_DIFF", "OP_GIT_STATUS", "OP_GIT_CHECKOUT", "OP_GIT_BRANCH", "OP_GIT_MERGE", "OP_GIT_REBASE"]},
    "Build": {"total": 5, "ops": ["OP_BUILD_COMPILE", "OP_BUILD_LINK", "OP_BUILD_TEST", "OP_BUILD_DEPLOY", "OP_BUILD_CLEAN"]},
    "Network": {"total": 6, "ops": ["OP_NET_HTTP_GET", "OP_NET_HTTP_POST", "OP_NET_TCP_CONNECT", "OP_NET_TCP_SEND", "OP_NET_TCP_RECV", "OP_NET_TCP_CLOSE"]},
    "Process": {"total": 4, "ops": ["OP_PROC_SPAWN", "OP_PROC_WAIT", "OP_PROC_KILL", "OP_PROC_SIGNAL"]},
    "Utility": {"total": 6, "ops": ["OP_HASH_SHA256", "OP_HASH_MD5", "OP_COMPRESS_GZIP", "OP_DECOMPRESS_GZIP", "OP_ENCRYPT_AES", "OP_DECRYPT_AES"]},
    "System": {"total": 9, "ops": ["OP_SYS_EXEC", "OP_SYS_ENV_GET", "OP_SYS_ENV_SET", "OP_SYS_FILE_EXISTS", "OP_SYS_DIR_CREATE", "OP_SYS_DIR_REMOVE", "OP_SYS_FILE_COPY", "OP_SYS_FILE_MOVE", "OP_SYS_FILE_DELETE"]},
}

IMPLEMENTED_OPCODES = {
    "Memory": 5,
    "I/O": 0,
    "Git": 11,
    "Build": 0,
    "Network": 0,
    "Process": 0,
    "Utility": 0,
    "System": 2,
}

KNOWLEDGE_DOMAINS = {
    "Distributed Systems": {"min_repos": 3, "repos": ["etcd-io/etcd", "kubernetes/kubernetes", "cockroachdb/cockroach", "istio/istio", "grpc/grpc-go", "linkerd/linkerd2"]},
    "Database/Storage": {"min_repos": 3, "repos": ["postgres/postgres", "redis/redis", "confluentinc/kafka", "mongodb/mongo", "apache/cassandra", "pingcap/tidb"]},
    "Network/Proxy": {"min_repos": 3, "repos": ["nginx/nginx", "envoyproxy/envoy", "cilium/cilium"]},
    "Security/Identity": {"min_repos": 3, "repos": ["hashicorp/vault", "openssl/openssl", "google/tink"]},
    "Observability": {"min_repos": 3, "repos": ["prometheus/prometheus", "jaegertracing/jaeger", "grafana/loki", "grafana/grafana"]},
    "Build/Packaging": {"min_repos": 3, "repos": ["vercel/turbo", "nodejs/node", "npm/cli"]},
    "Runtime/Interpreter": {"min_repos": 3, "repos": ["golang/go", "python/cpython", "tokio-rs/tokio"]},
    "LLM/Inference": {"min_repos": 3, "repos": ["Mayveskii/vllm", "vllm-project/vllm", "ollama/ollama", "huggingface/transformers", "vllm-project/vllm"]},
    "Agent/AI": {"min_repos": 3, "repos": ["microsoft/autogen", "langchain-ai/langchain", "crewai/crewai", "openai/swarm", "agentops/agentops"]},
    "Git/VCS": {"min_repos": 3, "repos": ["git/git", "libgit2/libgit2", "go-git/go-git"]},
    "OS/System": {"min_repos": 3, "repos": ["torvalds/linux", "llvm/llvm-project", "docker/docker", "gdb/gdb"]},
    "Data/Compute": {"min_repos": 3, "repos": ["pandas-dev/pandas", "scipy/scipy"]},
}

BEHAVIOR_SOURCES_TOTAL = 141
BEHAVIOR_SOURCES_IMPLEMENTED = 1

RESEARCH_THRESHOLD = 0.80


def compute_opcode_coverage():
    total = sum(d["total"] for d in OPCODE_DOMAINS.values())
    implemented = sum(IMPLEMENTED_OPCODES.get(name, 0) for name in OPCODE_DOMAINS)
    per_domain = {}
    for name, domain in OPCODE_DOMAINS.items():
        impl = IMPLEMENTED_OPCODES.get(name, 0)
        per_domain[name] = {
            "implemented": impl,
            "total": domain["total"],
            "coverage": impl / domain["total"] if domain["total"] > 0 else 0,
            "ops": domain["ops"],
        }
    return {
        "implemented": implemented,
        "total": total,
        "coverage": implemented / total,
        "per_domain": per_domain,
    }


def compute_distillation_coverage(manifest_path=None):
    if manifest_path and Path(manifest_path).exists():
        try:
            import yaml
            with open(manifest_path) as f:
                manifest = yaml.safe_load(f)
            distilled = 0
            total = 0
            for category in ["current", "new", "recommended", "agentic", "git", "network", "python", "llm", "swe", "os", "security", "hardware"]:
                for repo in manifest.get(category, []):
                    total += 1
                    if repo.get("status") == "distilled":
                        distilled += 1
            return {"distilled": distilled, "total": total, "coverage": distilled / total if total > 0 else 0}
        except Exception:
            pass

    return {"distilled": 0, "total": 90, "coverage": 0.0}


def compute_behavior_coverage(behaviors_path=None):
    if behaviors_path and Path(behaviors_path).exists():
        try:
            import yaml
            with open(behaviors_path) as f:
                data = yaml.safe_load(f)
            total = 0
            implemented = 0
            for source in data.get("sources", []):
                for behavior in source.get("behaviors", []):
                    total += 1
                    if behavior.get("status") in ("done", "partial"):
                        implemented += 1
            return {"implemented": implemented, "total": total, "coverage": implemented / total if total > 0 else 0}
        except Exception:
            pass

    return {"implemented": BEHAVIOR_SOURCES_IMPLEMENTED, "total": BEHAVIOR_SOURCES_TOTAL, "coverage": BEHAVIOR_SOURCES_IMPLEMENTED / BEHAVIOR_SOURCES_TOTAL}


def compute_composite(opcode, distill, behavior):
    score = (opcode["coverage"] * 0.4 +
             min(distill["distilled"], 36) / 36 * 0.3 +
             behavior["coverage"] * 0.3)
    return round(score, 4)


def gap_analysis(opcode, distill, behavior):
    gaps = []

    for name, domain in opcode["per_domain"].items():
        if domain["coverage"] < 1.0:
            missing = domain["total"] - domain["implemented"]
            gaps.append({
                "type": "opcode",
                "domain": name,
                "gap": f"{missing} OpCodes not implemented",
                "priority": "P0" if name in {"Build", "I/O", "System"} else "P1" if name in {"Network", "Process"} else "P2",
                "impact": f"Blocks research tasks requiring {name} domain operations",
            })

    for name, domain in KNOWLEDGE_DOMAINS.items():
        distill_count = 0
        if distill["coverage"] > 0:
            distill_count = int(distill["coverage"] * len(domain["repos"]))
        if distill_count < domain["min_repos"]:
            gaps.append({
                "type": "distillation",
                "domain": name,
                "gap": f"{domain['min_repos'] - distill_count} repos need distillation",
                "priority": "P0" if name in {"Git/VCS", "LLM/Inference", "Distributed Systems"} else "P1",
                "impact": f"No patterns available for {name} research tasks",
            })

    if behavior["coverage"] < 0.5:
        gaps.append({
            "type": "behavior",
            "domain": "all",
            "gap": f"{int((1 - behavior['coverage']) * behavior['total'])} behaviors not implemented",
            "priority": "P0",
            "impact": "Mimic cannot select best behavior — falls back to guessing",
        })

    gaps.sort(key=lambda g: {"P0": 0, "P1": 1, "P2": 2}.get(g["priority"], 3))

    return gaps


def main():
    parser = argparse.ArgumentParser(description="Compute domain coverage and gap analysis")
    parser.add_argument("--manifest", default="mimicrya/repos-manifest.yaml")
    parser.add_argument("--behaviors", default="mimicrya/behavior-sources.yaml")
    parser.add_argument("--output", default="-")
    args = parser.parse_args()

    opcode = compute_opcode_coverage()
    distill = compute_distillation_coverage(args.manifest)
    behavior = compute_behavior_coverage(args.behaviors)
    composite = compute_composite(opcode, distill, behavior)
    gaps = gap_analysis(opcode, distill, behavior)

    report = {
        "composite_coverage": composite,
        "research_threshold": RESEARCH_THRESHOLD,
        "gap_to_threshold": round(max(RESEARCH_THRESHOLD - composite, 0), 4),
        "opcode": opcode,
        "distillation": distill,
        "behavior": behavior,
        "gaps": gaps,
        "priority_order": [f"{g['priority']}: {g['domain']} — {g['gap']}" for g in gaps[:10]],
    }

    output = json.dumps(report, indent=2)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)


if __name__ == "__main__":
    main()
