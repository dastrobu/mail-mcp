package richtext

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// BlockType constants for styled block types
const (
	BlockTypeParagraph      = "paragraph"
	BlockTypeHeading        = "heading"
	BlockTypeCodeBlock      = "code_block"
	BlockTypeBlockquote     = "blockquote"
	BlockTypeListItem       = "list_item"
	BlockTypeHorizontalRule = "horizontal_rule"
	BlockTypeMargin         = "margin"
)

// StyledBlock represents a block of text with styling information
type StyledBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	Font         string        `json:"font,omitempty"`
	Size         int           `json:"size,omitempty"`
	Color        *AppleRGB     `json:"color,omitempty"`
	InlineStyles []InlineStyle `json:"inline_styles,omitempty"`
	Level        int           `json:"level,omitempty"` // For headings and list nesting
}

// InlineStyle represents styling for a range of characters within a block
type InlineStyle struct {
	Start int       `json:"start"`
	End   int       `json:"end"`
	Font  string    `json:"font,omitempty"`
	Size  int       `json:"size,omitempty"`
	Color *AppleRGB `json:"color,omitempty"`
}

// ConvertMarkdownToStyledBlocks converts a Markdown AST to styled blocks
func ConvertMarkdownToStyledBlocks(root ast.Node, source []byte, config *PreparedConfig) ([]StyledBlock, error) {
	var blocks []StyledBlock
	isFirstBlock := true

	// Walk the AST
	err := ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		// Only process nodes when entering (not when leaving)
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			headingBlocks, err := convertHeading(n, source, config, isFirstBlock)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, headingBlocks...)
			return ast.WalkSkipChildren, nil

		case *ast.Paragraph:
			// Skip paragraphs that are children of list items or blockquotes
			// They will be handled by their parent
			if isChildOfListItem(node) || isChildOfBlockquote(node) {
				return ast.WalkContinue, nil
			}
			paraBlocks, err := convertParagraph(n, source, config)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, paraBlocks...)
			return ast.WalkSkipChildren, nil

		case *ast.CodeBlock, *ast.FencedCodeBlock:
			codeBlocks, err := convertCodeBlock(n, source, config, isFirstBlock)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, codeBlocks...)
			return ast.WalkSkipChildren, nil

		case *ast.Blockquote:
			blockquoteBlocks, err := convertBlockquote(n, source, config, isFirstBlock)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, blockquoteBlocks...)
			return ast.WalkSkipChildren, nil

		case *ast.List:
			listBlocks, err := convertList(n, source, config, 0, isFirstBlock)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, listBlocks...)
			return ast.WalkSkipChildren, nil

		case *ast.ThematicBreak:
			block, err := convertThematicBreak(config)
			if err != nil {
				return ast.WalkStop, err
			}
			isFirstBlock = false
			blocks = append(blocks, block)
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk AST: %w", err)
	}

	return blocks, nil
}

// convertHeading converts a heading node to styled blocks (including margin blocks)
func convertHeading(node *ast.Heading, source []byte, config *PreparedConfig, isFirst bool) ([]StyledBlock, error) {
	styleKey := fmt.Sprintf("h%d", node.Level)
	style := config.GetStyle(styleKey)

	text, inlineStyles, err := extractTextWithInlineStyles(node, source, config)
	if err != nil {
		return nil, err
	}

	var blocks []StyledBlock

	// Add margin_top block if specified (skip for first block to avoid empty lines at start)
	if !isFirst && style.MarginTop != nil && *style.MarginTop > 0 {
		blocks = append(blocks, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginTop,
		})
	}

	// Add the heading block
	blocks = append(blocks, StyledBlock{
		Type:         BlockTypeHeading,
		Text:         text + "\n",
		Font:         safeString(style.Font),
		Size:         safeInt(style.Size),
		Color:        style.Color,
		InlineStyles: inlineStyles,
		Level:        node.Level,
	})

	// Add margin_bottom block if specified
	if style.MarginBottom != nil && *style.MarginBottom > 0 {
		blocks = append(blocks, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginBottom,
		})
	}

	return blocks, nil
}

