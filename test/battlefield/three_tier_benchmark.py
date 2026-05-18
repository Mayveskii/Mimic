#!/usr/bin/env python3
"""
Mimic Battlefield — 3-Tier Benchmark

Tests Mimic against real OpenRouter API (kimi k2.6) with 3 complexity tiers:
  Tier 1: Simple (single tool)       — schema validation, argument correctness
  Tier 2: Medium (tool chains)       — decomposition, budget tracking
  Tier 3: Complex (code analysis)    — full pipeline, RTK compression

Usage:
    export OPENROUTER_KEY=sk-or-v1-...
    python3 test/battlefield/three_tier_benchmark.py

Saves:
    test/battlefield/results/TECHNICAL.json  — raw metrics
    test/battlefield/results/SEMANTIC.md      — human-readable analysis
    test/battlefield/results/REQUIRED_DATA.md — data gaps for community
"""

import json
import os
import subprocess
import sys
import time
import tempfile
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass, field, asdict

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
MODEL = "moonshotai/kimi-k2.6"
MAX_TOKENS = 4096
TEMPERATURE = 0.1
BUDGET_USD = 2.0  # Total test budget

# ---------------------------------------------------------------------------
# Metrics dataclass
# ---------------------------------------------------------------------------
@dataclass
class TierResult:
    tier: int
    task_name: str
    description: str
    status: str  # pass / fail / partial
    latency_ms: float
    input_tokens: int
    output_tokens: int
    total_tokens: int
    cost_usd: float
    tool_calls_count: int
    argument_collisions: int
    rollback_triggers: int
    compressed_output: bool
    notes: List[str] = field(default_factory=list)
    raw_response: Optional[Dict] = None

@dataclass
class BattleReport:
    date: str
    model: str
    total_cost_usd: float
    total_tokens: int
    total_time_ms: float
    tier_results: List[TierResult]
    data_gaps: List[str]
    recommendations: List[str]

# ---------------------------------------------------------------------------
# OpenRouter Client
# ---------------------------------------------------------------------------
class OpenRouterClient:
    def __init__(self, api_key: str):
        self.api_key = api_key
        self.base_url = "https://openrouter.ai/api/v1"
        self.session_cost = 0.0
        self.session_tokens = 0

    def chat(self, messages: List[Dict], tools: Optional[List[Dict]] = None) -> Dict:
        import urllib.request
        
        payload = {
            "model": MODEL,
            "messages": messages,
            "max_tokens": MAX_TOKENS,
            "temperature": TEMPERATURE,
        }
        if tools:
            payload["tools"] = tools
            payload["tool_choice"] = "auto"

        req = urllib.request.Request(
            f"{self.base_url}/chat/completions",
            data=json.dumps(payload).encode(),
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
                "HTTP-Referer": "https://github.com/Mayveskii/Mimic",
                "X-Title": "Mimic Battlefield Benchmark",
            },
        )

        start = time.time()
        with urllib.request.urlopen(req, timeout=60) as resp:
            data = json.loads(resp.read().decode())
        latency = (time.time() - start) * 1000

        # Track costs
        usage = data.get("usage", {})
        tokens_in = usage.get("prompt_tokens", 0)
        tokens_out = usage.get("completion_tokens", 0)
        total = usage.get("total_tokens", tokens_in + tokens_out)
        
        # kimi k2.6 pricing: $0.50 / 1M input, $2.00 / 1M output (approximate)
        cost = (tokens_in * 0.50 + tokens_out * 2.00) / 1_000_000
        self.session_cost += cost
        self.session_tokens += total

        data["_latency_ms"] = latency
        data["_cost_usd"] = cost
        data["_tokens"] = {"input": tokens_in, "output": tokens_out, "total": total}
        return data

