package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/list_drafts.js
var listDraftsScript string

// ListDraftsInput defines input parameters for list_drafts tool
type ListDraftsInput struct {
	Account string `json:"account,omitempty" jsonschema:"Optional: Name of the email account to filter drafts by" long:"account" description:"Optional: Name of the email account to filter drafts by"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Maximum number of drafts to return (1-1000, default: 50)" long:"limit" description:"Maximum number of drafts to return (1-1000, default: 50)"`
}

// RegisterListDrafts registers the list_drafts tool with the MCP server
func RegisterListDrafts(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_drafts",
			Description: "Lists draft messages from the global Drafts mailbox, optionally filtered by a specific account. Returns Message.id() values for persistent drafts saved in the Drafts mailbox. These are different from OutgoingMessage objects. Use list_outgoing_messages to see in-memory drafts instead.",
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
	if input.Limit == 0 {
		input.Limit = 50
	}

	// Validate limit
	if input.Limit < 1 || input.Limit > 1000 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 1000")
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	data, err := jxa.Execute(ctx, listDraftsScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute list_drafts: %w", err)
	}

	return nil, data, nil
}
