# Rich Text Support

## Principles

1. **Business logic in Go, not JXA**: All Markdown parsing, style application, and margin calculations happen in Go
2. **JXA only renders**: JXA receives fully-processed styled blocks and just creates paragraphs
3. **Rune-based positioning**: All inline style positions use rune (Unicode code point) offsets, not byte offsets
4. **Margin blocks**: Margins are represented as separate "margin" type blocks created in Go
5. **Character-level color styling**: All colors are applied at character level (via InlineStyles) for consistent dark mode behavior
6. **Newline splitting**: Code blocks, blockquotes, and list items are split on newlines to avoid Mail.app paragraph splitting breaking inline styles

## UTF-8 Character Handling

**CRITICAL**: Always use rune length, never byte length for character positions:

```go
// ❌ WRONG - byte length
currentPos += len(text)
prefixLen := len(prefix)

// ✅ CORRECT - rune length
currentPos += len([]rune(text))
prefixLen := len([]rune(prefix))
```

**Why**: JXA character arrays use Unicode code point indexing. Multibyte UTF-8 characters (•, emoji, etc.) are multiple bytes but single runes.

## Margin Handling

Margins are handled entirely in Go:

```go
func convertHeading(node *ast.Heading, source []byte, config *RenderingConfig, isFirst bool) ([]StyledBlock, error) {
    var blocks []StyledBlock
    
    // Add margin_top block (skip for first block)
    if !isFirst && style.MarginTop > 0 {
        blocks = append(blocks, StyledBlock{
            Type: "margin",
            Text: "\n",
            Font: style.Font,
            Size: style.MarginTop,
        })
    }
    
    // Add the heading block
    blocks = append(blocks, StyledBlock{...})
    
    // Add margin_bottom block
    if style.MarginBottom > 0 {
        blocks = append(blocks, StyledBlock{
            Type: "margin",
            Text: "\n",
            Font: style.Font,
            Size: style.MarginBottom,
        })
    }
    
    return blocks, nil
}
```

**JXA renders ALL blocks as paragraphs** (simplified):
```javascript
// All blocks rendered the same way - Go adds newlines to block.text
const props = {};
if (block.font) {
    props.font = block.font;
}
if (block.size) {
    props.size = block.size;
}
if (block.color) {
    props.color = block.color;
}

Mail.make({
    new: "paragraph",
    withData: block.text,  // Already includes newline from Go
    withProperties: props,
    at: msg.content,
});
```

**Key Points:**
- No type checking needed - render everything as paragraph
- Font, size, and color are optional (only set if present)
- Newlines handled in Go (appended to `block.text`)
- Margin blocks just have different size, rendered same way

## Character-Level Color Styling and Paragraph Splitting

**CRITICAL**: Mail.app automatically splits text on embedded newlines when creating paragraphs, which breaks absolute character positioning for inline styles.

**Solution**: Always split blocks on newlines before sending to JXA.

**Color Application**:
- All colors MUST be applied at character level via `InlineStyles`, never at paragraph level via `Color` property
- Mail.app auto-adapts character-level colors in dark mode but NOT paragraph-level colors
- This ensures consistent color behavior across light and dark modes

**Implementation Pattern**:
```go
// For code blocks, blockquotes, and any block with potential newlines
text := buf.String()
lines := strings.Split(text, "\n")

for i, line := range lines {
    lineText := line + "\n"
    lineRuneCount := len([]rune(line))
    
    var lineInlineStyles []InlineStyle
    
    // Apply color as character-level style covering entire line
    if style.Color != nil {
        colorStyle := InlineStyle{
            Start: 0,
            End:   lineRuneCount,
            Color: style.Color,
        }
        lineInlineStyles = append(lineInlineStyles, colorStyle)
    }
    
    // Add inline styles from content (bold, italic, etc.)
    // Adjust positions for this line...
    
    blocks = append(blocks, StyledBlock{
        Type:         BlockTypeCodeBlock,
        Text:         lineText,
        Font:         safeString(style.Font),
        Size:         safeInt(style.Size),
        InlineStyles: lineInlineStyles,
    })
}
```

**Key Points**:
- Split on `\n` to create one styled block per line
- Each line gets `lineText = line + "\n"` (add back the newline)
- Character-level color covers `0` to `lineRuneCount` (excludes trailing newline)
- When preserving existing inline styles (like in blockquotes), adjust positions for each line
- Font and size can be set at paragraph level
- Color MUST be set at character level (via InlineStyles)

**Why This Works**:
- One styled block per line = no embedded newlines
- Mail.app doesn't split the paragraph
- Inline style positions remain accurate
- Character-level colors get dark mode adaptation automatically

**Applied To**:
- Code blocks (split by line)
- Blockquotes (split by line, preserve inline styles from content)
- List items (already split by line)
- Regular paragraphs (split on hard line breaks `\n`)

**DON'T**:
- Set `Color` property at paragraph level for any block with inline styles
- Include newlines in blocks without splitting
- Use paragraph-level color for consistency - always use character-level

## Block Type Constants

Block types are defined for documentation/debugging but JXA doesn't check them:

```go
const (
    BlockTypeParagraph      = "paragraph"
    BlockTypeHeading        = "heading"
    BlockTypeCodeBlock      = "code_block"
    BlockTypeBlockquote     = "blockquote"
    BlockTypeListItem       = "list_item"
    BlockTypeHorizontalRule = "horizontal_rule"
)
```

**Note**: Margin blocks have empty Type field - they're distinguished by having only size and newline text.

## Styled Block Structure

```go
type StyledBlock struct {
    Type         string        `json:"type"`           // "paragraph", "heading", "code_block", etc.
    Text         string        `json:"text"`           // Text content (includes trailing \n)
    Font         string        `json:"font,omitempty"` // Optional - paragraph level
    Size         int           `json:"size,omitempty"` // Optional - paragraph level
    Color        *AppleRGB     `json:"color,omitempty"` // DEPRECATED - use InlineStyles instead
    InlineStyles []InlineStyle `json:"inline_styles,omitempty"` // Character-level styling
    Level        int           `json:"level,omitempty"` // For headings and list nesting
}
```

**Key Points:**
- `Text` always includes trailing `\n` (added in Go, not JXA)
- Font and Size are applied at paragraph level
- **Color is DEPRECATED** - always use InlineStyles for color to ensure consistent dark mode behavior
- InlineStyles apply formatting at character level (including color)
- Margin blocks have empty Type, just size and `"\n"` text
- No `MarginTop`/`MarginBottom` fields - margins are separate blocks
- For blocks with newlines (code blocks, blockquotes), split into one block per line
**Example usage**:
```go
return StyledBlock{
    Type: BlockTypeHeading,  // ✅ Use constant
    Text: text,
    // ...
}

// ❌ Don't use magic strings
return StyledBlock{
    Type: "heading",
    // ...
}
```

**Note**: MarginTop/MarginBottom are set on blocks but margins are rendered as separate "margin" type blocks by the Go converter.

## Content Format Enum

Use constants for content format validation:

```go
const (
    ContentFormatPlain    = "plain"
    ContentFormatMarkdown = "markdown"
    ContentFormatDefault  = ContentFormatMarkdown // Default is markdown
)

// Use switch for format handling
switch contentFormat {
case ContentFormatMarkdown:
    // Parse and render
case ContentFormatPlain:
    // Plain text
default:
    return nil, nil, fmt.Errorf("invalid content_format")
}
```

**Never panic on user input** - return errors instead.