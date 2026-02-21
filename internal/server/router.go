package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// resolvedPathKeyType is the type for the resolved-path context key.
type resolvedPathKeyType struct{}

var resolvedPathKey = resolvedPathKeyType{}

// withResolvedPath attaches the validated file-system path to the request context.
func withResolvedPath(r *http.Request, fsPath string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), resolvedPathKey, fsPath))
}

// resolvedPathFrom retrieves the file-system path stored by the router.
func resolvedPathFrom(r *http.Request) string {
	v, _ := r.Context().Value(resolvedPathKey).(string)
	return v
}

type requestRouter struct {
	docRoot       string
	mdHandler     http.Handler
	dirHandler    http.Handler
	staticHandler http.Handler
}

// NewRequestRouter creates an http.Handler that resolves the URL path to a
// real file-system path, validates it against docRoot (path-traversal
// prevention), and delegates to the appropriate sub-handler.
func NewRequestRouter(
	docRoot string,
	mdHandler http.Handler,
	dirHandler http.Handler,
	staticHandler http.Handler,
) http.Handler {
	return &requestRouter{
		docRoot:       filepath.Clean(docRoot),
		mdHandler:     mdHandler,
		dirHandler:    dirHandler,
		staticHandler: staticHandler,
	}
}

func (rtr *requestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	// Build the candidate file-system path and clean it.
	fsPath := filepath.Join(rtr.docRoot, filepath.FromSlash(urlPath))
	fsPath = filepath.Clean(fsPath)

	// First security check: cleaned path must be inside docRoot.
	if !withinRoot(fsPath, rtr.docRoot) {
		http.NotFound(w, r)
		return
	}

	// Resolve symlinks to prevent traversal via symlinks.
	resolved, err := filepath.EvalSymlinks(fsPath)
	if err != nil {
		// File does not exist or is inaccessible.
		http.NotFound(w, r)
		return
	}

	// Second security check: resolved (real) path must also be inside docRoot.
	if !withinRoot(resolved, rtr.docRoot) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(resolved)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Attach resolved path for downstream handlers.
	r = withResolvedPath(r, resolved)

	switch {
	case info.IsDir():
		// Redirect to add a trailing slash for consistent relative-link behaviour.
		if !strings.HasSuffix(urlPath, "/") {
			http.Redirect(w, r, urlPath+"/", http.StatusMovedPermanently)
			return
		}
		rtr.dirHandler.ServeHTTP(w, r)
	case strings.HasSuffix(strings.ToLower(info.Name()), ".md"):
		rtr.mdHandler.ServeHTTP(w, r)
	default:
		rtr.staticHandler.ServeHTTP(w, r)
	}
}

// withinRoot returns true if path equals root or is directly inside root.
// Both arguments must be cleaned paths (no trailing separator).
func withinRoot(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
