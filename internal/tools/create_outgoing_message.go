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
}

func RegisterCreateOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_outgoing_message",
			Description: "Creates a new outgoing message (open window), then pastes content into its body using the Accessibility API. Returns the new Outgoing Message ID. NOTE: Mail.app may auto-save this message as a draft. If replacing this message, check for and delete the old outgoing message first.",
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
	// 1. Input Validation & Setup
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

	// 2. Prepare content for clipboard and JXA
	htmlContent, plainContent, err := ToClipboardContent(input.Content, contentFormat)
	if err != nil {
		return nil, nil, err
	}

	// 3. Execute JXA to create and save the draft
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	resultAny, err := jxa.Execute(ctx, createOutgoingMessageScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("JXA execution failed: %w", err)
	}

	resultMap, ok := resultAny.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("invalid JXA result format")
	}

	// Extract data for pasting
	outgoingID, idOk := resultMap["outgoing_id"].(float64)
	resultSubject, subjectOk := resultMap["subject"].(string)
	mailPID, pidOk := resultMap["pid"].(float64)

	if !idOk || !subjectOk || !pidOk {
		return nil, nil, fmt.Errorf("JXA result is missing required fields (outgoing_id, subject, pid)")
	}

	// 4. Paste content
	if err := mac.PasteIntoWindow(ctx, int(mailPID), resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}
	time.Sleep(250 * time.Millisecond) // Allow Mail.app to process the paste event.

	// 5. Return success
	finalResult := map[string]any{
		"outgoing_id": outgoingID,
		"subject":     resultSubject,
		"message":     "Outgoing message created and content pasted. Note: Paste success is not verified.",
	}

	return nil, finalResult, nil
}
