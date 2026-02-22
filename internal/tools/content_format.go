package tools

import (
	"fmt"
	"strings"

	"github.com/dastrobu/apple-mail-mcp/internal/md"
)

// ContentFormat constants for email content formatting
const (
	ContentFormatPlain    = "plain"
	ContentFormatMarkdown = "markdown"
	// ContentFormatDefault is the default content format
	ContentFormatDefault = ContentFormatMarkdown
)

// ValidateAndNormalizeContentFormat checks if the provided format is valid and returns the normalized version.
// If the input is nil or empty, it returns the default format.
func ValidateAndNormalizeContentFormat(format *string) (string, error) {
	if format == nil {
		return ContentFormatDefault, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*format))
	if normalized == "" {
		return ContentFormatDefault, nil
	}

	switch normalized {
	case ContentFormatPlain:
		return ContentFormatPlain, nil
	case ContentFormatMarkdown:
		return ContentFormatMarkdown, nil
	default:
		return "", fmt.Errorf("invalid content_format: %s", normalized)
	}
}

// IsValidContentFormat returns true if the format is supported.
func IsValidContentFormat(format string) bool {
	switch format {
	case ContentFormatPlain, ContentFormatMarkdown:
		return true
	default:
		return false
	}
}

// ToClipboardContent takes raw content and a format, and returns the content to paste and whether it is HTML.
// It panics if an unknown content format is provided.
func ToClipboardContent(content string, contentFormat string) (contentToPaste string, isHTML bool, err error) {
	switch contentFormat {
	case ContentFormatMarkdown:
		contentToPaste, err = md.Render(content)
		if err != nil {
			return "", false, err
		}
		return contentToPaste, true, nil
	case ContentFormatPlain:
		return content, false, nil
	default:
		panic(fmt.Sprintf("unknown content format: %s", contentFormat))
	}
}
