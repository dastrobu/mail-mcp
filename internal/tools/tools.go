package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// RegisterAll registers all available tools with the MCP server
func RegisterAll(srv *mcp.Server) {
	RegisterListAccounts(srv)
	RegisterListMailboxes(srv)
	RegisterGetMessageContent(srv)
	RegisterGetSelectedMessages(srv)
	RegisterReplyToMessage(srv)
	RegisterCreateDraft(srv)
	RegisterReplaceDraft(srv)
}
