# MCP Server for Splunk
[![smithery badge](https://smithery.ai/badge/@jkosik/mcp-server-splunk)](https://smithery.ai/server/@jkosik/mcp-server-splunk)

A Go implementation of the MCP server for Splunk.
Supports STDIO and SSE (Server-Sent Events HTTP API). Uses github.com/mark3labs/mcp-go SDK.

## MCP Tools implemented
- `list_splunk_saved_searches`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 100)
        - `offset` (number, optional): Offset for pagination (default 0)
- `list_splunk_alerts`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)
        - `title` (string, optional): Case-insensitive substring to filter alert titles
- `list_splunk_fired_alerts`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)
        - `ss_name` (string, optional): Search name pattern to filter alerts (default "*")
        - `earliest` (string, optional): Time range to look back (default "-24h")
- `list_splunk_indexes`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)
- `list_splunk_macros`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)

## MCP Prompts and Resources
- `internal/splunk/prompt.go` implements an MCP Prompt to find Splunk alerts for a specific keyword (e.g. GitHub or OKTA) and instructs Cursor to utilise multiple MCP tools to review all Splunk alerts, indexes and macros first to provide the best answer.
- `cmd/mcp/server/main.go` implements MCP Resource in the form of local CSV file with Splunk related content, providing further context to the chat.

## Usage
### Installing via Smithery

To install mcp-server-splunk for Claude Desktop automatically via [Smithery](https://smithery.ai/server/@jkosik/mcp-server-splunk):

```bash
npx -y @smithery/cli install @jkosik/mcp-server-splunk --client claude
```

### STDIO mode (default)
```bash
export SPLUNK_URL=https://your-splunk-instance
export SPLUNK_TOKEN=your-splunk-token

# List available tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | go run cmd/mcp-server-splunk/main.go | jq

# Call list_splunk_saved_searches tool
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_splunk_saved_searches","arguments":{}}}' | go run cmd/mcp-server-splunk/main.go | jq
```

## SSE mode (Server-Sent Events HTTP API)
```bash
export SPLUNK_URL=https://your-splunk-instance
export SPLUNK_TOKEN=your-splunk-token

# Start the server
go run cmd/mcp-server-splunk/main.go -transport sse -port 3001

# Call the server and get Session ID from the output. Do not terminate the session.
curl http://localhost:3001/sse

# Keep session running and and use different terminal window for the final MCP call
curl -X POST "http://localhost:3001/message?sessionId=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | jq
```

## Cursor integration
By configuring MCP Settings in Cursor, you can include remote data directly into the LLM context.

![Demo](docs/mcp-short.gif)

Integrate STDIO or SSE MCP Servers (see below) and use Cursor Chat.
Cursor will automatically try to use MCP Tools, Prompts or Re
Sample prompts:
- `How many MCP tools for Splunk are available?`
- `How many Splunk indexes do we have?`
- `Can you list first 5 Splunk macros including underlying queries?`
- `How many alers with "Alert_CRITICAL" in the name were fired in the last day?`
- `Read the MCP Resource "Data Dictionary" and find the contact person for the Splunk index XYZ.`

### STDIO mode
Build the server:
```
go build -o cmd/mcp-server-splunk/mcp-server-splunk cmd/mcp-server-splunk/main.go
```

Update `~/.cursor/mcp.json`
```json
{
  "mcpServers": {
    "splunk_stdio": {
      "name": "Splunk MCP Server (STDIO)",
      "description": "MCP server for Splunk integration",
      "type": "stdio",
      "command": "/Users/juraj/data/github.com/jkosik/mcp-server-splunk/cmd/mcp-server-splunk/mcp-server-splunk",
      "env": {
        "SPLUNK_URL": "https://your-splunk-instance:8089",
        "SPLUNK_TOKEN": "your-splunk-token"
      }
    }
  }
}
```

### SSE mode
Start the server:
```bash
export SPLUNK_URL=https://your-splunk-instance
export SPLUNK_TOKEN=your-splunk-token

# Start the server
go run cmd/mcp-server-splunk/main.go -transport sse -port 3001
```

Update `~/.cursor/mcp.json`
```json
{
  "mcpServers": {
    "splunk_sse": {
      "name": "Splunk MCP Server (SSE)",
      "description": "MCP server for Splunk integration (SSE mode)",
      "type": "sse",
      "url": "http://localhost:3001/sse"
    }
  }
}
```

## Smithery
`Dockerfile` and `smithery.yaml` are used to host this MCP implementation at https://smithery.ai

### Local build and run
```
docker build -t mcp-server-splunk .

echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | \
docker run --rm -i \
  -e SPLUNK_URL=https://your-splunk-instance \
  -e SPLUNK_TOKEN=your-splunk-token \
  mcp-server-splunk | jq
```