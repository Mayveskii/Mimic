# MCP Setup

Connect the public GonkaGate Docs MCP server in Cursor, VS Code / Copilot Agent, Claude Code, or any client that supports a remote HTTP MCP server.

Use GonkaGate Docs MCP when you want your assistant to read the live GonkaGate docs instead of answering from stale model memory. The server is public, read-only, and does not require auth.

## Endpoint

Use this URL:
**Endpoint**
```
https://mcp.gonkagate.com/v1
```

You do not need an API key or OAuth flow for this server.

## Cursor

Create `.cursor/mcp.json` in the repository root:
 Cursor
```
{
  "mcpServers": {
    "gonkagate-docs": {
      "url": "https://mcp.gonkagate.com/v1"
    }
  }
}
```

Reload Cursor. Then open Agent mode and make sure the GonkaGate docs tools appear.

## VS Code / Copilot Agent

Create `.vscode/mcp.json` in the repository root:
 VS Code / Copilot Agent
```
{
  "servers": {
    "gonkagate-docs": {
      "url": "https://mcp.gonkagate.com/v1"
    }
  }
}
```

Then open Copilot Chat, switch to Agent mode, and confirm the GonkaGate docs tools appear in the MCP tools list.

## Claude Code

Run this command:
 Command
```
claude mcp add --transport http gonkagate-docs https://mcp.gonkagate.com/v1
```

Then run `/mcp` inside Claude Code and confirm the server is connected.

If you want a shared project config, add `--scope project`. Claude Code will store it in `.mcp.json`.

## Other MCP clients

If your MCP client supports a remote HTTP or Streamable HTTP MCP server, point it to:
 Other MCP clients
```
https://mcp.gonkagate.com/v1
```

## Check that it works

After setup, try this prompt:
 Check that it works
```
Use the GonkaGate Docs MCP server to find the Quickstart page and summarize the first request flow.
```

What you should see:

- the answer is grounded in GonkaGate docs
- the answer references a GonkaGate docs page

## Troubleshooting

- If the client cannot connect, make sure you used `https://mcp.gonkagate.com/v1`, not the site root.
- If the server connects but no tools appear, reload the client and make sure MCP tools are enabled.
- If you want Russian docs, ask for the Russian page explicitly or request `locale: ru`.

## See also

- [Quickstart](/en/docs/quickstart)
- [GonkaGate FAQ](/en/docs/faq)