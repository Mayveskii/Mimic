#!/usr/bin/env python3
"""
Quick interactive test of Exa tools through Mimic MCP server.
"""
import subprocess, json, time, sys, os

env = os.environ.copy()
env["EXA_API_KEY"] = open("/home/cisco/mimic/project_context_main/secrets/tokens.env").read().split("EXA_API_KEY=")[1].split("\n")[0]

proc = subprocess.Popen(
    ["/home/cisco/mimic/bin/mimic", "serve"],
    stdin=subprocess.PIPE, stdout=subprocess.PIPE,
    stderr=subprocess.PIPE, env=env
)

def rpc(method, params=None, msg_id=1):
    req = json.dumps({"jsonrpc": "2.0", "id": msg_id, "method": method, "params": params or {}}) + "\n"
    proc.stdin.write(req.encode())
    proc.stdin.flush()
    line = proc.stdout.readline().decode()
    return json.loads(line)

def toolcall(name, args):
    return rpc("tools/call", {"name": name, "arguments": args})

# Initialize
resp = rpc("initialize", {"protocolVersion": "2024-11-05", "clientInfo": {"name": "test"}})
print("initialize:", "OK" if "result" in resp else "FAIL")

# Tools list
resp = rpc("tools/list", {})
tools = [t["name"] for t in resp.get("result", {}).get("tools", []) if "EXA" in t["name"] or t["name"] == "MIMIC_RESEARCH"]
print("Exa tools found:", tools)

# EXA_SEARCH
start = time.time()
resp = toolcall("EXA_SEARCH", {"query": "etcd RAFT Go implementation", "numResults": 3})
lat = (time.time() - start) * 1000
err = resp.get("result", {}).get("isError")
print(f"EXA_SEARCH: {'FAIL' if err else 'PASS'}  {lat:.0f}ms")

# EXA_FETCH
start = time.time()
resp = toolcall("EXA_FETCH", {"urls": ["https://github.com/etcd-io/etcd"], "maxChars": 500})
lat = (time.time() - start) * 1000
err = resp.get("result", {}).get("isError")
print(f"EXA_FETCH:  {'FAIL' if err else 'PASS'}  {lat:.0f}ms")

# MIMIC_RESEARCH shallow
start = time.time()
resp = toolcall("MIMIC_RESEARCH", {"topic": "Go RAFT consensus", "depth": "shallow"})
lat = (time.time() - start) * 1000
err = resp.get("result", {}).get("isError")
print(f"MIMIC_RESEARCH(shallow): {'FAIL' if err else 'PASS'}  {lat:.0f}ms")

# MIMIC_RESEARCH deep
start = time.time()
resp = toolcall("MIMIC_RESEARCH", {"topic": "Go RAFT consensus", "depth": "deep"})
lat = (time.time() - start) * 1000
err = resp.get("result", {}).get("isError")
print(f"MIMIC_RESEARCH(deep):    {'FAIL' if err else 'PASS'}  {lat:.0f}ms")

proc.terminate()
proc.wait(timeout=3)
print("\nDone.")
