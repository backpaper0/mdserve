package renderer_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"mdserve/internal/renderer"
)

// writeTemp creates a temporary .md file with the given content and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return f.Name()
}

// renderStr renders the given Markdown string and returns the HTML output.
func renderStr(t *testing.T, md string) string {
	t.Helper()
	path := writeTemp(t, md)
	r := renderer.New()
	got, err := r.Render(path)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	return string(got)
}

// --- Task 2.1: 標準Markdown構文のHTMLレンダリング ---

func TestRender_Heading(t *testing.T) {
	got := renderStr(t, "# Hello\n")
	if !strings.Contains(got, "<h1>Hello</h1>") {
		t.Errorf("expected <h1>Hello</h1>, got:\n%s", got)
	}
}

func TestRender_Paragraph(t *testing.T) {
	got := renderStr(t, "Hello world\n")
	if !strings.Contains(got, "<p>Hello world</p>") {
		t.Errorf("expected <p>Hello world</p>, got:\n%s", got)
	}
}

func TestRender_List(t *testing.T) {
	got := renderStr(t, "- item1\n- item2\n")
	if !strings.Contains(got, "<ul>") {
		t.Errorf("expected <ul>, got:\n%s", got)
	}
	if !strings.Contains(got, "item1") || !strings.Contains(got, "item2") {
		t.Errorf("expected list items, got:\n%s", got)
	}
}

func TestRender_Table(t *testing.T) {
	md := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	got := renderStr(t, md)
	if !strings.Contains(got, "<table>") {
		t.Errorf("expected <table>, got:\n%s", got)
	}
	if !strings.Contains(got, "<th>") {
		t.Errorf("expected <th>, got:\n%s", got)
	}
}

func TestRender_CodeBlock(t *testing.T) {
	md := "```\nhello code\n```\n"
	got := renderStr(t, md)
	if !strings.Contains(got, "hello code") {
		t.Errorf("expected code block content, got:\n%s", got)
	}
}

func TestRender_Link(t *testing.T) {
	got := renderStr(t, "[example](http://example.com)\n")
	if !strings.Contains(got, `href="http://example.com"`) {
		t.Errorf("expected href attribute, got:\n%s", got)
	}
	if !strings.Contains(got, "example") {
		t.Errorf("expected link text, got:\n%s", got)
	}
}

func TestRender_Image(t *testing.T) {
	got := renderStr(t, "![alt text](image.png)\n")
	if !strings.Contains(got, `src="image.png"`) {
		t.Errorf("expected src attribute, got:\n%s", got)
	}
	if !strings.Contains(got, `alt="alt text"`) {
		t.Errorf("expected alt attribute, got:\n%s", got)
	}
}

func TestRender_Bold(t *testing.T) {
	got := renderStr(t, "**bold text**\n")
	if !strings.Contains(got, "<strong>bold text</strong>") {
		t.Errorf("expected <strong>bold text</strong>, got:\n%s", got)
	}
}

func TestRender_Italic(t *testing.T) {
	got := renderStr(t, "*italic text*\n")
	if !strings.Contains(got, "<em>italic text</em>") {
		t.Errorf("expected <em>italic text</em>, got:\n%s", got)
	}
}

func TestRender_Strikethrough(t *testing.T) {
	got := renderStr(t, "~~strikethrough~~\n")
	if !strings.Contains(got, "<del>strikethrough</del>") {
		t.Errorf("expected <del>strikethrough</del>, got:\n%s", got)
	}
}

func TestRender_TaskList(t *testing.T) {
	md := "- [ ] pending task\n- [x] done task\n"
	got := renderStr(t, md)
	if !strings.Contains(got, `type="checkbox"`) {
		t.Errorf("expected checkbox input in task list, got:\n%s", got)
	}
}

// --- Task 2.2: YAML Front Matter除去 ---

func TestRender_FrontMatterExcluded(t *testing.T) {
	md := "---\ntitle: My Title\nauthor: Alice\n---\n\n# Content\n"
	got := renderStr(t, md)
	if strings.Contains(got, "title: My Title") || strings.Contains(got, "author: Alice") {
		t.Errorf("front matter YAML should not appear in HTML output, got:\n%s", got)
	}
	if !strings.Contains(got, "<h1>Content</h1>") {
		t.Errorf("content after front matter should be rendered, got:\n%s", got)
	}
}

func TestRender_FrontMatterOnlyBody(t *testing.T) {
	md := "---\nkey: value\n---\n\nBody text.\n"
	got := renderStr(t, md)
	if strings.Contains(got, "key: value") {
		t.Errorf("front matter key should not appear in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Body text.") {
		t.Errorf("body should be rendered, got:\n%s", got)
	}
}

// --- Task 2.3: コードブロックのシンタックスハイライト ---

func TestRender_SyntaxHighlightHasChromaClass(t *testing.T) {
	md := "```go\nfunc main() {}\n```\n"
	got := renderStr(t, md)
	// goldmark-highlighting with WithClasses(true) wraps code in elements with chroma CSS classes
	if !strings.Contains(got, "chroma") {
		t.Errorf("expected chroma CSS class for syntax highlighting, got:\n%s", got)
	}
}

func TestRender_SyntaxHighlightPreservesCode(t *testing.T) {
	md := "```python\nprint('hello')\n```\n"
	got := renderStr(t, md)
	if !strings.Contains(got, "print") {
		t.Errorf("expected code content preserved, got:\n%s", got)
	}
}

// --- Task 2.4: Mermaid.jsダイアグラムのHTMLレンダリング ---

func TestRender_MermaidBlockConverted(t *testing.T) {
	md := "```mermaid\ngraph TD\n  A --> B\n```\n"
	got := renderStr(t, md)
	if !strings.Contains(got, `<div class="mermaid">`) {
		t.Errorf("expected <div class=\"mermaid\">, got:\n%s", got)
	}
	if !strings.Contains(got, "graph TD") {
		t.Errorf("expected mermaid content preserved, got:\n%s", got)
	}
}

func TestRender_MermaidBlockNotInPreTag(t *testing.T) {
	md := "```mermaid\nsequenceDiagram\n  A->>B: Hello\n```\n"
	got := renderStr(t, md)
	// The mermaid content must be in div.mermaid, not in a <pre><code> block
	divIdx := strings.Index(got, `<div class="mermaid">`)
	if divIdx < 0 {
		t.Fatalf("expected <div class=\"mermaid\">, got:\n%s", got)
	}
	// Ensure it's inside div.mermaid, not pre/code
	preIdx := strings.Index(got, "<pre")
	seqIdx := strings.Index(got, "sequenceDiagram")
	if preIdx >= 0 && seqIdx > preIdx {
		// sequenceDiagram appears after a <pre> tag — likely inside it
		closePreIdx := strings.Index(got, "</pre>")
		if seqIdx < closePreIdx {
			t.Errorf("mermaid content should not be inside <pre>, got:\n%s", got)
		}
	}
}

func TestRender_NonMermaidCodeBlockUnchanged(t *testing.T) {
	md := "```bash\necho hello\n```\n"
	got := renderStr(t, md)
	// Non-mermaid code blocks should NOT be in div.mermaid
	if strings.Contains(got, `<div class="mermaid">`) {
		t.Errorf("non-mermaid block should not use mermaid div, got:\n%s", got)
	}
}

// --- エラーケース ---

func TestRender_FileNotFound(t *testing.T) {
	r := renderer.New()
	_, err := r.Render("/non/existent/file.md")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	var re *renderer.RenderError
	if !errors.As(err, &re) {
		t.Errorf("expected *renderer.RenderError, got %T: %v", err, err)
	}
}
