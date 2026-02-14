package richtext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertMarkdownToStyledBlocks(t *testing.T) {
	// Create a test config without margins for most tests
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_styles.yaml")

	configYAML := `defaults:
  font: "Helvetica"
  size: 12
  color: "#000000"

styles:
  h1:
    font: "Helvetica-Bold"
    size: 24
    color: "#000000"
  h2:
    font: "Helvetica-Bold"
    size: 20
    color: "#000000"
  bold:
    font: "Helvetica-Bold"
  italic:
    font: "Helvetica-Oblique"
  strikethrough:
    font: "Helvetica"
    color: "#6A737D"
  bold_italic:
    font: "Helvetica-BoldOblique"
    color: "#000000"
  code:
    font: "Menlo-Regular"
    size: 11
    color: "#D73A49"
  code_block:
    font: "Menlo-Regular"
    size: 11
    color: "#24292E"
  blockquote:
    font: "Helvetica-Oblique"
    size: 12
    color: "#6A737D"
    prefix:
      content: "> "
  list_item:
    font: "Helvetica"
    size: 12
    color: "#000000"
  horizontal_rule:
    font: "Helvetica"
    size: 1
    color: "#E1E4E8"
  link:
    color: "#0366D6"
  list:
    margin_top: 0
    margin_bottom: 0
  paragraph:
    font: "Helvetica"
    size: 12
    color: "#000000"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	tests := []struct {
		name       string
		markdown   string
		wantBlocks int
		checkFunc  func(*testing.T, []StyledBlock)
	}{
		{
			name:       "simple paragraph",
			markdown:   "Hello world",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Expected type paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "Hello world\n" {
					t.Errorf("Expected text 'Hello world\\n', got %q", blocks[0].Text)
				}
			},
		},
		{
			name:       "heading level 1",
			markdown:   "# Main Title",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Type != BlockTypeHeading {
					t.Errorf("Expected type heading, got %s", blocks[0].Type)
				}
				if blocks[0].Level != 1 {
					t.Errorf("Expected level 1, got %d", blocks[0].Level)
				}
				if blocks[0].Text != "Main Title\n" {
					t.Errorf("Expected text 'Main Title\\n', got %q", blocks[0].Text)
				}
				if blocks[0].Font != "Helvetica-Bold" {
					t.Errorf("Expected font Helvetica-Bold, got %s", blocks[0].Font)
				}
			},
		},
		{
			name:       "bold text",
			markdown:   "This is **bold** text",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Expected type paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "This is bold text\n" {
					t.Errorf("Expected text 'This is bold text\\n', got %q", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					if style.Start != 8 || style.End != 12 {
						t.Errorf("Expected inline style at 8-12, got %d-%d", style.Start, style.End)
					}
					if style.Font != "Helvetica-Bold" {
						t.Errorf("Expected font Helvetica-Bold, got %s", style.Font)
					}
				}
			},
		},
		{
			name:       "italic text",
			markdown:   "This is *italic* text",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					if style.Font != "Helvetica-Oblique" {
						t.Errorf("Expected font Helvetica-Oblique, got %s", style.Font)
					}
				}
			},
		},
		{
			name:       "strikethrough text",
			markdown:   "This is ~~strikethrough~~ text",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Text != "This is strikethrough text\n" {
					t.Errorf("Expected text 'This is strikethrough text\\n', got %q", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					if style.Font != "Helvetica" {
						t.Errorf("Expected font Helvetica, got %s", style.Font)
					}
				}
			},
		},
		{
			name:       "inline code",
			markdown:   "Use `code` here",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Text != "Use code here\n" {
					t.Errorf("Expected text 'Use code here\\n', got %q", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					if style.Font != "Menlo-Regular" {
						t.Errorf("Expected font Menlo-Regular, got %s", style.Font)
					}
				}
			},
		},
		{
			name:       "code block",
			markdown:   "```\nfunction test() {\n  return true;\n}\n```",
			wantBlocks: 4, // One block per line (split on newlines)
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// Check first line
				if blocks[0].Type != BlockTypeCodeBlock {
					t.Errorf("Expected type code_block, got %s", blocks[0].Type)
				}
				if blocks[0].Font != "Menlo-Regular" {
					t.Errorf("Expected font Menlo-Regular, got %s", blocks[0].Font)
				}
			},
		},
		{
			name:       "blockquote",
			markdown:   "> This is a quote",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// Blockquotes are now containers - child paragraphs render as paragraphs with prefix
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Expected type paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "> This is a quote\n" {
					t.Errorf("Expected text '> This is a quote\\n', got %q", blocks[0].Text)
				}
				// Paragraph inherits blockquote font since it uses default styling
				if blocks[0].Font != "Helvetica-Oblique" {
					t.Errorf("Expected font Helvetica-Oblique, got %s", blocks[0].Font)
				}
			},
		},
		{
			name:       "unordered list",
			markdown:   "- Item 1\n- Item 2\n- Item 3",
			wantBlocks: 3,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				for i, block := range blocks {
					if block.Type != BlockTypeListItem {
						t.Errorf("Block %d: Expected type list_item, got %s", i, block.Type)
					}
					if block.Level != 0 {
						t.Errorf("Block %d: Expected level 0, got %d", i, block.Level)
					}
				}
				if blocks[0].Text != "• Item 1\n" {
					t.Errorf("Expected text '• Item 1\\n', got %q", blocks[0].Text)
				}
			},
		},
		{
			name:       "ordered list",
			markdown:   "1. First\n2. Second\n3. Third",
			wantBlocks: 3,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Text != "1. First\n" {
					t.Errorf("Expected text '1. First\\n', got %q", blocks[0].Text)
				}
				if blocks[1].Text != "2. Second\n" {
					t.Errorf("Expected text '2. Second\\n', got %q", blocks[1].Text)
				}
				if blocks[2].Text != "3. Third\n" {
					t.Errorf("Expected text '3. Third\\n', got %q", blocks[2].Text)
				}
			},
		},
		{
			name:       "nested list",
			markdown:   "- Item 1\n  - Nested 1\n  - Nested 2\n- Item 2",
			wantBlocks: 4,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Level != 0 {
					t.Errorf("Expected level 0, got %d", blocks[0].Level)
				}
				if blocks[1].Level != 1 {
					t.Errorf("Expected level 1, got %d", blocks[1].Level)
				}
				if blocks[2].Level != 1 {
					t.Errorf("Expected level 1, got %d", blocks[2].Level)
				}
				if blocks[3].Level != 0 {
					t.Errorf("Expected level 0, got %d", blocks[3].Level)
				}
			},
		},
		{
			name:       "horizontal rule",
			markdown:   "---",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Type != BlockTypeHorizontalRule {
					t.Errorf("Expected type horizontal_rule, got %s", blocks[0].Type)
				}
			},
		},
		{
			name:       "link",
			markdown:   "Visit [GitHub](https://github.com)",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Text != "Visit GitHub (https://github.com)\n" {
					t.Errorf("Expected text 'Visit GitHub (https://github.com)\\n', got %q", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				}
			},
		},
		{
			name:       "multiple paragraphs",
			markdown:   "First paragraph\n\nSecond paragraph",
			wantBlocks: 2,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Text != "First paragraph\n" {
					t.Errorf("Expected text 'First paragraph\\n', got %q", blocks[0].Text)
				}
				if blocks[1].Text != "Second paragraph\n" {
					t.Errorf("Expected text 'Second paragraph\\n', got %q", blocks[1].Text)
				}
			},
		},
		{
			name:       "mixed inline styles",
			markdown:   "This has **bold** and *italic* and `code` and ~~strike~~",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if len(blocks[0].InlineStyles) != 4 {
					t.Errorf("Expected 4 inline styles, got %d", len(blocks[0].InlineStyles))
				}
			},
		},
		{
			name:       "list item with bold",
			markdown:   "- **Bold text** using double asterisks",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				expectedText := "• Bold text using double asterisks\n"
				if blocks[0].Text != expectedText {
					t.Errorf("Expected text %q, got %q", expectedText, blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					// "• " is 2 runes (bullet + space), so bold should start at position 2
					// "Bold text" is 9 runes, so it should end at position 11
					if style.Start != 2 {
						t.Errorf("Expected style start at 2, got %d", style.Start)
					}
					if style.End != 11 {
						t.Errorf("Expected style end at 11, got %d", style.End)
					}
					// Verify it covers "Bold text"
					runes := []rune(blocks[0].Text)
					styledText := string(runes[style.Start:style.End])
					if styledText != "Bold text" {
						t.Errorf("Expected styled text 'Bold text', got '%s'", styledText)
					}
				}
			},
		},
		{
			name:       "bold and italic combined",
			markdown:   "This is ***bold and italic*** text",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				expectedText := "This is bold and italic text\n"
				if blocks[0].Text != expectedText {
					t.Errorf("Expected text '%s', got '%s'", expectedText, blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style (merged bold+italic), got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					// "This is " is 8 runes, "bold and italic" is 15 runes
					if style.Start != 8 {
						t.Errorf("Expected style start at 8, got %d", style.Start)
					}
					if style.End != 23 {
						t.Errorf("Expected style end at 23, got %d", style.End)
					}
					if style.Font != "Helvetica-BoldOblique" {
						t.Errorf("Expected font Helvetica-BoldOblique, got %s", style.Font)
					}
					// Verify it covers "bold and italic"
					runes := []rune(blocks[0].Text)
					styledText := string(runes[style.Start:style.End])
					if styledText != "bold and italic" {
						t.Errorf("Expected styled text 'bold and italic', got '%s'", styledText)
					}
				}
			},
		},
		{
			name:       "hard line breaks",
			markdown:   "**To:** Test Recipient  \n**From:** sender@example.com  \n",
			wantBlocks: 2, // Hard line breaks split into separate blocks
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First block: "To: Test Recipient"
				if blocks[0].Text != "To: Test Recipient\n" {
					t.Errorf("Expected first block text %q, got %q", "To: Test Recipient\n", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style in first block, got %d", len(blocks[0].InlineStyles))
				} else {
					style1 := blocks[0].InlineStyles[0]
					if style1.Start != 0 || style1.End != 3 {
						t.Errorf("Expected first style at 0-3, got %d-%d", style1.Start, style1.End)
					}
				}

				// Second block: "From: sender@example.com"
				if blocks[1].Text != "From: sender@example.com\n" {
					t.Errorf("Expected second block text %q, got %q", "From: sender@example.com\n", blocks[1].Text)
				}
				if len(blocks[1].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style in second block, got %d", len(blocks[1].InlineStyles))
				} else {
					style2 := blocks[1].InlineStyles[0]
					if style2.Start != 0 || style2.End != 5 {
						t.Errorf("Expected second style at 0-5, got %d-%d", style2.Start, style2.End)
					}
				}
			},
		},
		{
			name:       "UTF-8 multibyte characters with bold",
			markdown:   "• **Bold text** using double asterisks",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				expectedText := "• Bold text using double asterisks\n"
				if blocks[0].Text != expectedText {
					t.Errorf("Expected text '%s', got '%s'", expectedText, blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
				} else {
					style := blocks[0].InlineStyles[0]
					// "• " is 2 runes (bullet + space), so bold should start at position 2
					// "Bold text" is 9 runes, so it should end at position 11
					if style.Start != 2 {
						t.Errorf("Expected style start at 2, got %d", style.Start)
					}
					if style.End != 11 {
						t.Errorf("Expected style end at 11, got %d", style.End)
					}
					// Verify it covers "Bold text"
					runes := []rune(blocks[0].Text)
					styledText := string(runes[style.Start:style.End])
					if styledText != "Bold text" {
						t.Errorf("Expected styled text 'Bold text', got '%s'", styledText)
					}
				}
			},
		},
		{
			name:       "multiple bold with hard line breaks",
			markdown:   "**To:** Test Recipient  \n**From:** sender@example.com  \n**Subject:** Markdown Test",
			wantBlocks: 3, // Hard line breaks split into separate blocks
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First block: "To: Test Recipient"
				if blocks[0].Text != "To: Test Recipient\n" {
					t.Errorf("Expected first block text 'To: Test Recipient\\n', got '%s'", blocks[0].Text)
				}
				if len(blocks[0].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style in first block, got %d", len(blocks[0].InlineStyles))
				} else {
					style1 := blocks[0].InlineStyles[0]
					text1 := string([]rune(blocks[0].Text)[style1.Start:style1.End])
					if text1 != "To:" {
						t.Errorf("First bold should be 'To:', got '%s' (positions %d-%d)", text1, style1.Start, style1.End)
					}
				}

				// Second block: "From: sender@example.com"
				if blocks[1].Text != "From: sender@example.com\n" {
					t.Errorf("Expected second block text 'From: sender@example.com\\n', got '%s'", blocks[1].Text)
				}
				if len(blocks[1].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style in second block, got %d", len(blocks[1].InlineStyles))
				} else {
					style2 := blocks[1].InlineStyles[0]
					text2 := string([]rune(blocks[1].Text)[style2.Start:style2.End])
					if text2 != "From:" {
						t.Errorf("Second bold should be 'From:', got '%s' (positions %d-%d)", text2, style2.Start, style2.End)
					}
				}

				// Third block: "Subject: Markdown Test"
				if blocks[2].Text != "Subject: Markdown Test\n" {
					t.Errorf("Expected third block text 'Subject: Markdown Test\\n', got '%s'", blocks[2].Text)
				}
				if len(blocks[2].InlineStyles) != 1 {
					t.Errorf("Expected 1 inline style in third block, got %d", len(blocks[2].InlineStyles))
				} else {
					style3 := blocks[2].InlineStyles[0]
					text3 := string([]rune(blocks[2].Text)[style3.Start:style3.End])
					if text3 != "Subject:" {
						t.Errorf("Third bold should be 'Subject:', got '%s' (positions %d-%d)", text3, style3.Start, style3.End)
					}
				}
			},
		},
		{
			name:       "soft line breaks become spaces",
			markdown:   "one\ntwo\nthree",
			wantBlocks: 1,
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Expected type paragraph, got %s", blocks[0].Type)
				}
				// Soft line breaks should become spaces
				if blocks[0].Text != "one two three\n" {
					t.Errorf("Expected text 'one two three\\n', got %q", blocks[0].Text)
				}
			},
		},
		{
			name:       "code block in list item",
			markdown:   "1. Install tool:\n    ```bash\n    brew install apple-mail-mcp\n    ```",
			wantBlocks: 3, // List item + code block line + trailing empty line from code block
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First block is the list item itself
				if blocks[0].Type != BlockTypeListItem {
					t.Errorf("Block 0: expected list_item, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "1. Install tool:\n" {
					t.Errorf("Block 0: expected text '1. Install tool:\\n', got %q", blocks[0].Text)
				}

				// Second block is the code block line, indented to align with list text
				if blocks[1].Type != BlockTypeCodeBlock {
					t.Errorf("Block 1: expected code_block, got %s", blocks[1].Type)
				}
				// "1. " is 3 chars, so sub-indent is 3 spaces.
				// Code block line starts with sub-indent (3 spaces)
				// The code block prefix (2 spaces) is NOT added automatically because we use
				// a custom config for these tests where code_block prefix content is empty
				expectedCodeLine := "   brew install apple-mail-mcp\n"
				if blocks[1].Text != expectedCodeLine {
					t.Errorf("Block 1: expected text %q, got %q", expectedCodeLine, blocks[1].Text)
				}

				// Third block is the trailing empty line from the code block
				if blocks[2].Type != BlockTypeCodeBlock {
					t.Errorf("Block 2: expected code_block (empty line), got %s", blocks[2].Type)
				}
				if blocks[2].Text != "   \n" {
					t.Errorf("Block 2: expected sub-indented empty line '   \\n', got %q", blocks[2].Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseMarkdown([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ParseMarkdown() error = %v", err)
			}

			blocks, err := ConvertMarkdownToStyledBlocks(doc, []byte(tt.markdown), config)
			if err != nil {
				t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
			}

			if len(blocks) != tt.wantBlocks {
				t.Errorf("Expected %d blocks, got %d", tt.wantBlocks, len(blocks))
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, blocks)
			}
		})
	}
}

func TestConvertMarkdownToStyledBlocks_WithMargins(t *testing.T) {
	// Create a test config WITH margins
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_styles_margins.yaml")

	configYAML := `defaults:
  font: "Helvetica"
  size: 12
  color: "#000000"

styles:
  h1:
    font: "Helvetica-Bold"
    size: 24
    color: "#000000"
    margin_top: 12
    margin_bottom: 6
  paragraph:
    font: "Helvetica"
    size: 12
    color: "#000000"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	tests := []struct {
		name       string
		markdown   string
		wantBlocks int
		checkFunc  func(*testing.T, []StyledBlock)
	}{
		{
			name:       "heading with margins",
			markdown:   "# Title With Margins",
			wantBlocks: 2, // heading + margin_bottom (no margin_top for first block)
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// Check heading block (first block, no margin_top)
				if blocks[0].Type != "heading" {
					t.Errorf("Expected first block type heading, got %s", blocks[0].Type)
				}
				// Text now includes trailing newline
				if blocks[0].Text != "Title With Margins\n" {
					t.Errorf("Expected text 'Title With Margins\\n', got %q", blocks[0].Text)
				}

				// Check margin_bottom block (no Type field set)
				// Margin block has just newline
				if blocks[1].Text != "\n" {
					t.Errorf("Expected margin text '\\n', got %q", blocks[1].Text)
				}
				// Margin blocks have no Type field (empty string)
				if blocks[1].Type != "" {
					t.Errorf("Expected margin Type to be empty, got %q", blocks[1].Type)
				}
				if blocks[1].Size != 6 {
					t.Errorf("Expected margin_bottom size 6, got %d", blocks[1].Size)
				}
				if blocks[1].Font != "Helvetica-Bold" {
					t.Errorf("Expected margin font Helvetica-Bold, got %s", blocks[1].Font)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseMarkdown([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ParseMarkdown() error = %v", err)
			}

			blocks, err := ConvertMarkdownToStyledBlocks(doc, []byte(tt.markdown), config)
			if err != nil {
				t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
			}

			if len(blocks) != tt.wantBlocks {
				t.Errorf("Expected %d blocks, got %d", tt.wantBlocks, len(blocks))
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, blocks)
			}
		})
	}
}

