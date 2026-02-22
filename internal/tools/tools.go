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
	RegisterCreateReplyDraft(srv, richtextConfig)
	RegisterReplaceReplyDraft(srv, richtextConfig)
	RegisterCreateOutgoingMessage(srv, richtextConfig)
	RegisterReplaceOutgoingMessage(srv, richtextConfig)
	RegisterListDrafts(srv)
	RegisterListOutgoingMessages(srv)
	RegisterFindMessages(srv)
}
