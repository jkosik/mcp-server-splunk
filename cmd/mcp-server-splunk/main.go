package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jkosik/mcp-server-splunk/internal/splunk"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Parse transport flag
	transport := flag.String("transport", "stdio", "Transport type: stdio or sse")
	port := flag.Int("port", 3001, "Port for SSE mode")
	flag.Parse()

	// Read Splunk credentials from environment variables
	baseURL := os.Getenv("SPLUNK_URL")
	authToken := os.Getenv("SPLUNK_TOKEN")
	if baseURL == "" || authToken == "" {
		fmt.Println("SPLUNK_URL and SPLUNK_TOKEN environment variables are required")
		return
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"Splunk MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	//////////////////////
	// SAVED SEARCHES //
	//////////////////////
	splunkTool := mcp.NewTool("list_splunk_saved_searches",
		mcp.WithDescription("List Splunk saved searches (paginated by count and offset arguments)."),
		mcp.WithNumber("count", mcp.Description("Number of results to return (max 100, default 10)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (default 0)")),
	)

	s.AddTool(splunkTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := splunk.NewClient(baseURL, authToken)
		count := 10
		offset := 0
		// Limit the count parameter to 100
		if v, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(v)
			if count > 100 {
				count = 100
			}
		}
		if v, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(v)
		}

		// Run the Splunk client and get the Splunk API response
		searches, total, err := client.GetSavedSearches(ctx, count, offset)
		if err != nil {
			return mcp.NewToolResultError("failed to get saved searches: " + err.Error()), nil
		}

		// Populate the response for the calling app
		note := fmt.Sprintf("Showing up to %d saved searches (as requested). Use 'offset' to paginate. Maximum per call is 100.", count)
		result := map[string]interface{}{
			"searches": searches,
			"count":    count,
			"offset":   offset,
			"total":    total,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
		}
		return mcp.NewToolResultText(note + "\n\n" + string(data)), nil
	})

	//////////////////////
	// FIRED ALERTS //
	//////////////////////
	alertsTool := mcp.NewTool("list_splunk_fired_alerts",
		mcp.WithDescription("List Splunk fired alerts (paginated by count and offset arguments)"),
		mcp.WithNumber("count", mcp.Description("Number of results to return (max 100, default 10)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (default 0)")),
	)

	s.AddTool(alertsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := splunk.NewClient(baseURL, authToken)
		count := 10
		offset := 0
		if v, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(v)
			if count > 500 { // more generous limit, since we're using SPL in API and the entire json is returned already.
				count = 500
			}
		}
		if v, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(v)
		}
		alerts, total, err := client.GetFiredAlerts(ctx, count, offset)
		if err != nil {
			return mcp.NewToolResultError("failed to get fired alerts: " + err.Error()), nil
		}
		note := fmt.Sprintf("Showing up to %d fired alerts (as requested). Use 'offset' to paginate. Maximum per call is 100.", count)
		result := map[string]interface{}{
			"alerts": alerts,
			"count":  count,
			"offset": offset,
			"total":  total,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
		}
		return mcp.NewToolResultText(note + "\n\n" + string(data)), nil
	})

	//////////////////////
	// ALERTS (With actions, filterable by title. Using SPL in API and the entire json is returned - mimicking pagination in GetAlerts.) //
	//////////////////////
	alertsAllTool := mcp.NewTool("list_splunk_alerts",
		mcp.WithDescription("List all Splunk alerts (saved searches with actions). Supports pagination and optional case-insensitive title filter."),
		mcp.WithNumber("count", mcp.Description("Number of results to return (max 100, default 10)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (default 0)")),
		mcp.WithString("title", mcp.Description("Case-insensitive substring to filter alert titles (optional)")),
	)

	s.AddTool(alertsAllTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := splunk.NewClient(baseURL, authToken)
		count := 10
		offset := 0
		title := ""
		if v, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(v)
			if count > 100 {
				count = 100
			}
		}
		if v, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(v)
		}
		if v, ok := request.Params.Arguments["title"].(string); ok {
			title = v
		}
		alerts, total, err := client.GetAlerts(ctx, count, offset, title)
		if err != nil {
			return mcp.NewToolResultError("failed to get alerts: " + err.Error()), nil
		}
		note := fmt.Sprintf("Showing up to %d alerts (as requested). Use 'offset' to paginate. Maximum per call is 100.", count)
		result := map[string]interface{}{
			"alerts": alerts,
			"count":  count,
			"offset": offset,
			"total":  total,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
		}
		return mcp.NewToolResultText(note + "\n\n" + string(data)), nil
	})

	//////////////////////
	// INDEXES //
	//////////////////////
	indexesTool := mcp.NewTool("list_splunk_indexes",
		mcp.WithDescription("List Splunk indexes (paginated by count and offset arguments)"),
		mcp.WithNumber("count", mcp.Description("Number of results to return (max 100, default 10)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (default 0)")),
	)

	s.AddTool(indexesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := splunk.NewClient(baseURL, authToken)
		count := 10
		offset := 0
		if v, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(v)
			if count > 100 {
				count = 100
			}
		}
		if v, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(v)
		}

		indexes, total, err := client.GetIndexes(ctx, count, offset)
		if err != nil {
			return mcp.NewToolResultError("failed to get indexes: " + err.Error()), nil
		}

		note := fmt.Sprintf("Showing up to %d indexes (as requested). Use 'offset' to paginate. Maximum per call is 100.", count)
		result := map[string]interface{}{
			"indexes": indexes,
			"count":   count,
			"offset":  offset,
			"total":   total,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
		}
		return mcp.NewToolResultText(note + "\n\n" + string(data)), nil
	})

	//////////////////////
	// MACROS //
	//////////////////////
	macrosTool := mcp.NewTool("list_splunk_macros",
		mcp.WithDescription("List Splunk macros (paginated by count and offset arguments)."),
		mcp.WithNumber("count", mcp.Description("Number of results to return (max 100, default 10)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination (default 0)")),
	)

	s.AddTool(macrosTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client := splunk.NewClient(baseURL, authToken)
		count := 10
		offset := 0
		if v, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(v)
			if count > 100 {
				count = 100
			}
		}
		if v, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(v)
		}

		macros, total, err := client.GetMacros(ctx, count, offset)
		if err != nil {
			return mcp.NewToolResultError("failed to get macros: " + err.Error()), nil
		}

		note := fmt.Sprintf("Showing up to %d macros (as requested). Use 'offset' to paginate. Maximum per call is 100.", count)
		result := map[string]interface{}{
			"macros": macros,
			"count":  count,
			"offset": offset,
			"total":  total,
		}
		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
		}
		return mcp.NewToolResultText(note + "\n\n" + string(data)), nil
	})

	// Start the server
	if *transport == "sse" {
		addr := fmt.Sprintf(":%d", *port)
		log.Printf("Starting SSE server on %s", addr)
		sseServer := server.NewSSEServer(s)
		if err := sseServer.Start(addr); err != nil {
			log.Fatalf("SSE server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}
}
