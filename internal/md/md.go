package md

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Render converts markdown content to HTML string using goldmark.
func Render(content string) (string, error) {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)
	var buf bytes.Buffer
	if err := gm.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown: %w", err)
	}
	return buf.String(), nil
}
