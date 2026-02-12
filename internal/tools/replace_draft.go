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

//go:embed scripts/replace_draft.js
var replaceDraftScript string

// ReplaceDraftInput defines input parameters for replace_draft tool
type ReplaceDraftInput struct {
	DraftID       int      `json:"draft_id" jsonschema:"The ID of the draft message to edit (from Drafts mailbox)"`
	Subject       string   `json:"subject,omitempty" jsonschema:"New subject line. Leave empty to keep existing subject"`
	Content       string   `json:"content,omitempty" jsonschema:"New body text. Leave empty to keep existing content. Note: This replaces ALL existing content"`
	ToRecipients  []string `json:"to_recipients,omitempty" jsonschema:"New list of To recipients. Leave empty to keep existing. Provide empty array to clear all"`
	CcRecipients  []string `json:"cc_recipients,omitempty" jsonschema:"New list of CC recipients. Leave empty to keep existing. Provide empty array to clear all"`
	BccRecipients []string `json:"bcc_recipients,omitempty" jsonschema:"New list of BCC recipients. Leave empty to keep existing. Provide empty array to clear all"`
	Sender        string   `json:"sender,omitempty" jsonschema:"New sender email address. Leave empty to keep existing sender"`
	OpeningWindow *bool    `json:"opening_window,omitempty" jsonschema:"Whether to show the compose window. Default is false"`
}

// RegisterReplaceDraft registers the replace_draft tool with the MCP server
func RegisterReplaceDraft(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "replace_draft",
			Description: "Replaces an existing email draft by deleting it and creating a new one with updated properties. Note: The old draft is deleted and a new draft is created, so the draft_id will change and threading headers (In-Reply-To, References) will be lost if the original was a reply. Content replacement is complete - any HTML formatting from the original draft will be replaced with plain text. Recipients are updated automatically.",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Replace Email Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleReplaceDraft,
	)
}

func handleReplaceDraft(ctx context.Context, request *mcp.CallToolRequest, input ReplaceDraftInput) (*mcp.CallToolResult, any, error) {
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

	data, err := jxa.Execute(ctx, replaceDraftScript,
		fmt.Sprintf("%d", input.DraftID),
		subject,
		input.Content,
		toRecipientsJSON,
		ccRecipientsJSON,
		bccRecipientsJSON,
		input.Sender,
		fmt.Sprintf("%t", openingWindow))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute replace_draft: %w", err)
	}

	return nil, data, nil
}
