package tools

import (
	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all available tools with the MCP server
func RegisterAll(srv *mcp.Server, richtextConfig *richtext.PreparedConfig) {
	RegisterListAccounts(srv)
	RegisterListMailboxes(srv)
	RegisterGetMessageContent(srv)
	RegisterGetSelectedMessages(srv)
	RegisterReplyToMessage(srv, richtextConfig)
	RegisterListDrafts(srv)
	RegisterCreateOutgoingMessage(srv, richtextConfig)
	RegisterListOutgoingMessages(srv)
	RegisterReplaceOutgoingMessage(srv, richtextConfig)
	RegisterFindMessages(srv)
}
