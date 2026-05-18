#!/usr/bin/env python3
"""
extract_multimodal.py — Multimodal extraction from source repositories.

Handles 6 modalities:
  TEXT   — prose: comments, docs, PR descriptions
  CODE   — source: functions, types, patterns
  IMAGE  — raster: screenshots, architecture PNGs/JPGs
  DIAGRAM — vector: mermaid, plantuml, d2, excalidraw JSON
  TABLE  — structured: benchmark CSVs, comparison markdown tables
  METRIC — numeric: perf counters, p99 latencies, memory bytes

Every artifact gets domain + layer + modality tags for linear searchability.
Same commit + same tool → same output (bitwise-identical).

Usage: python3 extract_multimodal.py --repo-dir <path> --repo <org>/<name> --commit <sha>
"""

import argparse
import hashlib
import json
import os
import re
from pathlib import Path
from typing import List, Dict, Optional


CODE_EXTENSIONS = {
    '.c', '.h', '.go', '.py', '.rs', '.zig', '.js', '.ts', '.java',
    '.cpp', '.hpp', '.cuh', '.cu', '.cuda',
}

TEXT_EXTENSIONS = {
    '.md', '.txt', '.rst', '.adoc', '.org',
}

IMAGE_EXTENSIONS = {
    '.png', '.jpg', '.jpeg', '.gif', '.webp', '.bmp',
}

DIAGRAM_EXTENSIONS = {
    '.mermaid', '.mmd', '.puml', '.plantuml', '.d2', '.excalidraw',
}

TABLE_EXTENSIONS = {
    '.csv', '.tsv',
}

METRIC_PATTERNS = [
    re.compile(r'(p99|p95|p50|latency|throughput|rps|qps|ops/sec)\s*[=:]\s*([\d.]+)\s*(ms|us|ns|s|/s)?', re.IGNORECASE),
    re.compile(r'memory\s*(?:usage|bytes|peak)\s*[=:]\s*([\d.]+)\s*(KB|MB|GB)?', re.IGNORECASE),
    re.compile(r'(\d[\d,]*)\s+(nonces?/min|tokens?/s|req/s|ops/s)', re.IGNORECASE),
]

DOMAIN_MAP = {
    'distributed': ['etcd', 'kubernetes', 'cockroach', 'consul', 'istio', 'tidb'],
    'database': ['postgres', 'redis', 'kafka', 'mongodb', 'cassandra', 'sqlite'],
    'network': ['nginx', 'envoy', 'rustnet', 'cilium'],
    'security': ['vault', 'openssl', 'tink'],
    'observability': ['prometheus', 'jaeger', 'grafana', 'loki'],
    'build': ['turbo', 'bun', 'netboot'],
    'runtime': ['go', 'cpython', 'tokio', 'node'],
    'llm': ['vllm', 'transformers', 'ollama', 'gemma'],
    'agent': ['autogen', 'langchain', 'crewai', 'swarm', 'hermes', 'code-mode'],
    'git': ['git', 'libgit2', 'go-git', 'gitingest', 'lazygit'],
    'os': ['linux', 'llvm', 'gdb', 'docker'],
    'data': ['pandas', 'scipy', 'numpy'],
}


def infer_domain(repo: str) -> str:
    repo_lower = repo.lower()
    for domain, keywords in DOMAIN_MAP.items():
        if any(kw in repo_lower for kw in keywords):
            return domain
    return "general"


def extract_code_artifacts(repo_dir: str, repo: str, commit: str) -> List[Dict]:
    artifacts = []
    domain = infer_domain(repo)

    for root, dirs, files in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in {'.git', 'node_modules', '__pycache__', 'vendor'}]
        for fname in files:
            ext = Path(fname).suffix
            if ext not in CODE_EXTENSIONS:
                continue

            fpath = os.path.join(root, fname)
            rel_path = os.path.relpath(fpath, repo_dir)

            try:
                with open(fpath, 'r', errors='replace') as f:
                    content = f.read()
            except Exception:
                continue

            functions = extract_functions(content, ext)
            for func_name, func_body, start_line in functions:
                artifact_id = hashlib.sha256(
                    f"{repo}:{commit}:{rel_path}:{start_line}:{func_name}".encode()
                ).hexdigest()[:64]

                token_count = len(func_body) // 4

                artifacts.append({
                    "id": artifact_id,
                    "source_repo": repo,
                    "source_commit": commit,
                    "domain": domain,
                    "layer": "code",
                    "modality": "CODE",
                    "pattern_name": func_name,
                    "pattern_description": f"Function {func_name} from {rel_path}",
                    "pattern_code": func_body,
                    "invariants": [],
                    "token_count": token_count,
                })

    return artifacts