func TestConvertMarkdownToStyledBlocks_WithPrefix(t *testing.T) {
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name     string
		markdown string
		want     []StyledBlock
	}{
		{
			name:     "code_block_with_prefix",
			markdown: "```\nfirst line\nsecond line\n```",
			want: []StyledBlock{
				{
					Type: BlockTypeCodeBlock,
					Text: "  first line\n",
					Font: "Menlo-Regular",
					Size: 11,
				},
				{
					Type: BlockTypeCodeBlock,
					Text: "  second line\n",
					Font: "Menlo-Regular",
					Size: 11,
				},
				{
					Type: BlockTypeCodeBlock,
					Text: "  \n",
					Font: "Menlo-Regular",
					Size: 11,
				},
				// No margin_bottom block since config has margin_bottom: null
			},
		},
		// Note: Blockquote tests removed - blockquotes are now containers that recursively
		// process child elements, so they don't have prefix styling
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseMarkdown([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ParseMarkdown() error = %v", err)
			}

			got, err := ConvertMarkdownToStyledBlocks(doc, []byte(tt.markdown), config)
			if err != nil {
				t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %d blocks, want %d blocks", len(got), len(tt.want))
			}

			for i, gotBlock := range got {
				wantBlock := tt.want[i]

				if gotBlock.Type != wantBlock.Type {
					t.Errorf("block %d: got type %q, want %q", i, gotBlock.Type, wantBlock.Type)
				}
				if gotBlock.Text != wantBlock.Text {
					t.Errorf("block %d: got text %q, want %q", i, gotBlock.Text, wantBlock.Text)
				}
				if gotBlock.Font != wantBlock.Font {
					t.Errorf("block %d: got font %q, want %q", i, gotBlock.Font, wantBlock.Font)
				}
				if gotBlock.Size != wantBlock.Size {
					t.Errorf("block %d: got size %d, want %d", i, gotBlock.Size, wantBlock.Size)
				}

				// Only check inline styles if expected
				if len(wantBlock.InlineStyles) > 0 {
					if len(gotBlock.InlineStyles) != len(wantBlock.InlineStyles) {
						t.Errorf("block %d: got %d inline styles, want %d", i, len(gotBlock.InlineStyles), len(wantBlock.InlineStyles))
						continue
					}

					for j, gotStyle := range gotBlock.InlineStyles {
						wantStyle := wantBlock.InlineStyles[j]
						if gotStyle.Start != wantStyle.Start {
							t.Errorf("block %d, style %d: got start %d, want %d", i, j, gotStyle.Start, wantStyle.Start)
						}
						if gotStyle.End != wantStyle.End {
							t.Errorf("block %d, style %d: got end %d, want %d", i, j, gotStyle.End, wantStyle.End)
						}
						if gotStyle.Font != wantStyle.Font {
							t.Errorf("block %d, style %d: got font %q, want %q", i, j, gotStyle.Font, wantStyle.Font)
						}
					}
				}
			}
		})
	}
}

