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

//go:embed scripts/replace_outgoing_message.js
var replaceOutgoingMessageScript string

type ReplaceOutgoingMessageInput struct {
	OutgoingID    int       `json:"outgoing_id" jsonschema:"The ID of the outgoing message to replace" long:"outgoing-id" description:"The ID of the outgoing message to replace"`
	Subject       *string   `json:"subject,omitempty" jsonschema:"New subject line (optional, keeps existing if null)" long:"subject" description:"New subject line (optional, keeps existing if null)"`
	Content       string    `json:"content" jsonschema:"New email body content. Supports Markdown formatting." long:"content" description:"New email body content. Supports Markdown formatting."`
	ContentFormat *string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ToRecipients  *[]string `json:"to_recipients,omitempty" jsonschema:"New list of To recipients (optional, keeps existing if null, clears if empty array)" long:"to-recipients" description:"New list of To recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	CcRecipients  *[]string `json:"cc_recipients,omitempty" jsonschema:"New list of CC recipients (optional, keeps existing if null, clears if empty array)" long:"cc-recipients" description:"New list of CC recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	BccRecipients *[]string `json:"bcc_recipients,omitempty" jsonschema:"New list of BCC recipients (optional, keeps existing if null, clears if empty array)" long:"bcc-recipients" description:"New list of BCC recipients (optional, keeps existing if null, clears if empty array). Can be specified multiple times."`
	Sender        *string   `json:"sender,omitempty" jsonschema:"New sender email address (optional, keeps existing if null)" long:"sender" description:"New sender email address (optional, keeps existing if null)"`
}

func RegisterReplaceOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_outgoing_message",
			Description: "Replaces an existing outgoing message (draft) with new content using the Accessibility API. This tool is for standalone drafts (not replies). It deletes the old draft and creates a fresh instance before pasting the new content at the top, preserving the default signature. Returns the new OutgoingMessage ID.",
			InputSchema: GenerateSchema[ReplaceOutgoingMessageInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Replace Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input ReplaceOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
			return HandleReplaceOutgoingMessage(ctx, request, input, richtextConfig)
		},
	)
}

func HandleReplaceOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input ReplaceOutgoingMessageInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	if input.OutgoingID == 0 {
		return nil, nil, fmt.Errorf("outgoing_id is required")
	}

	if input.Content == "" {
		return nil, nil, fmt.Errorf("content is required")
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

	resultAny, err := jxa.Execute(ctx, replaceOutgoingMessageScript,
		fmt.Sprintf("%d", input.OutgoingID),
		subject,
		"",
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		sender)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to replace outgoing message: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	newOutgoingID, _ := resultMap["outgoing_id"].(float64)
	resultSubject, _ := resultMap["subject"].(string)

	// Atomically wait for the draft window, focus its body, and paste the content.
	if err := mac.PasteIntoWindow(ctx, mailPID, resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("failed to paste content into draft window: %w", err)
	}

	finalResult := map[string]any{
		"outgoing_id": newOutgoingID,
		"subject":     resultSubject,
		"message":     "Draft replaced and content pasted via Accessibility API.",
	}

	return nil, finalResult, nil
}
