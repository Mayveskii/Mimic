# .NET SDK

Send chat.completions requests to GonkaGate from .NET with Betalgo.Ranul.OpenAI.

Send one `chat.completions` request from .NET with `Betalgo.Ranul.OpenAI`, then reuse the same service shape in your app code.

## Minimum working example

Install the current package:
 Installation
```
dotnet add package Betalgo.Ranul.OpenAI
```

Then run one `chat.completions` request:
**Request Example**
```
using System;
using System.Collections.Generic;
using System.Linq;
using Betalgo.Ranul.OpenAI;
using Betalgo.Ranul.OpenAI.Managers;
using Betalgo.Ranul.OpenAI.ObjectModels;
using Betalgo.Ranul.OpenAI.ObjectModels.RequestModels;

var apiKey = Environment.GetEnvironmentVariable("GONKAGATE_API_KEY")
    ?? throw new InvalidOperationException("missing GONKAGATE_API_KEY");

var openAiService = new OpenAIService(new OpenAIOptions
{
    ApiKey = apiKey,
    BaseDomain = "https://api.gonkagate.com/"
});

var completionResult = await openAiService.ChatCompletion.CreateCompletion(
    new ChatCompletionCreateRequest
    {
        Messages = new List<ChatMessage>
        {
            ChatMessage.FromUser("Say hello from GonkaGate in one sentence.")
        },
        Model = "qwen/qwen3-235b-a22b-instruct-2507-fp8"
    });

if (!completionResult.Successful)
{
    Console.WriteLine($"{completionResult.Error?.Code}: {completionResult.Error?.Message}");
    return;
}

Console.WriteLine(completionResult.Choices.First().Message.Content);
```

Expected result: if this prints text, your .NET client is wired correctly.

Use a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models). The example value above is only illustrative.

## What you need before you run it

- A current .NET runtime supported by `Betalgo.Ranul.OpenAI`
- A `gp-...` API key stored in `GONKAGATE_API_KEY`
- Enough available balance for the request
- A current GonkaGate model ID

## .NET-specific notes

- Use `Betalgo.Ranul.OpenAI`, not the older `Betalgo.OpenAI` package or namespace.
- Set `BaseDomain` to `https://api.gonkagate.com/`, not `https://api.gonkagate.com/v1`. Betalgo appends `v1` through its `ApiVersion` setting.
- If your app already uses DI, register the service with `AddOpenAIService(...)` instead of constructing a new `OpenAIService` in every call path.
- If `completionResult.Successful` is `false`, inspect `completionResult.Error?.Code` and `completionResult.Error?.Message` before retrying.

## Common errors and limits

- `401 invalid_api_key` usually means the key is missing, malformed, revoked, or tied to an unavailable account state.
- `404 model_not_found` means the model ID is no longer available. Refresh it from [GET /v1/models](/en/docs/api/api-reference/models/get-models).
- `429 insufficient_quota` means the available prepaid USD balance is too low for the request.

## See also

- [OpenAI SDK compatibility](/en/docs/sdk/openai) for the shared base URL, key format, model-selection rules, and additive `usage` fields across runtimes.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, secure storage, and rotation.
- [GonkaGate API Reference Overview](/en/docs/api/reference/overview) when the first .NET request works and you need exact request fields, streaming behavior, and failure policy.
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) for the current machine-readable model IDs.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching already-working OpenAI code instead of starting from zero.