func TestConvertMarkdownToStyledBlocks_ComplexDocument(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_styles_complex.yaml")

	configYAML := `defaults:
  font: "Helvetica"
  size: 12
  color: "#000000"

styles:
  h1:
    font: "Helvetica-Bold"
    size: 24
    color: "#000000"
  h2:
    font: "Helvetica-Bold"
    size: 20
    color: "#000000"
  bold:
    font: "Helvetica-Bold"
  italic:
    font: "Helvetica-Oblique"
  code:
    font: "Menlo-Regular"
    size: 11
    color: "#D73A49"
  list_item:
    font: "Helvetica"
    size: 12
  list:
    margin_top: 0
    margin_bottom: 0
  paragraph:
    font: "Helvetica"
    size: 12
    color: "#000000"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	markdown := "# Main Title\n\n" +
		"This is a paragraph with **bold** text and *italic* text.\n\n" +
		"## Subsection\n\n" +
		"Another paragraph here.\n\n" +
		"- List item 1\n" +
		"- List item 2\n" +
		"  - Nested item\n" +
		"- List item 3\n\n" +
		"Some `inline code` in text."

	doc, err := ParseMarkdown([]byte(markdown))
	if err != nil {
		t.Fatalf("ParseMarkdown() error = %v", err)
	}

	blocks, err := ConvertMarkdownToStyledBlocks(doc, []byte(markdown), config)
	if err != nil {
		t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
	}

	// Should have: 1 h1, 1 para, 1 h2, 1 para, 4 list items, 1 para
	if len(blocks) < 5 {
		t.Errorf("Expected at least 5 blocks, got %d", len(blocks))
	}

	// Check first block is h1
	if blocks[0].Type != BlockTypeHeading || blocks[0].Level != 1 {
		t.Errorf("First block should be h1, got type=%s level=%d", blocks[0].Type, blocks[0].Level)
	}
}

func TestConvertMarkdownToStyledBlocks_ComplexBlockquotes(t *testing.T) {
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name       string
		markdown   string
		wantBlocks int
		checkFunc  func(t *testing.T, blocks []StyledBlock)
	}{
		{
			name: "nested_blockquote",
			markdown: `> This is the first level of quoting.
>
> > This is nested blockquote.
>
> Back to the first level.`,
			wantBlocks: 7, // 3 paragraphs + 2 empty lines + 2 margins (nested margins get prefixed!)
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// Debug: print all blocks
				for i, block := range blocks {
					t.Logf("Block %d: type=%q, text=%q", i, block.Type, block.Text)
				}
				// First paragraph with "> " prefix
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Block 0: expected paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "> This is the first level of quoting.\n" {
					t.Errorf("Block 0: unexpected text: %q", blocks[0].Text)
				}

				// Block 1: empty line
				if blocks[1].Text != "> \n" {
					t.Errorf("Block 1: expected empty line '> \\n', got %q", blocks[1].Text)
				}

				// Block 2: margin_top from nested blockquote (gets prefixed!)
				if blocks[2].Text != "> > \n" {
					t.Errorf("Block 2: expected nested margin '> > \\n', got %q", blocks[2].Text)
				}

				// Block 3: nested paragraph
				if blocks[3].Text != "> > This is nested blockquote.\n" {
					t.Errorf("Block 3: expected nested paragraph, got %q", blocks[3].Text)
				}

				// Block 4: margin_bottom from nested blockquote (gets prefixed!)
				if blocks[4].Text != "> \n" {
					t.Errorf("Block 4: expected margin with prefix '> \\n', got %q", blocks[4].Text)
				}

				// Block 5: back to first level
				if blocks[5].Type != BlockTypeParagraph {
					t.Errorf("Block 5: expected paragraph, got %s", blocks[5].Type)
				}
				if blocks[5].Text != "> Back to the first level.\n" {
					t.Errorf("Block 5: unexpected text: %q", blocks[5].Text)
				}
			},
		},
		{
			name: "blockquote_with_header",
			markdown: `> ## This is a header.
>
> This is a paragraph.`,
			wantBlocks: 4, // Header + margin + paragraph + margin_bottom
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// Header
				if blocks[0].Type != BlockTypeHeading {
					t.Errorf("Block 0: expected heading, got %s", blocks[0].Type)
				}
				if blocks[0].Level != 2 {
					t.Errorf("Block 0: expected level 2, got %d", blocks[0].Level)
				}
				if blocks[0].Text != "> This is a header.\n" {
					t.Errorf("Block 0: unexpected text: %q", blocks[0].Text)
				}

				// Paragraph (block 2, after header's margin_bottom)
				if len(blocks) < 3 {
					t.Fatal("Not enough blocks")
				}
				if blocks[2].Type != BlockTypeParagraph {
					t.Errorf("Block 2: expected paragraph, got %s", blocks[2].Type)
				}
				if blocks[2].Text != "> This is a paragraph.\n" {
					t.Errorf("Block 2: unexpected text: %q", blocks[2].Text)
				}
			},
		},
		{
			name: "blockquote_with_list",
			markdown: `> 1. This is the first list item.
> 2. This is the second list item.`,
			wantBlocks: 4, // Two list items + list margin_bottom + blockquote margin_bottom
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First list item
				if blocks[0].Type != BlockTypeListItem {
					t.Errorf("Block 0: expected list_item, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "> 1. This is the first list item.\n" {
					t.Errorf("Block 0: unexpected text: %q", blocks[0].Text)
				}

				// Second list item
				if blocks[1].Type != BlockTypeListItem {
					t.Errorf("Block 1: expected list_item, got %s", blocks[1].Type)
				}
				if blocks[1].Text != "> 2. This is the second list item.\n" {
					t.Errorf("Block 1: unexpected text: %q", blocks[1].Text)
				}
			},
		},
		{
			name: "blockquote_with_code",
			markdown: `> Here's some example code:
>
>     return shell_exec("echo $input");`,
			wantBlocks: 6, // Paragraph + empty line + code blocks + blockquote margin_bottom
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First block is paragraph
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Block 0: expected paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "> Here's some example code:\n" {
					t.Errorf("Block 0: unexpected text: %q", blocks[0].Text)
				}

				// Block 1 should be empty line
				if blocks[1].Text != "> \n" {
					t.Errorf("Block 1: expected empty line with prefix '> \\n', got %q", blocks[1].Text)
				}

				// Find a code block (should be after the empty line)
				foundCode := false
				for i := 2; i < len(blocks)-1; i++ {
					if blocks[i].Type == BlockTypeCodeBlock {
						foundCode = true
						// Code block should have both blockquote and code prefixes
						if !strings.HasPrefix(blocks[i].Text, ">   ") {
							t.Errorf("Code block should have '>   ' prefix, got %q", blocks[i].Text)
						}
						break
					}
				}
				if !foundCode {
					t.Error("Expected to find a code_block")
				}
			},
		},
		{
			name: "blockquote_with_multiple_paragraphs",
			markdown: `> First paragraph.
>
> Second paragraph.`,
			wantBlocks: 4, // paragraph + empty line + paragraph + margin_bottom
			checkFunc: func(t *testing.T, blocks []StyledBlock) {
				// First paragraph with prefix
				if blocks[0].Type != BlockTypeParagraph {
					t.Errorf("Block 0: expected paragraph, got %s", blocks[0].Type)
				}
				if blocks[0].Text != "> First paragraph.\n" {
					t.Errorf("Block 0: unexpected text: %q", blocks[0].Text)
				}

				// Empty line with prefix for visual separation
				if len(blocks) < 2 {
					t.Fatal("Not enough blocks")
				}
				if blocks[1].Type != BlockTypeParagraph {
					t.Errorf("Block 1: expected paragraph (empty line), got %s", blocks[1].Type)
				}
				if blocks[1].Text != "> \n" {
					t.Errorf("Block 1: expected empty line with prefix '> \\n', got %q", blocks[1].Text)
				}

				// Second paragraph with prefix
				if len(blocks) < 3 {
					t.Fatal("Not enough blocks")
				}
				if blocks[2].Type != BlockTypeParagraph {
					t.Errorf("Block 2: expected paragraph, got %s", blocks[2].Type)
				}
				if blocks[2].Text != "> Second paragraph.\n" {
					t.Errorf("Block 2: unexpected text: %q", blocks[2].Text)
				}

				// Last block should be blockquote margin_bottom
				if len(blocks) < 4 {
					t.Fatal("Not enough blocks")
				}
				if blocks[3].Type != "" {
					t.Errorf("Block 3: expected margin (empty type), got %s", blocks[3].Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseMarkdown([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ParseMarkdown() error = %v", err)
			}

			blocks, err := ConvertMarkdownToStyledBlocks(doc, []byte(tt.markdown), config)
			if err != nil {
				t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
			}

			if len(blocks) != tt.wantBlocks {
				t.Errorf("Expected %d blocks, got %d", tt.wantBlocks, len(blocks))
				for i, block := range blocks {
					t.Logf("Block %d: type=%s, text=%q", i, block.Type, block.Text)
				}
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, blocks)
			}
		})
	}
}

func TestExtractTextWithInlineStyles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_styles_inline.yaml")

	configYAML := `defaults:
  font: "Helvetica"
  size: 12
  color: "#000000"

styles:
  bold:
    font: "Helvetica-Bold"
  paragraph:
    font: "Helvetica"
    size: 12
    color: "#000000"
`

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	markdown := "This is **bold** text"
	doc, err := ParseMarkdown([]byte(markdown))
	if err != nil {
		t.Fatalf("ParseMarkdown() error = %v", err)
	}

	blocks, err := ConvertMarkdownToStyledBlocks(doc, []byte(markdown), config)
	if err != nil {
		t.Fatalf("ConvertMarkdownToStyledBlocks() error = %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Text != "This is bold text\n" {
		t.Errorf("Expected 'This is bold text\\n', got %q", blocks[0].Text)
	}

	if len(blocks[0].InlineStyles) != 1 {
		t.Fatalf("Expected 1 inline style, got %d", len(blocks[0].InlineStyles))
	}

	style := blocks[0].InlineStyles[0]
	if style.Start != 8 {
		t.Errorf("Expected start=8, got %d", style.Start)
	}
	if style.End != 12 {
		t.Errorf("Expected end=12, got %d", style.End)
	}
}
