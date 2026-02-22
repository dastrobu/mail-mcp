package tools

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/list_accounts.js
var listAccountsScript string

// ListAccountsInput defines input parameters for list_accounts tool
type ListAccountsInput struct {
	Enabled bool `json:"enabled" long:"enabled" description:"Filter by enabled status"`
}

// RegisterListAccounts registers the list_accounts tool with the MCP server
func RegisterListAccounts(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_accounts",
			Description: "Lists all configured email accounts in Apple Mail with their properties.",
			InputSchema: GenerateSchema[ListAccountsInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "List Mail Accounts",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		HandleListAccounts,
	)
}

func HandleListAccounts(ctx context.Context, request *mcp.CallToolRequest, input ListAccountsInput) (*mcp.CallToolResult, any, error) {
	// Execute JXA script with enabled filter
	enabledStr := "false"
	if input.Enabled {
		enabledStr = "true"
	}

	data, err := jxa.Execute(ctx, listAccountsScript, enabledStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute list_accounts: %w", err)
	}

	return nil, data, nil
}
