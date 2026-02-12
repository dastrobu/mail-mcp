package tools

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/get_message_content.js
var getMessageContentScript string

// GetMessageContentInput defines input parameters for get_message_content tool
type GetMessageContentInput struct {
	Account   string `json:"account" jsonschema:"Name of the email account"`
	Mailbox   string `json:"mailbox" jsonschema:"Name of the mailbox (e.g. INBOX, Sent, Drafts)"`
	MessageID int    `json:"message_id" jsonschema:"The unique ID of the message to retrieve"`
}

// RegisterGetMessageContent registers the get_message_content tool with the MCP server
func RegisterGetMessageContent(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_message_content",
			Description: "Retrieves the full content (body) of a specific message by its ID from a specific account and mailbox. PERFORMANCE NOTE: For very large mailboxes (1000+ messages), searching for a message by ID can be slow. The tool searches only the first 1000 messages. Use smaller mailboxes or more recent messages for better performance.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Get Message Content",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleGetMessageContent,
	)
}

func handleGetMessageContent(ctx context.Context, request *mcp.CallToolRequest, input GetMessageContentInput) (*mcp.CallToolResult, any, error) {
	data, err := jxa.Execute(ctx, getMessageContentScript, input.Account, input.Mailbox, fmt.Sprintf("%d", input.MessageID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute get_message_content: %w", err)
	}

	return nil, data, nil
}
