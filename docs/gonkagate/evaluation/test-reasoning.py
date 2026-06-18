#!/usr/bin/env python3
"""Test reasoning/troubleshooting requests against all 3 GonkaGate models."""

import json
import time
from pathlib import Path
import urllib.request
import urllib.error

BASE_DIR = Path(__file__).parent
RESULTS_DIR = BASE_DIR / "04-reasoning-requests"
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


def chat_completion(base_url, api_key, model, messages, max_tokens=1500, temperature=0.2):
    url = f"{base_url}/chat/completions"
    payload = {
        "model": model,
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": temperature,
    }
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )

    start = time.time()
    try:
        with urllib.request.urlopen(req, timeout=90) as resp:
            body = resp.read().decode("utf-8")
            status = resp.status
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8")
        status = e.code
    elapsed = time.time() - start

    return {
        "model": model,
        "status": status,
        "elapsed_seconds": round(elapsed, 3),
        "response": json.loads(body) if body else None,
    }


def main():
    env = load_env()
    base_url = env.get("GONKAGATE_BASE_URL", "https://api.gonkagate.com/v1")
    api_key = env["GONKAGATE_API_KEY"]

    prompt = """Ты помощник руководителя службы технической поддержки.

Ситуация:
- Бот ТП работает на Bitrix24 через webhook
- Хранилище сессий использует Redis + файловый fallback (dual-write)
- Сегодня утром пользователи жаловались, что бот не реагирует на сообщения
- В логах support-сервиса вижу:
  [STORE] File sync error
  [RECONCILER] Цикл завершён
- Контейнеры: b24-support, b24-redis, b24-gateway — все healthy
- curl http://localhost:3001/health возвращает OK
- Redis отвечает на PING

Вопросы:
1. Какие 3 наиболее вероятные причины такого поведения?
2. В каком порядке проводить диагностику?
3. Какие команды выполнить, чтобы локализовать проблему?
4. Какое временное решение позволит восстановить работу, если Redis недоступен?

Отвечай структурированно, по пунктам."""

    models = [
        env.get("GONKAGATE_FAST_MODEL", "minimaxai/minimax-m2.7"),
        env.get("GONKAGATE_BALANCED_MODEL", "moonshotai/kimi-k2.6"),
        env.get("GONKAGATE_REASONING_MODEL", "qwen/qwen3-235b-a22b-instruct-2507-fp8"),
    ]

    summary_lines = ["=== Reasoning / Troubleshooting Test ===", ""]

    for model in models:
        print(f"Testing {model}...")
        result = chat_completion(
            base_url=base_url,
            api_key=api_key,
            model=model,
            messages=[{"role": "user", "content": prompt}],
            max_tokens=1500,
            temperature=0.2,
        )

        safe_model = model.replace("/", "_")
        out_path = RESULTS_DIR / f"{safe_model}.json"
        with open(out_path, "w", encoding="utf-8") as f:
            json.dump(result, f, ensure_ascii=False, indent=2)

        content = ""
        if result["response"] and "choices" in result["response"]:
            content = result["response"]["choices"][0]["message"].get("content", "")

        summary_lines.extend([
            f"--- {model} ---",
            f"Status: {result['status']}, Time: {result['elapsed_seconds']}s",
            "Response:",
            content,
            "",
        ])

    summary_path = RESULTS_DIR / "summary.md"
    with open(summary_path, "w", encoding="utf-8") as f:
        f.write("\n".join(summary_lines))

    print(f"\nResults saved to {RESULTS_DIR}")


if __name__ == "__main__":
    main()
