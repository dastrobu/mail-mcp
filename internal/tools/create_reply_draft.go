// Package tools implements the MCP tools that form the core functionality of
// the server, allowing programmatic interaction with the macOS Mail.app.
package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/dastrobu/apple-mail-mcp/internal/mac"
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/create_reply_draft.js
var createReplyDraftScript string

type CreateReplyDraftInput struct {
	AccountName    string   `json:"account_name" jsonschema:"The name of the account to reply from" long:"account" description:"The name of the account to reply from"`
	MailboxPath    []string `json:"mailbox_path" jsonschema:"Path to the mailbox containing the message (e.g., [\"Inbox\", \"My Subfolder\"])" long:"mailbox-path" description:"Path to the mailbox containing the message (e.g., [\"Inbox\", \"My Subfolder\"]). Can be specified multiple times."`
	MessageID      int      `json:"message_id" jsonschema:"The ID of the message to reply to" long:"message-id" description:"The ID of the message to reply to"`
	ReplyContent   string   `json:"reply_content" jsonschema:"The content of the reply body. Supports Markdown formatting." long:"reply-content" description:"The content of the reply body. Supports Markdown formatting."`
	ContentFormat  *string  `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ReplyToAll     *bool    `json:"reply_to_all,omitempty" jsonschema:"Set to true to reply to all recipients. Default is false." long:"reply-to-all" description:"Set to true to reply to all recipients. Default is false."`
	VipsOnlyReply  *bool    `json:"vips_only_reply,omitempty" jsonschema:"(Not implemented) Set to true to only reply to VIPs" long:"vips-only-reply" description:"(Not implemented) Set to true to only reply to VIPs"`
	InsertAsQuote  *bool    `json:"insert_as_quote,omitempty" jsonschema:"(Not implemented) Set to true to insert content as a quote" long:"insert-as-quote" description:"(Not implemented) Set to true to insert content as a quote"`
	PrependContent *bool    `json:"prepend_content,omitempty" jsonschema:"(Not implemented) Set to false to append content instead of prepending" long:"prepend-content" description:"(Not implemented) Set to false to append content instead of prepending"`
}

func RegisterCreateReplyDraft(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_reply_draft",
			Description: "Creates a reply to a specific message and pastes content at the top of the message body using the Accessibility API. Returns the new Draft ID.",
			InputSchema: GenerateSchema[CreateReplyDraftInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Reply Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyDraftInput) (*mcp.CallToolResult, any, error) {
			return HandleCreateReplyDraft(ctx, request, input, richtextConfig)
		},
	)
}

func HandleCreateReplyDraft(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyDraftInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation
	if input.AccountName == "" || input.MessageID == 0 || len(input.MailboxPath) == 0 || input.ReplyContent == "" {
		return nil, nil, fmt.Errorf("account_name, message_id, mailbox_path, and reply_content are required")
	}
	contentFormat, err := ValidateAndNormalizeContentFormat(input.ContentFormat)
	if err != nil {
		return nil, nil, err
	}
	if err := mac.EnsureAccessibility(); err != nil {
		return nil, nil, err
	}
	mailPID := mac.GetMailPID()
	if mailPID == 0 {
		return nil, nil, fmt.Errorf("mail.app is not running. Please start Mail.app and try again")
	}

	// 2. Prepare content for clipboard and JXA
	htmlContent, plainContent, err := ToClipboardContent(input.ReplyContent, contentFormat)
	if err != nil {
		return nil, nil, err
	}
	mailboxPathJSON, err := json.Marshal(input.MailboxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
	}
	replyToAll := false
	if input.ReplyToAll != nil {
		replyToAll = *input.ReplyToAll
	}

	// 3. Execute JXA to create the reply window
	resultAny, err := jxa.Execute(ctx, createReplyDraftScript, input.AccountName, string(mailboxPathJSON), fmt.Sprintf("%d", input.MessageID), fmt.Sprintf("%t", replyToAll))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create reply draft: %w", err)
	}
	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format from create_reply_draft.js")
	}
	subject, _ := resultMap["subject"].(string)
	draftID, _ := resultMap["draft_id"].(float64)

	// 4. Fire-and-forget paste operation using Accessibility API
	if err := mac.PasteIntoWindow(ctx, mailPID, subject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}

	// Give Mail.app a moment to process the paste event before we exit.
	time.Sleep(250 * time.Millisecond)

	// 5. Return success without verification
	finalResult := map[string]any{
		"draft_id":            draftID,
		"subject":             subject,
		"message":             "Reply created and content pasted via Accessibility API. Note: Paste success is not verified.",
		"original_message_id": input.MessageID,
		"account":             input.AccountName,
		"mailbox_path":        input.MailboxPath,
	}
	return nil, finalResult, nil
}
