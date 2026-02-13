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

//go:embed scripts/replace_outgoing_message.js
var replaceOutgoingMessageScript string

// ReplaceOutgoingMessageInput defines input parameters for replace_outgoing_message tool
type ReplaceOutgoingMessageInput struct {
	OutgoingID    int      `json:"outgoing_id" jsonschema:"The ID of the OutgoingMessage to replace (from reply_to_message or create_outgoing_message)"`
	Subject       string   `json:"subject,omitempty" jsonschema:"New subject line. Leave empty to keep existing subject"`
	Content       string   `json:"content" jsonschema:"New body text (supports Markdown formatting). REQUIRED: Cannot preserve rich text from existing message"`
	ContentFormat string   `json:"content_format,omitempty" jsonschema:"Content format: 'plain' or 'markdown'. Default is 'markdown'"`
	ToRecipients  []string `json:"to_recipients,omitempty" jsonschema:"New list of To recipients. Leave empty to keep existing. Provide empty array to clear all"`
	CcRecipients  []string `json:"cc_recipients,omitempty" jsonschema:"New list of CC recipients. Leave empty to keep existing. Provide empty array to clear all"`
	BccRecipients []string `json:"bcc_recipients,omitempty" jsonschema:"New list of BCC recipients. Leave empty to keep existing. Provide empty array to clear all"`
	Sender        string   `json:"sender,omitempty" jsonschema:"New sender email address. Leave empty to keep existing sender"`
	OpeningWindow *bool    `json:"opening_window,omitempty" jsonschema:"Whether to show the compose window. Default is false"`
}

// RegisterReplaceOutgoingMessage registers the replace_outgoing_message tool with the MCP server
func RegisterReplaceOutgoingMessage(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_outgoing_message",
			Description: "Replaces an OutgoingMessage (reply draft or new draft) by deleting it and creating a new one with updated properties. Supports Markdown formatting for content. This tool works with OutgoingMessage IDs returned by reply_to_message or create_outgoing_message. Note: Only works while the OutgoingMessage is still in memory (before Mail.app is closed). The old message is deleted and a new one is created, so the outgoing_id will change.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Replace Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		func(ctx context.Context, request *mcp.CallToolRequest, input ReplaceOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
			return handleReplaceOutgoingMessage(ctx, request, input, richtextConfig)
		},
	)
}

func handleReplaceOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input ReplaceOutgoingMessageInput, richtextConfig *richtext.PreparedConfig) (*mcp.CallToolResult, any, error) {
	// Validate required content parameter
	if input.Content == "" {
		return nil, nil, fmt.Errorf("content is required (cannot preserve rich text from existing message)")
	}

	// Trim subject to avoid Mail.app search issues with whitespace
	subject := input.Subject
	if subject != "" {
		subject = strings.TrimSpace(subject)
		// Validate subject is not whitespace-only
		if subject == "" {
			return nil, nil, fmt.Errorf("subject cannot be whitespace-only")
		}
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
	// Empty string means "keep existing", non-empty JSON array means "replace"
	toRecipientsJSON := ""
	if input.ToRecipients != nil {
		encoded, err := json.Marshal(input.ToRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode To recipients: %w", err)
		}
		toRecipientsJSON = string(encoded)
	}

	ccRecipientsJSON := ""
	if input.CcRecipients != nil {
		encoded, err := json.Marshal(input.CcRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode CC recipients: %w", err)
		}
		ccRecipientsJSON = string(encoded)
	}

	bccRecipientsJSON := ""
	if input.BccRecipients != nil {
		encoded, err := json.Marshal(input.BccRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode BCC recipients: %w", err)
		}
		bccRecipientsJSON = string(encoded)
	}

	data, err := jxa.Execute(ctx, replaceOutgoingMessageScript,
		fmt.Sprintf("%d", input.OutgoingID),
		subject,
		input.Content,
		contentFormat,
		contentJSON,
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		input.Sender,
		fmt.Sprintf("%t", openingWindow))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute replace_outgoing_message: %w", err)
	}

	return nil, data, nil
}
