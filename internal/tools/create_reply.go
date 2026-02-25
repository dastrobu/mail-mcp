// Package tools implements the MCP tools that form the core functionality of
// the server, allowing programmatic interaction with the macOS Mail.app.
package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dastrobu/mail-mcp/internal/jxa"
	"github.com/dastrobu/mail-mcp/internal/mac"
	"github.com/dastrobu/mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/create_reply.js
var createReplyScript string

type CreateReplyInput struct {
	MessageID     int      `json:"message_id" jsonschema:"The ID of the message to reply to" long:"message-id" description:"The ID of the message to reply to"`
	Account       string   `json:"account" jsonschema:"The name of the account the original message is in" long:"account" description:"The name of the account the original message is in"`
	MailboxPath   []string `json:"mailbox_path" jsonschema:"The full path to the mailbox of the original message (e.g., [\"Inbox\", \"Subfolder\"])" long:"mailbox-path" description:"The full path to the mailbox of the original message (e.g., [\"Inbox\", \"Subfolder\"]). Can be specified multiple times."`
	Content       string   `json:"content" jsonschema:"Email body content for the reply. Supports Markdown formatting." long:"content" description:"Email body content for the reply. Supports Markdown formatting."`
	ContentFormat *string  `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ReplyToAll    bool     `json:"reply_to_all,omitempty" jsonschema:"Reply to all recipients. Default is false." long:"reply-to-all" description:"Reply to all recipients. Default is false."`
}

func RegisterCreateReply(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_reply",
			Description: "Creates a reply to a specific message, opens it as a new window, and pastes in content. Returns the new Outgoing Message ID. NOTE: Mail.app may auto-save this message as a draft. If replacing this reply, check for and delete the old outgoing message first.",
			InputSchema: GenerateSchema[CreateReplyInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Reply",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true), // Creates a new message window
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyInput) (*mcp.CallToolResult, any, error) {
			return HandleCreateReply(ctx, request, input, richtextConfig)
		},
	)
}

func HandleCreateReply(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation and Setup
	if input.Account == "" || input.MessageID == 0 || input.Content == "" || len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("account, message_id, content, and mailbox_path are required")
	}
	contentFormat, err := ValidateAndNormalizeContentFormat(input.ContentFormat)
	if err != nil {
		return nil, nil, err
	}
	if err := mac.EnsureAccessibility(); err != nil {
		return nil, nil, err
	}

	// 2. Prepare content for clipboard and JXA
	htmlContent, plainContent, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}

	// 3. Execute JXA to create the reply
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	resultAny, err := jxa.Execute(ctx, createReplyScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("JXA execution failed: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	// 4. Extract data for pasting
	outgoingID, idOk := resultMap["outgoing_id"].(float64)
	resultSubject, subjectOk := resultMap["subject"].(string)
	mailPID, pidOk := resultMap["pid"].(float64)

	if !idOk || !subjectOk || !pidOk {
		return nil, nil, fmt.Errorf("JXA result is missing required fields (outgoing_id, subject, pid)")
	}

	// 5. Paste content
	if err := mac.PasteIntoWindow(ctx, int(mailPID), resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}
	time.Sleep(250 * time.Millisecond)

	// 6. Return success
	finalResult := map[string]any{
		"outgoing_id": outgoingID,
		"subject":     resultSubject,
		"message":     "Reply created and content pasted.",
	}

	return nil, finalResult, nil
}