// convertParagraph converts a paragraph node to a styled block
func convertParagraph(node *ast.Paragraph, source []byte, config *PreparedConfig) ([]StyledBlock, error) {
	style := config.GetStyle("paragraph")

	text, inlineStyles, err := extractTextWithInlineStyles(node, source, config)
	if err != nil {
		return nil, err
	}

	// Split on hard line breaks (newlines in the text)
	// Mail.app automatically creates separate paragraphs for embedded newlines,
	// which breaks absolute character positioning for inline styles
	lines := strings.Split(text, "\n")

	// If no newlines or only trailing newline, return single block
	if len(lines) <= 1 || (len(lines) == 2 && lines[1] == "") {
		return []StyledBlock{{
			Type:         BlockTypeParagraph,
			Text:         text + "\n",
			Font:         safeString(style.Font),
			Size:         safeInt(style.Size),
			Color:        style.Color,
			InlineStyles: inlineStyles,
		}}, nil
	}

	// Split into separate blocks for each line
	blocks := make([]StyledBlock, 0, len(lines))
	currentPos := 0

	for _, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		lineLength := len([]rune(line))
		lineEnd := currentPos + lineLength

		// Find inline styles that apply to this line
		var lineStyles []InlineStyle
		for _, style := range inlineStyles {
			// Check if style overlaps with this line
			if style.End <= currentPos || style.Start >= lineEnd {
				continue // Style is outside this line
			}

			// Adjust style positions to be relative to line start
			adjustedStyle := style
			if style.Start < currentPos {
				adjustedStyle.Start = 0
			} else {
				adjustedStyle.Start = style.Start - currentPos
			}

			if style.End > lineEnd {
				adjustedStyle.End = lineLength
			} else {
				adjustedStyle.End = style.End - currentPos
			}

			lineStyles = append(lineStyles, adjustedStyle)
		}

		blocks = append(blocks, StyledBlock{
			Type:         BlockTypeParagraph,
			Text:         line + "\n",
			Font:         safeString(style.Font),
			Size:         safeInt(style.Size),
			Color:        style.Color,
			InlineStyles: lineStyles,
		})

		currentPos += lineLength + 1 // +1 for the newline
	}

	return blocks, nil
}

// convertCodeBlock converts a code block to styled blocks (including margin blocks)
func convertCodeBlock(node ast.Node, source []byte, config *PreparedConfig, isFirst bool) ([]StyledBlock, error) {
	style := config.GetStyle("code_block")

	var buf bytes.Buffer
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		buf.Write(line.Value(source))
	}

	var blocks []StyledBlock

	// Add margin_top block if specified (skip for first block)
	if !isFirst && style.MarginTop != nil && *style.MarginTop > 0 {
		blocks = append(blocks, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginTop,
		})
	}

	// Get prefix (if any)
	prefix := ""
	prefixRuneCount := 0
	var prefixInlineStyle *InlineStyle
	if style.Prefix != nil && style.Prefix.Content != nil {
		prefix = *style.Prefix.Content
		prefixRuneCount = len([]rune(prefix))

		// Create inline style for prefix if it has custom styling
		if style.Prefix.Font != nil || style.Prefix.Size != nil || style.Prefix.Color != nil {
			prefixInlineStyle = &InlineStyle{
				Start: 0,
				End:   prefixRuneCount,
			}
			if style.Prefix.Font != nil {
				prefixInlineStyle.Font = *style.Prefix.Font
			}
			if style.Prefix.Size != nil {
				prefixInlineStyle.Size = *style.Prefix.Size
			}
			if style.Prefix.Color != nil {
				prefixInlineStyle.Color = style.Prefix.Color
			}
		}
	}

	// Split on newlines to avoid Mail.app paragraph splitting breaking inline styles
	text := buf.String()
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		lineText := prefix + line + "\n"
		lineRuneCount := prefixRuneCount + len([]rune(line))

		var lineInlineStyles []InlineStyle

		// Add prefix inline style if it has custom styling
		if prefixInlineStyle != nil {
			lineInlineStyles = append(lineInlineStyles, *prefixInlineStyle)
		}

		// Add code block color as character-level style covering entire line (including prefix)
		if style.Color != nil {
			colorStyle := InlineStyle{
				Start: 0,
				End:   lineRuneCount,
				Color: style.Color,
			}
			lineInlineStyles = append(lineInlineStyles, colorStyle)
		}

		blocks = append(blocks, StyledBlock{
			Type:         BlockTypeCodeBlock,
			Text:         lineText,
			Font:         safeString(style.Font),
			Size:         safeInt(style.Size),
			InlineStyles: lineInlineStyles,
		})
	}

	// Add margin_bottom block if specified
	if style.MarginBottom != nil && *style.MarginBottom > 0 {
		blocks = append(blocks, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginBottom,
		})
	}

	return blocks, nil
}

