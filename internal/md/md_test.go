package md

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		contains []string // Strings we expect to see in the rendered HTML
	}{
		{
			name:     "heading 1",
			markdown: "# Heading 1",
			contains: []string{"<h1>Heading 1</h1>"},
		},
		{
			name:     "bold text",
			markdown: "This is **bold**",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "italic text",
			markdown: "This is *italic*",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "blockquote",
			markdown: "> This is a quote",
			contains: []string{"<blockquote>", "This is a quote", "</blockquote>"},
		},
		{
			name:     "inline code",
			markdown: "Use `go test` to run tests",
			contains: []string{"<code>go test</code>"},
		},
		{
			name:     "code block",
			markdown: "```go\nfunc main() {}\n```",
			contains: []string{"<pre><code class=\"language-go\">func main() {}", "</code></pre>"},
		},
		{
			name:     "unordered list",
			markdown: "- Item 1\n- Item 2",
			contains: []string{"<ul>", "<li>Item 1</li>", "<li>Item 2</li>", "</ul>"},
		},
		{
			name:     "ordered list",
			markdown: "1. First\n2. Second",
			contains: []string{"<ol>", "<li>First</li>", "<li>Second</li>", "</ol>"},
		},
		{
			name:     "link",
			markdown: "[GitHub](https://github.com)",
			contains: []string{"<a href=\"https://github.com\">GitHub</a>"},
		},
		{
			name:     "horizontal rule",
			markdown: "---",
			contains: []string{"<hr>"},
		},
		{
			name:     "table",
			markdown: "| Header 1 | Header 2 |\n| --- | --- |\n| Cell 1 | Cell 2 |",
			contains: []string{"<table>", "<thead>", "<th>Header 1</th>", "<tbody>", "<td>Cell 1</td>"},
		},
		{
			name:     "task list",
			markdown: "- [x] Done\n- [ ] Todo",
			contains: []string{"<li><input checked=\"\" disabled=\"\" type=\"checkbox\"> Done</li>", "<li><input disabled=\"\" type=\"checkbox\"> Todo</li>"},
		},
		{
			name:     "strikethrough",
			markdown: "~~deleted~~",
			contains: []string{"<del>deleted</del>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.markdown)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Render() output does not contain expected string.\nGot: %q\nWant: %q", got, want)
				}
			}
		})
	}
}
