package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/dastrobu/apple-mail-mcp/internal/mac"
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/create_outgoing_message.js
var createOutgoingMessageScript string

type CreateOutgoingMessageInput struct {
	Subject       string    `json:"subject" jsonschema:"Subject line of the email" long:"subject" description:"Subject line of the email"`
	Content       string    `json:"content" jsonschema:"Email body content. Supports Markdown formatting." long:"content" description:"Email body content. Supports Markdown formatting."`
	ContentFormat *string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'." long:"content-format" description:"Content format: 'plain' or 'markdown'. Default is 'markdown'."`
	ToRecipients  []string  `json:"to_recipients" jsonschema:"List of To recipient email addresses" long:"to-recipients" description:"List of To recipient email addresses. Can be specified multiple times."`
	CcRecipients  *[]string `json:"cc_recipients,omitempty" jsonschema:"List of CC recipient email addresses (optional)" long:"cc-recipients" description:"List of CC recipient email addresses (optional). Can be specified multiple times."`
	BccRecipients *[]string `json:"bcc_recipients,omitempty" jsonschema:"List of BCC recipient email addresses (optional)" long:"bcc-recipients" description:"List of BCC recipient email addresses (optional). Can be specified multiple times."`
	Sender        *string   `json:"sender,omitempty" jsonschema:"Sender email address (optional, uses default account if omitted)" long:"sender" description:"Sender email address (optional, uses default account if omitted)"`
}

func RegisterCreateOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_outgoing_message",
			Description: "Creates a new outgoing email message using the Accessibility API to support rich text content. The message is created and opened in a new window, and content is pasted into the body. Returns the OutgoingMessage ID immediately. The message remains in drafts.",
			InputSchema: GenerateSchema[CreateOutgoingMessageInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
			return HandleCreateOutgoingMessage(ctx, request, input, richtextConfig)
		},
	)
}

func HandleCreateOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		return nil, nil, fmt.Errorf("subject is required and cannot be empty or whitespace-only")
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
		return nil, nil, fmt.Errorf("Mail.app is not running. Please start Mail.app and try again")
	}

	contentToPaste, isHTML, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}

	toRecipientsJSON, err := json.Marshal(input.ToRecipients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode To recipients: %w", err)
	}

	ccRecipientsJSON := ""
	if input.CcRecipients != nil {
		encoded, err := json.Marshal(*input.CcRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode CC recipients: %w", err)
		}
		ccRecipientsJSON = string(encoded)
	}

	bccRecipientsJSON := ""
	if input.BccRecipients != nil {
		encoded, err := json.Marshal(*input.BccRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode BCC recipients: %w", err)
		}
		bccRecipientsJSON = string(encoded)
	}

	sender := ""
	if input.Sender != nil {
		sender = *input.Sender
	}

	resultAny, err := jxa.Execute(ctx, createOutgoingMessageScript,
		subject,
		"",
		string(toRecipientsJSON),
		ccRecipientsJSON,
		bccRecipientsJSON,
		sender)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create outgoing message: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	draftID, _ := resultMap["draft_id"].(float64)
	resultSubject, _ := resultMap["subject"].(string)

	if _, err := mac.WaitForWindowFocus(ctx, mailPID, resultSubject, 5*time.Second); err != nil {
		return nil, nil, fmt.Errorf("failed to focus compose window: %w. Cannot paste content safely", err)
	}

	if err := mac.FocusBody(mailPID); err != nil {
		return nil, nil, fmt.Errorf("failed to find or focus message body (make sure window is visible): %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := mac.PasteContent(contentToPaste, isHTML); err != nil {
		return nil, nil, fmt.Errorf("failed to paste content: %w", err)
	}

	finalResult := map[string]any{
		"draft_id": draftID,
		"subject":  resultSubject,
		"message":  "Draft created and content pasted via Accessibility API.",
	}

	return nil, finalResult, nil
}