// applyPrefixToBlocks adds a prefix to all non-margin blocks
func applyPrefixToBlocks(blocks []StyledBlock, prefixContent string, prefixStyle *PreparedPrefix) []StyledBlock {
	if prefixContent == "" {
		return blocks
	}

	prefixRuneCount := len([]rune(prefixContent))
	result := make([]StyledBlock, 0, len(blocks))

	for _, block := range blocks {
		// Handle margin blocks (empty lines between paragraphs)
		if block.Type == "" && block.Text == "\n" {
			// Apply prefix to margin blocks too (empty lines in blockquotes should have ">")
			result = append(result, StyledBlock{
				Type: block.Type,
				Text: prefixContent + "\n",
				Font: block.Font,
				Size: block.Size,
			})
			continue
		}

		// Split block text into lines and prepend prefix to each line
		lines := strings.Split(block.Text, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1] // Remove trailing empty string from split
		}

		for _, line := range lines {
			prefixedText := prefixContent + line + "\n"

			// Adjust inline styles to account for prefix
			var adjustedInlineStyles []InlineStyle

			// Add prefix inline style if it has custom styling
			if prefixStyle != nil && (prefixStyle.Font != nil || prefixStyle.Size != nil || prefixStyle.Color != nil) {
				prefixInlineStyle := InlineStyle{
					Start: 0,
					End:   prefixRuneCount,
				}
				if prefixStyle.Font != nil {
					prefixInlineStyle.Font = *prefixStyle.Font
				}
				if prefixStyle.Size != nil {
					prefixInlineStyle.Size = *prefixStyle.Size
				}
				if prefixStyle.Color != nil {
					prefixInlineStyle.Color = prefixStyle.Color
				}
				adjustedInlineStyles = append(adjustedInlineStyles, prefixInlineStyle)
			}

			// Adjust existing inline styles by shifting positions
			for _, style := range block.InlineStyles {
				adjustedStyle := style
				adjustedStyle.Start += prefixRuneCount
				adjustedStyle.End += prefixRuneCount
				adjustedInlineStyles = append(adjustedInlineStyles, adjustedStyle)
			}

			result = append(result, StyledBlock{
				Type:         block.Type,
				Text:         prefixedText,
				Font:         block.Font,
				Size:         block.Size,
				Color:        block.Color,
				InlineStyles: adjustedInlineStyles,
				Level:        block.Level,
			})
		}
	}

	return result
}

