package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all available tools with the MCP server.
func RegisterAll(srv *mcp.Server) {
	// Informational tools
	RegisterListAccounts(srv)
	RegisterListMailboxes(srv)
	RegisterGetMessageContent(srv)
	RegisterFindMessages(srv)
	RegisterGetSelectedMessages(srv)
	RegisterListOutgoingMessages(srv)
	RegisterListDrafts(srv)

	// Message creation and manipulation tools
	RegisterCreateReply(srv)
	RegisterReplaceReply(srv)
	RegisterCreateOutgoingMessage(srv)
	RegisterReplaceOutgoingMessage(srv)
	RegisterDeleteOutgoingMessage(srv)
	RegisterDeleteDraft(srv)
}
