#!/usr/bin/env python3
"""Test error handling for GonkaGate API."""

import json
import time
from pathlib import Path
import urllib.request
import urllib.error
import socket

BASE_DIR = Path(__file__).parent
RESULTS_DIR = BASE_DIR / "05-error-handling"
RESULTS_DIR.mkdir(exist_ok=True)


def load_env():
    env_path = BASE_DIR / ".." / ".." / ".." / ".env"
    env = {}
    with open(env_path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line and "=" in line and not line.startswith("#"):
                key, value = line.split("=", 1)
                env[key] = value
    return env


def make_request(base_url, api_key=None, model=None, timeout=30, path="/chat/completions"):
    url = f"{base_url}{path}"
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"

    payload = {
        "model": model or "moonshotai/kimi-k2.6",
        "messages": [{"role": "user", "content": "Hello"}],
        "max_tokens": 10,
    }
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")

    start = time.time()
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            body = resp.read().decode("utf-8")
            status = resp.status
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8")
        status = e.code
    except urllib.error.URLError as e:
        body = str(e.reason)
        status = "URL_ERROR"
    except socket.timeout:
        body = "SOCKET_TIMEOUT"
        status = "TIMEOUT"
    except Exception as e:
        body = str(e)
        status = "EXCEPTION"
    elapsed = time.time() - start

    return {
        "status": status,
        "elapsed_seconds": round(elapsed, 3),
        "body": body,
    }


def main():
    env = load_env()
    base_url = env.get("GONKAGATE_BASE_URL", "https://api.gonkagate.com/v1")
    api_key = env["GONKAGATE_API_KEY"]

    tests = {
        "invalid_key": lambda: make_request(base_url, api_key="gp-invalid-key", timeout=10),
        "invalid_model": lambda: make_request(base_url, api_key=api_key, model="nonexistent/model", timeout=30),
        "timeout": lambda: make_request(base_url, api_key=api_key, timeout=0.001),
    }

    results = {}
    summary_lines = ["=== Error Handling Test ===", ""]

    for name, test_fn in tests.items():
        print(f"Testing {name}...")
        result = test_fn()
        results[name] = result

        summary_lines.extend([
            f"--- {name} ---",
            f"Status: {result['status']}, Time: {result['elapsed_seconds']}s",
            f"Body: {result['body'][:500]}",
            "",
        ])

    # Rate limit test: send 10 rapid requests
    print("Testing rate limit with 10 rapid requests...")
    rate_results = []
    for i in range(10):
        res = make_request(base_url, api_key=api_key, timeout=10)
        rate_results.append({"index": i, **res})
        time.sleep(0.1)

    results["rate_limit_burst"] = rate_results
    non_200 = [r for r in rate_results if r["status"] != 200]
    summary_lines.extend([
        "--- rate_limit_burst (10 requests) ---",
        f"Non-200 responses: {len(non_200)}",
        f"Statuses: {[r['status'] for r in rate_results]}",
        "",
    ])

    out_path = RESULTS_DIR / "results.json"
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(results, f, ensure_ascii=False, indent=2)

    summary_path = RESULTS_DIR / "summary.md"
    with open(summary_path, "w", encoding="utf-8") as f:
        f.write("\n".join(summary_lines))

    print(f"\nResults saved to {RESULTS_DIR}")


if __name__ == "__main__":
    main()
