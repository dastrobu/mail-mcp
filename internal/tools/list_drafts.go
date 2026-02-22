package tools

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/list_drafts.js
var listDraftsScript string

// ListDraftsInput defines input parameters for list_drafts tool
type ListDraftsInput struct {
	Account string `json:"account" jsonschema:"Name of the email account" long:"account" description:"Name of the email account"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Maximum number of drafts to return (1-1000, default: 50)" long:"limit" description:"Maximum number of drafts to return (1-1000, default: 50)"`
}

// RegisterListDrafts registers the list_drafts tool with the MCP server
func RegisterListDrafts(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_drafts",
			Description: "Lists draft messages from the Drafts mailbox for a specific account. Returns Message.id() values for persistent drafts saved in the Drafts mailbox. These are different from OutgoingMessage objects. Use list_outgoing_messages to see in-memory drafts instead.",
			InputSchema: GenerateSchema[ListDraftsInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "List Draft Messages",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		HandleListDrafts,
	)
}

func HandleListDrafts(ctx context.Context, request *mcp.CallToolRequest, input ListDraftsInput) (*mcp.CallToolResult, any, error) {
	// Apply default limit
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	// Validate limit
	if limit < 1 || limit > 1000 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 1000")
	}

	data, err := jxa.Execute(ctx, listDraftsScript,
		input.Account,
		fmt.Sprintf("%d", limit))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute list_drafts: %w", err)
	}

	return nil, data, nil
}
