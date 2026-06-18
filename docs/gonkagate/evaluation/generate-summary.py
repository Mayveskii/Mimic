#!/usr/bin/env python3
"""Generate evaluation summary from test results."""

import json
from pathlib import Path

BASE_DIR = Path(__file__).parent


def load_json(path):
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def extract_content(result):
    if result.get("response") and "choices" in result["response"]:
        return result["response"]["choices"][0]["message"].get("content", "")
    return ""


def main():
    models = [
        "minimaxai_minimax-m2.7",
        "moonshotai_kimi-k2.6",
        "qwen_qwen3-235b-a22b-instruct-2507-fp8",
    ]
    model_names = {
        "minimaxai_minimax-m2.7": "minimaxai/minimax-m2.7",
        "moonshotai_kimi-k2.6": "moonshotai/kimi-k2.6",
        "qwen_qwen3-235b-a22b-instruct-2507-fp8": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    }

    # Simple
    simple_times = {}
    for m in models:
        data = load_json(BASE_DIR / "02-simple-requests" / f"{m}.json")
        simple_times[m] = data["elapsed_seconds"]

    # Analytics
    analytics_times = {}
    analytics_content = {}
    for m in models:
        data = load_json(BASE_DIR / "03-analytics-requests" / f"{m}.json")
        analytics_times[m] = data["elapsed_seconds"]
        analytics_content[m] = extract_content(data)

    # Reasoning
    reasoning_times = {}
    reasoning_content = {}
    for m in models:
        data = load_json(BASE_DIR / "04-reasoning-requests" / f"{m}.json")
        reasoning_times[m] = data["elapsed_seconds"]
        reasoning_content[m] = extract_content(data)

    # Errors
    error_data = load_json(BASE_DIR / "05-error-handling" / "results.json")

    # Check for think blocks
    has_think = {m: "<think>" in analytics_content.get(m, "") or "<think>" in reasoning_content.get(m, "") for m in models}

    # Check truncation
    truncated = {m: reasoning_content.get(m, "").endswith(("|", "```", "-")) for m in models}

    lines = []
    lines.append("# GonkaGate Models Evaluation Summary")
    lines.append("")
    lines.append("**Date:** 2026-05-28")
    lines.append("")
    lines.append("**Models tested:**")
    for name in model_names.values():
        lines.append(f"- `{name}`")
    lines.append("")

    lines.append("## Response Times")
    lines.append("")
    lines.append("| Model | Simple | Analytics | Reasoning |")
    lines.append("|-------|--------|-----------|-----------|")
    for m in models:
        lines.append(f"| `{model_names[m]}` | {simple_times[m]}s | {analytics_times[m]}s | {reasoning_times[m]}s |")
    lines.append("")

    lines.append("## Accuracy Check — Analytics Task")
    lines.append("")
    lines.append("Correct answers: 10 sessions; Корнилов and Орлов tied with 3 each; password_reset and software_bug tied with 3 each; average rating 4.2; peak 09:00-10:00 or 10:00-11:00.")
    lines.append("")
    lines.append("| Model | Count | Top Operator | Top Category | Avg Rating | Peak | Verdict |")
    lines.append("|-------|-------|--------------|--------------|------------|------|---------|")

    verdicts = {
        "minimaxai_minimax-m2.7": "✅ Correct on all counts, noted ties",
        "moonshotai_kimi-k2.6": "❌ Avg rating wrong (4.3 instead of 4.2), noted ties",
        "qwen_qwen3-235b-a22b-instruct-2507-fp8": "❌ Avg rating wrong, peak wrong, ignored ties",
    }

    # Extract metrics manually from content
    def guess_metrics(content):
        text = content.lower()
        count = "10" if "10" in text else "?"
        top_op = "Корнилов" if "корнилов" in text else "?"
        if "корнилов" in text and "орлов" in text:
            top_op = "Корнилов + Орлов (tie)"
        top_cat = "?"
        if "password_reset" in text and "software_bug" in text:
            top_cat = "password_reset + software_bug (tie)"
        elif "password_reset" in text:
            top_cat = "password_reset"
        elif "software_bug" in text:
            top_cat = "software_bug"
        avg = "4.2" if "4.2" in text else ("4.3" if "4.3" in text else "?")
        peak = "?"
        if "09:00" in text:
            peak = "09:00-10:00"
        elif "11:00" in text:
            peak = "11:00-12:00"
        return count, top_op, top_cat, avg, peak

    for m in models:
        count, top_op, top_cat, avg, peak = guess_metrics(analytics_content.get(m, ""))
        lines.append(f"| `{model_names[m]}` | {count} | {top_op} | {top_cat} | {avg} | {peak} | {verdicts[m]} |")
    lines.append("")

    lines.append("## Behavioral Observations")
    lines.append("")
    lines.append("| Model | Think Block | Identity | Notes |")
    lines.append("|-------|-------------|----------|-------|")
    lines.append(f"| minimax | {'Yes' if has_think['minimaxai_minimax-m2.7'] else 'No'} | MiniMax-M2.7 | Detailed, accurate, emits reasoning tags |")
    lines.append(f"| kimi | {'Yes' if has_think['moonshotai_kimi-k2.6'] else 'No'} | Claims to be Claude (Anthropic) | System-prompt override of identity; arithmetic error |")
    lines.append(f"| qwen | {'Yes' if has_think['qwen_qwen3-235b-a22b-instruct-2507-fp8'] else 'No'} | Qwen (Alibaba) | Fast on analytics but very slow on reasoning; accuracy issues |")
    lines.append("")

    lines.append("## Error Handling")
    lines.append("")
    lines.append(f"- Invalid key: `{error_data['invalid_key']['status']}` — `{error_data['invalid_key']['body'][:80]}`")
    lines.append(f"- Invalid model: `{error_data['invalid_model']['status']}` — `{error_data['invalid_model']['body'][:80]}`")
    lines.append(f"- Timeout (1ms): `{error_data['timeout']['status']}` — `{error_data['timeout']['body']}`")
    non_200 = [r['status'] for r in error_data['rate_limit_burst'] if r['status'] != 200]
    lines.append(f"- Rate-limit burst (10 req × 0.1s): {len(non_200)} non-200 responses")
    lines.append("")

    lines.append("## Recommendations")
    lines.append("")
    lines.append("### Default model for TP assistant")
    lines.append("Use `minimaxai/minimax-m2.7` as the default model.")
    lines.append("- Fast enough for real-time bot responses")
    lines.append("- Highest accuracy on analytics task")
    lines.append("- Provides reasoning in `<think>` blocks (strip or log separately)")
    lines.append("")
    lines.append("### Fallback / alternative roles")
    lines.append("- `moonshotai/kimi-k2.6`: use for concise analytical summaries when speed is less critical")
    lines.append("- `qwen/qwen3-235b-a22b-instruct-2507-fp8`: use only for offline/deep-reasoning tasks; 54s is too slow for interactive bot")
    lines.append("")
    lines.append("### Implementation notes")
    lines.append("1. Strip `<think>...</think>` blocks from responses before sending to users")
    lines.append("2. Increase `max_tokens` to 2048+ for reasoning/troubleshooting")
    lines.append("3. Always validate arithmetic claims from LLM against computed data")
    lines.append("4. Implement timeout handling and fallback to 'передам оператору'")
    lines.append("5. Add per-user rate limiting to avoid abuse")
    lines.append("")

    out_path = BASE_DIR / "SUMMARY.md"
    with open(out_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    print(f"Summary written to {out_path}")


if __name__ == "__main__":
    main()