// convertBlockquote converts a blockquote to styled blocks by recursively processing children
func convertBlockquote(node *ast.Blockquote, source []byte, config *PreparedConfig, isFirst bool) ([]StyledBlock, error) {
	style := config.GetStyle("blockquote")
	var result []StyledBlock

	// Check if we're nested inside another blockquote
	isNested := isChildOfBlockquote(node)

	// Collect content blocks that will get the prefix
	var contentBlocks []StyledBlock

	// Add margin_top block if specified (skip for first block)
	// Add to contentBlocks so nested blockquotes get prefixed properly
	// For root-level, we'll strip the prefix after applying it
	if !isFirst && style.MarginTop != nil && *style.MarginTop > 0 {
		contentBlocks = append(contentBlocks, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginTop,
		})
	}

	// Recursively convert child nodes
	isFirstChild := true
	var prevChild ast.Node
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		var childBlocks []StyledBlock
		var err error

		// Add empty line between certain block elements for visual separation
		if !isFirstChild && prevChild != nil && shouldAddEmptyLineBetween(prevChild, child) {
			// Add empty line with just prefix (will be prefixed later)
			contentBlocks = append(contentBlocks, StyledBlock{
				Type: BlockTypeParagraph,
				Text: "\n",
				Font: safeString(style.Font),
				Size: safeInt(style.Size),
			})
		}

		switch n := child.(type) {
		case *ast.Heading:
			childBlocks, err = convertHeading(n, source, config, isFirstChild)
		case *ast.Paragraph:
			childBlocks, err = convertParagraph(n, source, config)
		case *ast.CodeBlock, *ast.FencedCodeBlock:
			childBlocks, err = convertCodeBlock(n, source, config, isFirstChild)
		case *ast.Blockquote:
			// Nested blockquote
			childBlocks, err = convertBlockquote(n, source, config, isFirstChild)
		case *ast.List:
			childBlocks, err = convertList(n, source, config, 0, isFirstChild)
		case *ast.ThematicBreak:
			block, convertErr := convertThematicBreak(config)
			if convertErr != nil {
				err = convertErr
			} else {
				childBlocks = []StyledBlock{block}
			}
		default:
			// Skip unsupported node types
			continue
		}

		if err != nil {
			return nil, err
		}

		isFirstChild = false
		prevChild = child
		contentBlocks = append(contentBlocks, childBlocks...)
	}

	// Apply blockquote prefix to content blocks
	var prefixContent string
	var prefixStyle *PreparedPrefix
	if style.Prefix != nil && style.Prefix.Content != nil {
		prefixContent = *style.Prefix.Content
		prefixStyle = style.Prefix
	}
	prefixedBlocks := applyPrefixToBlocks(contentBlocks, prefixContent, prefixStyle)

	// Strip prefix from our own margin_top ONLY if we're at root level (not nested)
	// Nested blockquotes need the prefix on margin_top so parent can add its prefix too
	startIdx := 0
	if !isNested && !isFirst && style.MarginTop != nil && *style.MarginTop > 0 && len(prefixedBlocks) > 0 {
		firstBlock := prefixedBlocks[0]
		// Check if first block is our margin_top (has our margin size and got prefixed)
		if firstBlock.Type == "" && firstBlock.Size == *style.MarginTop && firstBlock.Text != "\n" {
			// Strip the prefix we just added (for root level only)
			result = append(result, StyledBlock{
				Type: firstBlock.Type,
				Text: "\n",
				Font: firstBlock.Font,
				Size: firstBlock.Size,
			})
			startIdx = 1
		}
	}
	result = append(result, prefixedBlocks[startIdx:]...)

	// Add margin_bottom block if specified
	// This margin comes AFTER the blockquote, so it should NOT get the prefix
	if style.MarginBottom != nil && *style.MarginBottom > 0 {
		result = append(result, StyledBlock{
			Text: "\n",
			Font: safeString(style.Font),
			Size: *style.MarginBottom,
		})
	}

	return result, nil
}

// shouldAddEmptyLineBetween determines if an empty line should be added between two block elements
func shouldAddEmptyLineBetween(prev, curr ast.Node) bool {
	// Add empty line between paragraphs
	if _, prevIsPara := prev.(*ast.Paragraph); prevIsPara {
		if _, currIsPara := curr.(*ast.Paragraph); currIsPara {
			return true
		}
	}

	// Add empty line before headings (unless first element)
	if _, currIsHeading := curr.(*ast.Heading); currIsHeading {
		return true
	}

	// Add empty line before code blocks
	if _, currIsCode := curr.(*ast.CodeBlock); currIsCode {
		return true
	}
	if _, currIsFenced := curr.(*ast.FencedCodeBlock); currIsFenced {
		return true
	}

	// Add empty line before nested blockquotes
	if _, currIsBlockquote := curr.(*ast.Blockquote); currIsBlockquote {
		return true
	}

	// Add empty line before lists
	if _, currIsList := curr.(*ast.List); currIsList {
		return true
	}

	return false
}

// convertList converts a list to styled blocks (including margin blocks)
func convertList(node *ast.List, source []byte, config *PreparedConfig, level int, isFirst bool) ([]StyledBlock, error) {
	var blocks []StyledBlock

	// Limit nesting depth to 4 levels
	if level > 3 {
		level = 3
	}

	listStyle := config.GetStyle("list")
	itemStyle := config.GetStyle("list_item")

	// Add margin_top block if specified (skip for first block and nested lists)
	// Add margin_top block if specified (only for top-level lists, skip for first block)
	if level == 0 && !isFirst && listStyle.MarginTop != nil && *listStyle.MarginTop > 0 {
		blocks = append(blocks, StyledBlock{
			Type: "margin",
			Text: "\n",
			Font: safeString(itemStyle.Font),
			Size: *listStyle.MarginTop,
		})
	}

	itemNum := 1
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			itemBlocks, err := convertListItem(item, source, config, node.IsOrdered(), itemNum, level)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, itemBlocks...)
			itemNum++
		}
	}

	// Add margin_bottom block if specified (only for top-level lists)
	if level == 0 && listStyle.MarginBottom != nil && *listStyle.MarginBottom > 0 {
		blocks = append(blocks, StyledBlock{
			Type: "margin",
			Text: "\n",
			Font: safeString(itemStyle.Font),
			Size: *listStyle.MarginBottom,
		})
	}

	return blocks, nil
}

