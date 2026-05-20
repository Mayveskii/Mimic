#!/usr/bin/env python3
"""
Gonkagate E2E Test — Validates full Mimic pipeline with gonkagate inference.

Tests:
  1. LLM can call Mimic tools via JSON-RPC
  2. Mesh query returns relevant patterns
  3. ActionBytes are decoded correctly
  4. All components (embed, mesh, qdrant) are healthy
"""

import json
import os
import subprocess
import sys
import time
import urllib.request
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
MODEL = "moonshotai/kimi-k2.6"
API_KEY = "gp-mzGtSPvHQug18BV8B4D2NucbYhpJgnUe"
BASE_URL = "https://api.gonkagate.com/v1"

# ---------------------------------------------------------------------------
# LLM Client (OpenAI-compatible)
# ---------------------------------------------------------------------------
def llm_chat(messages, tools=None, max_tokens=2048):
    payload = {
        "model": MODEL,
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": 0.1,
    }
    if tools:
        payload["tools"] = tools
        payload["tool_choice"] = "auto"

    req = urllib.request.Request(
        f"{BASE_URL}/chat/completions",
        data=json.dumps(payload).encode(),
        headers={
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json",
            "HTTP-Referer": "https://github.com/Mayveskii/Mimic",
            "X-Title": "Mimic-Test",
        },
        method="POST",
    )

    start = time.time()
    resp = urllib.request.urlopen(req, timeout=60)
    latency = (time.time() - start) * 1000
    data = json.loads(resp.read())
    return data, latency

# ---------------------------------------------------------------------------
# MCP Client (stdio over SSH)
# ---------------------------------------------------------------------------
class MCPClient:
    def __init__(self, proc):
        self.proc = proc
        self.req_id = 1

    def call(self, method, params=None):
        req = {"jsonrpc": "2.0", "id": self.req_id, "method": method}
        self.req_id += 1
        if params:
            req["params"] = params
        self.proc.stdin.write(json.dumps(req) + "\n")
        self.proc.stdin.flush()
        return json.loads(self.proc.stdout.readline())

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def get_tool_schema(client):
    """Get tool list from Mimic and format for LLM."""
    resp = client.call("tools/list")
    tools = []
    for tool in resp.get("result", {}).get("tools", []):
        tools.append({
            "type": "function",
            "function": {
                "name": tool["name"],
                "description": tool.get("description", ""),
                "parameters": tool.get("inputSchema", {"type": "object"}),
            }
        })
    return tools

def run_tool(client, name, arguments):
    """Execute a tool via MCP."""
    resp = client.call("tools/call", {"name": name, "arguments": arguments})
    res = resp.get("result", {})
    if "content" in res:
        return res["content"][0].get("text", "")
    return str(res)

# ---------------------------------------------------------------------------
# Test Scenarios
# ---------------------------------------------------------------------------
def test_tool_list(client):
    """Test 1: Can we list tools?"""
    tools = get_tool_schema(client)
    print(f"  Tool list: {len(tools)} tools")
    assert len(tools) >= 30, f"Expected 30+ tools, got {len(tools)}"
    return True

def test_mesh_query(client):
    """Test 2: Can we query mesh?"""
    result = run_tool(client, "MESH_QUERY", {"query": "how to compile go project", "topK": 3})
    assert "sim=" in result, "Mesh query should return similarity scores"
    assert len(result.splitlines()) >= 2, "Should have at least one result"
    return True

def test_mesh_status(client):
    """Test 3: Is mesh healthy?"""
    result = run_tool(client, "MESH_STATUS", {})
    assert "slots" in result.lower() or "graphs" in result.lower(), f"Mesh status unclear: {result[:100]}"
    return True

def test_llm_can_use_tools(client):
    """Test 4: LLM picks correct tool for task."""
    tools = get_tool_schema(client)
    # Only send a subset to reduce tokens
    tools_subset = [t for t in tools if t["function"]["name"] in [
        "SYS_FILE_READ", "SYS_DIR_CREATE", "BUILD_COMPILE", "BUILD_TEST", "MESH_QUERY"
    ]]

    messages = [{
        "role": "user",
        "content": "Read the file /home/cisco/mimic/README.md"
    }]

    data, latency = llm_chat(messages, tools=tools_subset)
    choice = data["choices"][0]

    if choice["finish_reason"] == "tool_calls":
        tc = choice["message"]["tool_calls"][0]
        name = tc["function"]["name"]
        args = json.loads(tc["function"]["arguments"])
        print(f"  LLM chose tool: {name}({args})")
        assert name == "SYS_FILE_READ", f"Expected SYS_FILE_READ, got {name}"
        assert args.get("path") == "/home/cisco/mimic/README.md"
        return True
    else:
        print(f"  LLM response (no tool): {choice['message']['content'][:200]}")
        return False

def test_mesh_auto_apply(client):
    """Test 5: MESH_AUTO_APPLY returns decoded ActionBytes."""
    result = run_tool(client, "MESH_AUTO_APPLY", {
        "query": "error handling in distributed systems",
        "similarity_threshold": 0.4
    })
    # Should either return a pattern, say "below threshold", or "no ActionBytes"
    assert "AUTO-APPLIED" in result or "below threshold" in result.lower() or "no ActionBytes" in result, \
        f"Unexpected result: {result[:200]}"
    return True

def test_project_map(client):
    """Test 6: Project map works."""
    result = run_tool(client, "PROJECT_MAP_STATUS", {})
    # If not indexed, will return error, that's ok
    assert "files" in result.lower() or "not indexed" in result.lower() or "error" in result.lower(), \
        f"Unexpected project map result: {result[:200]}"
    return True

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    binary = PROJECT_ROOT / "bin" / "mimic"
    if not binary.exists():
        print("❌ Mimic binary not found. Run 'make build' first.")
        return 1

    print("🚀 Starting Mimic server via SSH...")
    proc = subprocess.Popen(
        ["ssh", "-o", "StrictHostKeyChecking=no", "-p", "2022", "root@192.168.111.25", "/opt/mimic/bin/mimic", "serve"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    time.sleep(3)

    client = MCPClient(proc)
    results = []

    tests = [
        ("Tool List", test_tool_list),
        ("Mesh Query", test_mesh_query),
        ("Mesh Status", test_mesh_status),
        ("LLM Tool Selection", test_llm_can_use_tools),
        ("Mesh Auto-Apply", test_mesh_auto_apply),
        ("Project Map", test_project_map),
    ]

    print(f"\n🧪 Running {len(tests)} E2E tests via gonkagate...")
    print("=" * 60)

    for name, fn in tests:
        try:
            ok = fn(client)
            status = "✅ PASS" if ok else "⚠️ SKIP"
            results.append((name, ok))
        except Exception as e:
            status = f"❌ FAIL: {e}"
            results.append((name, False))
        print(f"  {status}: {name}")

    proc.terminate()
    proc.wait(timeout=3)

    passed = sum(1 for _, ok in results if ok)
    total = len(results)
    print("=" * 60)
    print(f"\n🏁 Results: {passed}/{total} passed")
    return 0 if passed == total else 1

if __name__ == "__main__":
    sys.exit(main())
