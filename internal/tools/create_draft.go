package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/create_draft.js
var createDraftScript string

// CreateDraftInput defines input parameters for create_draft tool
type CreateDraftInput struct {
	Subject       string   `json:"subject" jsonschema:"The subject line of the email"`
	Content       string   `json:"content" jsonschema:"The body text of the email"`
	ToRecipients  []string `json:"to_recipients,omitempty" jsonschema:"List of To recipient email addresses"`
	CcRecipients  []string `json:"cc_recipients,omitempty" jsonschema:"List of CC recipient email addresses"`
	BccRecipients []string `json:"bcc_recipients,omitempty" jsonschema:"List of BCC recipient email addresses"`
	Sender        string   `json:"sender,omitempty" jsonschema:"Sender email address. If not provided, the default account will be used"`
	OpeningWindow *bool    `json:"opening_window,omitempty" jsonschema:"Whether to show the compose window. Default is false"`
}

// RegisterCreateDraft registers the create_draft tool with the MCP server
func RegisterCreateDraft(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "create_draft",
			Description: "Creates a new email draft with the specified subject, content, and recipients. The draft is automatically saved to the Drafts mailbox and is not sent automatically. This tool creates a brand new email message (not a reply to an existing message). Recipients are added automatically to the draft.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Create Email Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleCreateDraft,
	)
}

func handleCreateDraft(ctx context.Context, request *mcp.CallToolRequest, input CreateDraftInput) (*mcp.CallToolResult, any, error) {
	// Trim subject to avoid Mail.app search issues with whitespace
	subject := strings.TrimSpace(input.Subject)

	// Validate subject is not empty or whitespace-only
	if subject == "" {
		return nil, nil, fmt.Errorf("subject cannot be empty or whitespace-only")
	}

	// Apply defaults for optional parameters
	openingWindow := false
	if input.OpeningWindow != nil {
		openingWindow = *input.OpeningWindow
	}

	// Encode recipient arrays as JSON strings
	toRecipientsJSON := "[]"
	if len(input.ToRecipients) > 0 {
		encoded, err := json.Marshal(input.ToRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode To recipients: %w", err)
		}
		toRecipientsJSON = string(encoded)
	}

	ccRecipientsJSON := "[]"
	if len(input.CcRecipients) > 0 {
		encoded, err := json.Marshal(input.CcRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode CC recipients: %w", err)
		}
		ccRecipientsJSON = string(encoded)
	}

	bccRecipientsJSON := "[]"
	if len(input.BccRecipients) > 0 {
		encoded, err := json.Marshal(input.BccRecipients)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode BCC recipients: %w", err)
		}
		bccRecipientsJSON = string(encoded)
	}

	data, err := jxa.Execute(ctx, createDraftScript,
		subject,
		input.Content,
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		input.Sender,
		fmt.Sprintf("%t", openingWindow))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute create_draft: %w", err)
	}

	return nil, data, nil
}
