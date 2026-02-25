package tools

import (
	"github.com/dastrobu/mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all available tools with the MCP server.
func RegisterAll(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	// Informational tools
	RegisterListAccounts(srv)
	RegisterListMailboxes(srv)
	RegisterGetMessageContent(srv)
	RegisterFindMessages(srv)
	RegisterGetSelectedMessages(srv)
	RegisterListOutgoingMessages(srv)
	RegisterListDrafts(srv)

	// Message creation and manipulation tools that require rich text processing
	RegisterCreateReply(srv, richtextConfig)
	RegisterReplaceReply(srv, richtextConfig)
	RegisterCreateOutgoingMessage(srv, richtextConfig)
	RegisterReplaceOutgoingMessage(srv, richtextConfig)
	RegisterDeleteOutgoingMessage(srv)
	RegisterDeleteDraft(srv)
}
