package splunk

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterPrompts registers all MCP prompts for Splunk
func RegisterPrompts(s *server.MCPServer, client *Client) {
	s.AddPrompt(mcp.NewPrompt("bt_alerts_by_keyword",
		mcp.WithPromptDescription("List all BT_Alert alerts that reference a given keyword (e.g., OKTA, GITLAB, etc.). You must check all alerts and macros, paginating with count=100 as many times as needed to cover all results."),
		mcp.WithArgument("keyword",
			mcp.ArgumentDescription("The keyword to search for in alert titles, descriptions, or SPL (e.g., okta, gitlab, cloudflare)"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		keywordRaw, ok := request.Params.Arguments["keyword"]
		if !ok {
			return nil, fmt.Errorf("missing required argument: keyword")
		}
		keyword := fmt.Sprintf("%v", keywordRaw)
		keyword = strings.ToLower(keyword)

		// Fetch all BT_Alert alerts with pagination
		var alerts []Alert
		offset := 0
		for {
			batch, total, err := client.GetAlerts(ctx, 100, offset, "BT_Alert")
			if err != nil {
				return nil, fmt.Errorf("failed to get alerts: %w", err)
			}
			alerts = append(alerts, batch...)
			if offset+100 >= total || len(batch) == 0 {
				break
			}
			offset += 100
		}

		// Fetch all macros with pagination
		var macros []Macro
		macroOffset := 0
		for {
			batch, total, err := client.GetMacros(ctx, 100, macroOffset)
			if err != nil {
				return nil, fmt.Errorf("failed to get macros: %w", err)
			}
			macros = append(macros, batch...)
			if macroOffset+100 >= total || len(batch) == 0 {
				break
			}
			macroOffset += 100
		}
		macroMap := map[string]string{}
		for _, macro := range macros {
			macroMap[macro.Name] = macro.Definition
		}

		var matchingAlerts []Alert
		for _, alert := range alerts {
			if strings.Contains(strings.ToLower(alert.Title), keyword) ||
				strings.Contains(strings.ToLower(alert.Description), keyword) {
				matchingAlerts = append(matchingAlerts, alert)
				continue
			}

			searchLower := strings.ToLower(alert.Search)
			if strings.Contains(searchLower, keyword) {
				matchingAlerts = append(matchingAlerts, alert)
				continue
			}

			// Check for macros in the search field (e.g., `macro_name`)
			for macroName, macroDef := range macroMap {
				macroPattern := "`" + macroName + "`"
				if strings.Contains(searchLower, macroPattern) && strings.Contains(strings.ToLower(macroDef), keyword) {
					matchingAlerts = append(matchingAlerts, alert)
					break
				}
			}
		}

		var b strings.Builder
		b.WriteString(fmt.Sprintf("Found %d BT_Alert alerts referencing '%s':\n", len(matchingAlerts), keyword))
		for _, alert := range matchingAlerts {
			b.WriteString(fmt.Sprintf("- %s\n", alert.Title))
		}

		messages := []mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent("You must check all alerts and macros, paginating with count=100 as many times as needed to cover all results."),
			),
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent(b.String()),
			),
		}

		return mcp.NewGetPromptResult(
			fmt.Sprintf("BT_Alert alerts referencing '%s'", keyword),
			messages,
		), nil
	})
}