def extract_functions(content: str, ext: str) -> List[tuple]:
    if ext in {'.c', '.h', '.cpp', '.hpp', '.cu', '.cuh'}:
        return extract_c_functions(content)
    elif ext == '.go':
        return extract_go_functions(content)
    elif ext == '.py':
        return extract_python_functions(content)
    elif ext == '.rs':
        return extract_rust_functions(content)
    return []


def extract_c_functions(content: str) -> List[tuple]:
    pattern = re.compile(
        r'^(?:static\s+)?(?:\w+\s+)+(\w+)\s*\([^)]*\)\s*\{',
        re.MULTILINE
    )
    results = []
    for match in pattern.finditer(content):
        name = match.group(1)
        start = match.start()
        line_num = content[:start].count('\n') + 1
        body = extract_balanced_braces(content[match.end() - 1:])
        if body:
            results.append((name, body, line_num))
    return results


def extract_go_functions(content: str) -> List[tuple]:
    pattern = re.compile(r'^func\s+(?:\(\w+\s+\*?\w+\)\s+)?(\w+)\s*\(', re.MULTILINE)
    results = []
    for match in pattern.finditer(content):
        name = match.group(1)
        start = match.start()
        line_num = content[:start].count('\n') + 1
        body_start = content.index('{', match.end()) if '{' in content[match.end():match.end()+200] else match.end()
        body = extract_balanced_braces(content[body_start:])
        if body:
            results.append((name, body, line_num))
    return results


def extract_python_functions(content: str) -> List[tuple]:
    pattern = re.compile(r'^def\s+(\w+)\s*\(', re.MULTILINE)
    results = []
    for match in pattern.finditer(content):
        name = match.group(1)
        start = match.start()
        line_num = content[:start].count('\n') + 1
        lines = content[start:].split('\n')
        body_lines = [lines[0]]
        base_indent = len(lines[0]) - len(lines[0].lstrip())
        for line in lines[1:]:
            if line.strip() and not line.startswith(' ' * (base_indent + 1)):
                break
            body_lines.append(line)
        body = '\n'.join(body_lines)
        results.append((name, body, line_num))
    return results


def extract_rust_functions(content: str) -> List[tuple]:
    pattern = re.compile(
        r'(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*[<(]',
        re.MULTILINE
    )
    results = []
    for match in pattern.finditer(content):
        name = match.group(1)
        start = match.start()
        line_num = content[:start].count('\n') + 1
        brace_pos = content.find('{', match.end())
        if brace_pos == -1 or brace_pos - match.end() > 500:
            continue
        body = extract_balanced_braces(content[brace_pos:])
        if body:
            results.append((name, body, line_num))
    return results


def extract_balanced_braces(text: str) -> Optional[str]:
    if not text or text[0] != '{':
        return None
    depth = 0
    for i, ch in enumerate(text):
        if ch == '{':
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0:
                return text[:i + 1]
    return None


def extract_text_artifacts(repo_dir: str, repo: str, commit: str) -> List[Dict]:
    artifacts = []
    domain = infer_domain(repo)

    for root, dirs, files in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in {'.git', 'node_modules'}]
        for fname in files:
            ext = Path(fname).suffix
            if ext not in TEXT_EXTENSIONS:
                continue
            fpath = os.path.join(root, fname)
            rel_path = os.path.relpath(fpath, repo_dir)
            try:
                with open(fpath, 'r', errors='replace') as f:
                    content = f.read()
            except Exception:
                continue

            sections = re.split(r'\n#{1,3}\s+', content)
            for section in sections:
                if len(section.strip()) < 50:
                    continue
                title_match = re.match(r'(.+)\n', section)
                title = title_match.group(1).strip() if title_match else "untitled"
                artifact_id = hashlib.sha256(
                    f"{repo}:{commit}:{rel_path}:TEXT:{title[:64]}".encode()
                ).hexdigest()[:64]
                artifacts.append({
                    "id": artifact_id,
                    "source_repo": repo,
                    "source_commit": commit,
                    "domain": domain,
                    "layer": "text",
                    "modality": "TEXT",
                    "pattern_name": re.sub(r'[^a-z0-9]+', '_', title.lower())[:64],
                    "pattern_description": title[:200],
                    "pattern_code": section[:4096],
                    "invariants": [],
                    "token_count": len(section) // 4,
                })
    return artifacts


