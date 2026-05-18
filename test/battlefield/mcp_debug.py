#!/usr/bin/env python3
"""
mcp_debug.py — Interactive MCP Server Debugger

Usage:
    python3 test/battlefield/mcp_debug.py

Allows interactive exploration of Mimic MCP server:
    > tools          — list all tools
    > call <name>    — call a tool with interactive args
    > chain <list>   — execute a chain of tools
    > inspect <name> — show tool schema
    > test <tier>    — run tier 1/2/3 tests
    > quit           — exit
"""

import json
import subprocess
import sys
import time
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent

class MCPDebugger:
    def __init__(self):
        self.proc = None
        self.tools = []
        self.tool_map = {}
        
    def start(self):
        binary = PROJECT_ROOT / "bin" / "mimic"
        if not binary.exists():
            print("❌ Binary not found. Run 'make build' first.")
            sys.exit(1)
            
        print("🚀 Starting Mimic MCP server...")
        self.proc = subprocess.Popen(
            [str(binary), "serve"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        time.sleep(1.5)
        
        # Get tools
        req = {"jsonrpc": "2.0", "id": 1, "method": "tools/list"}
        self.send(req)
        resp = self.recv()
        self.tools = resp.get("result", {}).get("tools", [])
        self.tool_map = {t["name"]: t for t in self.tools}
        
        print(f"✅ {len(self.tools)} tools loaded")
        print(f"   Type 'help' for commands\n")
        
    def send(self, msg):
        line = json.dumps(msg) + "\n"
        self.proc.stdin.write(line)
        self.proc.stdin.flush()
        
    def recv(self):
        line = self.proc.stdout.readline()
        return json.loads(line)
        
    def call_tool(self, name, arguments):
        req = {
            "jsonrpc": "2.0",
            "id": int(time.time() * 1000),
            "method": "tools/call",
            "params": {"name": name, "arguments": arguments},
        }
        self.send(req)
        return self.recv()
        
    def cmd_tools(self, args):
        print(f"\n📋 {len(self.tools)} registered tools:")
        for t in self.tools:
            schema = t.get("inputSchema", {})
            required = schema.get("required", [])
            print(f"  • {t['name']:<20} {t.get('description', '')[:50]}")
            if required:
                print(f"    Required: {', '.join(required)}")
        print()
        
    def cmd_inspect(self, args):
        if not args:
            print("Usage: inspect <tool_name>")
            return
        name = args[0]
        tool = self.tool_map.get(name)
        if not tool:
            print(f"❌ Tool '{name}' not found")
            return
        print(f"\n🔍 {name}")
        print(f"   Description: {tool.get('description', 'N/A')}")
        schema = tool.get("inputSchema", {})
        print(f"   Schema:")
        print(json.dumps(schema, indent=4))
        print()
        
    def cmd_call(self, args):
        if not args:
            print("Usage: call <tool_name> [json_args]")
            return
        name = args[0]
        if name not in self.tool_map:
            print(f"❌ Tool '{name}' not found")
            return
            
        # Parse arguments
        if len(args) > 1:
            try:
                arguments = json.loads(" ".join(args[1:]))
            except json.JSONDecodeError:
                print("❌ Invalid JSON arguments")
                return
        else:
            # Interactive prompt for args
            schema = self.tool_map[name].get("inputSchema", {})
            props = schema.get("properties", {})
            required = schema.get("required", [])
            arguments = {}
            if props:
                print(f"\nEnter arguments for {name} (press Enter to skip optional):")
                for prop, spec in props.items():
                    req = "*" if prop in required else " "
                    prompt = f"  [{req}] {prop} ({spec.get('type', 'any')}): "
                    val = input(prompt).strip()
                    if val:
                        # Convert type
                        ptype = spec.get("type", "string")
                        if ptype == "integer":
                            val = int(val)
                        elif ptype == "boolean":
                            val = val.lower() in ("true", "1", "yes")
                        arguments[prop] = val
                    elif prop in required:
                        print(f"   ⚠️  {prop} is required but empty")
                        
        print(f"\n📤 Calling {name}({json.dumps(arguments)})...")
        result = self.call_tool(name, arguments)
        
        print(f"\n📥 Result:")
        print(json.dumps(result, indent=2))
        
        # Analyze result quality
        if "error" in result:
            print(f"\n⚠️  ERROR: {result['error']}")
        elif "result" in result:
            content = result["result"].get("content", "")
            if isinstance(content, list):
                text = " ".join(c.get("text", "") for c in content)
            else:
                text = str(content)
            print(f"\n📊 Output length: {len(text)} chars")
            if len(text) == 0:
                print("⚠️  WARNING: Empty output!")
            elif len(text) < 50:
                print(f"   Content: '{text}'")
                
    def cmd_chain(self, args):
        if not args:
            print("Usage: chain tool1 arg1 tool2 arg2 ...")
            print("Example: chain SYS_FILE_EXISTS '{\"path\":\"README.md\"}' IO_READ '{\"path\":\"README.md\"}'")
            return
            
        # Parse chain: tool_name json_args tool_name json_args ...
        i = 0
        chain = []
        while i < len(args):
            name = args[i]
            if name not in self.tool_map:
                print(f"❌ Unknown tool: {name}")
                return
            i += 1
            if i < len(args) and args[i].startswith("{"):
                try:
                    arguments = json.loads(args[i])
                    i += 1
                except:
                    arguments = {}
            else:
                arguments = {}
            chain.append((name, arguments))
            
        print(f"\n⛓️  Executing chain of {len(chain)} tools...")
        for i, (name, args) in enumerate(chain, 1):
            print(f"\n[{i}/{len(chain)}] {name}({json.dumps(args)})...")
            result = self.call_tool(name, args)
            print(f"Result: {json.dumps(result, indent=2)[:200]}...")
            if "error" in result:
                print(f"❌ Chain broken at step {i}: {result['error']}")
                return
                
        print(f"\n✅ Chain completed: {len(chain)} steps")
        
    def cmd_test_tier1(self, args):
        print("\n🧪 Tier 1: Simple tool calls")
        tests = [
            ("SYS_FILE_EXISTS", {"path": "README.md"}, "exists: true"),
            ("HASH_SHA256", {"data": "hello"}, None),
            ("GIT_STATUS", {}, None),
        ]
        for name, arguments, expected_substring in tests:
            print(f"\n  {name}...")
            result = self.call_tool(name, arguments)
            if "error" in result:
                print(f"    ❌ FAIL: {result['error']}")
            else:
                content = str(result.get("result", {}).get("content", ""))
                if expected_substring and expected_substring not in content:
                    print(f"    ⚠️  Expected '{expected_substring}' in output")
                    print(f"    Got: '{content[:100]}'")
                else:
                    print(f"    ✅ OK (output: {len(content)} chars)")
                    
    def cmd_test_tier2(self, args):
        print("\n🧪 Tier 2: Tool chains")
        print("\nScenario: 'Check git status, then read recent commits'")
        
        # Step 1
        print("\n[1] GIT_STATUS...")
        r1 = self.call_tool("GIT_STATUS", {})
        print(f"    Status: {'✅ OK' if 'error' not in r1 else '❌ FAIL'}")
        
        # Step 2
        print("\n[2] GIT_LOG (limit=3)...")
        r2 = self.call_tool("GIT_LOG", {"limit": 3})
        print(f"    Status: {'✅ OK' if 'error' not in r2 else '❌ FAIL'}")
        
        # Analyze if this would work for a model
        print("\n📊 Analysis:")
        print("   Model sees step 1 result, should decide to call step 2")
        print("   But in single API call, model only makes ONE tool_call")
        print("   Need: external loop that feeds result back to model")
        
    def cmd_test_tier3(self, args):
        print("\n🧪 Tier 3: Complex analysis")
        print("\nScenario: 'Read core/ops.c and analyze error handling'")
        
        # Step 1: Read file
        print("\n[1] IO_READ core/ops.c (first 50 lines)...")
        r1 = self.call_tool("IO_READ", {"path": "core/ops.c", "limit": 50})
        content = str(r1.get("result", {}).get("content", ""))
        print(f"    Read {len(content)} chars")
        
        # Step 2: Search for error patterns
        print("\n[2] Search for ERR_ constants...")
        err_count = content.count("ERR_")
        print(f"    Found {err_count} ERR_ references")
        
        print("\n📊 Analysis:")
        print("   Output is truncated to 50 lines")
        print("   RTK compression would strip comments, keep signatures")
        print("   For full analysis, need: either larger limit OR grep first")
        
    def cmd_help(self, args):
        print("""
📖 Commands:
  tools                    — List all registered tools
  inspect <name>           — Show tool JSON schema
  call <name> [args]       — Call a tool (interactive if no args)
  chain <tool> <args> ...  — Execute multiple tools sequentially
  test 1                   — Run Tier 1 tests (simple calls)
  test 2                   — Run Tier 2 tests (chains)
  test 3                   — Run Tier 3 tests (complex)
  help                     — Show this help
  quit                     — Exit debugger
""")
        
    def run(self):
        self.start()
        
        while True:
            try:
                cmdline = input("mcp> ").strip()
            except EOFError:
                break
            except KeyboardInterrupt:
                print("\nUse 'quit' to exit")
                continue
                
            if not cmdline:
                continue
                
            parts = cmdline.split()
            cmd = parts[0].lower()
            args = parts[1:]
            
            if cmd in ("quit", "exit", "q"):
                break
            elif cmd == "help":
                self.cmd_help(args)
            elif cmd == "tools":
                self.cmd_tools(args)
            elif cmd == "inspect":
                self.cmd_inspect(args)
            elif cmd == "call":
                self.cmd_call(args)
            elif cmd == "chain":
                self.cmd_chain(args)
            elif cmd == "test":
                tier = args[0] if args else "1"
                if tier == "1":
                    self.cmd_test_tier1(args)
                elif tier == "2":
                    self.cmd_test_tier2(args)
                elif tier == "3":
                    self.cmd_test_tier3(args)
                else:
                    print(f"Unknown tier: {tier}")
            else:
                print(f"Unknown command: {cmd}. Type 'help' for available commands.")
                
        print("\n👋 Stopping MCP server...")
        self.proc.terminate()
        self.proc.wait(timeout=3)
        print("Done.")

if __name__ == "__main__":
    MCPDebugger().run()
