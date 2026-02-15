package tools

import (
	"context"
	_ "embed"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/list_outgoing_messages.js
var listOutgoingMessagesScript string

// RegisterListOutgoingMessages registers the list_outgoing_messages tool with the MCP server
func RegisterListOutgoingMessages(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_outgoing_messages",
			Description: "Lists all OutgoingMessage objects currently in memory in Mail.app. These are unsent messages that were created with create_outgoing_message or create_reply_draft. Returns outgoing_id for each message which can be used with replace_outgoing_message or replace_reply_draft. Note: Only shows messages in the current Mail.app session - messages are lost when Mail.app is closed or messages are sent.",
			InputSchema: GenerateSchema[struct{}](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "List Outgoing Messages",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleListOutgoingMessages,
	)
}

func handleListOutgoingMessages(ctx context.Context, request *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	data, err := jxa.Execute(ctx, listOutgoingMessagesScript)
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}