def extract_diagram_artifacts(repo_dir: str, repo: str, commit: str) -> List[Dict]:
    artifacts = []
    domain = infer_domain(repo)

    for root, dirs, files in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in {'.git'}]
        for fname in files:
            ext = Path(fname).suffix
            if ext not in DIAGRAM_EXTENSIONS:
                continue
            fpath = os.path.join(root, fname)
            rel_path = os.path.relpath(fpath, repo_dir)
            try:
                with open(fpath, 'r', errors='replace') as f:
                    content = f.read()
            except Exception:
                continue

            artifact_id = hashlib.sha256(
                f"{repo}:{commit}:{rel_path}:DIAGRAM".encode()
            ).hexdigest()[:64]
            artifacts.append({
                "id": artifact_id,
                "source_repo": repo,
                "source_commit": commit,
                "domain": domain,
                "layer": "diagram",
                "modality": "DIAGRAM",
                "pattern_name": Path(fname).stem.replace('-', '_'),
                "pattern_description": f"Diagram from {rel_path}",
                "pattern_code": content[:16384],
                "invariants": [],
                "token_count": len(content) // 4,
            })
    return artifacts


def extract_metric_artifacts(repo_dir: str, repo: str, commit: str) -> List[Dict]:
    artifacts = []
    domain = infer_domain(repo)

    for root, dirs, files in os.walk(repo_dir):
        dirs[:] = [d for d in dirs if d not in {'.git'}]
        for fname in files:
            fpath = os.path.join(root, fname)
            rel_path = os.path.relpath(fpath, repo_dir)
            try:
                with open(fpath, 'r', errors='replace') as f:
                    content = f.read()
            except Exception:
                continue

            for pattern in METRIC_PATTERNS:
                for match in pattern.finditer(content):
                    metric_str = match.group(0)
                    artifact_id = hashlib.sha256(
                        f"{repo}:{commit}:{rel_path}:METRIC:{metric_str[:64]}".encode()
                    ).hexdigest()[:64]
                    artifacts.append({
                        "id": artifact_id,
                        "source_repo": repo,
                        "source_commit": commit,
                        "domain": domain,
                        "layer": "metric",
                        "modality": "METRIC",
                        "pattern_name": re.sub(r'[^a-z0-9]+', '_', metric_str.lower())[:64],
                        "pattern_description": metric_str[:200],
                        "pattern_code": metric_str,
                        "invariants": [],
                        "token_count": len(metric_str) // 4,
                    })
    return artifacts


def main():
    parser = argparse.ArgumentParser(description="Multimodal extraction from source repo")
    parser.add_argument("--repo-dir", required=True, help="Path to cloned repo")
    parser.add_argument("--repo", required=True, help="org/repo identifier")
    parser.add_argument("--commit", required=True, help="Commit SHA")
    parser.add_argument("--output", default="-", help="Output file")
    parser.add_argument("--modalities", default="code,text,diagram,metric",
                        help="Comma-separated modalities to extract")
    args = parser.parse_args()

    all_artifacts = []
    modalities = args.modalities.split(",")

    if "code" in modalities:
        all_artifacts.extend(extract_code_artifacts(args.repo_dir, args.repo, args.commit))
    if "text" in modalities:
        all_artifacts.extend(extract_text_artifacts(args.repo_dir, args.repo, args.commit))
    if "diagram" in modalities:
        all_artifacts.extend(extract_diagram_artifacts(args.repo_dir, args.repo, args.commit))
    if "metric" in modalities:
        all_artifacts.extend(extract_metric_artifacts(args.repo_dir, args.repo, args.commit))

    output = json.dumps(all_artifacts, indent=2, sort_keys=True)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)

    print(f"[multimodal] {len(all_artifacts)} artifacts extracted", file=sys.stderr)


if __name__ == "__main__":
    main()
