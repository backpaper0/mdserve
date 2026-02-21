package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"mdserve/internal/server"
)

func TestNewAssetHandler_ServesCSS(t *testing.T) {
	testFS := fstest.MapFS{
		"assets/github-markdown.css": {
			Data: []byte("body { color: black; }"),
		},
	}
	handler := server.NewAssetHandler(testFS)

	req := httptest.NewRequest("GET", "/assets/github-markdown.css", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/css") {
		t.Errorf("expected text/css content type, got %s", ct)
	}
	if !strings.Contains(rr.Body.String(), "color: black") {
		t.Errorf("expected CSS content in body, got: %s", rr.Body.String())
	}
}

func TestNewAssetHandler_ServesJS(t *testing.T) {
	testFS := fstest.MapFS{
		"assets/mermaid.min.js": {
			Data: []byte("var mermaid = {};"),
		},
	}
	handler := server.NewAssetHandler(testFS)

	req := httptest.NewRequest("GET", "/assets/mermaid.min.js", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestNewAssetHandler_NotFound(t *testing.T) {
	testFS := fstest.MapFS{}
	handler := server.NewAssetHandler(testFS)

	req := httptest.NewRequest("GET", "/assets/nonexistent.js", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestNewAssetHandler_MultipleFiles(t *testing.T) {
	testFS := fstest.MapFS{
		"assets/github-markdown.css": {Data: []byte(".md { }")},
		"assets/highlight.css":       {Data: []byte(".chroma { }")},
		"assets/mermaid.min.js":      {Data: []byte("var m={};")},
	}
	handler := server.NewAssetHandler(testFS)

	paths := []string{
		"/assets/github-markdown.css",
		"/assets/highlight.css",
		"/assets/mermaid.min.js",
	}
	for _, path := range paths {
		req := httptest.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", path, rr.Code)
		}
	}
}
