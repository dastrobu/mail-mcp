package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/find_messages.js
var findMessagesScript string

// FindMessagesInput defines input parameters for find_messages tool
type FindMessagesInput struct {
	Account     string   `json:"account" jsonschema:"Name of the email account" long:"account" description:"Name of the email account"`
	MailboxPath []string `json:"mailboxPath" jsonschema:"Mailbox path array (e.g., ['Inbox'] or ['Inbox', 'GitHub']). Note: Mailbox names are case-sensitive." long:"mailbox-path" description:"Mailbox path (can be specified multiple times for nested mailboxes). Note: Mailbox names are case-sensitive."`
	Subject     string   `json:"subject,omitempty" jsonschema:"Filter by subject (substring match)" long:"subject" description:"Filter by subject (substring match)"`
	Sender      string   `json:"sender,omitempty" jsonschema:"Filter by sender email address (substring match)" long:"sender" description:"Filter by sender email address (substring match)"`
	ReadStatus  *bool    `json:"readStatus,omitempty" jsonschema:"Filter by read status (true for read, false for unread)" long:"read-status" description:"Filter by read status (true for read, false for unread)"`
	FlaggedOnly bool     `json:"flaggedOnly,omitempty" jsonschema:"Filter for flagged messages only" long:"flagged-only" description:"Filter for flagged messages only"`
	DateAfter   string   `json:"dateAfter,omitempty" jsonschema:"Filter for messages received after this ISO date (e.g., '2024-01-01T00:00:00Z')" long:"date-after" description:"Filter for messages received after this ISO date (e.g., '2024-01-01T00:00:00Z')"`
	DateBefore  string   `json:"dateBefore,omitempty" jsonschema:"Filter for messages received before this ISO date (e.g., '2024-12-31T23:59:59Z')" long:"date-before" description:"Filter for messages received before this ISO date (e.g., '2024-12-31T23:59:59Z')"`
	Limit       int      `json:"limit,omitempty" jsonschema:"Maximum number of messages to return (1-1000, default: 50)" long:"limit" description:"Maximum number of messages to return (1-1000, default: 50)"`
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
		HandleFindMessages,
	)
}

func HandleFindMessages(ctx context.Context, request *mcp.CallToolRequest, input FindMessagesInput) (*mcp.CallToolResult, any, error) {
	// Apply default limit
	if input.Limit == 0 {
		input.Limit = 50
	}

	// Validate limit
	if input.Limit < 1 || input.Limit > 1000 {
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

	// Marshal input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	data, err := jxa.Execute(ctx, findMessagesScript, string(inputJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute find_messages: %w", err)
	}

	return nil, data, nil
}
