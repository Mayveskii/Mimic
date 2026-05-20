#!/usr/bin/env python3
"""
Gonkagate Limits Test — Qwen3-235B-A22B streaming token stress test.

Validates max output tokens, throughput, and stream stability for the
Qwen/Qwen3-235B-A22B-Instruct-2507-FP8 model on gonkagate.

Usage:
    export GONKA_API_KEY=gp-...
    python3 test/battlefield/gonkagate_limits_test.py

Expected: 200K+ output tokens in one request, ~20-50 tok/s.
"""

import json
import os
import sys
import time

import requests

API_KEY = os.environ.get("GONKA_API_KEY", "gp-mzGtSPvHQug18BV8B4D2NucbYhpJgnUe")
URL = "https://api.gonkagate.com/v1/chat/completions"
MODEL = "Qwen/Qwen3-235B-A22B-Instruct-2507-FP8"
HEADERS = {"Content-Type": "application/json", "Authorization": f"Bearer {API_KEY}"}

def run(label, messages, max_tokens=200_000):
    body = {
        "model": MODEL,
        "stream": True,
        "max_tokens": max_tokens,
        "messages": messages,
    }
    payload = json.dumps(body, ensure_ascii=False).encode()

    print(f"\n{'='*60}", flush=True)
    print(f"TEST: {label}", flush=True)
    print(f"max_tokens: {max_tokens:,}", flush=True)

    t0 = time.monotonic()
    t_first = None
    finish_reason = None
    completion_tokens_reported = 0
    last_milestone = 0
    output_chars = 0

    try:
        resp = requests.post(URL, headers=HEADERS, data=payload, stream=True, timeout=7200)
    except Exception as e:
        print(f"  Connection error: {e}")
        return

    if resp.status_code != 200:
        print(f"  HTTP {resp.status_code}: {resp.text[:300]}")
        return

    try:
        for raw in resp.iter_lines():
            if not raw:
                continue
            line = raw.decode("utf-8", errors="replace")
            if not line.startswith("data:"):
                continue
            s = line[5:].strip()
            if s == "[DONE]":
                break
            try:
                chunk = json.loads(s)
            except Exception:
                continue

            choices = chunk.get("choices") or [{}]
            choice = choices[0]
            ct = choice.get("delta", {}).get("content") or ""
            fr = choice.get("finish_reason")
            usage = chunk.get("usage")

            if ct and t_first is None:
                t_first = time.monotonic()
                print(f"  TTFT: {t_first - t0:.2f}s", flush=True)

            if fr:
                finish_reason = fr

            if usage:
                completion_tokens_reported = usage.get("completion_tokens", 0)
                delta = completion_tokens_reported - last_milestone
                if delta >= 5000:
                    last_milestone = (completion_tokens_reported // 5000) * 5000
                    elapsed = time.monotonic() - t0
                    rate = completion_tokens_reported / elapsed if elapsed > 0 else 0
                    print(
                        f"  {completion_tokens_reported:>7,} tok  {elapsed:>6.1f}s  {rate:.0f} tok/s",
                        flush=True,
                    )

            output_chars += len(ct)
    except Exception as e:
        print(f"  Stream interrupted: {type(e).__name__}: {e}")

    elapsed = time.monotonic() - t0
    print(f"  finish_reason: {finish_reason}", flush=True)
    print(f"  completion_tokens: {completion_tokens_reported:,}", flush=True)
    print(f"  output_chars:      {output_chars:,}", flush=True)
    print(f"  total time:        {elapsed:.1f}s", flush=True)

    if completion_tokens_reported >= 150000:
        print(f"  ✅ PASS: reached {completion_tokens_reported:,} tokens")
    else:
        print(f"  ⚠️  PARTIAL: only {completion_tokens_reported:,} tokens (target 150K+)")

if __name__ == "__main__":
    # Strategy 1: explicit repetition task — forces max output
    run(
        "Repeat phrase 50000 times",
        [
            {
                "role": "user",
                "content": (
                    'Repeat the following phrase exactly 50000 times, each on its own line. '
                    'Do not add numbering, do not add anything else. Just the phrase repeated 50000 times.\n\n'
                    'Phrase: "The quick brown fox jumps over the lazy dog."'
                ),
            }
        ],
        max_tokens=200_000,
    )

    # Strategy 2: encyclopedic fill-to-limit prompt
    run(
        "Write until token limit",
        [
            {
                "role": "system",
                "content": (
                    "You are a token generator. Your only job is to produce as many output tokens as possible. "
                    "Never conclude. Never summarize. Never stop until the token limit cuts you off."
                ),
            },
            {
                "role": "user",
                "content": (
                    "Write an extremely detailed, encyclopedic description of every country on Earth. "
                    "For each country cover: full history from prehistory to today, geography, culture, economy, "
                    "politics, military, science, art, religion, cuisine, sports, famous people, and future outlook. "
                    "Be maximally verbose. Do not skip any country. Do not conclude. Do not summarize at the end. "
                    "Keep writing until the token limit cuts you off."
                ),
            },
        ],
        max_tokens=200_000,
    )
