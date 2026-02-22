package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/apple-mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/list_mailboxes.js
var listMailboxesScript string

// ListMailboxesInput defines input parameters for list_mailboxes tool
type ListMailboxesInput struct {
	Account     string   `json:"account" jsonschema:"Name of the email account" long:"account" description:"Name of the email account"`
	MailboxPath []string `json:"mailboxPath,omitempty" jsonschema:"Optional path to a mailbox to list its sub-mailboxes (e.g. ['Inbox'] to list mailboxes under Inbox). If omitted, lists top-level mailboxes. Note: Mailbox names are case-sensitive." long:"mailbox-path" description:"Optional path to a mailbox to list its sub-mailboxes (e.g. Inbox to list mailboxes under Inbox). Can be specified multiple times for nested paths."`
}

// RegisterListMailboxes registers the list_mailboxes tool with the MCP server
func RegisterListMailboxes(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_mailboxes",
			Description: "Lists mailboxes (folders) for a specific account in Apple Mail. By default lists top-level mailboxes. Optionally provide mailboxPath to list sub-mailboxes of a specific mailbox. Returns mailboxPath for each mailbox to support nested mailbox navigation.",
			InputSchema: GenerateSchema[ListMailboxesInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "List Mailboxes",
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: new(false),
				OpenWorldHint:   new(true),
			},
		},
		HandleListMailboxes,
	)
}

func HandleListMailboxes(ctx context.Context, request *mcp.CallToolRequest, input ListMailboxesInput) (*mcp.CallToolResult, any, error) {
	// Ensure mailboxPath is never nil - use empty array instead
	mailboxPath := input.MailboxPath
	if mailboxPath == nil {
		mailboxPath = []string{}
	}

	// Marshal mailboxPath to JSON for passing to JXA script
	mailboxPathJSON, err := json.Marshal(mailboxPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
	}

	data, err := jxa.Execute(ctx, listMailboxesScript, input.Account, string(mailboxPathJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute list_mailboxes: %w", err)
	}

	return nil, data, nil
}
