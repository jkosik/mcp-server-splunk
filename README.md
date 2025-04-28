# MCP Server for Splunk

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
- `list_splunk_indexes`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)
- `list_splunk_macros`
    - Parameters:
        - `count` (number, optional): Number of results to return (max 100, default 10)
        - `offset` (number, optional): Offset for pagination (default 0)


## Usage
### STDIO mode (default)
```bash
SPLUNK_URL=https://your-splunk-instance
SPLUNK_TOKEN=your-splunk-token

# List available tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | go run cmd/mcp-server-splunk/main.go | jq

# Call list_splunk_saved_searches tool
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_splunk_saved_searches","arguments":{}}}' | go run cmd/mcp-server-splunk/main.go | jq
```

## SSE mode (Server-Sent Events HTTP API)
```bash
SPLUNK_URL=https://your-splunk-instance
SPLUNK_TOKEN=your-splunk-token

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
        "SPLUNK_URL": "https://your-splunk-instance",
        "SPLUNK_TOKEN": "your-splunk-token"
      }
    }
  }
}
```

### SSE mode
Start the server:
```
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