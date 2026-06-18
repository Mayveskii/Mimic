# GonkaGate Documentation (Offline Mirror)

**Source:** https://gonkagate.com/en/docs  
**Downloaded:** 2026-05-28  
**Pages:** 52  
**Formats:** HTML (`html/`) + Markdown (`markdown/`)

---

## Structure

```
docs/gonkagate/
├── html/          # Original HTML pages
├── markdown/      # Extracted Markdown (easier to read/search)
├── urls.txt       # List of downloaded URLs
├── sitemap.xml    # Original sitemap
└── README.md      # This file
```

---

## Index

### Getting Started
- [Quickstart](markdown/en/docs/quickstart.md)
- [FAQ](markdown/en/docs/faq.md)
- [Principles](markdown/en/docs/guides/overview/principles.md)
- [Models & Routing](markdown/en/docs/guides/overview/models.md)
- [Migration from OpenAI](markdown/en/docs/guides/overview/migration.md)

### Authentication
- [API Keys](markdown/en/docs/authentication/api-keys.md)
- [Management API Keys](markdown/en/docs/authentication/management-api-keys.md)

### API Reference
- [Overview](markdown/en/docs/api/reference/overview.md)
- [Parameters](markdown/en/docs/api/reference/parameters.md)
- [Streaming](markdown/en/docs/api/reference/streaming.md)
- [Rate Limits](markdown/en/docs/api/reference/rate-limits.md)
- [Error Handling](markdown/en/docs/api/reference/error-handling.md)
- [Chat Completions](markdown/en/docs/api/api-reference/chat/send-chat-completion-request.md)
- [Models](markdown/en/docs/api/api-reference/models/get-models.md)
- [Dashboard Models](markdown/en/docs/api/api-reference/dashboard/get-dashboard-models.md)
- [Public Pricing](markdown/en/docs/api/api-reference/public/get-public-pricing.md)
- [Whoami](markdown/en/docs/api/api-reference/whoami/get-whoami.md)

### Features
- [Presets](markdown/en/docs/guides/features/presets.md)
- [Tool Calling](markdown/en/docs/guides/features/tool-calling.md)
- [Structured Outputs](markdown/en/docs/guides/features/structured-outputs.md)
- [Plugins Overview](markdown/en/docs/guides/features/plugins/overview.md)
- [Web Search Plugin](markdown/en/docs/guides/features/plugins/web-search.md)
- [PDF Inputs Plugin](markdown/en/docs/guides/features/plugins/pdf-inputs.md)
- [Response Healing Plugin](markdown/en/docs/guides/features/plugins/response-healing.md)
- [Privacy Sanitization Plugin](markdown/en/docs/guides/features/plugins/privacy-sanitization.md)

### Coding Agents Guides
- [Claude Code](markdown/en/docs/guides/coding-agents/claude-code.md)
- [Cursor](markdown/en/docs/guides/coding-agents/cursor.md)
- [Hermes Agent](markdown/en/docs/guides/coding-agents/hermes-agent.md)
- [Kilo Code](markdown/en/docs/guides/coding-agents/kilo-code.md)
- [MiMoCode](markdown/en/docs/guides/coding-agents/mimo-code.md)
- [OpenClaw](markdown/en/docs/guides/coding-agents/openclaw.md)
- [OpenCode](markdown/en/docs/guides/coding-agents/opencode.md)

### Community & Integrations
- [Awesome GonkaGate](markdown/en/docs/guides/community/awesome-gonkagate.md)
- [Aider](markdown/en/docs/guides/community/aider.md)
- [Cline](markdown/en/docs/guides/community/cline.md)
- [Roo Code](markdown/en/docs/guides/community/roo-code.md)
- [Frameworks Overview](markdown/en/docs/guides/community/frameworks-and-integrations-overview.md)
- [LangChain](markdown/en/docs/guides/community/langchain.md)
- [LlamaIndex](markdown/en/docs/guides/community/llamaindex.md)
- [n8n Setup](markdown/en/docs/guides/community/n8n.md)
- [PydanticAI](markdown/en/docs/guides/community/pydanticai.md)
- [TanStack AI](markdown/en/docs/guides/community/tanstack-ai.md)
- [Vercel AI SDK](markdown/en/docs/guides/community/vercel-ai-sdk.md)

### SDKs & MCP
- [SDK Overview](markdown/en/docs/sdk.md)
- [OpenAI SDK](markdown/en/docs/sdk/openai.md)
- [Python SDK](markdown/en/docs/sdk/python.md)
- [TypeScript SDK](markdown/en/docs/sdk/typescript.md)
- [Go SDK](markdown/en/docs/sdk/go.md)
- [Java SDK](markdown/en/docs/sdk/java.md)
- [.NET SDK](markdown/en/docs/sdk/dotnet.md)
- [MCP Setup](markdown/en/docs/mcp.md)

### Other
- [Gonka API](markdown/en/gonka-api.md)

---

## Search

```bash
# Search in markdown
grep -r "your query" docs/gonkagate/markdown/

# Search in HTML
grep -r "your query" docs/gonkagate/html/
```

---

## Notes

- Markdown was automatically extracted from HTML.
- Code blocks, tables, headings, lists, and links are preserved.
- Some visual styling and interactive elements may be lost in Markdown conversion.
- For exact original rendering, use the `html/` version.
