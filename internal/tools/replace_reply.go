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

//go:embed scripts/replace_reply.js
var replaceReplyScript string

type ReplaceReplyInput struct {
	OutgoingID  int      `json:"outgoing_id" jsonschema:"The ID of the outgoing reply message to replace" long:"outgoing-id" description:"The ID of the outgoing reply message to replace"`
	MessageID   int      `json:"message_id" jsonschema:"The ID of the original message to reply to" long:"message-id" description:"The ID of the original message to reply to"`
	Account     string   `json:"account" jsonschema:"The account of the original message" long:"account" description:"The account of the original message"`
	MailboxPath []string `json:"mailbox_path" jsonschema:"The mailbox path of the original message" long:"mailbox-path" description:"The mailbox path of the original message. Can be specified multiple times."`

	Content       string  `json:"content" jsonschema:"New email body content for the reply. Supports Markdown formatting." long:"content" description:"New email body content for the reply. Supports Markdown formatting."`
	ContentFormat *string `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ReplyToAll    bool    `json:"reply_to_all,omitempty" jsonschema:"Reply to all recipients. Default is false." long:"reply-to-all" description:"Reply to all recipients. Default is false."`

	// Optional overrides for the new reply
	Subject       *string   `json:"subject,omitempty" jsonschema:"New subject line (optional, keeps existing if null)" long:"subject" description:"New subject line (optional, keeps existing if null)"`
	ToRecipients  *[]string `json:"to_recipients,omitempty" jsonschema:"New list of To recipients (optional, replaces reply recipients)" long:"to-recipients" description:"New list of To recipients (optional, replaces reply recipients). Can be specified multiple times."`
	CcRecipients  *[]string `json:"cc_recipients,omitempty" jsonschema:"New list of CC recipients (optional, replaces reply recipients)" long:"cc-recipients" description:"New list of CC recipients (optional, replaces reply recipients). Can be specified multiple times."`
	BccRecipients *[]string `json:"bcc_recipients,omitempty" jsonschema:"New list of BCC recipients (optional, replaces reply recipients)" long:"bcc-recipients" description:"New list of BCC recipients (optional, replaces reply recipients). Can be specified multiple times."`
}

func RegisterReplaceReply(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_reply",
			Description: "Replaces an existing reply with new content. Deletes the old reply window, creates a new one, and pastes in the new content. NOTE: Mail.app may auto-save messages as drafts. If replacing again, use the new outgoing_id.",
			InputSchema: GenerateSchema[ReplaceReplyInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Replace Reply",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input ReplaceReplyInput) (*mcp.CallToolResult, any, error) {
			return HandleReplaceReply(ctx, request, input, richtextConfig)
		},
	)
}

func HandleReplaceReply(ctx context.Context, request *mcp.CallToolRequest, input ReplaceReplyInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation and Setup
	if input.OutgoingID == 0 || input.MessageID == 0 || input.Account == "" || len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("outgoing_id, message_id, account, and mailbox_path are required")
	}
	if err := mac.EnsureAccessibility(); err != nil {
		return nil, nil, err
	}

	contentFormat, err := ValidateAndNormalizeContentFormat(input.ContentFormat)
	if err != nil {
		return nil, nil, err
	}
	htmlContent, plainContent, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}

	// 2. Prepare arguments for JXA
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	// 3. Execute JXA to replace the reply
	resultAny, err := jxa.Execute(ctx, replaceReplyScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("JXA execution failed: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	// 4. Extract data for pasting
	newOutgoingID, idOk := resultMap["outgoing_id"].(float64)
	resultSubject, subjectOk := resultMap["subject"].(string)
	mailPID, pidOk := resultMap["pid"].(float64)

	if !idOk || !subjectOk || !pidOk {
		return nil, nil, fmt.Errorf("JXA result is missing required fields (outgoing_id, subject, pid)")
	}

	// 5. Paste content into the new reply window
	if err := mac.PasteIntoWindow(ctx, int(mailPID), resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}

	time.Sleep(250 * time.Millisecond) // Allow Mail.app to process the paste event.

	// 6. Return success
	finalResult := map[string]any{
		"outgoing_id": newOutgoingID,
		"subject":     resultSubject,
		"message":     "Reply replaced and content pasted.",
	}

	return nil, finalResult, nil
}
