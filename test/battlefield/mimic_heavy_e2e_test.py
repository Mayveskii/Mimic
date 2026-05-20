#!/usr/bin/env python3
"""
Mimic Heavy E2E Tests — Enterprise-grade scenarios via Gonkagate inference.

Scenarios (tier-3 complexity):
  1. K8s Debug — pod crash loop, OOMKilled diagnosis
  2. Large Refactor — cross-package error handling in Go
  3. Data Flow Architecture — microservices topology design
  4. Production Migration — zero-downtime database migration

Runs against gonkagate API (moonshotai/kimi-k2.6) with budget tracking.

Usage:
    python3 test/battlefield/mimic_heavy_e2e_test.py
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
MODEL = "Qwen/Qwen3-235B-A22B-Instruct-2507-FP8"
API_KEY = "gp-mzGtSPvHQug18BV8B4D2NucbYhpJgnUe"
BASE_URL = "https://api.gonkagate.com/v1"
BUDGET_USD = 2.0
COST_PER_1K_TOKENS = 0.00025  # Qwen3-235B rate

# ---------------------------------------------------------------------------
# LLM Client
# ---------------------------------------------------------------------------
class LLMClient:
    def __init__(self):
        self.total_input = 0
        self.total_output = 0
        self.session_cost = 0.0

    def chat(self, messages, tools=None, max_tokens=4096):
        payload = {
            "model": MODEL,
            "messages": messages,
            "max_tokens": max_tokens,
            "temperature": 0.2,
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
                "X-Title": "Mimic-Heavy-Test",
            },
            method="POST",
        )

        start = time.time()
        resp = urllib.request.urlopen(req, timeout=120)
        latency = (time.time() - start) * 1000
        data = json.loads(resp.read())

        usage = data.get("usage", {})
        self.total_input += usage.get("prompt_tokens", 0)
        self.total_output += usage.get("completion_tokens", 0)
        cost = (usage.get("prompt_tokens", 0) + usage.get("completion_tokens", 0)) * COST_PER_1K_TOKENS / 1000
        self.session_cost += cost

        return data, latency, cost

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

    def tool(self, name, arguments):
        return self.call("tools/call", {"name": name, "arguments": arguments})

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def get_tools(mcp):
    resp = mcp.call("tools/list")
    return resp.get("result", {}).get("tools", [])

def format_tools_for_llm(tools):
    out = []
    for t in tools:
        out.append({
            "type": "function",
            "function": {
                "name": t["name"],
                "description": t.get("description", ""),
                "parameters": t.get("inputSchema", {"type": "object"}),
            }
        })
    return out

def extract_text(result):
    if "content" in result:
        return result["content"][0].get("text", "")
    return str(result)

# ---------------------------------------------------------------------------
# Scenario 1: Kubernetes Debug (OOMKilled pod)
# ---------------------------------------------------------------------------
def scenario_k8s_debug(mcp, llm, tools):
    print("\n🔴 SCENARIO 1: Kubernetes Debug — OOMKilled pod")
    print("-" * 70)

    system_prompt = """You are a DevOps engineer debugging Kubernetes.
You have access to tools: EXA_SEARCH (web research), MESH_QUERY (find patterns), SYS_FILE_READ (read files).
Step by step: diagnose → research → apply fix → verify."""

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "A Kubernetes pod is in CrashLoopBackOff with OOMKilled status. The pod runs a Python app processing large datasets. Diagnose why, suggest memory limits/resources fix, and create a plan."}
    ]

    conversation_cost = 0.0
    tool_calls_made = 0

    for turn in range(5):
        data, latency, cost = llm.chat(messages, tools=tools)
        conversation_cost += cost
        choice = data["choices"][0]

        if choice["finish_reason"] != "tool_calls":
            print(f"  LLM final answer (turn {turn}): {choice['message']['content'][:200]}...")
            break

        # Execute tools
        for tc in choice["message"].get("tool_calls", []):
            name = tc["function"]["name"]
            args = json.loads(tc["function"]["arguments"])
            print(f"  Tool call: {name}({args})")
            result = mcp.tool(name, args)
            text = extract_text(result.get("result", {}))
            print(f"    → {text[:150]}...")
            messages.append({"role": "assistant", "content": None, "tool_calls": [tc]})
            messages.append({"role": "tool", "tool_call_id": tc["id"], "name": name, "content": text[:1000]})
            tool_calls_made += 1

    print(f"  Tool calls: {tool_calls_made} | Cost: ${conversation_cost:.4f}")
    assert tool_calls_made >= 2, "Should make multiple tool calls"
    assert conversation_cost < 0.5, "Should stay under $0.50"
    return True

# ---------------------------------------------------------------------------
# Scenario 2: Large Go Refactor (error handling)
# ---------------------------------------------------------------------------
def scenario_go_refactor(mcp, llm, tools):
    print("\n🟡 SCENARIO 2: Large Go Refactor — error handling across packages")
    print("-" * 70)

    system_prompt = """You are refactoring error handling in a large Go project.