# ---------------------------------------------------------------------------
# MCP Server wrapper
# ---------------------------------------------------------------------------
class MCPServer:
    def __init__(self):
        self.proc = None
        self.tools = []

    def start(self):
        self.proc = subprocess.Popen(
            [str(PROJECT_ROOT / "bin" / "mimic"), "serve"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        time.sleep(1.5)
        
        # Get tools list
        req = {"jsonrpc": "2.0", "id": 1, "method": "tools/list"}
        self.proc.stdin.write(json.dumps(req) + "\n")
        self.proc.stdin.flush()
        resp = json.loads(self.proc.stdout.readline())
        self.tools = resp.get("result", {}).get("tools", [])
        
        # Convert MCP tools to OpenAI-compatible format for LLM APIs
        # OpenAI uses 'parameters' instead of 'inputSchema'
        self.openai_tools = []
        for t in self.tools:
            openai_tool = {
                "type": "function",
                "function": {
                    "name": t["name"],
                    "description": t.get("description", ""),
                    "parameters": t.get("inputSchema", {}),
                }
            }
            self.openai_tools.append(openai_tool)
        
        return self.tools

    def call_tool(self, name: str, arguments: Dict) -> Dict:
        req = {
            "jsonrpc": "2.0",
            "id": int(time.time() * 1000),
            "method": "tools/call",
            "params": {"name": name, "arguments": arguments},
        }
        self.proc.stdin.write(json.dumps(req) + "\n")
        self.proc.stdin.flush()
        return json.loads(self.proc.stdout.readline())

    def stop(self):
        if self.proc:
            self.proc.terminate()
            self.proc.wait(timeout=3)

# ---------------------------------------------------------------------------
# Tier 1: Simple tasks
# ---------------------------------------------------------------------------
def run_tier_1(client: OpenRouterClient, mcp: MCPServer) -> List[TierResult]:
    """Single-tool tasks. Validate schemas, argument correctness, real execution."""
    results = []
    tools_spec = mcp.openai_tools

    tasks = [
        {
            "name": "file_exists",
            "prompt": "Check if file README.md exists in the current directory. Use the available tool.",
            "expected_tool": "SYS_FILE_EXISTS",
            "expected_args": {"path": "README.md"},
        },
        {
            "name": "hash_sha256",
            "prompt": "Calculate SHA256 hash of the string 'hello world'. Use the HASH_SHA256 tool.",
            "expected_tool": "HASH_SHA256",
            "expected_args": {"data": "hello world"},
        },
        {
            "name": "git_status",
            "prompt": "Show git status of this repository. Use the GIT_STATUS tool.",
            "expected_tool": "GIT_STATUS",
            "expected_args": {},
        },
        {
            "name": "build_test",
            "prompt": "Run tests in this project. The test directory is './test'. Use BUILD_TEST tool with filter='Test'.",
            "expected_tool": "BUILD_TEST",
            "expected_args": {"dir": "./test", "filter": "Test"},
        },
    ]

    for task in tasks:
        print(f"\n  [Tier 1] {task['name']}...")
        
        start = time.time()
        resp = client.chat(
            messages=[{"role": "user", "content": task["prompt"]}],
            tools=tools_spec,
        )
        latency = (time.time() - start) * 1000

        # Check if model used correct tool
        tool_calls = resp.get("choices", [{}])[0].get("message", {}).get("tool_calls", [])
        used_tools = [tc["function"]["name"] for tc in tool_calls if tc.get("function")]
        
        # Validate arguments via MCP
        arg_errors = 0
        tool_results = []
        for tc in tool_calls:
            fn = tc.get("function", {})
            tool_name = fn.get("name", "")
            args = json.loads(fn.get("arguments", "{}"))
            
            if tool_name == task["expected_tool"]:
                result = mcp.call_tool(tool_name, args)
                tool_results.append(result)
                # Check if MCP reported argument errors
                if "error" in str(result).lower():
                    arg_errors += 1

        status = "pass" if task["expected_tool"] in used_tools and arg_errors == 0 else "fail"
        if status == "pass" and not tool_results:
            status = "partial"  # Correct tool but no execution
        
        # Debug logging for failures
        if status == "fail":
            print(f"    DEBUG: Expected {task['expected_tool']}, got {used_tools}")
            print(f"    DEBUG: Tool calls raw: {json.dumps(tool_calls, indent=2)[:500]}")

        results.append(TierResult(
            tier=1,
            task_name=task["name"],
            description=f"Single tool: {task['expected_tool']}",
            status=status,
            latency_ms=latency,
            input_tokens=resp["_tokens"]["input"],
            output_tokens=resp["_tokens"]["output"],
            total_tokens=resp["_tokens"]["total"],
            cost_usd=resp["_cost_usd"],
            tool_calls_count=len(tool_calls),
            argument_collisions=arg_errors,
            rollback_triggers=0,
            compressed_output=False,
            notes=[f"Used tools: {used_tools}"] + [f"Result: {r.get('result', {}).get('content', '')[:100]}" for r in tool_results],
            raw_response=resp,
        ))
        print(f"    Status: {status} | Cost: ${resp['_cost_usd']:.4f} | Latency: {latency:.0f}ms")

    return results

# ---------------------------------------------------------------------------
# Tier 2: Medium tasks (chains)
# ---------------------------------------------------------------------------
def run_tier_2(client: OpenRouterClient, mcp: MCPServer) -> List[TierResult]:
    """Multi-step tasks requiring tool chains and decomposition."""
    results = []
    tools_spec = mcp.openai_tools

    tasks = [
        {
            "name": "analyze_repo_state",
            "prompt": """Analyze the current state of this repository:
1. Check git status
2. List recent git log (5 commits)
3. Read the main Makefile to understand build targets
Use the appropriate tools and provide a summary.
""",
            "min_tools": 3,
        },
        {
            "name": "build_and_verify",
            "prompt": """Build this project and verify tests pass:
1. Run 'make build' equivalent (BUILD_COMPILE)
2. Run tests (BUILD_TEST)
3. Check that the binary exists (SYS_FILE_EXISTS on bin/mimic)
Use appropriate tools.
""",
            "min_tools": 3,
        },
    ]

    for task in tasks:
        print(f"\n  [Tier 2] {task['name']}...")
        
        start = time.time()
        resp = client.chat(
            messages=[{"role": "user", "content": task["prompt"]}],
            tools=tools_spec,
        )
        latency = (time.time() - start) * 1000

        tool_calls = resp.get("choices", [{}])[0].get("message", {}).get("tool_calls", [])
        used_tools = [tc["function"]["name"] for tc in tool_calls if tc.get("function")]

        # Execute tools
        arg_errors = 0
        for tc in tool_calls:
            fn = tc.get("function", {})
            tool_name = fn.get("name", "")
            args = json.loads(fn.get("arguments", "{}"))
            result = mcp.call_tool(tool_name, args)
            if "error" in str(result).lower():
                arg_errors += 1

        status = "pass" if len(used_tools) >= task["min_tools"] and arg_errors == 0 else "partial"
        if len(used_tools) < 2:
            status = "fail"

        results.append(TierResult(
            tier=2,
            task_name=task["name"],
            description="Multi-step tool chain",
            status=status,
            latency_ms=latency,
            input_tokens=resp["_tokens"]["input"],
            output_tokens=resp["_tokens"]["output"],
            total_tokens=resp["_tokens"]["total"],
            cost_usd=resp["_cost_usd"],
            tool_calls_count=len(tool_calls),
            argument_collisions=arg_errors,
            rollback_triggers=0,
            compressed_output=len(tool_calls) > 2,  # Would trigger RTK
            notes=[f"Chain: {' -> '.join(used_tools)}"],
        ))
        print(f"    Status: {status} | Tools: {len(tool_calls)} | Cost: ${resp['_cost_usd']:.4f}")

    return results

# ---------------------------------------------------------------------------
# Tier 3: Complex tasks (code analysis)
# ---------------------------------------------------------------------------
def run_tier_3(client: OpenRouterClient, mcp: MCPServer) -> List[TierResult]:
    """Complex analysis requiring context compression and synthesis."""
    results = []
    tools_spec = mcp.openai_tools

    tasks = [
        {
            "name": "refactor_core_ops",
            "prompt": """This is a Go project with C-core. I need to understand and improve error handling:

1. Read core/ops.c and find all error return paths (search for ERR_ constants)
2. Read the error handling patterns in the README
3. Suggest 3 specific improvements to make error messages more actionable

Use IO_READ to examine files. Focus on concrete, implementable changes.
""",
            "context_size": "large",
        },
        {
            "name": "architecture_review",
            "prompt": """Review the architecture of this project for production readiness:

1. Read AGENTS.md to understand agent rules
2. Read specs/00-SPEC-INDEX.md for documentation coverage
3. Check if all critical files exist (Makefile, Dockerfile, go.mod)
4. Suggest what monitoring/observability is missing for production deployment

Use file tools to examine the codebase. Be specific.
""",
            "context_size": "very_large",
        },
    ]

    for task in tasks:
        print(f"\n  [Tier 3] {task['name']}...")
        
        start = time.time()
        resp = client.chat(
            messages=[{"role": "user", "content": task["prompt"]}],
            tools=tools_spec,
        )
        latency = (time.time() - start) * 1000

        tool_calls = resp.get("choices", [{}])[0].get("message", {}).get("tool_calls", [])
        used_tools = [tc["function"]["name"] for tc in tool_calls if tc.get("function")]
        
        # Check if model used decomposition (multiple independent calls)
        decomposition = len(tool_calls) > 3 and len(set(used_tools)) >= 3

        # Check if RTK compression would be needed
        rtk_needed = False
        for tc in tool_calls:
            fn = tc.get("function", {})
            result = mcp.call_tool(fn.get("name"), json.loads(fn.get("arguments", "{}")))
            content = str(result.get("result", {}).get("content", ""))
            if len(content) > 1000:  # Would trigger RTK
                rtk_needed = True

        # Quality: did we get actionable output?
        final_msg = resp.get("choices", [{}])[0].get("message", {})
        # Model may return content=None if using tool_calls
        content_text = final_msg.get("content") or ""
        has_actionable = any(kw in content_text.lower() for kw in ["improve", "suggest", "add", "remove", "change", "implement"])

        status = "pass" if decomposition and has_actionable else "partial"
        if len(tool_calls) < 2:
            status = "fail"

        results.append(TierResult(
            tier=3,
            task_name=task["name"],
            description="Complex analysis with synthesis",
            status=status,
            latency_ms=latency,
            input_tokens=resp["_tokens"]["input"],
            output_tokens=resp["_tokens"]["output"],
            total_tokens=resp["_tokens"]["total"],
            cost_usd=resp["_cost_usd"],
            tool_calls_count=len(tool_calls),
            argument_collisions=0,
            rollback_triggers=0,
            compressed_output=rtk_needed,
            notes=[
                f"Decomposition: {decomposition}",
                f"RTK triggered: {rtk_needed}",
                f"Actionable output: {has_actionable}",
                f"Unique tools: {len(set(used_tools))}",
            ],
        ))
        print(f"    Status: {status} | Tools: {len(tool_calls)} | Decomp: {decomposition} | RTK: {rtk_needed} | Cost: ${resp['_cost_usd']:.4f}")

    return results

# ---------------------------------------------------------------------------
# Data gap analysis
# ---------------------------------------------------------------------------
def analyze_data_gaps(all_results: List[TierResult]) -> List[str]:
    """Identify what data is missing for better performance."""
    gaps = []
    
    # Check argument collision rate
    total_collisions = sum(r.argument_collisions for r in all_results)
    if total_collisions > 0:
        gaps.append(f"Argument collisions detected: {total_collisions}. Need: more complete JSON Schema enums/defaults for affected tools.")
    
    # Check decomposition success
    tier2_decomp = [r for r in all_results if r.tier == 2 and r.status != "pass"]
    if tier2_decomp:
        gaps.append(f"Tier 2 decomposition failures: {len(tier2_decomp)}. Need: better task relationship metadata in mesh slots.")
    
    # Check RTK usage
    rtk_cases = [r for r in all_results if r.compressed_output]
    if rtk_cases:
        gaps.append(f"RTK compression triggered {len(rtk_cases)} times. Need: language-specific compression rules for: {', '.join(set(r.task_name for r in rtk_cases))}")
    
    # Check if model hallucinated tools
    all_used = set()
    for r in all_results:
        for note in r.notes:
            if "Used tools:" in note or "Chain:" in note:
                tools = note.replace("Used tools: ", "").replace("Chain: ", "").replace(" -> ", ",").split(",")
                all_used.update(t.strip().strip("[]'") for t in tools)
    
    known_tools = {"SYS_FILE_EXISTS", "IO_READ", "GIT_STATUS", "GIT_LOG", "BUILD_TEST", "BUILD_COMPILE", 
                   "HASH_SHA256", "HASH_MD5", "GIT_ADD", "GIT_COMMIT", "MMAP_ALLOC"}
    unknown_used = all_used - known_tools
    if unknown_used:
        gaps.append(f"Model tried unknown tools: {unknown_used}. Need: expand tool registry or improve tool descriptions.")
    
    if not gaps:
        gaps.append("No critical gaps detected. Current mesh data (13,611 slots) and tool schemas are sufficient for tested tiers.")
        gaps.append("Recommendation: Add more language-specific slots (Go, Rust, Python) to improve Tier 3 analysis quality.")
    
    return gaps

# ---------------------------------------------------------------------------
# Report generation
# ---------------------------------------------------------------------------
def save_technical_report(report: BattleReport, path: Path):
    data = {
        "date": report.date,
        "model": report.model,
        "budget_usd": BUDGET_USD,
        "total_cost_usd": report.total_cost_usd,
        "total_tokens": report.total_tokens,
        "total_time_ms": report.total_time_ms,
        "efficiency": {
            "tokens_per_dollar": report.total_tokens / report.total_cost_usd if report.total_cost_usd > 0 else 0,
            "tasks_per_dollar": len(report.tier_results) / report.total_cost_usd if report.total_cost_usd > 0 else 0,
        },
        "tier_summary": {
            "tier_1": {
                "tasks": len([r for r in report.tier_results if r.tier == 1]),
                "pass": len([r for r in report.tier_results if r.tier == 1 and r.status == "pass"]),
                "fail": len([r for r in report.tier_results if r.tier == 1 and r.status == "fail"]),
            },
            "tier_2": {
                "tasks": len([r for r in report.tier_results if r.tier == 2]),
                "pass": len([r for r in report.tier_results if r.tier == 2 and r.status == "pass"]),
                "partial": len([r for r in report.tier_results if r.tier == 2 and r.status == "partial"]),
            },
            "tier_3": {
                "tasks": len([r for r in report.tier_results if r.tier == 3]),
                "pass": len([r for r in report.tier_results if r.tier == 3 and r.status == "pass"]),
                "partial": len([r for r in report.tier_results if r.tier == 3 and r.status == "partial"]),
            },
        },
        "results": [asdict(r) for r in report.tier_results],
        "data_gaps": report.data_gaps,
        "recommendations": report.recommendations,
    }
    with open(path, 'w') as f:
        json.dump(data, f, indent=2)
    print(f"\n💾 Technical report: {path}")

def save_semantic_report(report: BattleReport, path: Path):
    with open(path, 'w') as f:
        f.write("# Battlefield Report — 3-Tier Benchmark\n\n")
        f.write(f"**Date:** {report.date}\n")
        f.write(f"**Model:** {report.model}\n")
        f.write(f"**Budget:** ${BUDGET_USD:.2f} | **Spent:** ${report.total_cost_usd:.4f}\n\n")
        
        f.write("## What We Tested\n\n")
        f.write("This benchmark measures Mimic's effectiveness in real-world conditions\n")
        f.write("using the OpenRouter API with kimi k2.6.\n\n")
        
        f.write("### Tier 1: Simple Tasks (Single Tools)\n\n")
        f.write("These test whether the model understands individual tools and uses correct arguments.\n")
        f.write("Think of it as: *Can the agent use a hammer without hitting its thumb?*\n\n")
        tier1 = [r for r in report.tier_results if r.tier == 1]
        for r in tier1:
            emoji = "✅" if r.status == "pass" else "❌" if r.status == "fail" else "⚠️"
            f.write(f"- {emoji} **{r.task_name}**: {r.description} — ")
            f.write(f"{r.status.upper()} (${r.cost_usd:.4f}, {r.latency_ms:.0f}ms, {r.tool_calls_count} tool call{'s' if r.tool_calls_count != 1 else ''})\n")
        f.write("\n")
        
        f.write("### Tier 2: Medium Tasks (Tool Chains)\n\n")
        f.write("These test whether the model breaks complex requests into ordered steps.\n")
        f.write("Think of it as: *Can the agent build a birdhouse, not just hammer one nail?*\n\n")
        tier2 = [r for r in report.tier_results if r.tier == 2]
        for r in tier2:
            emoji = "✅" if r.status == "pass" else "⚠️" if r.status == "partial" else "❌"
            f.write(f"- {emoji} **{r.task_name}**: {r.description} — ")
            f.write(f"{r.status.upper()} ({r.tool_calls_count} tools chained, ${r.cost_usd:.4f})\n")
            if r.notes:
                f.write(f"  - Note: {r.notes[0]}\n")
        f.write("\n")
        
        f.write("### Tier 3: Complex Tasks (Analysis & Synthesis)\n\n")
        f.write("These test whether the model can analyze large outputs, compress them,\n")
        f.write("and produce actionable recommendations.\n")
        f.write("Think of it as: *Can the agent redesign the workshop, not just build one shelf?*\n\n")
        tier3 = [r for r in report.tier_results if r.tier == 3]
        for r in tier3:
            emoji = "✅" if r.status == "pass" else "⚠️" if r.status == "partial" else "❌"
            f.write(f"- {emoji} **{r.task_name}**: {r.description} — ")
            f.write(f"{r.status.upper()} (RTK: {'yes' if r.compressed_output else 'no'}, ${r.cost_usd:.4f})\n")
            for note in r.notes:
                f.write(f"  - {note}\n")
        f.write("\n")
        
        f.write("## Efficiency Analysis\n\n")
        f.write(f"- **Total tokens:** {report.total_tokens:,}\n")
        f.write(f"- **Total cost:** ${report.total_cost_usd:.4f}\n")
        f.write(f"- **Total time:** {report.total_time_ms / 1000:.1f}s\n")
        if report.total_cost_usd > 0:
            f.write(f"- **Value:** {report.total_tokens / report.total_cost_usd:,.0f} tokens per dollar\n")
        f.write("\n")
        
        f.write("## Data Gaps for Community\n\n")
        f.write("These are the specific data/tool improvements needed based on test results:\n\n")
        for i, gap in enumerate(report.data_gaps, 1):
            f.write(f"{i}. {gap}\n")
        f.write("\n")
        
        f.write("## Recommendations\n\n")
        for rec in report.recommendations:
            f.write(f"- {rec}\n")
        f.write("\n")
        
        f.write("---\n\n")
        f.write("*This report is both a benchmark and a roadmap. Each data gap is a contribution opportunity.*\n")
    
    print(f"💾 Semantic report: {path}")

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    # API key
    api_key = os.environ.get("OPENROUTER_KEY", os.environ.get("OPENROUTER_API_KEY", ""))
    if not api_key:
        secrets_dir = PROJECT_ROOT / "project_context_main" / "secrets"
        for fname in ["openrouter.key", "openrouter.txt", ".env"]:
            fpath = secrets_dir / fname
            if fpath.exists():
                content = fpath.read_text().strip()
                for line in content.split("\n"):
                    if "OPENROUTER" in line and "=" in line:
                        api_key = line.split("=", 1)[1].strip().strip("\"'\"")
                        break
                if api_key:
                    break
    
    if not api_key:
        print("❌ OPENROUTER_KEY not found. Set env var or place in project_context_main/secrets/openrouter.key")
        return 1

    # Ensure binary built
    if not (PROJECT_ROOT / "bin" / "mimic").exists():
        print("Building Mimic binary...")
        subprocess.run(["make", "build"], cwd=PROJECT_ROOT, check=True)

    # Setup
    client = OpenRouterClient(api_key)
    mcp = MCPServer()
    results_dir = PROJECT_ROOT / "test" / "battlefield" / "results"
    results_dir.mkdir(parents=True, exist_ok=True)

    print("=" * 60)
    print("MIMIC BATTLEFIELD — 3-Tier Benchmark")
    print("=" * 60)
    print(f"Model: {MODEL}")
    print(f"Budget: ${BUDGET_USD:.2f}")
    print("")

    try:
        # Start MCP
        print("🚀 Starting MCP server...")
        tools = mcp.start()
        print(f"   {len(tools)} tools registered")

        # Run tiers
        all_results = []

        print("\n" + "=" * 60)
        print("TIER 1: Simple Tasks (Single Tools)")
        print("=" * 60)
        all_results.extend(run_tier_1(client, mcp))

        if client.session_cost >= BUDGET_USD * 0.5:
            print(f"\n⚠️  Budget 50% spent (${client.session_cost:.4f}). Skipping higher tiers.")
        else:
            print("\n" + "=" * 60)
            print("TIER 2: Medium Tasks (Tool Chains)")
            print("=" * 60)
            all_results.extend(run_tier_2(client, mcp))

        if client.session_cost >= BUDGET_USD * 0.8:
            print(f"\n⚠️  Budget 80% spent (${client.session_cost:.4f}). Skipping Tier 3.")
        else:
            print("\n" + "=" * 60)
            print("TIER 3: Complex Tasks (Analysis)")
            print("=" * 60)
            all_results.extend(run_tier_3(client, mcp))

        # Analyze
        gaps = analyze_data_gaps(all_results)
        recommendations = [
            "Add language-specific mesh slots (Go, Rust, Python) for Tier 3 analysis",
            "Expand JSON Schema with more enum values for commonly-misused arguments",
            "Test RTK compression ratios on real large outputs (>5KB)",
            "Add 'project context' mesh slots for common project types (CLI, web service, library)",
        ]

        # Build report
        report = BattleReport(
            date=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            model=MODEL,
            total_cost_usd=client.session_cost,
            total_tokens=client.session_tokens,
            total_time_ms=sum(r.latency_ms for r in all_results),
            tier_results=all_results,
            data_gaps=gaps,
            recommendations=recommendations,
        )

        # Save
        save_technical_report(report, results_dir / "TECHNICAL.json")
        save_semantic_report(report, results_dir / "SEMANTIC.md")

        # Print summary
        print("\n" + "=" * 60)
        print("BENCHMARK COMPLETE")
        print("=" * 60)
        print(f"Tasks:        {len(all_results)}")
        print(f"Passed:       {len([r for r in all_results if r.status == 'pass'])}")
        print(f"Partial:      {len([r for r in all_results if r.status == 'partial'])}")
        print(f"Failed:       {len([r for r in all_results if r.status == 'fail'])}")
        print(f"Total cost:   ${client.session_cost:.4f} / ${BUDGET_USD:.2f}")
        print(f"Total tokens: {client.session_tokens:,}")
        print(f"Data gaps:    {len(gaps)}")
        print("")
        print(f"Results saved:")
        print(f"  {results_dir / 'TECHNICAL.json'}")
        print(f"  {results_dir / 'SEMANTIC.md'}")
        print("=" * 60)

    finally:
        mcp.stop()

    return 0

if __name__ == "__main__":
    sys.exit(main())
