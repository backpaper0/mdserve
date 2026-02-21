// Package renderer provides Markdown to HTML conversion.
package renderer

import (
	"bytes"
	"fmt"
	"os"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/frontmatter"
)

// Renderer converts a Markdown file to an HTML fragment.
type Renderer interface {
	Render(filePath string) ([]byte, error)
}

// RenderError is returned when rendering a file fails.
type RenderError struct {
	FilePath string
	Cause    error
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("render %s: %v", e.FilePath, e.Cause)
}

func (e *RenderError) Unwrap() error { return e.Cause }

type goldmarkRenderer struct {
	md goldmark.Markdown
}

// New creates a Renderer with the full pipeline:
//   - Standard extensions: Table, Strikethrough, TaskList
//   - YAML Front Matter removal (go.abhg.dev/goldmark/frontmatter)
//   - Syntax highlighting with CSS classes (goldmark-highlighting + Chroma github style)
//   - Mermaid code fence → <div class="mermaid"> (custom extension)
func New() Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
			&MermaidExtension{},
		),
	)
	return &goldmarkRenderer{md: md}
}

// Render reads the Markdown file at filePath and returns an HTML fragment.
func (r *goldmarkRenderer) Render(filePath string) ([]byte, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, &RenderError{FilePath: filePath, Cause: err}
	}
	var buf bytes.Buffer
	if err := r.md.Convert(src, &buf); err != nil {
		return nil, &RenderError{FilePath: filePath, Cause: err}
	}
	return buf.Bytes(), nil
}
