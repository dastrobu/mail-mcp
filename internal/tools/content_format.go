package tools

import (
	"fmt"
	"strings"

	"github.com/dastrobu/mail-mcp/internal/md"
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

// ToClipboardContent takes raw content and a format, and returns the HTML content (optional), the plain text content, and an error.
// If the format is Markdown, the HTML is the rendered Markdown and the plain text is the original Markdown.
// If the format is Plain, the HTML is nil and the plain text is the raw content.
func ToClipboardContent(content string, contentFormat string) (htmlContent *string, plainContent string, err error) {
	switch contentFormat {
	case ContentFormatMarkdown:
		html, err := md.Render(content)
		if err != nil {
			return nil, "", err
		}
		return &html, content, nil
	case ContentFormatPlain:
		return nil, content, nil
	default:
		panic(fmt.Sprintf("unknown content format: %s", contentFormat))
	}
}