// convertListItem converts a list item to styled blocks
func convertListItem(node *ast.ListItem, source []byte, config *PreparedConfig, ordered bool, num int, level int) ([]StyledBlock, error) {
	style := config.GetStyle("list_item")

	var blocks []StyledBlock

	// Build the bullet/number prefix
	indent := strings.Repeat("  ", level)
	var prefix string
	if ordered {
		prefix = fmt.Sprintf("%s%d. ", indent, num)
	} else {
		prefix = fmt.Sprintf("%s• ", indent)
	}

	// Extract text from child nodes
	var text string
	var inlineStyles []InlineStyle

	// List items can contain paragraphs or text blocks directly
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			var err error
			text, inlineStyles, err = extractTextWithInlineStyles(para, source, config)
			if err != nil {
				return nil, err
			}
			break
		} else if textBlock, ok := child.(*ast.TextBlock); ok {
			var err error
			text, inlineStyles, err = extractTextWithInlineStyles(textBlock, source, config)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	// Adjust inline style positions to account for prefix
	// Use rune length because the prefix may contain multibyte UTF-8 characters (e.g., •)
	prefixLen := len([]rune(prefix))
	for i := range inlineStyles {
		inlineStyles[i].Start += prefixLen
		inlineStyles[i].End += prefixLen
	}

	blocks = append(blocks, StyledBlock{
		Type:         BlockTypeListItem,
		Text:         prefix + text + "\n",
		Font:         safeString(style.Font),
		Size:         safeInt(style.Size),
		Color:        style.Color,
		InlineStyles: inlineStyles,
		Level:        level,
	})

	// Handle nested lists (always pass isFirst=false for nested lists)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			nestedBlocks, err := convertList(list, source, config, level+1, false)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, nestedBlocks...)
		}
	}

	return blocks, nil
}

// convertThematicBreak converts a horizontal rule to a styled block
func convertThematicBreak(config *PreparedConfig) (StyledBlock, error) {
	style := config.GetStyle("horizontal_rule")

	return StyledBlock{
		Type:  BlockTypeHorizontalRule,
		Text:  "─────────────────────────────────────\n",
		Font:  safeString(style.Font),
		Size:  safeInt(style.Size),
		Color: style.Color,
	}, nil
}

