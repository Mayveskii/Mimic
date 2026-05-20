#!/usr/bin/env python3
"""
Integration test: launch mimic serve, call Exa tools via MCP JSON-RPC.

Usage:
    export EXA_API_KEY=...
    python3 test/battlefield/exa_integration_test.py

Metrics collected:
- Latency per tool call (ms)
- Result count
- Token reduction ratio (for MIMIC_RESEARCH deep)
- Error rate
"""

import json
import subprocess
import sys
import time
from pathlib import Path

# Load secrets from local project_context_main
SECRETS_FILE = Path("/home/cisco/mimic/project_context_main/secrets/tokens.env")
MIMIC_BINARY = Path("/home/cisco/mimic/bin/mimic")


def load_secrets():
    """Load EXA_API_KEY from local secrets file."""
    if not SECRETS_FILE.exists():
        return {}
    secrets = {}
    with open(SECRETS_FILE) as f:
        for line in f:
            line = line.strip()
            if line.startswith("#") or not line:
                continue
            if "=" in line:
                k, v = line.split("=", 1)
                secrets[k] = v
    return secrets


def send_jsonrpc(proc, method, params=None, msg_id=1):
    """Send JSON-RPC request to mimic stdio and read response."""
    req = {
        "jsonrpc": "2.0",
        "id": msg_id,
        "method": method,
        "params": params or {}
    }
    req_line = json.dumps(req) + "\n"
    proc.stdin.write(req_line.encode())
    proc.stdin.flush()
    
    # Read response line (with timeout)
    start = time.time()
    while proc.poll() is None:
        try:
            line = proc.stdout.readline().decode()
            if line.strip():
                resp = json.loads(line)
                latency_ms = (time.time() - start) * 1000
                if "result" in resp:
                    return resp["result"], latency_ms, None
                elif "error" in resp:
                    return None, latency_ms, resp["error"]
            if time.time() - start > 30:
                return None, (time.time() - start) * 1000, "Timeout"
        except json.JSONDecodeError:
            continue
    return None, 0, "Process exited"


def result_text(result):
    """Extract text content from MCP tool result."""
    if not result:
        return ""
    content = result.get("content", [])
    texts = []
    for c in content:
        if isinstance(c, dict) and c.get("type") == "text":
            texts.append(c.get("text", ""))
    return "\n".join(texts)


def count_lines(text):
    return len(text.splitlines())


def count_chars(text):
    return len(text)


