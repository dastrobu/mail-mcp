// Package tools implements the MCP tools that form the core functionality of
// the server, allowing programmatic interaction with the macOS Mail.app.
package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/dastrobu/mail-mcp/internal/jxa"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/delete_outgoing_message.js
var deleteOutgoingMessageScript string

type DeleteOutgoingMessageInput struct {
	OutgoingID int `json:"outgoing_id" jsonschema:"The ID of the outgoing message to delete" long:"outgoing-id" description:"The ID of the outgoing message to delete"`
}

func RegisterDeleteOutgoingMessage(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "delete_outgoing_message",
			Description: "Deletes an outgoing message (draft or open composition window) by its ID. This action is irreversible.",
			InputSchema: GenerateSchema[DeleteOutgoingMessageInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Delete Outgoing Message",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		HandleDeleteOutgoingMessage,
	)
}

func HandleDeleteOutgoingMessage(ctx context.Context, request *mcp.CallToolRequest, input DeleteOutgoingMessageInput) (*mcp.CallToolResult, any, error) {
	// Prepare arguments for JXA
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	// Execute JXA
	data, err := jxa.Execute(ctx, deleteOutgoingMessageScript, string(inputJSON))
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}
