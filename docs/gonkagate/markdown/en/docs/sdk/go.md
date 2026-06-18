# Go SDK

Send chat.completions requests to GonkaGate from Go with sashabaranov/go-openai.

Send one `chat.completions` request from Go with `github.com/sashabaranov/go-openai`, then reuse the same client shape in your service code.

## Minimum working example

Install the client and export your API key:
 Installation
```
go get github.com/sashabaranov/go-openai
export GONKAGATE_API_KEY="gp-your-api-key"
```

Then run one `chat.completions` request:
**Request Example**
```
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/sashabaranov/go-openai"
)

func main() {
    apiKey := os.Getenv("GONKAGATE_API_KEY")
    if apiKey == "" {
        log.Fatal("missing GONKAGATE_API_KEY")
    }

    config := openai.DefaultConfig(apiKey)
    config.BaseURL = "https://api.gonkagate.com/v1"

    client := openai.NewClientWithConfig(config)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := client.CreateChatCompletion(
        ctx,
        openai.ChatCompletionRequest{
            Model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: "Say hello from GonkaGate in one sentence.",
                },
            },
        },
    )

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

Expected result: if this prints text, your Go client is wired correctly.

Use a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models). The example value above is only illustrative.

## What you need before you run it

- Go `1.18+`, because that is the minimum version required by `github.com/sashabaranov/go-openai`
- A `gp-...` API key stored in `GONKAGATE_API_KEY`
- Enough available balance for the request
- A current GonkaGate model ID

## Go-specific notes

- Set `config.BaseURL = "https://api.gonkagate.com/v1"` before you build the client.
- Reuse one configured client in your process or service instead of recreating it on every request.
- Use `context.WithTimeout(...)` on each request so slow calls fail predictably.
- If you need custom proxy or transport behavior, pass a custom `http.Client` through `config.HTTPClient`.

## Common errors and limits

- If you forget `config.BaseURL = "https://api.gonkagate.com/v1"`, the client still points at the default OpenAI endpoint.
- `401 invalid_api_key` usually means the key is missing, malformed, revoked, or tied to an unavailable account state.
- `404 model_not_found` means the model ID is no longer available. Refresh it from [GET /v1/models](/en/docs/api/api-reference/models/get-models).
- `context deadline exceeded` means your timeout budget is too small for the request path or current network conditions.
- `429 insufficient_quota` means the available prepaid USD balance is too low for the request.

## See also

- [OpenAI SDK compatibility](/en/docs/sdk/openai) for the shared `BaseURL`, key format, model-selection rules, and additive `usage` fields across runtimes.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, secure storage, and rotation.
- [GonkaGate API Reference Overview](/en/docs/api/reference/overview) when the first Go request works and you need exact request fields, streaming behavior, and failure policy.
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) for the current machine-readable model IDs.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching already-working OpenAI code instead of starting from zero.