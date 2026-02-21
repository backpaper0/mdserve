package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mdserve/internal/dirlist"
	"mdserve/internal/renderer"
	"mdserve/internal/server"
	"mdserve/internal/tmpl"
)

// newTestServer creates an httptest.Server with the full handler stack.
func newTestServer(t *testing.T, docRoot string) *httptest.Server {
	t.Helper()
	r := renderer.New()
	tmplEngine := tmpl.New()
	lister := dirlist.New()

	mdH := server.NewMarkdownHandler(r, tmplEngine, false)
	dirH := server.NewDirectoryHandler(lister, r, tmplEngine, docRoot, false)
	staticH := server.NewStaticFileHandler(docRoot)
	router := server.NewRequestRouter(docRoot, mdH, dirH, staticH)

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	return ts
}

// writeDoc creates a file under docRoot.
func writeDoc(t *testing.T, docRoot, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(docRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("writeDoc MkdirAll: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writeDoc WriteFile %s: %v", relPath, err)
	}
}

// --- Task 5.1: RequestRouter ---

func TestRouter_MDFileRoutedToMarkdownHandler(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "hello.md", "# Hello\n\nWorld")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/hello.md")
	if err != nil {
		t.Fatalf("GET /hello.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html content-type, got %s", ct)
	}
}

func TestRouter_DirectoryRoutedToDirHandler(t *testing.T) {
	docRoot := t.TempDir()
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for root directory, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html content-type for directory, got %s", ct)
	}
}

func TestRouter_NonExistentPathReturns404(t *testing.T) {
	docRoot := t.TempDir()
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/nonexistent.md")
	if err != nil {
		t.Fatalf("GET /nonexistent.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRouter_PathTraversalBlocked(t *testing.T) {
	docRoot := t.TempDir()
	ts := newTestServer(t, docRoot)

	// HTTP client follows redirects and path-normalizes, so use a raw-ish attempt.
	// The Go HTTP client normalizes /../ in URLs, so we test with a URL-encoded path.
	resp, err := http.Get(ts.URL + "/..%2F..%2Fetc%2Fpasswd")
	if err != nil {
		t.Fatalf("GET path traversal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("path traversal should not return 200; body: %s", string(body))
	}
}

func TestRouter_StaticFileServed(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "image.png", "PNG_DATA")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/image.png")
	if err != nil {
		t.Fatalf("GET /image.png: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for static file, got %d", resp.StatusCode)
	}
}

// --- Task 5.2: MarkdownHandler ---

func TestMarkdownHandler_RendersHTMLContent(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "doc.md", "# My Document\n\nSome content here.")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/doc.md")
	if err != nil {
		t.Fatalf("GET /doc.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected full HTML page with DOCTYPE")
	}
	if !strings.Contains(html, "My Document") {
		t.Error("expected markdown heading in HTML output")
	}
	if !strings.Contains(html, "Some content here") {
		t.Error("expected markdown content in HTML output")
	}
}

func TestMarkdownHandler_ContentTypeIsHTML(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "file.md", "# Test")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/file.md")
	if err != nil {
		t.Fatalf("GET /file.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestMarkdownHandler_ContainsAssetLinks(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "page.md", "# Page")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/page.md")
	if err != nil {
		t.Fatalf("GET /page.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "/assets/") {
		t.Error("expected /assets/ references in HTML")
	}
}

func TestMarkdownHandler_SubdirMDFile(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "sub/page.md", "# Sub Page")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/sub/page.md")
	if err != nil {
		t.Fatalf("GET /sub/page.md: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Sub Page") {
		t.Error("expected 'Sub Page' in response")
	}
}

// --- Task 5.3: DirectoryHandler ---

func TestDirHandler_EmptyDirShowsListing(t *testing.T) {
	docRoot := t.TempDir()
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for empty dir listing, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<!DOCTYPE html>") {
		t.Error("expected full HTML page for directory listing")
	}
}

func TestDirHandler_WithREADMERendersMarkdown(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "README.md", "# Welcome\n\nThis is the README.")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "Welcome") {
		t.Error("expected README content 'Welcome' in response")
	}
	if !strings.Contains(html, "This is the README") {
		t.Error("expected README body text in response")
	}
}

func TestDirHandler_WithIndexMDRendersMarkdown(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "index.md", "# Index Page")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Index Page") {
		t.Error("expected 'Index Page' from index.md in response")
	}
}

func TestDirHandler_ListingContainsMDLinks(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "alpha.md", "# Alpha")
	writeDoc(t, docRoot, "beta.md", "# Beta")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "alpha.md") {
		t.Error("expected alpha.md in directory listing")
	}
	if !strings.Contains(html, "beta.md") {
		t.Error("expected beta.md in directory listing")
	}
}

func TestDirHandler_ListingExcludesNonMDFiles(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "doc.md", "# Doc")
	writeDoc(t, docRoot, "image.png", "PNG data")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// image.png should not appear in the listing as a link
	if strings.Contains(html, `href="/image.png"`) {
		t.Error("non-.md file image.png should not appear as link in directory listing")
	}
}

func TestDirHandler_SubdirectoryRequest(t *testing.T) {
	docRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(docRoot, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeDoc(t, docRoot, "docs/guide.md", "# Guide")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/docs/")
	if err != nil {
		t.Fatalf("GET /docs/: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "guide.md") {
		t.Error("expected guide.md in /docs/ listing")
	}
}

// --- Task 5.4: StaticFileHandler ---

func TestStaticHandler_ServesNonMDFile(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "style.css", "body { margin: 0; }")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/style.css")
	if err != nil {
		t.Fatalf("GET /style.css: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "margin: 0") {
		t.Error("expected CSS content in response")
	}
}

func TestStaticHandler_ServesImageFile(t *testing.T) {
	docRoot := t.TempDir()
	writeDoc(t, docRoot, "logo.png", "\x89PNG\r\n\x1a\n")
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/logo.png")
	if err != nil {
		t.Fatalf("GET /logo.png: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestStaticHandler_NotFoundForMissingFile(t *testing.T) {
	docRoot := t.TempDir()
	ts := newTestServer(t, docRoot)

	resp, err := http.Get(ts.URL + "/missing.pdf")
	if err != nil {
		t.Fatalf("GET /missing.pdf: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
