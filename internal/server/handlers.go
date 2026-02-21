package server

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"mdserve/internal/dirlist"
	"mdserve/internal/renderer"
	"mdserve/internal/sse"
	"mdserve/internal/tmpl"
)

// --- MarkdownHandler ---

type markdownHandler struct {
	renderer   renderer.Renderer
	tmplEngine tmpl.TemplateEngine
	liveReload bool
}

// NewMarkdownHandler returns an http.Handler that renders a .md file as a
// complete HTML page using the given Renderer and TemplateEngine.
func NewMarkdownHandler(r renderer.Renderer, t tmpl.TemplateEngine, liveReload bool) http.Handler {
	return &markdownHandler{renderer: r, tmplEngine: t, liveReload: liveReload}
}

func (h *markdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsPath := resolvedPathFrom(r)
	if fsPath == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	content, err := h.renderer.Render(fsPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	title := strings.TrimSuffix(filepath.Base(fsPath), filepath.Ext(fsPath))
	crumbs := breadcrumbsForURL(r.URL.Path)

	data := tmpl.PageData{
		Title:       title,
		Content:     template.HTML(content), //nolint:gosec // trusted renderer output
		Breadcrumbs: crumbs,
		LiveReload:  h.liveReload,
	}

	html, err := h.tmplEngine.RenderPage(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

// breadcrumbsForURL builds a breadcrumb trail for the parent directories of urlPath.
// The first element is always {Label: "Root", URL: "/"}.
func breadcrumbsForURL(urlPath string) []dirlist.Breadcrumb {
	crumbs := []dirlist.Breadcrumb{{Label: "Root", URL: "/"}}

	dir := path.Dir(urlPath)
	if dir == "/" || dir == "." {
		return crumbs
	}

	trimmed := strings.Trim(dir, "/")
	parts := strings.Split(trimmed, "/")
	u := "/"
	for _, part := range parts {
		if part == "" {
			continue
		}
		u += part + "/"
		crumbs = append(crumbs, dirlist.Breadcrumb{Label: part, URL: u})
	}
	return crumbs
}

// --- DirectoryHandler ---

type directoryHandler struct {
	lister     dirlist.DirectoryLister
	renderer   renderer.Renderer
	tmplEngine tmpl.TemplateEngine
	docRoot    string
	liveReload bool
}

// NewDirectoryHandler returns an http.Handler that serves directory listings
// and index files (README.md / index.md).
func NewDirectoryHandler(
	lister dirlist.DirectoryLister,
	r renderer.Renderer,
	t tmpl.TemplateEngine,
	docRoot string,
	liveReload bool,
) http.Handler {
	cleanRoot := filepath.Clean(docRoot)
	if resolved, err := filepath.EvalSymlinks(cleanRoot); err == nil {
		cleanRoot = resolved
	}
	return &directoryHandler{
		lister:     lister,
		renderer:   r,
		tmplEngine: t,
		docRoot:    cleanRoot,
		liveReload: liveReload,
	}
}

func (h *directoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsPath := resolvedPathFrom(r)
	if fsPath == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	listing, err := h.lister.List(fsPath, h.docRoot)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	forceList := r.URL.Query().Has("list")

	if listing.IndexFile != "" && !forceList {
		h.serveIndexFile(w, r, listing)
		return
	}

	h.serveDirList(w, r, listing)
}

func (h *directoryHandler) serveIndexFile(w http.ResponseWriter, r *http.Request, listing *dirlist.Listing) {
	content, err := h.renderer.Render(listing.IndexFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := tmpl.PageData{
		Title:       listing.Title,
		Content:     template.HTML(content), //nolint:gosec // trusted renderer output
		Breadcrumbs: listing.Breadcrumbs,
		LiveReload:  h.liveReload,
		DirListURL:  r.URL.Path + "?list",
	}

	html, err := h.tmplEngine.RenderPage(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

func (h *directoryHandler) serveDirList(w http.ResponseWriter, r *http.Request, listing *dirlist.Listing) {
	indexURL := ""
	if listing.IndexFile != "" {
		indexURL = r.URL.Path
	}

	data := tmpl.DirListData{
		Title:       listing.Title,
		Breadcrumbs: listing.Breadcrumbs,
		Entries:     listing.Entries,
		LiveReload:  h.liveReload,
		IndexURL:    indexURL,
	}

	html, err := h.tmplEngine.RenderDirList(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

// --- StaticFileHandler ---

// NewStaticFileHandler returns an http.Handler that serves files directly
// from docRoot using http.FileServer.
func NewStaticFileHandler(docRoot string) http.Handler {
	return http.FileServer(http.Dir(docRoot))
}

// --- SSEHandler ---

const sseKeepaliveInterval = 15 * time.Second

// NewSSEHandler returns an http.HandlerFunc that manages Server-Sent Events connections.
// It registers each connecting client with the broker, forwards reload events, and
// sends keepalive comments every 15 seconds. Returns 500 if the ResponseWriter does
// not implement http.Flusher.
func NewSSEHandler(broker sse.Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		ch := broker.Register()
		defer broker.Unregister(ch)

		ticker := time.NewTicker(sseKeepaliveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
				if _, err := fmt.Fprintf(w, "data: reload\n\n"); err != nil {
					return
				}
				flusher.Flush()
			case <-ticker.C:
				if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}