def main():
    print("=" * 60)
    print("Exa Integration Test — Mimic MCP Tools")
    print("=" * 60)
    
    secrets = load_secrets()
    exa_key = secrets.get("EXA_API_KEY", "")
    if not exa_key:
        print("ERROR: EXA_API_KEY not found in project_context_main/secrets/tokens.env")
        sys.exit(1)
    
    print(f"Loaded EXA_API_KEY (len={len(exa_key)})")
    print(f"Mimic binary: {MIMIC_BINARY}")
    print("")
    
    # Launch mimic serve
    env = {"EXA_API_KEY": exa_key, "PATH": "/usr/local/bin:/usr/bin:/bin"}
    proc = subprocess.Popen(
        [str(MIMIC_BINARY), "serve"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=env
    )
    
    # Give it time to start
    time.sleep(1)
    
    if proc.poll() is not None:
        stderr = proc.stderr.read().decode()
        print(f"ERROR: mimic serve exited early: {stderr}")
        sys.exit(1)
    
    print("mimic serve started (PID={})".format(proc.pid))
    print("")
    
    results = []
    
    # Initialize MCP
    init_params = {
        "protocolVersion": "2024-11-05",
        "clientInfo": {"name": "test", "version": "0.1"}
    }
    result, latency, err = send_jsonrpc(proc, "initialize", init_params, 1)
    print(f"[initialize] latency={latency:.1f}ms result={'OK' if result else 'FAIL'}")
    
    # List tools
    result, latency, err = send_jsonrpc(proc, "tools/list", {}, 2)
    tools = result.get("tools", []) if result else []
    exa_tools = [t for t in tools if t.get("name", "").startswith("EXA") or t.get("name", "") == "MIMIC_RESEARCH"]
    print(f"[tools/list] latency={latency:.1f}ms total_tools={len(tools)} exa_tools={len(exa_tools)}")
    for t in exa_tools:
        print(f"  - {t['name']} ({t.get('group', 'unknown')})")
    print("")
    
    # Test 1: EXA_SEARCH
    print("--- Test 1: EXA_SEARCH ---")
    result, latency, err = send_jsonrpc(proc, "tools/call", {
        "name": "EXA_SEARCH",
        "arguments": {
            "query": "etcd RAFT consensus algorithm Go implementation",
            "numResults": 5,
            "type": "auto"
        }
    }, 3)
    text = result_text(result)
    lines = count_lines(text)
    chars = count_chars(text)
    status = "PASS" if result and not result.get("isError") else "FAIL"
    print(f"  Status: {status} | Latency: {latency:.1f}ms")
    print(f"  Results: {lines} lines, {chars} chars")
    if status == "FAIL" and err:
        print(f"  Error: {err}")
    print("")
    results.append(("EXA_SEARCH", status, latency, lines, chars))
    
    # Test 2: EXA_FETCH
    print("--- Test 2: EXA_FETCH ---")
    result, latency, err = send_jsonrpc(proc, "tools/call", {
        "name": "EXA_FETCH",
        "arguments": {
            "urls": ["https://github.com/etcd-io/etcd"],
            "maxChars": 2000
        }
    }, 4)
    text = result_text(result)
    lines = count_lines(text)
    chars = count_chars(text)
    status = "PASS" if result and not result.get("isError") else "FAIL"
    print(f"  Status: {status} | Latency: {latency:.1f}ms")
    print(f"  Results: {lines} lines, {chars} chars")
    if status == "FAIL" and err:
        print(f"  Error: {err}")
    print("")
    results.append(("EXA_FETCH", status, latency, lines, chars))
    
    # Test 3: MIMIC_RESEARCH (shallow)
    print("--- Test 3: MIMIC_RESEARCH (shallow) ---")
    result, latency, err = send_jsonrpc(proc, "tools/call", {
        "name": "MIMIC_RESEARCH",
        "arguments": {
            "topic": "Go RAFT consensus implementation patterns 2026",
            "depth": "shallow"
        }
    }, 5)
    text = result_text(result)
    lines = count_lines(text)
    chars = count_chars(text)
    status = "PASS" if result and not result.get("isError") else "FAIL"
    print(f"  Status: {status} | Latency: {latency:.1f}ms")
    print(f"  Results: {lines} lines, {chars} chars")
    if status == "FAIL" and err:
        print(f"  Error: {err}")
    print("")
    results.append(("MIMIC_RESEARCH_shallow", status, latency, lines, chars))
    
    # Test 4: MIMIC_RESEARCH (deep)
    print("--- Test 4: MIMIC_RESEARCH (deep) ---")
    result, latency, err = send_jsonrpc(proc, "tools/call", {
        "name": "MIMIC_RESEARCH",
        "arguments": {
            "topic": "Go RAFT consensus implementation patterns 2026",
            "depth": "deep"
        }
    }, 6)
    text = result_text(result)
    lines = count_lines(text)
    chars = count_chars(text)
    status = "PASS" if result and not result.get("isError") else "FAIL"
    print(f"  Status: {status} | Latency: {latency:.1f}ms")
    print(f"  Results: {lines} lines, {chars} chars")
    
    # Estimate compression: raw fetch vs compressed research
    # A typical search result set would be ~5K chars raw HTML
    raw_estimate = 5000 * 5  # 5 search results, ~5KB each
    compression_ratio = raw_estimate / chars if chars > 0 else 0
    print(f"  Estimated compression ratio: {compression_ratio:.1f}x ({100 - 100/compression_ratio:.0f}% reduction)")
    if status == "FAIL" and err:
        print(f"  Error: {err}")
    print("")
    results.append(("MIMIC_RESEARCH_deep", status, latency, lines, chars, compression_ratio))
    
    # Test 5: Exa handler disabled (no key)
    print("--- Test 5: EXA handler disabled simulation ---")
    # This would require restarting mimic without EXA_API_KEY
    # Skip for now — unit test covers this
    print("  SKIPPED (covered by Go unit test)")
    print("")
    
    # Cleanup
    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()
    
    # Summary
    print("=" * 60)
    print("SUMMARY")
    print("=" * 60)
    passes = sum(1 for r in results if r[1] == "PASS")
    fails = len(results) - passes
    total_latency = sum(r[2] for r in results)
    
    print(f"Total: {len(results)} tests | PASS: {passes} | FAIL: {fails}")
    print(f"Total latency: {total_latency:.1f}ms | Avg: {total_latency/len(results):.1f}ms")
    print("")
    
    for name, status, latency, lines, chars, *extra in results:
        print(f"  {name:30} {status:6} {latency:8.1f}ms {lines:4} lines {chars:5} chars", end="")
        if extra:
            print(f"  compression={extra[0]:.1f}x")
        else:
            print()
    
    return 0 if fails == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
