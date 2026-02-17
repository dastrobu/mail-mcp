package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/create_outgoing_message.js
var createOutgoingMessageScript string

// ContentFormat constants for email content formatting
const (
	ContentFormatPlain    = "plain"
	ContentFormatMarkdown = "markdown"
	// Default content format
	ContentFormatDefault = ContentFormatMarkdown
)

// CreateOutgoingMessageInput defines input parameters for create_outgoing_message tool
type CreateOutgoingMessageInput struct {
	Subject       string   `json:"subject" jsonschema:"Subject line of the email"`
	Content       string   `json:"content" jsonschema:"Email body content (supports Markdown formatting: headings, bold, italic, code blocks, blockquotes, lists, links, horizontal rules. Tables and Mermaid diagrams are not supported)"`
	ContentFormat string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'"`
	ToRecipients  []string `json:"to_recipients" jsonschema:"List of To recipient email addresses"`
	CcRecipients  []string `json:"cc_recipients,omitempty" jsonschema:"List of CC recipient email addresses (optional)"`
	BccRecipients []string `json:"bcc_recipients,omitempty" jsonschema:"List of BCC recipient email addresses (optional)"`
	Sender        string   `json:"sender,omitempty" jsonschema:"Sender email address (optional, uses default account if omitted)"`
	OpeningWindow *bool    `json:"opening_window,omitempty" jsonschema:"Whether to show the compose window. Default is false"`
}

// RegisterCreateOutgoingMessage registers the create_outgoing_message tool with the MCP server
func RegisterCreateOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_outgoing_message",
			Description: "Creates a new outgoing email message with optional Markdown formatting and returns its OutgoingMessage ID immediately (no delay). The message is saved but not sent. Use replace_outgoing_message to modify it. Returns OutgoingMessage.id() which works with replace_outgoing_message. Note: The OutgoingMessage only exists in memory while Mail.app is running. If you need persistent drafts that survive Mail.app restart, use reply_to_message instead.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
			return handleCreateOutgoingMessage(ctx, request, input, richtextConfig)
		},
	)
}

func handleCreateOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input CreateOutgoingMessageInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// Trim and validate subject
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		return nil, nil, fmt.Errorf("subject is required and cannot be empty or whitespace-only")
	}

	// Validate content
	if input.Content == "" {
		return nil, nil, fmt.Errorf("content is required")
	}

	// Apply defaults for optional parameters
	openingWindow := false
	if input.OpeningWindow != nil {
		openingWindow = *input.OpeningWindow
	}

	// Determine content format (default to markdown)
	contentFormat := strings.ToLower(strings.TrimSpace(input.ContentFormat))
	if contentFormat == "" {
		contentFormat = ContentFormatDefault
	}

	// Process content based on format
	var contentJSON string
	switch contentFormat {
	case ContentFormatMarkdown:
		// Parse Markdown and convert to styled blocks
		doc, err := richtext.ParseMarkdown([]byte(input.Content))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse Markdown: %w", err)
		}

		styledBlocks, err := richtext.ConvertMarkdownToStyledBlocks(doc, []byte(input.Content), richtextConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert Markdown to styled blocks: %w", err)
		}

		// Encode styled blocks as JSON
		encoded, err := json.Marshal(styledBlocks)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode styled blocks: %w", err)
		}
		contentJSON = string(encoded)

	case ContentFormatPlain:
		// Plain text - just pass the content directly
		contentJSON = ""

	default:
		return nil, nil, fmt.Errorf("invalid content_format: %s (must be '%s' or '%s')", contentFormat, ContentFormatPlain, ContentFormatMarkdown)
	}

	// Encode recipient arrays as JSON strings
	toRecipientsJSON, err := json.Marshal(input.ToRecipients)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode To recipients: %w", err)
	}

	ccRecipientsJSON := ""
	if len(input.CcRecipients) > 0 {
		encoded, err := json.Marshal(input.CcRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode CC recipients: %w", err)
		}
		ccRecipientsJSON = string(encoded)
	}

	bccRecipientsJSON := ""
	if len(input.BccRecipients) > 0 {
		encoded, err := json.Marshal(input.BccRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode BCC recipients: %w", err)
		}
		bccRecipientsJSON = string(encoded)
	}

	data, err := jxa.Execute(ctx, createOutgoingMessageScript,
		subject,
		input.Content,
		contentFormat,
		contentJSON,
		string(toRecipientsJSON),
		ccRecipientsJSON,
		bccRecipientsJSON,
		input.Sender,
		fmt.Sprintf("%t", openingWindow))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute create_outgoing_message: %w", err)
	}

	return nil, data, nil
}