Available: PROJECT_MAP_QUERY_SYMBOL, MESH_QUERY, MESH_AUTO_APPLY, SYS_FILE_READ.
Find all functions that return errors, then apply consistent wrapping pattern."""

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "In /home/cisco/mimic/internal/ find all Go functions that return errors. Then use mesh to find the best error wrapping pattern and suggest changes."}
    ]

    conversation_cost = 0.0
    tool_calls_made = 0

    for turn in range(5):
        data, latency, cost = llm.chat(messages, tools=tools)
        conversation_cost += cost
        choice = data["choices"][0]

        if choice["finish_reason"] != "tool_calls":
            print(f"  LLM final answer (turn {turn}): {choice['message']['content'][:200]}...")
            break

        for tc in choice["message"].get("tool_calls", []):
            name = tc["function"]["name"]
            args = json.loads(tc["function"]["arguments"])
            print(f"  Tool call: {name}({args})")
            result = mcp.tool(name, args)
            text = extract_text(result.get("result", {}))
            print(f"    → {text[:150]}...")
            messages.append({"role": "assistant", "content": None, "tool_calls": [tc]})
            messages.append({"role": "tool", "tool_call_id": tc["id"], "name": name, "content": text[:1000]})
            tool_calls_made += 1

    print(f"  Tool calls: {tool_calls_made} | Cost: ${conversation_cost:.4f}")
    assert tool_calls_made >= 2
    assert conversation_cost < 0.5
    return True

# ---------------------------------------------------------------------------
# Scenario 3: Data Flow Architecture (microservices)
# ---------------------------------------------------------------------------
def scenario_data_flow(mcp, llm, tools):
    print("\n🟢 SCENARIO 3: Data Flow Architecture — microservices topology")
    print("-" * 70)

    system_prompt = """You are designing data flow architecture.
Available: EXA_SEARCH (research patterns), MESH_QUERY (find distributed patterns), PLAN_GENERATE (create validated plan).
Design ingress → gateway → mesh → database flow."""

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "Design data flow for a microservices platform: ingress → API gateway → service mesh (Istio) → PostgreSQL cluster. Use web search and mesh patterns, then generate a validated plan."}
    ]

    conversation_cost = 0.0
    tool_calls_made = 0
    plan_generated = False

    for turn in range(6):
        data, latency, cost = llm.chat(messages, tools=tools)
        conversation_cost += cost
        choice = data["choices"][0]

        if choice["finish_reason"] != "tool_calls":
            print(f"  LLM final answer (turn {turn}): {choice['message']['content'][:200]}...")
            break

        for tc in choice["message"].get("tool_calls", []):
            name = tc["function"]["name"]
            args = json.loads(tc["function"]["arguments"])
            print(f"  Tool call: {name}({args})")
            result = mcp.tool(name, args)
            text = extract_text(result.get("result", {}))
            print(f"    → {text[:200]}...")
            messages.append({"role": "assistant", "content": None, "tool_calls": [tc]})
            messages.append({"role": "tool", "tool_call_id": tc["id"], "name": name, "content": text[:1000]})
            tool_calls_made += 1
            if name == "PLAN_GENERATE":
                plan_generated = True

    print(f"  Tool calls: {tool_calls_made} | Plan generated: {plan_generated} | Cost: ${conversation_cost:.4f}")
    assert tool_calls_made >= 3
    assert plan_generated, "Should generate a plan"
    assert conversation_cost < 0.6
    return True

# ---------------------------------------------------------------------------
# Scenario 4: Production DB Migration (zero downtime)
# ---------------------------------------------------------------------------
def scenario_db_migration(mcp, llm, tools):
    print("\n🔵 SCENARIO 4: Production Migration — zero-downtime PostgreSQL")
    print("-" * 70)

    system_prompt = """You are planning a zero-downtime database migration.
