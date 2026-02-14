package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/find_messages.js
var findMessagesScript string

// FindMessagesInput defines input parameters for find_messages tool
type FindMessagesInput struct {
	Account     string   `json:"account" jsonschema:"Name of the email account"`
	MailboxPath []string `json:"mailboxPath" jsonschema:"Mailbox path array (e.g., ['Inbox'] or ['Inbox', 'GitHub']). Note: Mailbox names are case-sensitive."`
	Subject     string   `json:"subject,omitempty" jsonschema:"Filter by subject (substring match)"`
	Sender      string   `json:"sender,omitempty" jsonschema:"Filter by sender email address (substring match)"`
	ReadStatus  *bool    `json:"readStatus,omitempty" jsonschema:"Filter by read status (true for read, false for unread)"`
	FlaggedOnly bool     `json:"flaggedOnly,omitempty" jsonschema:"Filter for flagged messages only"`
	DateAfter   string   `json:"dateAfter,omitempty" jsonschema:"Filter for messages received after this ISO date (e.g., '2024-01-01T00:00:00Z')"`
	DateBefore  string   `json:"dateBefore,omitempty" jsonschema:"Filter for messages received before this ISO date (e.g., '2024-12-31T23:59:59Z')"`
	Limit       int      `json:"limit,omitempty" jsonschema:"Maximum number of messages to return (1-1000, default: 50)"`
}

// RegisterFindMessages registers the find_messages tool with the MCP server
func RegisterFindMessages(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "find_messages",
			Description: "Find messages in a mailbox. At least one filter criterion must be specified.",
			InputSchema: GenerateSchema[FindMessagesInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Find Messages",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		handleFindMessages,
	)
}

func handleFindMessages(ctx context.Context, request *mcp.CallToolRequest, input FindMessagesInput) (*mcp.CallToolResult, any, error) {
	// Apply default limit
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	// Validate limit
	if limit < 1 || limit > 1000 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 1000")
	}

	// Validate mailbox path
	if len(input.MailboxPath) == 0 {
		return nil, nil, fmt.Errorf("mailboxPath is required")
	}

	// Require at least one filter criterion
	hasFilter := input.Subject != "" ||
		input.Sender != "" ||
		input.ReadStatus != nil ||
		input.FlaggedOnly ||
		input.DateAfter != "" ||
		input.DateBefore != ""

	if !hasFilter {
		return nil, nil, fmt.Errorf("at least one filter criterion is required (subject, sender, readStatus, flaggedOnly, dateAfter, or dateBefore)")
	}

	// Marshal mailbox path to JSON
	mailboxPathJSON, err := json.Marshal(input.MailboxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
	}

	// Build filter options JSON
	filterOptions := map[string]any{}
	if input.Subject != "" {
		filterOptions["subject"] = input.Subject
	}
	if input.Sender != "" {
		filterOptions["sender"] = input.Sender
	}
	if input.ReadStatus != nil {
		filterOptions["readStatus"] = *input.ReadStatus
	}
	if input.FlaggedOnly {
		filterOptions["flaggedOnly"] = true
	}
	if input.DateAfter != "" {
		filterOptions["dateAfter"] = input.DateAfter
	}
	if input.DateBefore != "" {
		filterOptions["dateBefore"] = input.DateBefore
	}

	filterOptionsJSON, err := json.Marshal(filterOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal filter options: %w", err)
	}

	data, err := jxa.Execute(ctx, findMessagesScript,
		input.Account,
		string(mailboxPathJSON),
		string(filterOptionsJSON),
		fmt.Sprintf("%d", limit))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute find_messages: %w", err)
	}

	return nil, data, nil
}
