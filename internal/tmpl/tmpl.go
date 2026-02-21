// Package tmpl provides HTML template rendering for Markdown and directory listing pages.
package tmpl

import (
	"bytes"
	"embed"
	"html/template"
	"sync"

	"mdserve/internal/dirlist"
)

// PageData holds the data needed to render a Markdown content page.
type PageData struct {
	Title       string
	Content     template.HTML // Trusted HTML fragment from Renderer; not escaped again
	Breadcrumbs []dirlist.Breadcrumb
	LiveReload  bool
}

// DirListData holds the data needed to render a directory listing page.
type DirListData struct {
	Title       string
	Breadcrumbs []dirlist.Breadcrumb
	Entries     []dirlist.Entry
	LiveReload  bool
}

// TemplateEngine renders full HTML pages from structured data.
type TemplateEngine interface {
	RenderPage(data PageData) ([]byte, error)
	RenderDirList(data DirListData) ([]byte, error)
}

//go:embed templates
var templateFS embed.FS

type engine struct {
	once    sync.Once
	tmpl    *template.Template
	initErr error
}

func (e *engine) init() {
	e.once.Do(func() {
		t, err := template.ParseFS(templateFS, "templates/*.html")
		if err != nil {
			e.initErr = err
			return
		}
		e.tmpl = t
	})
}

// New creates a new TemplateEngine backed by the embedded HTML templates.
func New() TemplateEngine {
	return &engine{}
}

// RenderPage renders a full HTML page for a Markdown document.
func (e *engine) RenderPage(data PageData) ([]byte, error) {
	e.init()
	if e.initErr != nil {
		return nil, e.initErr
	}
	var buf bytes.Buffer
	if err := e.tmpl.ExecuteTemplate(&buf, "page.html", data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderDirList renders a full HTML page for a directory listing.
func (e *engine) RenderDirList(data DirListData) ([]byte, error) {
	e.init()
	if e.initErr != nil {
		return nil, e.initErr
	}
	var buf bytes.Buffer
	if err := e.tmpl.ExecuteTemplate(&buf, "dirlist.html", data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
