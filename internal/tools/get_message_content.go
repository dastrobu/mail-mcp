package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/get_message_content.js
var getMessageContentScript string

// GetMessageContentInput defines input parameters for get_message_content tool
type GetMessageContentInput struct {
	Account     string   `json:"account" jsonschema:"Name of the email account" long:"account" description:"Name of the email account"`
	MailboxPath []string `json:"mailboxPath" jsonschema:"Path to the mailbox as an array (e.g. ['Inbox'] for top-level or ['Inbox','GitHub'] for nested mailbox). Use the mailboxPath field from get_selected_messages. Note: Mailbox names are case-sensitive." long:"mailbox-path" description:"Path to the mailbox. Can be specified multiple times for nested paths."`
	MessageID   int      `json:"message_id" jsonschema:"The unique ID of the message to retrieve" long:"message-id" description:"The unique ID of the message to retrieve"`
}

// RegisterGetMessageContent registers the get_message_content tool with the MCP server
func RegisterGetMessageContent(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_message_content",
			Description: "Retrieves the full content (body) of a specific message by its ID from a specific account and mailbox. Supports nested mailboxes via mailboxPath array. IMPORTANT: Use the mailboxPath field from get_selected_messages output, not the mailbox field.",
			InputSchema: GenerateSchema[GetMessageContentInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Get Message Content",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		HandleGetMessageContent,
	)
}

func HandleGetMessageContent(ctx context.Context, request *mcp.CallToolRequest, input GetMessageContentInput) (*mcp.CallToolResult, any, error) {
	// Validate mailboxPath
	if len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("mailboxPath is required and must be a non-empty array")
	}

	// Marshal input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	// Execute JXA script with input as JSON string
	data, err := jxa.Execute(ctx, getMessageContentScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute get_message_content: %w", err)
	}

	return nil, data, nil
}
