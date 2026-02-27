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

func RegisterReplaceOutgoingMessage(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_outgoing_message",
			Description: "Replaces an outgoing message (draft or open window) with new content. Deletes the old message, creates a new one with updated properties, and pastes new content. NOTE: Mail.app may auto-save this message as a draft. If replacing this message again, check for and delete the old outgoing message first.",
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
			return HandleReplaceOutgoingMessage(ctx, request, input)
		},
	)
}

func HandleReplaceOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input ReplaceOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
	// 1. Input Validation and Setup
	if input.OutgoingID == 0 {
		return nil, nil, fmt.Errorf("outgoing_id is required")
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

	// 3. Execute JXA to replace the message
	resultAny, err := jxa.Execute(ctx, replaceOutgoingMessageScript, string(inputJSON))
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

	// 5. Paste content into the new message window
	if err := mac.PasteIntoWindow(ctx, int(mailPID), resultSubject, 5*time.Second, htmlContent, plainContent); err != nil {
		return nil, nil, fmt.Errorf("accessibility paste operation failed: %w", err)
	}

	time.Sleep(250 * time.Millisecond) // Allow Mail.app to process the paste event.

	// 6. Return success
	finalResult := map[string]any{
		"outgoing_id": newOutgoingID,
		"subject":     resultSubject,
		"message":     "Outgoing message replaced and content pasted.",
	}

	return nil, finalResult, nil
}
