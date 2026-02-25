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

//go:embed scripts/delete_draft.js
var deleteDraftScript string

type DeleteDraftInput struct {
	DraftID int `json:"draft_id" jsonschema:"The ID of the draft to delete" long:"draft-id" description:"The ID of the draft to delete"`
}

func RegisterDeleteDraft(srv *mcp.Server) {
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "delete_draft",
			Description: "Deletes a draft message by its ID. This action is irreversible.",
			InputSchema: GenerateSchema[DeleteDraftInput](),
			Annotations: &mcp.ToolAnnotations{
				Title:           "Delete Draft",
				ReadOnlyHint:    false,
				IdempotentHint:  false,
				DestructiveHint: new(true),
				OpenWorldHint:   new(true),
			},
		},
		HandleDeleteDraft,
	)
}

func HandleDeleteDraft(ctx context.Context, request *mcp.CallToolRequest, input DeleteDraftInput) (*mcp.CallToolResult, any, error) {
	// Prepare arguments for JXA
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal input for JXA: %w", err)
	}

	// Execute JXA
	data, err := jxa.Execute(ctx, deleteDraftScript, string(inputJSON))
	if err != nil {
		return nil, nil, err
	}

	return nil, data, nil
}
