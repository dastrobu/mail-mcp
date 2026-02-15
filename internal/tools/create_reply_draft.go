package tools

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/dastrobu/apple-mail-mcp/internal/mac"
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuin/goldmark"
)

//go:embed scripts/create_reply_draft.js
var createReplyDraftScript string

// CreateReplyDraftInput defines input parameters for create_reply_draft tool
type CreateReplyDraftInput struct {
	Account       string   `json:"account" jsonschema:"Name of the email account"`
	MailboxPath   []string `json:"mailboxPath" jsonschema:"Path to the mailbox as an array (e.g. ['Inbox'] for top-level or ['Inbox','GitHub'] for nested mailbox). Use the mailboxPath field from get_selected_messages. Note: Mailbox names are case-sensitive."`
	MessageID     int      `json:"message_id" jsonschema:"The unique ID of the message to reply to"`
	ReplyContent  string   `json:"reply_content" jsonschema:"Reply message content. Will be pasted as plain text."`
	ContentFormat *string  `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ReplyToAll    *bool    `json:"reply_to_all,omitempty" jsonschema:"Whether to reply to all recipients. Default is false (reply to sender only)."`
}

// RegisterCreateReplyDraft registers the create_reply_draft tool with the MCP server
func RegisterCreateReplyDraft(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_reply_draft",
			Description: "Creates a reply to a specific message using the Accessibility API. This approach preserves the original message quote and avoids blockquote wrapping issues. It requires Accessibility permissions for the apple-mail-mcp binary. This tool will ALWAYS open a Mail.app window and simulate a paste operation to insert content. IMPORTANT: Use the mailboxPath field from get_selected_messages output.",
			InputSchema: GenerateSchema[CreateReplyDraftInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Reply Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyDraftInput) (*mcp.CallToolResult, any, error) {
			return handleCreateReplyDraft(ctx, request, input, richtextConfig)
		},
	)
}

func handleCreateReplyDraft(ctx context.Context, request *mcp.CallToolRequest, input CreateReplyDraftInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation (including ContentFormat)
	if len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("mailboxPath is required and must be a non-empty array")
	}

	contentFormat, err := ValidateAndNormalizeContentFormat(input.ContentFormat)
	if err != nil {
		return nil, nil, err
	}

	// 2. Environment Checks
	if err := mac.EnsureAccessibility(); err != nil {
		return nil, nil, err
	}

	mailPID := mac.GetMailPID()
	if mailPID == 0 {
		return nil, nil, fmt.Errorf("Mail.app is not running. Please start Mail.app and try again")
	}

	// 3. Preparation
	mailboxPathJSON, err := json.Marshal(input.MailboxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
	}

	replyToAll := false
	if input.ReplyToAll != nil {
		replyToAll = *input.ReplyToAll
	}

	var contentToPaste string
	isHTML := false

	if contentFormat == ContentFormatMarkdown {
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(input.ReplyContent), &buf); err != nil {
			return nil, nil, fmt.Errorf("failed to convert markdown: %w", err)
		}
		contentToPaste = buf.String()
		isHTML = true
	} else {
		contentToPaste = input.ReplyContent
		isHTML = false
	}

	// 4. JXA Execution
	resultAny, err := jxa.Execute(ctx, createReplyDraftScript,
		input.Account,
		string(mailboxPathJSON),
		fmt.Sprintf("%d", input.MessageID),
		fmt.Sprintf("%t", replyToAll))

	if err != nil {
		return nil, nil, fmt.Errorf("failed to open reply window: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	subject, _ := resultMap["subject"].(string)
	draftID, _ := resultMap["draft_id"].(float64)

	// 5. Accessibility Operations (Paste)
	if _, err := mac.WaitForWindowFocus(ctx, mailPID, subject, 5*time.Second); err != nil {
		return nil, nil, fmt.Errorf("failed to focus reply window: %w. Cannot paste content safely", err)
	}

	if err := mac.FocusBody(mailPID); err != nil {
		return nil, nil, fmt.Errorf("failed to find or focus message body (make sure window is visible): %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := mac.PasteContent(contentToPaste, isHTML); err != nil {
		return nil, nil, fmt.Errorf("failed to paste content: %w", err)
	}

	finalResult := map[string]any{
		"draft_id":            draftID,
		"subject":             subject,
		"original_message_id": input.MessageID,
		"account":             input.Account,
		"mailbox_path":        input.MailboxPath,
		"message":             "Reply created and content pasted via Accessibility API.",
	}

	return nil, finalResult, nil
}
