# PDF Inputs

Choose pdf-text or native for PDF inputs with the file-parser plugin.

Use `file-parser` only when a `/v1/chat/completions` request must choose a PDF engine explicitly. If bare PDF inputs already work for the request, you can usually omit the plugin and let GonkaGate use server-managed PDF handling. For authenticated requests with plugin settings policy enabled, omitting `file-parser` can still apply a saved `file-parser` default.

## Minimum working request
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "file",
          "file": {
            "filename": "sample.pdf",
            "file_data": "data:application/pdf;base64,<base64-pdf>"
          }
        },
        {
          "type": "text",
          "text": "Summarize the attached PDF."
        }
      ]
    }
  ],
  "plugins": [
    {
      "id": "file-parser",
      "pdf": { "engine": "pdf-text" }
    }
  ]
}
```

Result: this request pins local `pdf-text` parsing for one PDF input. GonkaGate extracts text before the upstream model call.

## Choose the PDF engine

| Option | What GonkaGate does | Use it when | Do not choose it when |
|---|---|---|---|
| Omit pdf.engine inside file-parser | Uses server-managed auto selection. Today that usually resolves to local pdf-text. | The request includes file-parser but does not need a fixed engine. | The request must pin local parsing or explicitly ask for native. |
| Omit file-parser entirely | Bare PDF handling stays enabled without the plugin. In most requests GonkaGate uses server-managed auto selection, but authenticated requests with plugin settings policy enabled can inherit a saved file-parser default instead. | Bare PDF support is enough and the request does not need an explicit per-request override. | The request must pin local parsing or explicitly ask for native, or the caller must ignore saved plugin defaults. |
| pdf-text | Parses the PDF locally, extracts text before the upstream call, and can attach file annotations in non-stream responses. | You want predictable text extraction from public http/https URLs or data:application/pdf;base64,.... | The selected model path should receive the original PDF file for native handling. |
| native | Forwards the original PDF file and the file-parser plugin upstream unchanged. | The selected model path is expected to support native PDF input and you do not want gateway-side extraction. | You expect fallback to local parsing. Explicit native fails with pdf_native_unavailable when native handling is unavailable. |

## What changes in the request and response

- `pdf-text` is local-only. GonkaGate extracts text before the upstream model call and strips the local-only override from the forwarded payload.
- Non-stream `pdf-text` responses can include reusable file annotations in `choices[].message.annotations[]`.
- Streaming responses keep the normal chat-completions SSE shape and do not add PDF annotations.
- `native` keeps the original file part and forwards `plugins: [{ "id": "file-parser", "pdf": { "engine": "native" } }]` upstream unchanged.

## Supported sources and practical limits

- `pdf-text` supports public `http/https` URLs and `data:application/pdf;base64,...`.
- Private or local URLs are blocked by design.
- OCR is not supported yet. Textless or scanned PDFs can fail with `pdf_input_text_unavailable`.
- Current default request limits are up to `5` PDF files per request and up to `1 MiB` per file.
- Oversized request bodies can fail with HTTP `413 payload_too_large` before PDF parsing starts.

## How saved plugin settings affect the request

Authenticated requests can inherit a saved `file-parser` default when plugin settings policy is enabled.

- An enabled saved setting can supply the default PDF engine even when the request omits `file-parser`.
- For authenticated requests with plugin settings policy enabled, omitting `file-parser` does not guarantee `server_auto`. A saved `file-parser` default can still become the effective engine.
- Locked settings can reject an incompatible request override with `400 invalid_request_error` and `code=plugin_override_blocked`.
- A disabled saved `file-parser` setting means “no saved engine preference”. It does not turn bare PDF support off.

## Common failures

| Error | What it means | What to do next |
|---|---|---|
| 400 pdf_engine_not_supported | The request asked for an unsupported engine such as mistral-ocr. | Use pdf-text or native only. |
| 400 pdf_native_unavailable | The request explicitly asked for native, but the selected model/runtime cannot use native PDF handling right now. | Omit the engine for auto handling or switch to pdf-text. |
| 400 plugin_override_blocked | Locked saved plugin settings conflict with the request override. | Remove the override or align the request with the saved account policy. |
| 400 pdf_input_not_supported or 400 pdf_input_invalid_data_url | The PDF source is not valid for pdf-text. | Use a public http/https URL or a valid data:application/pdf;base64,... payload. |
| 400 pdf_input_text_unavailable | The PDF does not expose usable text and OCR is not available. | Send a text-based PDF or preprocess it outside GonkaGate. |
| 413 payload_too_large or 400 pdf_input_too_large | The body or file exceeds the active size budget. | Reduce file size/count or prefer URL mode instead of embedding large data URLs. |

## See also

- [Plugins Overview](./overview) to compare plugin activation rules.
- [Structured Outputs](/en/docs/guides/features/structured-outputs) if the next step is returning JSON after PDF extraction.
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact `file` content-part and `plugins` schema.