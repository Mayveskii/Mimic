# GonkaGate Models Evaluation Summary

**Date:** 2026-05-28

**Models tested:**
- `minimaxai/minimax-m2.7`
- `moonshotai/kimi-k2.6`
- `qwen/qwen3-235b-a22b-instruct-2507-fp8`

## Response Times

| Model | Simple | Analytics | Reasoning |
|-------|--------|-----------|-----------|
| `minimaxai/minimax-m2.7` | 0.685s | 8.878s | 19.89s |
| `moonshotai/kimi-k2.6` | 0.512s | 8.499s | 15.916s |
| `qwen/qwen3-235b-a22b-instruct-2507-fp8` | 0.667s | 3.061s | 54.237s |

## Accuracy Check — Analytics Task

Correct answers: 10 sessions; Корнилов and Орлов tied with 3 each; password_reset and software_bug tied with 3 each; average rating 4.2; peak 09:00-10:00 or 10:00-11:00.

| Model | Count | Top Operator | Top Category | Avg Rating | Peak | Verdict |
|-------|-------|--------------|--------------|------------|------|---------|
| `minimaxai/minimax-m2.7` | 10 | Корнилов + Орлов (tie) | password_reset + software_bug (tie) | 4.2 | 09:00-10:00 | ✅ Correct on all counts, noted ties |
| `moonshotai/kimi-k2.6` | 10 | Корнилов + Орлов (tie) | password_reset + software_bug (tie) | ? | 09:00-10:00 | ❌ Avg rating wrong (4.3 instead of 4.2), noted ties |
| `qwen/qwen3-235b-a22b-instruct-2507-fp8` | 10 | Корнилов | software_bug | 4.3 | 11:00-12:00 | ❌ Avg rating wrong, peak wrong, ignored ties |

## Behavioral Observations

| Model | Think Block | Identity | Notes |
|-------|-------------|----------|-------|
| minimax | Yes | MiniMax-M2.7 | Detailed, accurate, emits reasoning tags |
| kimi | No | Claims to be Claude (Anthropic) | System-prompt override of identity; arithmetic error |
| qwen | No | Qwen (Alibaba) | Fast on analytics but very slow on reasoning; accuracy issues |

## Error Handling

- Invalid key: `401` — `{"error":{"message":"Invalid credentials.","type":"authentication_error","code":`
- Invalid model: `404` — `{"error":{"message":"Model not available.","type":"invalid_request_error","code"`
- Timeout (1ms): `URL_ERROR` — `timed out`
- Rate-limit burst (10 req × 0.1s): 0 non-200 responses

## Recommendations

### Default model for TP assistant
Use `minimaxai/minimax-m2.7` as the default model.
- Fast enough for real-time bot responses
- Highest accuracy on analytics task
- Provides reasoning in `<think>` blocks (strip or log separately)

### Fallback / alternative roles
- `moonshotai/kimi-k2.6`: use for concise analytical summaries when speed is less critical
- `qwen/qwen3-235b-a22b-instruct-2507-fp8`: use only for offline/deep-reasoning tasks; 54s is too slow for interactive bot

### Implementation notes
1. Strip `<think>...</think>` blocks from responses before sending to users
2. Increase `max_tokens` to 2048+ for reasoning/troubleshooting
3. Always validate arithmetic claims from LLM against computed data
4. Implement timeout handling and fallback to 'передам оператору'
5. Add per-user rate limiting to avoid abuse
