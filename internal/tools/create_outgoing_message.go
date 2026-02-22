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

//go:embed scripts/create_outgoing_message.js
var createOutgoingMessageScript string

type CreateOutgoingMessageInput struct {
	Account       string    `json:"account" jsonschema:"The name of the account to send from" long:"account" description:"The name of the account to send from"`
	Subject       string    `json:"subject" jsonschema:"Subject line of the email" long:"subject" description:"Subject line of the email"`
	Content       string    `json:"content" jsonschema:"Email body content. Supports Markdown formatting." long:"content" description:"Email body content. Supports Markdown formatting."`
	ContentFormat *string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ToRecipients  *[]string `json:"to_recipients,omitempty" jsonschema:"List of To recipients" long:"to-recipients" description:"List of To recipients. Can be specified multiple times."`
	CcRecipients  *[]string `json:"cc_recipients,omitempty" jsonschema:"List of CC recipients" long:"cc-recipients" description:"List of CC recipients. Can be specified multiple times."`
	BccRecipients *[]string `json:"bcc_recipients,omitempty" jsonschema:"List of BCC recipients" long:"bcc-recipients" description:"List of BCC recipients. Can be specified multiple times."`
	Sender        *string   `json:"sender,omitempty" jsonschema:"The sender email address (optional, overrides account default)" long:"sender" description:"The sender email address (optional, overrides account default)"`
}

func RegisterCreateOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_outgoing_message",
			Description: "Creates a new outgoing message (draft) and pastes content at the top of the message body using the Accessibility API. Returns the new Draft ID.",
			InputSchema: GenerateSchema[CreateOutgoingMessageInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
			return HandleCreateOutgoingMessage(ctx, request, input, richtextConfig)
		},
	)
}

func HandleCreateOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation
	if input.Account == "" || input.Subject == "" || input.Content == "" {
		return nil, nil, fmt.Errorf("account, subject, and content are required")
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
	htmlContent, plainContent, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}
	toRecipientsJSON, err := json.Marshal(input.ToRecipients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal to-recipients: %w", err)
	}
	ccRecipientsJSON, err := json.Marshal(input.CcRecipients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal cc-recipients: %w", err)
	}
	bccRecipientsJSON, err := json.Marshal(input.BccRecipients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal bcc-recipients: %w", err)
	}

	// 3. Execute JXA to create the draft window
	sender := ""
	if input.Sender != nil {
		sender = *input.Sender
	}

	resultAny, err := jxa.Execute(ctx, createOutgoingMessageScript,
		input.Subject,
		string(toRecipientsJSON),
		string(ccRecipientsJSON),
		string(bccRecipientsJSON),
		input.Account,
		sender)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create outgoing message: %w", err)
	}
	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format from create_outgoing_message.js")
	}
	draftID, _ := resultMap["draft_id"].(float64)
	resultSubject, _ := resultMap["subject"].(string)

	// 4. Fire-and-forget paste operation using Accessibility API
	if err := mac.PasteIntoWindow(ctx, mailPID, resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}

	// Give Mail.app a moment to process the paste event before we exit.
	time.Sleep(250 * time.Millisecond)

	// 5. Return success without verification
	finalResult := map[string]any{
		"draft_id": draftID,
		"subject":  resultSubject,
		"message":  "Draft created and content pasted via Accessibility API. Note: Paste success is not verified.",
	}
	return nil, finalResult, nil
}
