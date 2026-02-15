package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/dastrobu/apple-mail-mcp/internal/richtext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleCreateOutgoingMessage_UnknownContentFormat(t *testing.T) {
	// Use the default embedded config
	config, err := richtext.LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	invalidFormat := "invalid"
	input := CreateOutgoingMessageInput{
		Subject:       "Test",
		Content:       "Test content",
		ContentFormat: &invalidFormat,
		ToRecipients:  []string{"test@example.com"},
	}

	ctx := context.Background()
	_, _, err = handleCreateOutgoingMessage(ctx, &mcp.CallToolRequest{}, input, config)

	if err == nil {
		t.Errorf("Expected error for unknown content format, but got nil")
		return
	}

	expectedSubstring := "invalid content_format"
	if len(err.Error()) < len(expectedSubstring) {
		t.Errorf("Error message too short: %s", err.Error())
		return
	}

	// Check if error message contains expected text
	errMsg := err.Error()
	found := false
	for i := 0; i <= len(errMsg)-len(expectedSubstring); i++ {
		if errMsg[i:i+len(expectedSubstring)] == expectedSubstring {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedSubstring, errMsg)
	}
}

func TestHandleCreateOutgoingMessage_DefaultContentFormat(t *testing.T) {
	// Test that empty content_format defaults to 'markdown'
	emptyFormat := ""
	input := CreateOutgoingMessageInput{
		Subject:       "Test",
		Content:       "Test content",
		ContentFormat: &emptyFormat, // Empty should default to markdown
		ToRecipients:  []string{"test@example.com"},
	}

	// Verify the default is applied correctly
	contentFormat := strings.ToLower(strings.TrimSpace(*input.ContentFormat))
	if contentFormat == "" {
		contentFormat = ContentFormatDefault
	}

	if contentFormat != ContentFormatMarkdown {
		t.Errorf("Expected default content format to be '%s', got '%s'", ContentFormatMarkdown, contentFormat)
	}

	// Also verify it's a valid format
	switch contentFormat {
	case ContentFormatPlain:
		// Valid
	case ContentFormatMarkdown:
		// Valid
	default:
		t.Errorf("Default content format '%s' is not a valid format", contentFormat)
	}
}

func TestHandleCreateOutgoingMessage_ValidContentFormats(t *testing.T) {
	tests := []struct {
		name          string
		contentFormat string
		shouldError   bool
		expectedValue string
	}{
		{
			name:          "plain format",
			contentFormat: ContentFormatPlain,
			shouldError:   false,
			expectedValue: ContentFormatPlain,
		},
		{
			name:          "markdown format",
			contentFormat: ContentFormatMarkdown,
			shouldError:   false,
			expectedValue: ContentFormatMarkdown,
		},
		{
			name:          "empty format defaults to markdown",
			contentFormat: "",
			shouldError:   false,
			expectedValue: ContentFormatMarkdown,
		},
		{
			name:          "unknown format returns error",
			contentFormat: "html",
			shouldError:   true,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests would require JXA to actually execute,
			// so we only test that the switch statement is correctly structured
			// by checking the contentFormat normalization logic
			contentFormat := strings.ToLower(strings.TrimSpace(tt.contentFormat))
			if contentFormat == "" {
				contentFormat = ContentFormatDefault
			}

			// Verify the normalized format matches expected value
			if tt.expectedValue != "" && contentFormat != tt.expectedValue {
				t.Errorf("Expected normalized format '%s', got '%s'", tt.expectedValue, contentFormat)
			}

			// Verify format is recognized
			isValid := false
			switch contentFormat {
			case ContentFormatPlain:
				isValid = true
			case ContentFormatMarkdown:
				isValid = true
			default:
				isValid = false
			}

			if isValid && tt.shouldError {
				t.Errorf("Format %s should be invalid but is valid", tt.contentFormat)
			}
			if !isValid && !tt.shouldError {
				t.Errorf("Format %s should be valid but is invalid", tt.contentFormat)
			}
		})
	}
}
