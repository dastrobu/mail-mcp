package tools

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/get_selected_messages.js
var getSelectedMessagesScript string

// GetSelectedMessagesInput defines input parameters for get_selected_messages tool
type GetSelectedMessagesInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum number of messages to return (1-100, default 5)" long:"limit" description:"Maximum number of messages to return (1-100, default 5)"`
}

// RegisterGetSelectedMessages registers the get_selected_messages tool with the MCP server
func RegisterGetSelectedMessages(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_selected_messages",
			Description: "Gets the currently selected message(s) in Mail.app.",
			InputSchema: GenerateSchema[GetSelectedMessagesInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Get Selected Messages",
				ReadOnlyHint:    true,
				IdempotentHint:  false, // Selection can change between calls
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		HandleGetSelectedMessages,
	)
}

func HandleGetSelectedMessages(ctx context.Context, request *mcp.CallToolRequest, input GetSelectedMessagesInput) (*mcp.CallToolResult, any, error) {
	// Apply default for limit if not specified
	limit := input.Limit
	if limit == 0 {
		limit = 5 // default
	}

	data, err := jxa.Execute(ctx, getSelectedMessagesScript, fmt.Sprintf("%d", limit))
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}
