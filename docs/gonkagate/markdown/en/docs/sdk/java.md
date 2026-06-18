# Java SDK

Send chat.completions requests to GonkaGate from Java with openai-java.

Send one `chat.completions` request from Java with `openai-java`, then reuse the same client shape in your application code.

## Minimum working example

Add the SDK:
 Minimum working example
```
implementation("com.openai:openai-java:4.29.0")
```

Then run one `chat.completions` request:
**Request Example**
```
import com.openai.client.OpenAIClient;
import com.openai.client.okhttp.OpenAIOkHttpClient;
import com.openai.models.chat.completions.ChatCompletion;
import com.openai.models.chat.completions.ChatCompletionCreateParams;

public class BasicUsage {
    public static void main(String[] args) {
        String apiKey = System.getenv("GONKAGATE_API_KEY");
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalStateException("missing GONKAGATE_API_KEY");
        }

        OpenAIClient client = OpenAIOkHttpClient.builder()
            .apiKey(apiKey)
            .baseUrl("https://api.gonkagate.com/v1")
            .build();

        ChatCompletionCreateParams params = ChatCompletionCreateParams.builder()
            .addUserMessage("Say hello from GonkaGate in one sentence.")
            .model("qwen/qwen3-235b-a22b-instruct-2507-fp8")
            .build();

        ChatCompletion completion = client.chat().completions().create(params);

        completion.choices().forEach(choice ->
            choice.message().content().ifPresent(System.out::println)
        );
    }
}
```

Expected result: if this prints text, your Java client is wired correctly.

Use a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models). The example value above is only illustrative.

## What you need before you run it

- Java `8+`, because that is the minimum version required by `openai-java`
- A `gp-...` API key stored in `GONKAGATE_API_KEY`
- Enough available balance for the request
- A current GonkaGate model ID
- If you use Maven instead of Gradle, use the same coordinates: `com.openai:openai-java:4.29.0`

## Java-specific notes

- Set `.baseUrl("https://api.gonkagate.com/v1")` before you build the client.
- Reuse one client per application where possible. The SDK maintains connection pools and thread pools internally.
- The default client is synchronous. If you need `CompletableFuture`-based execution, switch to `client.async()` or create an `OpenAIOkHttpClientAsync`.
- If you copy older snippets, update the dependency version together with the rest of the example.

## Common errors and limits

- If you forget `.baseUrl("https://api.gonkagate.com/v1")`, the client still points at the default OpenAI endpoint.
- `401 invalid_api_key` usually means the key is missing, malformed, revoked, or tied to an unavailable account state.
- `404 model_not_found` means the model ID is no longer available. Refresh it from [GET /v1/models](/en/docs/api/api-reference/models/get-models).
- `429 insufficient_quota` means the available prepaid USD balance is too low for the request.

## See also

- [OpenAI SDK compatibility](/en/docs/sdk/openai) for the shared base URL, key format, model-selection rules, and additive `usage` fields across runtimes.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, secure storage, and rotation.
- [GonkaGate API Reference Overview](/en/docs/api/reference/overview) when the first Java request works and you need exact request fields, streaming behavior, and failure policy.
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) for the current machine-readable model IDs.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching already-working OpenAI code instead of starting from zero.