// extractTextWithInlineStyles extracts text from a node and identifies inline styles
func extractTextWithInlineStyles(node ast.Node, source []byte, config *PreparedConfig) (string, []InlineStyle, error) {
	var buf bytes.Buffer
	var styles []InlineStyle
	currentPos := 0

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			text := string(n.Segment.Value(source))
			buf.WriteString(text)
			currentPos += len([]rune(text))

			// Handle hard line break (two spaces at end of line)
			if n.HardLineBreak() {
				buf.WriteString("\n")
				currentPos += 1
			} else if n.SoftLineBreak() {
				// Soft line break: add a space after this text node
				buf.WriteString(" ")
				currentPos += 1
			}

		case *ast.Emphasis:
			// Check if this is nested bold+italic (*** text ***)
			// by looking for child emphasis with different level
			hasNestedEmphasis := false
			var nestedEmphasisLevel int
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if childEmph, ok := child.(*ast.Emphasis); ok {
					hasNestedEmphasis = true
					nestedEmphasisLevel = childEmph.Level
					break
				}
			}

			// Detect bold+italic combination
			isBoldItalic := false
			if hasNestedEmphasis {
				// Level 1 (italic) containing Level 2 (bold) = bold+italic
				// OR Level 2 (bold) containing Level 1 (italic) = bold+italic
				if (n.Level == 1 && nestedEmphasisLevel == 2) || (n.Level == 2 && nestedEmphasisLevel == 1) {
					isBoldItalic = true
				}
			}

			// Extract the text inside emphasis (may contain nested emphasis)
			emphText, nestedStyles, err := extractTextWithInlineStyles(n, source, config)
			if err != nil {
				return "", nil, err
			}

			start := currentPos
			end := start + len([]rune(emphText))

			if isBoldItalic {
				// Use combined bold+italic style
				style := config.GetStyle("bold_italic")

				// Add single bold_italic style (don't add nested styles separately)
				inlineStyle := InlineStyle{
					Start: start,
					End:   end,
					Font:  safeString(style.Font),
					Size:  safeInt(style.Size),
					Color: style.Color,
				}
				styles = append(styles, inlineStyle)
			} else {
				// Regular emphasis (bold OR italic, not both)
				styleKey := "italic"
				if n.Level == 2 {
					styleKey = "bold"
				}
				style := config.GetStyle(styleKey)

				// Add this emphasis style
				inlineStyle := InlineStyle{
					Start: start,
					End:   end,
					Font:  safeString(style.Font),
					Size:  safeInt(style.Size),
					Color: style.Color,
				}
				styles = append(styles, inlineStyle)

				// Add nested styles with adjusted positions
				for _, nestedStyle := range nestedStyles {
					nestedStyle.Start += start
					nestedStyle.End += start
					styles = append(styles, nestedStyle)
				}
			}

			buf.WriteString(emphText)
			currentPos += len([]rune(emphText))

		case *extast.Strikethrough:
			// Extract the text inside strikethrough (may contain nested elements)
			strikeText, nestedStyles, err := extractTextWithInlineStyles(n, source, config)
			if err != nil {
				return "", nil, err
			}

			start := currentPos
			end := start + len([]rune(strikeText))

			style := config.GetStyle("strikethrough")

			// Add this strikethrough style
			inlineStyle := InlineStyle{
				Start: start,
				End:   end,
				Font:  safeString(style.Font),
				Size:  safeInt(style.Size),
				Color: style.Color,
			}
			styles = append(styles, inlineStyle)

			// Add nested styles with adjusted positions
			for _, nestedStyle := range nestedStyles {
				nestedStyle.Start += start
				nestedStyle.End += start
				styles = append(styles, nestedStyle)
			}

			buf.WriteString(strikeText)
			currentPos += len([]rune(strikeText))

		case *ast.CodeSpan:
			// Extract code span text
			codeText := string(n.Text(source))
			start := currentPos
			end := start + len([]rune(codeText))

			style := config.GetStyle("code")

			inlineStyle := InlineStyle{
				Start: start,
				End:   end,
				Font:  safeString(style.Font),
				Size:  safeInt(style.Size),
				Color: style.Color,
			}
			styles = append(styles, inlineStyle)

			buf.WriteString(codeText)
			currentPos += len([]rune(codeText))

		case *ast.Link:
			// Extract link text and URL (may contain nested emphasis)
			linkText, nestedStyles, err := extractTextWithInlineStyles(n, source, config)
			if err != nil {
				return "", nil, err
			}
			url := string(n.Destination)

			// Format as "text (url)"
			formatted := fmt.Sprintf("%s (%s)", linkText, url)
			start := currentPos
			linkTextEnd := start + len([]rune(linkText))

			style := config.GetStyle("link")

			// Style just the link text part, not the URL
			inlineStyle := InlineStyle{
				Start: start,
				End:   linkTextEnd,
				Font:  safeString(style.Font),
				Size:  safeInt(style.Size),
				Color: style.Color,
			}
			styles = append(styles, inlineStyle)

			// Add nested styles with adjusted positions
			for _, nestedStyle := range nestedStyles {
				nestedStyle.Start += start
				nestedStyle.End += start
				styles = append(styles, nestedStyle)
			}

			buf.WriteString(formatted)
			currentPos += len([]rune(formatted))
		}
	}

	return buf.String(), styles, nil
}

// safeString safely dereferences a string pointer, returning empty string if nil
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeInt safely dereferences an int pointer, returning 0 if nil
func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// isChildOfListItem checks if a node is a child of a list item
func isChildOfListItem(node ast.Node) bool {
	parent := node.Parent()
	for parent != nil {
		if _, ok := parent.(*ast.ListItem); ok {
			return true
		}
		parent = parent.Parent()
	}
	return false
}

// isChildOfBlockquote checks if a node is a child of a blockquote
func isChildOfBlockquote(node ast.Node) bool {
	parent := node.Parent()
	for parent != nil {
		if _, ok := parent.(*ast.Blockquote); ok {
			return true
		}
		parent = parent.Parent()
	}
	return false
}