Available: EXA_SEARCH (research), MESH_QUERY (find patterns), PLAN_GENERATE (validated plan).
Design blue-green migration with rollback strategy."""

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "Migrate a 2TB PostgreSQL database to a new cluster with zero downtime. Research best practices via web search, find mesh patterns for database migration, then generate a validated execution plan."}
    ]

    conversation_cost = 0.0
    tool_calls_made = 0
    has_research = False
    has_plan = False

    for turn in range(6):
        if llm.session_cost > BUDGET_USD * 0.5:
            print("  ⚠️  Budget halfway exhausted, stopping")
            break

        data, latency, cost = llm.chat(messages, tools=tools)
        conversation_cost += cost
        choice = data["choices"][0]

        if choice["finish_reason"] != "tool_calls":
            print(f"  LLM final answer (turn {turn}): {choice['message']['content'][:200]}...")
            break

        for tc in choice["message"].get("tool_calls", []):
            name = tc["function"]["name"]
            args = json.loads(tc["function"]["arguments"])
            print(f"  Tool call: {name}({args})")
            result = mcp.tool(name, args)
            text = extract_text(result.get("result", {}))
            print(f"    → {text[:200]}...")
            messages.append({"role": "assistant", "content": None, "tool_calls": [tc]})
            messages.append({"role": "tool", "tool_call_id": tc["id"], "name": name, "content": text[:1000]})
            tool_calls_made += 1
            if name == "EXA_SEARCH":
                has_research = True
            if name == "PLAN_GENERATE":
                has_plan = True

    print(f"  Tool calls: {tool_calls_made} | Research: {has_research} | Plan: {has_plan} | Cost: ${conversation_cost:.4f}")
    assert tool_calls_made >= 2
    assert has_research, "Should research via EXA_SEARCH"
    assert has_plan, "Should generate migration plan"
    assert conversation_cost < 0.6
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

    mcp = MCPClient(proc)
    tools_raw = get_tools(mcp)
    tools = format_tools_for_llm(tools_raw)
    print(f"📋 Loaded {len(tools)} tools from Mimic")

    # Filter relevant tools for heavy scenarios
    relevant = ["EXA_SEARCH", "MESH_QUERY", "MESH_AUTO_APPLY", "MESH_STATUS",
                "PROJECT_MAP_QUERY_SYMBOL", "PROJECT_MAP_INDEX", "PLAN_GENERATE",
                "SYS_FILE_READ", "SYS_DIR_CREATE", "BUILD_TEST"]
    tools_subset = [t for t in tools if any(r in t["function"]["name"] for r in relevant)]
    print(f"🔧 Using {len(tools_subset)} relevant tools")

    llm = LLMClient()
    results = []

    scenarios = [
        ("K8s Debug", scenario_k8s_debug),
        ("Go Refactor", scenario_go_refactor),
        ("Data Flow", scenario_data_flow),
        ("DB Migration", scenario_db_migration),
    ]

    print(f"\n💰 Budget: ${BUDGET_USD} | Running {len(scenarios)} heavy scenarios...")
    print("=" * 70)

    for name, fn in scenarios:
        try:
            ok = fn(mcp, llm, tools_subset)
            results.append((name, ok, llm.session_cost))
            print(f"  ✅ PASS: {name}")
        except Exception as e:
            results.append((name, False, llm.session_cost))
            print(f"  ❌ FAIL: {name} — {e}")

    proc.terminate()
    proc.wait(timeout=5)

    print("\n" + "=" * 70)
    print("🏁 HEAVY E2E BENCHMARK COMPLETE")
    print("=" * 70)
    passed = sum(1 for _, ok, _ in results if ok)
    total = len(results)
    print(f"Scenarios: {passed}/{total} passed")
    print(f"Total cost: ${llm.session_cost:.4f} / ${BUDGET_USD}")
    print(f"Tokens: {llm.total_input} in + {llm.total_output} out = {llm.total_input + llm.total_output} total")
    for name, ok, cost in results:
        emoji = "✅" if ok else "❌"
        print(f"  {emoji} {name}: ${cost:.4f}")

    return 0 if passed == total else 1

if __name__ == "__main__":
    sys.exit(main())
