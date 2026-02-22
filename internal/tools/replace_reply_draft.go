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

//go:embed scripts/replace_reply_draft.js
var replaceReplyDraftScript string

type ReplaceReplyDraftInput struct {
	OutgoingID        int       `json:"outgoing_id" jsonschema:"The ID of the reply draft to replace" long:"outgoing-id" description:"The ID of the reply draft to replace"`
	OriginalMessageID int       `json:"original_message_id" jsonschema:"The ID of the original message being replied to. This is used to re-create the reply with a clean quote." long:"original-message-id" description:"The ID of the original message being replied to. This is used to re-create the reply with a clean quote."`
	Account           string    `json:"account" jsonschema:"The account name of the original message" long:"account" description:"The account name of the original message"`
	MailboxPath       []string  `json:"mailbox_path" jsonschema:"The mailbox path of the original message as an array" long:"mailbox-path" description:"The mailbox path of the original message. Can be specified multiple times for nested paths."`
	Subject           *string   `json:"subject,omitempty" jsonschema:"New subject line (optional, keeps existing if null)" long:"subject" description:"New subject line (optional, keeps existing if null)"`
	Content           string    `json:"content" jsonschema:"New email body content. Supports Markdown formatting." long:"content" description:"New email body content. Supports Markdown formatting."`
	ContentFormat     *string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ToRecipients      *[]string `json:"to_recipients,omitempty" jsonschema:"New list of To recipients (optional, keeps existing if null, clears if empty array)" long:"to-recipients" description:"New list of To recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	CcRecipients      *[]string `json:"cc_recipients,omitempty" jsonschema:"New list of CC recipients (optional, keeps existing if null, clears if empty array)" long:"cc-recipients" description:"New list of CC recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	BccRecipients     *[]string `json:"bcc_recipients,omitempty" jsonschema:"New list of BCC recipients (optional, keeps existing if null, clears if empty array)" long:"bcc-recipients" description:"New list of BCC recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	Sender            *string   `json:"sender,omitempty" jsonschema:"New sender email address (optional, keeps existing if null)" long:"sender" description:"New sender email address (optional, keeps existing if null)"`
}

func RegisterReplaceReplyDraft(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_reply_draft",
			Description: "Replaces an existing reply draft with new content while preserving the original message quote and signature. It achieves this by deleting the old draft and creating a fresh reply to the original message before pasting the new content. This tool should be used when you want to update a draft that was previously created as a reply. Requires Accessibility permissions.",
			InputSchema: GenerateSchema[ReplaceReplyDraftInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Replace Reply Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input ReplaceReplyDraftInput) (*mcp.CallToolResult, any, error) {
			return HandleReplaceReplyDraft(ctx, request, input, richtextConfig)
		},
	)
}

func HandleReplaceReplyDraft(ctx context.Context, request *mcp.CallToolRequest, input ReplaceReplyDraftInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	if input.OutgoingID == 0 {
		return nil, nil, fmt.Errorf("outgoing_id is required")
	}
	if input.OriginalMessageID == 0 {
		return nil, nil, fmt.Errorf("original_message_id is required")
	}
	if len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("mailbox_path is required")
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

	htmlContent, plainContent, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}

	sentinel := "__KEEP__"

	subject := sentinel
	if input.Subject != nil {
		subject = *input.Subject
	}

	sender := sentinel
	if input.Sender != nil {
		sender = *input.Sender
	}

	toRecipientsJSON := sentinel
	if input.ToRecipients != nil {
		encoded, err := json.Marshal(*input.ToRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal to-recipients: %w", err)
		}
		toRecipientsJSON = string(encoded)
	}

	ccRecipientsJSON := sentinel
	if input.CcRecipients != nil {
		encoded, err := json.Marshal(*input.CcRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal cc-recipients: %w", err)
		}
		ccRecipientsJSON = string(encoded)
	}

	bccRecipientsJSON := sentinel
	if input.BccRecipients != nil {
		encoded, err := json.Marshal(*input.BccRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal bcc-recipients: %w", err)
		}
		bccRecipientsJSON = string(encoded)
	}

	mailboxPathJSON, err := json.Marshal(input.MailboxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode mailbox path: %w", err)
	}

	resultAny, err := jxa.Execute(ctx, replaceReplyDraftScript,
		fmt.Sprintf("%d", input.OutgoingID),
		subject,
		"",
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		sender,
		fmt.Sprintf("%d", input.OriginalMessageID),
		input.Account,
		string(mailboxPathJSON))

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fresh reply draft: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}
	newOutgoingID, _ := resultMap["outgoing_id"].(float64)
	subjectResult, _ := resultMap["subject"].(string)

	// Atomically wait for the reply window, focus its body, and paste the content.
	if err := mac.PasteIntoWindow(ctx, mailPID, subjectResult, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("failed to paste content into reply window: %w", err)
	}

	finalResult := map[string]any{
		"outgoing_id":         newOutgoingID,
		"subject":             subjectResult,
		"original_message_id": input.OriginalMessageID,
		"message":             "Reply draft replaced and content pasted via Accessibility API.",
	}

	return nil, finalResult, nil
}
