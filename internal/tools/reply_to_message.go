package tools

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/reply_to_message.js
var replyToMessageScript string

// ReplyToMessageInput defines input parameters for reply_to_message tool
type ReplyToMessageInput struct {
	Account       string `json:"account" jsonschema:"Name of the email account"`
	Mailbox       string `json:"mailbox" jsonschema:"Name of the mailbox containing the message to reply to"`
	MessageID     int    `json:"message_id" jsonschema:"The unique ID of the message to reply to"`
	ReplyContent  string `json:"reply_content" jsonschema:"The content/body of the reply message. Mail.app automatically includes the quoted original message."`
	OpeningWindow *bool  `json:"opening_window,omitempty" jsonschema:"Whether to show the window for the reply message. Default is false."`
	ReplyToAll    *bool  `json:"reply_to_all,omitempty" jsonschema:"Whether to reply to all recipients. Default is false (reply to sender only)."`
}

// RegisterReplyToMessage registers the reply_to_message tool with the MCP server
func RegisterReplyToMessage(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "reply_to_message",
			Description: "Creates a reply to a specific message and saves it as a draft in the Drafts mailbox. Mail.app automatically includes the quoted original message. The reply is not sent automatically - it remains in drafts for review and manual sending. WARNING: Do not use this tool to reply to draft messages (messages in the Drafts mailbox) as it will crash Mail.app. Use replace_draft to modify drafts instead. PERFORMANCE NOTE: For very large mailboxes (1000+ messages), searching for a message by ID can be slow. The tool searches only the first 1000 messages. Use smaller mailboxes or more recent messages for better performance.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Reply to Message (Draft)",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleReplyToMessage,
	)
}

func handleReplyToMessage(ctx context.Context, request *mcp.CallToolRequest, input ReplyToMessageInput) (*mcp.CallToolResult, any, error) {
	// Apply defaults for optional parameters
	openingWindow := false
	if input.OpeningWindow != nil {
		openingWindow = *input.OpeningWindow
	}

	replyToAll := false
	if input.ReplyToAll != nil {
		replyToAll = *input.ReplyToAll
	}

	data, err := jxa.Execute(ctx, replyToMessageScript,
		input.Account,
		input.Mailbox,
		fmt.Sprintf("%d", input.MessageID),
		input.ReplyContent,
		fmt.Sprintf("%t", openingWindow),
		fmt.Sprintf("%t", replyToAll))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute reply_to_message: %w", err)
	}

	return nil, data, nil
}
